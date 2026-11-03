package admin

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	managedProductListDefaultLimit  = 20
	managedProductListMaximumLimit  = 100
	managedProductListMaximumOffset = 10000
)

var ErrInvalidProductListInput = errors.New("invalid managed product list input")

// ManagedProductListInput is intentionally bounded so internal admin searches
// cannot turn into unbounded catalog scans or arbitrary SQL ordering.
type ManagedProductListInput struct {
	Query     string `json:"query,omitempty"`
	Status    string `json:"status,omitempty"` // active (default), archived, or all
	Source    string `json:"source,omitempty"`
	Category  string `json:"category,omitempty"`
	Brand     string `json:"brand,omitempty"`
	SortBy    string `json:"sortBy,omitempty"`
	SortOrder string `json:"sortOrder,omitempty"` // asc or desc
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}

type ManagedProductPage struct {
	Items  []ProductManagementRecord `json:"items"`
	Total  int64                     `json:"total"`
	Limit  int                       `json:"limit"`
	Offset int                       `json:"offset"`
}

// ListManagedProducts returns a bounded, administrator-only product work queue.
// Search and brand filtering use the effective editorial values, while source
// and provenance filters remain tied to the immutable supplier record.
func ListManagedProducts(
	ctx context.Context,
	pool *pgxpool.Pool,
	sessionToken string,
	input ManagedProductListInput,
) (ManagedProductPage, error) {
	filter, err := normalizeManagedProductListInput(input)
	if err != nil {
		return ManagedProductPage{}, err
	}

	queryCtx, cancel, tx, _, err := beginAdminProductTransaction(ctx, pool, sessionToken)
	if err != nil {
		return ManagedProductPage{}, err
	}
	defer cancel()
	defer func() { _ = tx.Rollback(queryCtx) }()

	const where = `
		FROM products AS p
		LEFT JOIN admin_product_overrides AS apo ON apo.product_id = p.id
		WHERE
			($1 = '' OR
				strpos(lower(COALESCE(apo.title_override, p.title)), lower($1)) > 0 OR
				strpos(lower(COALESCE(apo.description_override, p.description)), lower($1)) > 0 OR
				strpos(lower(COALESCE(apo.brand_override, p.brand)), lower($1)) > 0 OR
				strpos(lower(p.category), lower($1)) > 0 OR
				strpos(lower(p.source), lower($1)) > 0 OR
				strpos(lower(p.source_id), lower($1)) > 0 OR
				p.id::text = $1)
			AND ($2 = 'all' OR COALESCE(apo.visibility_status, 'active') = $2)
			AND ($3 = '' OR lower(p.source) = lower($3))
			AND ($4 = '' OR strpos(lower(p.category), lower($4)) > 0)
			AND ($5 = '' OR strpos(lower(COALESCE(apo.brand_override, p.brand)), lower($5)) > 0)
	`
	args := []any{filter.Query, filter.Status, filter.Source, filter.Category, filter.Brand}

	var total int64
	if err := tx.QueryRow(queryCtx, "SELECT count(*)::bigint "+where, args...).Scan(&total); err != nil {
		return ManagedProductPage{}, fmt.Errorf("count managed products: %w", err)
	}

	sortColumns := map[string]string{
		"updated_at":  "COALESCE(apo.updated_at, p.updated_at)",
		"imported_at": "p.imported_at",
		"title":       "lower(COALESCE(apo.title_override, p.title))",
		"price":       "p.price",
		"rating":      "p.rating",
		"stock":       "p.stock",
	}
	order := strings.ToUpper(filter.SortOrder)
	orderBy := sortColumns[filter.SortBy] + " " + order + ", p.id DESC"
	rows, err := tx.Query(queryCtx, `
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
	`+where+`
		ORDER BY `+orderBy+`
		LIMIT $6 OFFSET $7
	`, filter.Query, filter.Status, filter.Source, filter.Category, filter.Brand, filter.Limit, filter.Offset)
	if err != nil {
		return ManagedProductPage{}, fmt.Errorf("list managed products: %w", err)
	}
	defer rows.Close()

	items := make([]ProductManagementRecord, 0, filter.Limit)
	for rows.Next() {
		record, err := scanManagedProductRecord(rows)
		if err != nil {
			return ManagedProductPage{}, err
		}
		items = append(items, record)
	}
	if err := rows.Err(); err != nil {
		return ManagedProductPage{}, fmt.Errorf("iterate managed products: %w", err)
	}
	rows.Close()

	if err := tx.Commit(queryCtx); err != nil {
		return ManagedProductPage{}, fmt.Errorf("commit managed product list: %w", err)
	}
	return ManagedProductPage{
		Items: items, Total: total, Limit: filter.Limit, Offset: filter.Offset,
	}, nil
}

func normalizeManagedProductListInput(input ManagedProductListInput) (ManagedProductListInput, error) {
	normalized := input
	normalizeText := func(value string, maximum int) (string, error) {
		value = strings.TrimSpace(value)
		if !utf8.ValidString(value) || utf8.RuneCountInString(value) > maximum {
			return "", ErrInvalidProductListInput
		}
		return value, nil
	}

	var err error
	if normalized.Query, err = normalizeText(input.Query, 120); err != nil {
		return ManagedProductListInput{}, err
	}
	if normalized.Source, err = normalizeText(input.Source, 64); err != nil {
		return ManagedProductListInput{}, err
	}
	if normalized.Category, err = normalizeText(input.Category, 120); err != nil {
		return ManagedProductListInput{}, err
	}
	if normalized.Brand, err = normalizeText(input.Brand, 120); err != nil {
		return ManagedProductListInput{}, err
	}

	normalized.Status = strings.ToLower(strings.TrimSpace(input.Status))
	if normalized.Status == "" {
		normalized.Status = "active"
	}
	switch normalized.Status {
	case "active", "archived", "all":
	default:
		return ManagedProductListInput{}, ErrInvalidProductListInput
	}

	normalized.SortBy = strings.ToLower(strings.TrimSpace(input.SortBy))
	if normalized.SortBy == "" {
		normalized.SortBy = "updated_at"
	}
	switch normalized.SortBy {
	case "updated_at", "imported_at", "title", "price", "rating", "stock":
	default:
		return ManagedProductListInput{}, ErrInvalidProductListInput
	}

	normalized.SortOrder = strings.ToLower(strings.TrimSpace(input.SortOrder))
	if normalized.SortOrder == "" {
		normalized.SortOrder = "desc"
	}
	if normalized.SortOrder != "asc" && normalized.SortOrder != "desc" {
		return ManagedProductListInput{}, ErrInvalidProductListInput
	}

	if normalized.Limit < 0 || normalized.Limit > managedProductListMaximumLimit ||
		normalized.Offset < 0 || normalized.Offset > managedProductListMaximumOffset {
		return ManagedProductListInput{}, ErrInvalidProductListInput
	}
	if normalized.Limit == 0 {
		normalized.Limit = managedProductListDefaultLimit
	}
	return normalized, nil
}

type managedProductScanner interface {
	Scan(dest ...any) error
}

func scanManagedProductRecord(row managedProductScanner) (ProductManagementRecord, error) {
	var record ProductManagementRecord
	var title, description, brand pgtype.Text
	err := row.Scan(
		&record.ID, &record.OriginalTitle, &record.OriginalDescription, &record.OriginalBrand,
		&record.Title, &record.Description, &record.Brand, &record.Category,
		&record.Price, &record.Rating, &record.Stock, &record.Source, &record.SourceID,
		&record.ImportedAt, &record.VisibilityStatus, &record.Revision, &record.UpdatedAt,
		&title, &description, &brand,
	)
	if err != nil {
		return ProductManagementRecord{}, fmt.Errorf("scan managed product: %w", err)
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

