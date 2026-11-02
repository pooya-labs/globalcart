package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pooya16pg-spec/globalcart/backend/internal/auth"
)

const productManagementQueryTimeout = 5 * time.Second

var (
	ErrManagedProductNotFound = errors.New("managed product not found")
	ErrInvalidProductInput    = errors.New("invalid product management input")
	ErrNoProductChanges       = errors.New("product update contains no changes")
	ErrProductAlreadyArchived = errors.New("product is already archived")
	ErrProductNotArchived     = errors.New("product is not archived")
	ErrStaleProductRevision   = errors.New("product revision is stale")
)

type ProductManagementRecord struct {
	ID                  int64     `json:"id"`
	Title               string    `json:"title"`
	Description         string    `json:"description"`
	Brand               string    `json:"brand"`
	OriginalTitle       string    `json:"originalTitle"`
	OriginalDescription string    `json:"originalDescription"`
	OriginalBrand       string    `json:"originalBrand"`
	Category            string    `json:"category"`
	Price               float64   `json:"price"`
	Rating              float64   `json:"rating"`
	Stock               int32     `json:"stock"`
	Source              string    `json:"source"`
	SourceID            string    `json:"sourceId"`
	ImportedAt          time.Time `json:"importedAt"`
	VisibilityStatus    string    `json:"visibilityStatus"`
	Revision            int64     `json:"revision"`
	UpdatedAt           time.Time `json:"updatedAt"`
	OverriddenFields    []string  `json:"overriddenFields"`
}

// ProductUpdateInput uses nil to leave a field unchanged. ClearFields explicitly
// removes an override and restores the current supplier value for that field.
type ProductUpdateInput struct {
	ExpectedRevision int64    `json:"expectedRevision"`
	Title       *string  `json:"title,omitempty"`
	Description *string  `json:"description,omitempty"`
	Brand       *string  `json:"brand,omitempty"`
	ClearFields []string `json:"clearFields,omitempty"`
}

type productOverrideState struct {
	Title            *string `json:"title"`
	Description      *string `json:"description"`
	Brand            *string `json:"brand"`
	VisibilityStatus string  `json:"visibilityStatus"`
	Revision         int64   `json:"-"`
}

func (s productOverrideState) snapshot() ([]byte, error) {
	return json.Marshal(struct {
		Title            *string `json:"title"`
		Description      *string `json:"description"`
		Brand            *string `json:"brand"`
		VisibilityStatus string  `json:"visibilityStatus"`
	}{s.Title, s.Description, s.Brand, s.VisibilityStatus})
}

func beginAdminProductTransaction(
	ctx context.Context,
	pool *pgxpool.Pool,
	sessionToken string,
) (context.Context, context.CancelFunc, pgx.Tx, string, error) {
	if strings.TrimSpace(sessionToken) == "" {
		return nil, nil, nil, "", auth.ErrUnauthenticated
	}
	if pool == nil {
		return nil, nil, nil, "", ErrDatabaseUnavailable
	}
	queryCtx, cancel := context.WithTimeout(ctx, productManagementQueryTimeout)
	principal, err := auth.ResolveSession(queryCtx, pool, sessionToken)
	if err != nil {
		cancel()
		return nil, nil, nil, "", err
	}
	tx, err := pool.Begin(queryCtx)
	if err != nil {
		cancel()
		return nil, nil, nil, "", fmt.Errorf("begin admin product transaction: %w", err)
	}
	if err := auth.AuthorizeAdminTx(queryCtx, tx, principal, auth.AuthorizationAdminManageProducts); err != nil {
		_ = tx.Rollback(queryCtx)
		cancel()
		return nil, nil, nil, "", err
	}
	return queryCtx, cancel, tx, principal.UserID, nil
}

// ReadManagedProduct returns the current editorial view and its untouched
// supplier values. Archived products remain readable to administrators so they
// can be inspected and restored.
func ReadManagedProduct(
	ctx context.Context,
	pool *pgxpool.Pool,
	sessionToken string,
	productID int64,
) (ProductManagementRecord, error) {
	if productID <= 0 {
		return ProductManagementRecord{}, ErrInvalidProductInput
	}
	queryCtx, cancel, tx, _, err := beginAdminProductTransaction(ctx, pool, sessionToken)
	if err != nil {
		return ProductManagementRecord{}, err
	}
	defer cancel()
	defer func() { _ = tx.Rollback(queryCtx) }()

	record, err := readManagedProductTx(queryCtx, tx, productID, false)
	if err != nil {
		return ProductManagementRecord{}, err
	}
	if err := tx.Commit(queryCtx); err != nil {
		return ProductManagementRecord{}, fmt.Errorf("commit managed product read: %w", err)
	}
	return record, nil
}

// UpdateManagedProduct changes only the admin overlay. Raw imported title,
// description, brand, pricing, rating, stock and source provenance are never
// overwritten by an editorial edit.
func UpdateManagedProduct(
	ctx context.Context,
	pool *pgxpool.Pool,
	sessionToken string,
	productID int64,
	input ProductUpdateInput,
) (ProductManagementRecord, error) {
	if productID <= 0 || input.ExpectedRevision < 0 {
		return ProductManagementRecord{}, ErrInvalidProductInput
	}
	normalized, clear, err := normalizeProductUpdate(input)
	if err != nil {
		return ProductManagementRecord{}, err
	}
	if normalized.Title == nil && normalized.Description == nil && normalized.Brand == nil && len(clear) == 0 {
		return ProductManagementRecord{}, ErrNoProductChanges
	}

	queryCtx, cancel, tx, actorID, err := beginAdminProductTransaction(ctx, pool, sessionToken)
	if err != nil {
		return ProductManagementRecord{}, err
	}
	defer cancel()
	defer func() { _ = tx.Rollback(queryCtx) }()

	record, err := readManagedProductTx(queryCtx, tx, productID, true)
	if err != nil {
		return ProductManagementRecord{}, err
	}
	before, exists, err := loadProductOverrideState(queryCtx, tx, productID)
	if err != nil {
		return ProductManagementRecord{}, err
	}
	if !exists {
		before.VisibilityStatus = "active"
	}
	if before.Revision != input.ExpectedRevision {
		return ProductManagementRecord{}, ErrStaleProductRevision
	}

	after := before
	changed := make([]string, 0, 3)
	apply := func(name string, supplied *string, raw string, current **string) {
		if supplied == nil && !clear[name] {
			return
		}
		var next *string
		if supplied != nil && *supplied != raw {
			value := *supplied
			next = &value
		}
		if equalOptionalString(*current, next) {
			return
		}
		*current = next
		changed = append(changed, name)
	}
	apply("title", normalized.Title, record.OriginalTitle, &after.Title)
	apply("description", normalized.Description, record.OriginalDescription, &after.Description)
	apply("brand", normalized.Brand, record.OriginalBrand, &after.Brand)
	if len(changed) == 0 {
		return ProductManagementRecord{}, ErrNoProductChanges
	}

	beforeJSON, err := before.snapshot()
	if err != nil {
		return ProductManagementRecord{}, fmt.Errorf("encode product revision before state: %w", err)
	}
	after.Revision = before.Revision + 1
	afterJSON, err := after.snapshot()
	if err != nil {
		return ProductManagementRecord{}, fmt.Errorf("encode product revision after state: %w", err)
	}
	if err := persistProductOverrideState(queryCtx, tx, productID, actorID, after); err != nil {
		return ProductManagementRecord{}, err
	}
	if err := appendProductRevision(queryCtx, tx, productID, actorID, "updated", after.Revision, beforeJSON, afterJSON, changed, after.VisibilityStatus); err != nil {
		return ProductManagementRecord{}, err
	}

	record, err = readManagedProductTx(queryCtx, tx, productID, false)
	if err != nil {
		return ProductManagementRecord{}, err
	}
	if err := tx.Commit(queryCtx); err != nil {
		return ProductManagementRecord{}, fmt.Errorf("commit product editorial update: %w", err)
	}
	return record, nil
}

// ArchiveManagedProduct hides a source product from the read-only catalog API
// without deleting its supplier record or vendor offers.
func ArchiveManagedProduct(
	ctx context.Context,
	pool *pgxpool.Pool,
	sessionToken string,
	productID int64,
	expectedRevision int64,
) (ProductManagementRecord, error) {
	return setManagedProductArchived(ctx, pool, sessionToken, productID, expectedRevision, true)
}

// RestoreManagedProduct makes an archived source product visible again while
// preserving its supplier fields and editorial overrides.
func RestoreManagedProduct(
	ctx context.Context,
	pool *pgxpool.Pool,
	sessionToken string,
	productID int64,
	expectedRevision int64,
) (ProductManagementRecord, error) {
	return setManagedProductArchived(ctx, pool, sessionToken, productID, expectedRevision, false)
}

func setManagedProductArchived(
	ctx context.Context,
	pool *pgxpool.Pool,
	sessionToken string,
	productID int64,
	expectedRevision int64,
	archive bool,
) (ProductManagementRecord, error) {
	if productID <= 0 || expectedRevision < 0 {
		return ProductManagementRecord{}, ErrInvalidProductInput
	}
	queryCtx, cancel, tx, actorID, err := beginAdminProductTransaction(ctx, pool, sessionToken)
	if err != nil {
		return ProductManagementRecord{}, err
	}
	defer cancel()
	defer func() { _ = tx.Rollback(queryCtx) }()

	if _, err := readManagedProductTx(queryCtx, tx, productID, true); err != nil {
		return ProductManagementRecord{}, err
	}
	before, exists, err := loadProductOverrideState(queryCtx, tx, productID)
	if err != nil {
		return ProductManagementRecord{}, err
	}
	if !exists {
		before.VisibilityStatus = "active"
	}
	if before.Revision != expectedRevision {
		return ProductManagementRecord{}, ErrStaleProductRevision
	}
	if archive && before.VisibilityStatus == "archived" {
		return ProductManagementRecord{}, ErrProductAlreadyArchived
	}
	if !archive && before.VisibilityStatus != "archived" {
		return ProductManagementRecord{}, ErrProductNotArchived
	}

	after := before
	action := "archived"
	if archive {
		after.VisibilityStatus = "archived"
	} else {
		after.VisibilityStatus = "active"
		action = "restored"
	}
	beforeJSON, err := before.snapshot()
	if err != nil {
		return ProductManagementRecord{}, fmt.Errorf("encode product visibility before state: %w", err)
	}
	after.Revision = before.Revision + 1
	afterJSON, err := after.snapshot()
	if err != nil {
		return ProductManagementRecord{}, fmt.Errorf("encode product visibility after state: %w", err)
	}
	if err := persistProductOverrideState(queryCtx, tx, productID, actorID, after); err != nil {
		return ProductManagementRecord{}, err
	}
	if err := appendProductRevision(queryCtx, tx, productID, actorID, action, after.Revision, beforeJSON, afterJSON, []string{"visibilityStatus"}, after.VisibilityStatus); err != nil {
		return ProductManagementRecord{}, err
	}
	record, err := readManagedProductTx(queryCtx, tx, productID, false)
	if err != nil {
		return ProductManagementRecord{}, err
	}
	if err := tx.Commit(queryCtx); err != nil {
		return ProductManagementRecord{}, fmt.Errorf("commit product visibility change: %w", err)
	}
	return record, nil
}

func normalizeProductUpdate(input ProductUpdateInput) (ProductUpdateInput, map[string]bool, error) {
	normalized := ProductUpdateInput{}
	clear := make(map[string]bool, len(input.ClearFields))
	for _, field := range input.ClearFields {
		field = strings.TrimSpace(field)
		if field != "title" && field != "description" && field != "brand" {
			return normalized, nil, fmt.Errorf("%w: unsupported clear field %q", ErrInvalidProductInput, field)
		}
		if clear[field] {
			return normalized, nil, fmt.Errorf("%w: duplicate clear field %q", ErrInvalidProductInput, field)
		}
		clear[field] = true
	}
	normalize := func(name string, supplied *string, maxRunes int, allowEmpty bool) (*string, error) {
		if supplied == nil {
			return nil, nil
		}
		value := strings.TrimSpace(*supplied)
		if clear[name] {
			return nil, fmt.Errorf("%w: field %q cannot be set and cleared together", ErrInvalidProductInput, name)
		}
		if (!allowEmpty && value == "") || utf8.RuneCountInString(value) > maxRunes {
			requirement := "text"
			if !allowEmpty {
				requirement = "non-empty text"
			}
			return nil, fmt.Errorf("%w: field %q must contain %s and be at most %d characters", ErrInvalidProductInput, name, requirement, maxRunes)
		}
		normalizedValue := value
		return &normalizedValue, nil
	}
	var err error
	if normalized.Title, err = normalize("title", input.Title, 300, false); err != nil {
		return ProductUpdateInput{}, nil, err
	}
	if normalized.Description, err = normalize("description", input.Description, 10000, true); err != nil {
		return ProductUpdateInput{}, nil, err
	}
	if normalized.Brand, err = normalize("brand", input.Brand, 120, true); err != nil {
		return ProductUpdateInput{}, nil, err
	}
	return normalized, clear, nil
}

func equalOptionalString(left, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func optionalString(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}

func loadProductOverrideState(
	ctx context.Context,
	tx pgx.Tx,
	productID int64,
) (productOverrideState, bool, error) {
	var state productOverrideState
	var title, description, brand pgtype.Text
	err := tx.QueryRow(ctx, `
		SELECT title_override, description_override, brand_override,
		       visibility_status, revision
		FROM admin_product_overrides
		WHERE product_id = $1
		FOR UPDATE
	`, productID).Scan(&title, &description, &brand, &state.VisibilityStatus, &state.Revision)
	if errors.Is(err, pgx.ErrNoRows) {
		return productOverrideState{VisibilityStatus: "active"}, false, nil
	}
	if err != nil {
		return productOverrideState{}, false, fmt.Errorf("read product editorial state: %w", err)
	}
	state.Title = optionalString(title)
	state.Description = optionalString(description)
	state.Brand = optionalString(brand)
	return state, true, nil
}

func persistProductOverrideState(
	ctx context.Context,
	tx pgx.Tx,
	productID int64,
	actorID string,
	after productOverrideState,
) error {
	var archivedBy any
	if after.VisibilityStatus == "archived" {
		archivedBy = actorID
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO admin_product_overrides(
			product_id, title_override, description_override, brand_override,
			visibility_status, revision, created_by_user_id, updated_by_user_id,
			archived_by_user_id, archived_at
		)
		VALUES($1,$2,$3,$4,$5,$6,$7,$7,$8,
			CASE WHEN $5 = 'archived' THEN now() ELSE NULL END)
		ON CONFLICT(product_id) DO UPDATE SET
			title_override = EXCLUDED.title_override,
			description_override = EXCLUDED.description_override,
			brand_override = EXCLUDED.brand_override,
			visibility_status = EXCLUDED.visibility_status,
			revision = EXCLUDED.revision,
			updated_by_user_id = EXCLUDED.updated_by_user_id,
			archived_by_user_id = CASE
				WHEN EXCLUDED.visibility_status = 'archived'
				THEN COALESCE(admin_product_overrides.archived_by_user_id, EXCLUDED.archived_by_user_id)
				ELSE NULL
			END,
			archived_at = CASE
				WHEN EXCLUDED.visibility_status = 'archived'
				THEN COALESCE(admin_product_overrides.archived_at, EXCLUDED.archived_at)
				ELSE NULL
			END,
			updated_at = now()
	`, productID, beforeOrAfterString(after.Title), beforeOrAfterString(after.Description),
		beforeOrAfterString(after.Brand), after.VisibilityStatus, after.Revision, actorID, archivedBy)
	if err != nil {
		return fmt.Errorf("persist product editorial state: %w", err)
	}
	return nil
}

func beforeOrAfterString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func appendProductRevision(
	ctx context.Context,
	tx pgx.Tx,
	productID int64,
	actorID string,
	action string,
	revision int64,
	beforeJSON []byte,
	afterJSON []byte,
	changedFields []string,
	visibilityStatus string,
) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO admin_product_revisions(
			product_id, revision, action, before_state, after_state, actor_user_id
		)
		VALUES($1,$2,$3,$4::jsonb,$5::jsonb,$6)
	`, productID, revision, action, beforeJSON, afterJSON, actorID)
	if err != nil {
		return fmt.Errorf("append admin product revision: %w", err)
	}
	fieldsJSON, err := json.Marshal(changedFields)
	if err != nil {
		return fmt.Errorf("encode product audit changed fields: %w", err)
	}
	eventType := "admin.product.updated"
	switch action {
	case "updated":
	case "archived":
		eventType = "admin.product.archived"
	case "restored":
		eventType = "admin.product.restored"
	default:
		return fmt.Errorf("invalid admin product revision action %q", action)
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO audit_events(
			event_type, actor_user_id, resource_type, resource_id, metadata
		)
		VALUES($1,$2,'product',$3,
			jsonb_build_object('revision',$4::bigint,'changed_fields',$5::jsonb,'visibility_status',$6::text))
	`, eventType, actorID, strconv.FormatInt(productID, 10), revision, fieldsJSON, visibilityStatus)
	if err != nil {
		return fmt.Errorf("append admin product audit event: %w", err)
	}
	return nil
}

func readManagedProductTx(
	ctx context.Context,
	tx pgx.Tx,
	productID int64,
	lockProduct bool,
) (ProductManagementRecord, error) {
	query := `
		SELECT p.id, p.title, p.description, p.brand,
		       COALESCE(apo.title_override, p.title),
		       COALESCE(apo.description_override, p.description),
		       COALESCE(apo.brand_override, p.brand),
		       p.category, p.price::float8, p.rating::float8, p.stock,
		       p.source, p.source_id, p.imported_at,
		       COALESCE(apo.visibility_status, 'active'),
		       COALESCE(apo.revision, 0),
		       COALESCE(apo.updated_at, p.updated_at),
		       apo.title_override, apo.description_override, apo.brand_override
		FROM products AS p
		LEFT JOIN admin_product_overrides AS apo ON apo.product_id = p.id
		WHERE p.id = $1
	`
	if lockProduct {
		query += " FOR UPDATE OF p"
	}
	var record ProductManagementRecord
	var title, description, brand pgtype.Text
	err := tx.QueryRow(ctx, query, productID).Scan(
		&record.ID, &record.OriginalTitle, &record.OriginalDescription, &record.OriginalBrand,
		&record.Title, &record.Description, &record.Brand, &record.Category,
		&record.Price, &record.Rating, &record.Stock, &record.Source, &record.SourceID,
		&record.ImportedAt, &record.VisibilityStatus, &record.Revision, &record.UpdatedAt,
		&title, &description, &brand,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ProductManagementRecord{}, ErrManagedProductNotFound
	}
	if err != nil {
		return ProductManagementRecord{}, fmt.Errorf("read managed product: %w", err)
	}
	record.OverriddenFields = make([]string, 0, 3)
	if title.Valid {
		record.OverriddenFields = append(record.OverriddenFields, "title")
	}
	if description.Valid {
		record.OverriddenFields = append(record.OverriddenFields, "description")
	}
	if brand.Valid {
		record.OverriddenFields = append(record.OverriddenFields, "brand")
	}
	return record, nil
}
