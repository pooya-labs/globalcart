package commerce

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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pooya16pg-spec/globalcart/backend/internal/auth"
)

const sellerOfferTimeout = 5 * time.Second

var (
	ErrInvalidSellerOfferInput     = errors.New("invalid seller offer input")
	ErrSellerOfferConflict         = errors.New("seller offer or SKU already exists")
	ErrSellerOfferNotFound         = errors.New("seller offer not found")
	ErrSellerOfferArchived         = errors.New("seller offer is archived")
	ErrSellerOfferRevisionConflict = errors.New("seller offer revision is stale")
	ErrSellerOfferNoChanges        = errors.New("seller offer update contains no changes")
)

type SellerOfferCreateInput struct {
	ProductID int64  `json:"productId"`
	VariantID *int64 `json:"variantId,omitempty"`
	SellerSKU string `json:"sellerSku,omitempty"`
	Price     string `json:"price"`
	Stock     int32  `json:"stock"`
}

type SellerOfferUpdateInput struct {
	ExpectedRevision int64   `json:"expectedRevision"`
	SellerSKU        *string `json:"sellerSku,omitempty"`
	Price            *string `json:"price,omitempty"`
	Stock            *int32  `json:"stock,omitempty"`
	Status           *string `json:"status,omitempty"`
}

type SellerOfferRecord struct {
	OfferID   int64     `json:"offerId"`
	VendorID  int64     `json:"vendorId"`
	ProductID int64     `json:"productId"`
	VariantID *int64    `json:"variantId,omitempty"`
	SellerSKU string    `json:"sellerSku,omitempty"`
	Price     string    `json:"price"`
	Currency  string    `json:"currency"`
	Stock     int32     `json:"stock"`
	Status    string    `json:"status"`
	Revision  int64     `json:"revision"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func normalizeSellerOfferPrice(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	parts := strings.Split(raw, ".")
	if len(parts) < 1 || len(parts) > 2 || parts[0] == "" || len(parts[0]) > 10 ||
		(len(parts[0]) > 1 && parts[0][0] == '0') {
		return "", ErrInvalidSellerOfferInput
	}
	for _, r := range parts[0] {
		if r < '0' || r > '9' {
			return "", ErrInvalidSellerOfferInput
		}
	}
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || whole > 9_999_999_999 {
		return "", ErrInvalidSellerOfferInput
	}
	fraction := "00"
	if len(parts) == 2 {
		if len(parts[1]) < 1 || len(parts[1]) > 2 {
			return "", ErrInvalidSellerOfferInput
		}
		for _, r := range parts[1] {
			if r < '0' || r > '9' {
				return "", ErrInvalidSellerOfferInput
			}
		}
		fraction = parts[1]
		if len(fraction) == 1 {
			fraction += "0"
		}
	}
	cents, err := strconv.ParseInt(fraction, 10, 64)
	if err != nil || whole*100+cents <= 0 {
		return "", ErrInvalidSellerOfferInput
	}
	return fmt.Sprintf("%d.%02d", whole, cents), nil
}

func normalizeSellerOfferSKU(value string) (string, error) {
	value = strings.TrimSpace(value)
	if !utf8.ValidString(value) || strings.ContainsRune(value, 0) || utf8.RuneCountInString(value) > 120 {
		return "", ErrInvalidSellerOfferInput
	}
	return value, nil
}

func isSellerOfferUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func normalizeSellerOfferCreate(in SellerOfferCreateInput) (SellerOfferCreateInput, error) {
	if in.ProductID <= 0 || (in.VariantID != nil && *in.VariantID <= 0) || in.Stock < 0 {
		return SellerOfferCreateInput{}, ErrInvalidSellerOfferInput
	}
	var err error
	in.SellerSKU, err = normalizeSellerOfferSKU(in.SellerSKU)
	if err != nil {
		return SellerOfferCreateInput{}, err
	}
	in.Price, err = normalizeSellerOfferPrice(in.Price)
	if err != nil {
		return SellerOfferCreateInput{}, err
	}
	return in, nil
}

func normalizeSellerOfferUpdate(in SellerOfferUpdateInput) (SellerOfferUpdateInput, error) {
	if in.ExpectedRevision < 0 {
		return SellerOfferUpdateInput{}, ErrInvalidSellerOfferInput
	}
	if in.SellerSKU != nil {
		value, err := normalizeSellerOfferSKU(*in.SellerSKU)
		if err != nil {
			return SellerOfferUpdateInput{}, err
		}
		in.SellerSKU = &value
	}
	if in.Price != nil {
		value, err := normalizeSellerOfferPrice(*in.Price)
		if err != nil {
			return SellerOfferUpdateInput{}, err
		}
		in.Price = &value
	}
	if in.Stock != nil && *in.Stock < 0 {
		return SellerOfferUpdateInput{}, ErrInvalidSellerOfferInput
	}
	if in.Status != nil {
		status := strings.ToLower(strings.TrimSpace(*in.Status))
		if status != "draft" && status != "paused" {
			return SellerOfferUpdateInput{}, ErrInvalidSellerOfferInput
		}
		in.Status = &status
	}
	if in.SellerSKU == nil && in.Price == nil && in.Stock == nil && in.Status == nil {
		return SellerOfferUpdateInput{}, ErrSellerOfferNoChanges
	}
	return in, nil
}

func readSellerOfferTx(ctx context.Context, tx pgx.Tx, vendorID, offerID int64, lock bool) (SellerOfferRecord, error) {
	var out SellerOfferRecord
	var variant pgtype.Int8
	var sku pgtype.Text
	query := `SELECT id,vendor_id,product_id,variant_id,seller_sku,price::text,currency,
		stock,status,revision,created_at,updated_at
		FROM vendor_products WHERE vendor_id=$1 AND id=$2`
	if lock {
		query += " FOR UPDATE"
	}
	err := tx.QueryRow(ctx, query, vendorID, offerID).Scan(
		&out.OfferID, &out.VendorID, &out.ProductID, &variant, &sku, &out.Price,
		&out.Currency, &out.Stock, &out.Status, &out.Revision, &out.CreatedAt, &out.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return SellerOfferRecord{}, ErrSellerOfferNotFound
	}
	if err != nil {
		return SellerOfferRecord{}, err
	}
	if variant.Valid {
		id := variant.Int64
		out.VariantID = &id
	}
	if sku.Valid {
		out.SellerSKU = sku.String
	}
	return out, nil
}

func validateSellerOfferTargetTx(ctx context.Context, tx pgx.Tx, productID int64, variantID *int64) error {
	var found int64
	var err error
	if variantID == nil {
		err = tx.QueryRow(ctx, `
			SELECT p.id FROM products p
			LEFT JOIN admin_product_overrides apo ON apo.product_id=p.id
			WHERE p.id=$1 AND COALESCE(apo.visibility_status,'active')='active'
			FOR SHARE OF p
		`, productID).Scan(&found)
	} else {
		err = tx.QueryRow(ctx, `
			SELECT p.id FROM products p
			JOIN product_variants pv ON pv.product_id=p.id
			LEFT JOIN admin_product_overrides apo ON apo.product_id=p.id
			WHERE p.id=$1 AND pv.id=$2 AND pv.is_active
			  AND COALESCE(apo.visibility_status,'active')='active'
			FOR SHARE OF p,pv
		`, productID, *variantID).Scan(&found)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrInvalidProduct
	}
	if err != nil {
		return fmt.Errorf("validate seller offer target: %w", err)
	}
	return nil
}

func sellerOfferSnapshot(v SellerOfferRecord) map[string]any {
	return map[string]any{
		"offerId": v.OfferID, "vendorId": v.VendorID, "productId": v.ProductID,
		"variantId": v.VariantID, "sellerSku": v.SellerSKU, "price": v.Price,
		"currency": v.Currency, "stock": v.Stock, "status": v.Status,
	}
}

func appendSellerOfferHistoryTx(ctx context.Context, tx pgx.Tx, actorID, action string,
	before *SellerOfferRecord, after SellerOfferRecord, changed []string) error {
	beforeJSON := []byte("{}")
	var err error
	if before != nil {
		beforeJSON, err = json.Marshal(sellerOfferSnapshot(*before))
		if err != nil {
			return fmt.Errorf("encode seller offer before state: %w", err)
		}
	}
	afterJSON, err := json.Marshal(sellerOfferSnapshot(after))
	if err != nil {
		return fmt.Errorf("encode seller offer after state: %w", err)
	}
	fieldsJSON, err := json.Marshal(changed)
	if err != nil {
		return fmt.Errorf("encode seller offer changed fields: %w", err)
	}
	if _, err = tx.Exec(ctx, `
		INSERT INTO vendor_product_revisions(vendor_product_id,revision,action,before_state,after_state,actor_user_id)
		VALUES($1,$2,$3,$4::jsonb,$5::jsonb,$6)
	`, after.OfferID, after.Revision, action, beforeJSON, afterJSON, actorID); err != nil {
		return fmt.Errorf("append seller offer revision: %w", err)
	}
	eventType := "seller.offer.updated"
	if action == "created" {
		eventType = "seller.offer.created"
	}
	if action == "archived" {
		eventType = "seller.offer.archived"
	}
	if _, err = tx.Exec(ctx, `
		INSERT INTO audit_events(event_type,actor_user_id,resource_type,resource_id,metadata)
		VALUES($1,$2,'vendor_offer',$3,jsonb_build_object(
			'vendor_id',$4::bigint,'revision',$5::bigint,
			'changed_fields',$6::jsonb,'before',$7::jsonb,'after',$8::jsonb
		))
	`, eventType, actorID, strconv.FormatInt(after.OfferID, 10), after.VendorID,
		after.Revision, fieldsJSON, beforeJSON, afterJSON); err != nil {
		return fmt.Errorf("append seller offer audit: %w", err)
	}
	return nil
}

func sellerOfferTx(ctx context.Context, s *Service) (context.Context, context.CancelFunc, pgx.Tx, error) {
	if s == nil || s.pool == nil {
		return nil, nil, nil, errors.New("seller offer database is unavailable")
	}
	qctx, cancel := context.WithTimeout(ctx, sellerOfferTimeout)
	tx, err := s.pool.BeginTx(qctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		cancel()
		return nil, nil, nil, fmt.Errorf("begin seller offer transaction: %w", err)
	}
	return qctx, cancel, tx, nil
}

// New seller offers always begin in draft; a seller cannot self-publish.
func (s *Service) CreateVendorOfferForSeller(ctx context.Context, principal auth.SessionPrincipal,
	vendorID int64, input SellerOfferCreateInput) (SellerOfferRecord, error) {
	if vendorID <= 0 {
		return SellerOfferRecord{}, ErrInvalidSellerOfferInput
	}
	in, err := normalizeSellerOfferCreate(input)
	if err != nil {
		return SellerOfferRecord{}, err
	}
	qctx, cancel, tx, err := sellerOfferTx(ctx, s)
	if err != nil {
		return SellerOfferRecord{}, err
	}
	defer cancel()
	defer func() { _ = tx.Rollback(context.Background()) }()
	if err = auth.AuthorizeSellerVendorTx(qctx, tx, principal, auth.AuthorizationSellerWriteVendor, vendorID); err != nil {
		return SellerOfferRecord{}, err
	}
	if err = validateSellerOfferTargetTx(qctx, tx, in.ProductID, in.VariantID); err != nil {
		return SellerOfferRecord{}, err
	}
	var variantArg any
	if in.VariantID != nil {
		variantArg = *in.VariantID
	}
	var out SellerOfferRecord
	var variant pgtype.Int8
	var sku pgtype.Text
	err = tx.QueryRow(qctx, `
		INSERT INTO vendor_products(vendor_id,product_id,variant_id,seller_sku,price,currency,stock,status,revision)
		VALUES($1,$2,$3,$4,$5::numeric,'USD',$6,'draft',1)
		RETURNING id,vendor_id,product_id,variant_id,seller_sku,price::text,currency,
		          stock,status,revision,created_at,updated_at
	`, vendorID, in.ProductID, variantArg, nullIfEmpty(in.SellerSKU), in.Price, in.Stock).Scan(
		&out.OfferID, &out.VendorID, &out.ProductID, &variant, &sku, &out.Price,
		&out.Currency, &out.Stock, &out.Status, &out.Revision, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		if isSellerOfferUniqueViolation(err) {
			return SellerOfferRecord{}, ErrSellerOfferConflict
		}
		return SellerOfferRecord{}, fmt.Errorf("insert seller offer: %w", err)
	}
	if variant.Valid {
		id := variant.Int64
		out.VariantID = &id
	}
	if sku.Valid {
		out.SellerSKU = sku.String
	}
	if _, err = tx.Exec(qctx, `INSERT INTO vendor_price_history(vendor_product_id,old_price,new_price,currency,change_source)
		VALUES($1,NULL,$2::numeric,'USD','merchant_update')`, out.OfferID, out.Price); err != nil {
		return SellerOfferRecord{}, fmt.Errorf("record opening seller price: %w", err)
	}
	if _, err = tx.Exec(qctx, `INSERT INTO vendor_inventory_events(vendor_product_id,quantity_delta,resulting_stock,reason,reference)
		VALUES($1,$2,$2,'opening_balance','seller_offer_created')`, out.OfferID, out.Stock); err != nil {
		return SellerOfferRecord{}, fmt.Errorf("record opening seller stock: %w", err)
	}
	if err = appendSellerOfferHistoryTx(qctx, tx, principal.UserID, "created", nil, out,
		[]string{"productId", "variantId", "sellerSku", "price", "stock", "status"}); err != nil {
		return SellerOfferRecord{}, err
	}
	if err = tx.Commit(qctx); err != nil {
		return SellerOfferRecord{}, fmt.Errorf("commit seller offer creation: %w", err)
	}
	return out, nil
}

func (s *Service) UpdateVendorOfferForSeller(ctx context.Context, principal auth.SessionPrincipal,
	offerID int64, input SellerOfferUpdateInput) (SellerOfferRecord, error) {
	if offerID <= 0 {
		return SellerOfferRecord{}, ErrInvalidSellerOfferInput
	}
	in, err := normalizeSellerOfferUpdate(input)
	if err != nil {
		return SellerOfferRecord{}, err
	}
	qctx, cancel, tx, err := sellerOfferTx(ctx, s)
	if err != nil {
		return SellerOfferRecord{}, err
	}
	defer cancel()
	defer func() { _ = tx.Rollback(context.Background()) }()
	vendorID, err := vendorIDForOffer(qctx, tx, offerID)
	if errors.Is(err, ErrVendorProductNotFound) {
		return SellerOfferRecord{}, ErrSellerOfferNotFound
	}
	if err != nil {
		return SellerOfferRecord{}, fmt.Errorf("resolve offer vendor: %w", err)
	}
	if err = auth.AuthorizeSellerVendorTx(qctx, tx, principal, auth.AuthorizationSellerWriteVendor, vendorID); err != nil {
		return SellerOfferRecord{}, err
	}
	before, err := readSellerOfferTx(qctx, tx, vendorID, offerID, true)
	if err != nil {
		return SellerOfferRecord{}, err
	}
	if before.Status == "archived" {
		return SellerOfferRecord{}, ErrSellerOfferArchived
	}
	if before.Revision != in.ExpectedRevision {
		return SellerOfferRecord{}, ErrSellerOfferRevisionConflict
	}
	after := before
	changed := make([]string, 0, 4)
	if in.SellerSKU != nil && after.SellerSKU != *in.SellerSKU {
		after.SellerSKU = *in.SellerSKU
		changed = append(changed, "sellerSku")
	}
	if in.Price != nil && after.Price != *in.Price {
		after.Price = *in.Price
		changed = append(changed, "price")
	}
	if in.Stock != nil && after.Stock != *in.Stock {
		after.Stock = *in.Stock
		changed = append(changed, "stock")
	}
	if in.Status != nil && after.Status != *in.Status {
		after.Status = *in.Status
		changed = append(changed, "status")
	}
	if len(changed) == 0 {
		return SellerOfferRecord{}, ErrSellerOfferNoChanges
	}
	err = tx.QueryRow(qctx, `
		UPDATE vendor_products SET seller_sku=NULLIF($3,''),price=$4::numeric,stock=$5,status=$6,updated_at=now()
		WHERE id=$1 AND vendor_id=$2 AND revision=$7 RETURNING revision,updated_at
	`, offerID, vendorID, after.SellerSKU, after.Price, after.Stock, after.Status, in.ExpectedRevision).
		Scan(&after.Revision, &after.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SellerOfferRecord{}, ErrSellerOfferRevisionConflict
		}
		if isSellerOfferUniqueViolation(err) {
			return SellerOfferRecord{}, ErrSellerOfferConflict
		}
		return SellerOfferRecord{}, fmt.Errorf("update seller offer: %w", err)
	}
	if before.Price != after.Price {
		if _, err = tx.Exec(qctx, `INSERT INTO vendor_price_history(vendor_product_id,old_price,new_price,currency,change_source)
			VALUES($1,$2::numeric,$3::numeric,$4,'merchant_update')`, offerID, before.Price, after.Price, before.Currency); err != nil {
			return SellerOfferRecord{}, fmt.Errorf("record seller price change: %w", err)
		}
	}
	if before.Stock != after.Stock {
		delta := after.Stock - before.Stock
		if _, err = tx.Exec(qctx, `INSERT INTO vendor_inventory_events(vendor_product_id,quantity_delta,resulting_stock,reason,reference)
			VALUES($1,$2,$3,'stock_adjustment','seller_offer_update')`, offerID, delta, after.Stock); err != nil {
			return SellerOfferRecord{}, fmt.Errorf("record seller stock change: %w", err)
		}
	}
	if err = appendSellerOfferHistoryTx(qctx, tx, principal.UserID, "updated", &before, after, changed); err != nil {
		return SellerOfferRecord{}, err
	}
	if err = tx.Commit(qctx); err != nil {
		return SellerOfferRecord{}, fmt.Errorf("commit seller offer update: %w", err)
	}
	return after, nil
}

func (s *Service) ArchiveVendorOfferForSeller(ctx context.Context, principal auth.SessionPrincipal,
	offerID, expectedRevision int64) (SellerOfferRecord, error) {
	if offerID <= 0 || expectedRevision < 0 {
		return SellerOfferRecord{}, ErrInvalidSellerOfferInput
	}
	qctx, cancel, tx, err := sellerOfferTx(ctx, s)
	if err != nil {
		return SellerOfferRecord{}, err
	}
	defer cancel()
	defer func() { _ = tx.Rollback(context.Background()) }()
	vendorID, err := vendorIDForOffer(qctx, tx, offerID)
	if errors.Is(err, ErrVendorProductNotFound) {
		return SellerOfferRecord{}, ErrSellerOfferNotFound
	}
	if err != nil {
		return SellerOfferRecord{}, fmt.Errorf("resolve offer vendor: %w", err)
	}
	if err = auth.AuthorizeSellerVendorTx(qctx, tx, principal, auth.AuthorizationSellerWriteVendor, vendorID); err != nil {
		return SellerOfferRecord{}, err
	}
	before, err := readSellerOfferTx(qctx, tx, vendorID, offerID, true)
	if err != nil {
		return SellerOfferRecord{}, err
	}
	if before.Status == "archived" {
		return SellerOfferRecord{}, ErrSellerOfferArchived
	}
	if before.Revision != expectedRevision {
		return SellerOfferRecord{}, ErrSellerOfferRevisionConflict
	}
	after := before
	after.Status = "archived"
	err = tx.QueryRow(qctx, `UPDATE vendor_products SET status='archived',updated_at=now()
		WHERE id=$1 AND vendor_id=$2 AND revision=$3 RETURNING revision,updated_at`,
		offerID, vendorID, expectedRevision).Scan(&after.Revision, &after.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SellerOfferRecord{}, ErrSellerOfferRevisionConflict
		}
		return SellerOfferRecord{}, fmt.Errorf("archive seller offer: %w", err)
	}
	if err = appendSellerOfferHistoryTx(qctx, tx, principal.UserID, "archived", &before, after, []string{"status"}); err != nil {
		return SellerOfferRecord{}, err
	}
	if err = tx.Commit(qctx); err != nil {
		return SellerOfferRecord{}, fmt.Errorf("commit seller offer archive: %w", err)
	}
	return after, nil
}
