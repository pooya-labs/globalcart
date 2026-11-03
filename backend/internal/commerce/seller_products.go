package commerce

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pooya16pg-spec/globalcart/backend/internal/auth"
)

const (
	sellerProductListQueryTimeout  = 5 * time.Second
	sellerProductListDefaultLimit  = 20
	sellerProductListMaximumLimit  = 100
	sellerProductListMaximumOffset = 10000
)

var ErrInvalidSellerProductListInput = errors.New("invalid seller product list input")

// SellerProductListInput is bounded to keep seller catalog searches predictable.
// Vendor scope is deliberately supplied separately and is re-authorized in the
// same database transaction as the list query.
type SellerProductListInput struct {
	Query     string `json:"query,omitempty"`
	Status    string `json:"status,omitempty"` // all (default), active, draft, paused, or archived
	Category  string `json:"category,omitempty"`
	Source    string `json:"source,omitempty"`
	Brand     string `json:"brand,omitempty"`
	SortBy    string `json:"sortBy,omitempty"`
	SortOrder string `json:"sortOrder,omitempty"` // asc or desc
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}

// SellerProductRecord contains the effective storefront-facing product text
// alongside the seller's own offer values. Supplier provenance remains read-only.
type SellerProductRecord struct {
	OfferID           int64     `json:"offerId"`
	VendorID          int64     `json:"vendorId"`
	ProductID         int64     `json:"productId"`
	VariantID         *int64    `json:"variantId,omitempty"`
	SellerSKU         string    `json:"sellerSku,omitempty"`
	Title             string    `json:"title"`
	VariantTitle      string    `json:"variantTitle,omitempty"`
	Description       string    `json:"description"`
	Brand             string    `json:"brand,omitempty"`
	Category          string    `json:"category"`
	Source            string    `json:"source"`
	SourceID          string    `json:"sourceId"`
	Price             string    `json:"price"`
	Currency          string    `json:"currency"`
	Stock             int32     `json:"stock"`
	Status            string    `json:"status"`
	CatalogVisibility string    `json:"catalogVisibility"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type SellerProductPage struct {
	Items  []SellerProductRecord `json:"items"`
	Total  int64                 `json:"total"`
	Limit  int                   `json:"limit"`
	Offset int                   `json:"offset"`
}

const sellerProductListFrom = "FROM vendor_products AS vp\n" +
	"JOIN products AS p ON p.id = vp.product_id\n" +
	"LEFT JOIN product_variants AS pv ON pv.id = vp.variant_id AND pv.product_id = vp.product_id\n" +
	"LEFT JOIN admin_product_overrides AS apo ON apo.product_id = p.id\n" +
	"WHERE vp.vendor_id = $1\n" +
	"  AND ($2 = '' OR\n" +
	"    strpos(lower(COALESCE(apo.title_override, p.title)), lower($2)) > 0 OR\n" +
	"    strpos(lower(COALESCE(apo.description_override, p.description)), lower($2)) > 0 OR\n" +
	"    strpos(lower(COALESCE(apo.brand_override, p.brand)), lower($2)) > 0 OR\n" +
	"    strpos(lower(COALESCE(pv.title, '')), lower($2)) > 0 OR\n" +
	"    strpos(lower(p.category), lower($2)) > 0 OR\n" +
	"    strpos(lower(p.source), lower($2)) > 0 OR\n" +
	"    strpos(lower(p.source_id), lower($2)) > 0 OR\n" +
	"    strpos(lower(COALESCE(vp.seller_sku, '')), lower($2)) > 0 OR\n" +
	"    vp.id::text = $2 OR p.id::text = $2)\n" +
	"  AND ($3 = 'all' OR vp.status = $3)\n" +
	"  AND ($4 = '' OR strpos(lower(p.category), lower($4)) > 0)\n" +
	"  AND ($5 = '' OR lower(p.source) = lower($5))\n" +
	"  AND ($6 = '' OR strpos(lower(COALESCE(apo.brand_override, p.brand)), lower($6)) > 0)"

// ListSellerProductsForSeller lists only offers belonging to the selected vendor.
// VendorID is a selector, never proof of access: the live session, seller role,
// vendor status and exact active membership are revalidated under row locks in
// the same transaction that reads the products.
func (s *Service) ListSellerProductsForSeller(
	ctx context.Context,
	principal auth.SessionPrincipal,
	vendorID int64,
	input SellerProductListInput,
) (SellerProductPage, error) {
	if vendorID <= 0 {
		return SellerProductPage{}, ErrInvalidSellerProductListInput
	}
	filter, err := normalizeSellerProductListInput(input)
	if err != nil {
		return SellerProductPage{}, err
	}

	queryCtx, cancel := context.WithTimeout(ctx, sellerProductListQueryTimeout)
	defer cancel()

	tx, err := s.pool.BeginTx(queryCtx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return SellerProductPage{}, fmt.Errorf("begin seller product list: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	if err := auth.AuthorizeSellerVendorTx(
		queryCtx, tx, principal, auth.AuthorizationSellerReadVendor, vendorID,
	); err != nil {
		return SellerProductPage{}, err
	}

	args := []any{vendorID, filter.Query, filter.Status, filter.Category, filter.Source, filter.Brand}
	var total int64
	if err := tx.QueryRow(
		queryCtx, "SELECT count(*)::bigint "+sellerProductListFrom, args...,
	).Scan(&total); err != nil {
		return SellerProductPage{}, fmt.Errorf("count seller products: %w", err)
	}

	sortColumns := map[string]string{
		"updated_at": "vp.updated_at",
		"title":      "lower(COALESCE(apo.title_override, p.title))",
		"price":      "vp.price",
		"stock":      "vp.stock",
		"status":     "vp.status",
	}
	orderBy := sortColumns[filter.SortBy] + " " + strings.ToUpper(filter.SortOrder) + ", vp.id DESC"
	rows, err := tx.Query(queryCtx,
		"SELECT vp.id, vp.vendor_id, vp.product_id, vp.variant_id, "+
			"COALESCE(vp.seller_sku, ''), COALESCE(apo.title_override, p.title), "+
			"COALESCE(pv.title, ''), COALESCE(apo.description_override, p.description), "+
			"COALESCE(apo.brand_override, p.brand), p.category, p.source, p.source_id, "+
			"vp.price::text, vp.currency, vp.stock, vp.status, "+
			"COALESCE(apo.visibility_status, 'active'), vp.updated_at "+
			sellerProductListFrom+" ORDER BY "+orderBy+" LIMIT $7 OFFSET $8",
		vendorID, filter.Query, filter.Status, filter.Category, filter.Source, filter.Brand,
		filter.Limit, filter.Offset,
	)
	if err != nil {
		return SellerProductPage{}, fmt.Errorf("list seller products: %w", err)
	}
	defer rows.Close()

	items := make([]SellerProductRecord, 0, filter.Limit)
	for rows.Next() {
		var item SellerProductRecord
		var variantID pgtype.Int8
		if err := rows.Scan(
			&item.OfferID, &item.VendorID, &item.ProductID, &variantID,
			&item.SellerSKU, &item.Title, &item.VariantTitle, &item.Description,
			&item.Brand, &item.Category, &item.Source, &item.SourceID,
			&item.Price, &item.Currency, &item.Stock, &item.Status,
			&item.CatalogVisibility, &item.UpdatedAt,
		); err != nil {
			return SellerProductPage{}, fmt.Errorf("scan seller product: %w", err)
		}
		if variantID.Valid {
			id := variantID.Int64
			item.VariantID = &id
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return SellerProductPage{}, fmt.Errorf("iterate seller products: %w", err)
	}
	rows.Close()

	if err := tx.Commit(queryCtx); err != nil {
		return SellerProductPage{}, fmt.Errorf("commit seller product list: %w", err)
	}
	return SellerProductPage{Items: items, Total: total, Limit: filter.Limit, Offset: filter.Offset}, nil
}

func normalizeSellerProductListInput(input SellerProductListInput) (SellerProductListInput, error) {
	normalized := input
	normalizeText := func(value string, maximum int) (string, error) {
		value = strings.TrimSpace(value)
		if !utf8.ValidString(value) || strings.ContainsRune(value, rune(0)) ||
			utf8.RuneCountInString(value) > maximum {
			return "", ErrInvalidSellerProductListInput
		}
		return value, nil
	}

	var err error
	if normalized.Query, err = normalizeText(input.Query, 120); err != nil {
		return SellerProductListInput{}, err
	}
	if normalized.Category, err = normalizeText(input.Category, 120); err != nil {
		return SellerProductListInput{}, err
	}
	if normalized.Source, err = normalizeText(input.Source, 64); err != nil {
		return SellerProductListInput{}, err
	}
	if normalized.Brand, err = normalizeText(input.Brand, 120); err != nil {
		return SellerProductListInput{}, err
	}

	normalized.Status = strings.ToLower(strings.TrimSpace(input.Status))
	if normalized.Status == "" {
		normalized.Status = "all"
	}
	switch normalized.Status {
	case "all", "active", "draft", "paused", "archived":
	default:
		return SellerProductListInput{}, ErrInvalidSellerProductListInput
	}

	normalized.SortBy = strings.ToLower(strings.TrimSpace(input.SortBy))
	if normalized.SortBy == "" {
		normalized.SortBy = "updated_at"
	}
	switch normalized.SortBy {
	case "updated_at", "title", "price", "stock", "status":
	default:
		return SellerProductListInput{}, ErrInvalidSellerProductListInput
	}

	normalized.SortOrder = strings.ToLower(strings.TrimSpace(input.SortOrder))
	if normalized.SortOrder == "" {
		normalized.SortOrder = "desc"
	}
	if normalized.SortOrder != "asc" && normalized.SortOrder != "desc" {
		return SellerProductListInput{}, ErrInvalidSellerProductListInput
	}

	if normalized.Limit < 0 || normalized.Limit > sellerProductListMaximumLimit ||
		normalized.Offset < 0 || normalized.Offset > sellerProductListMaximumOffset {
		return SellerProductListInput{}, ErrInvalidSellerProductListInput
	}
	if normalized.Limit == 0 {
		normalized.Limit = sellerProductListDefaultLimit
	}
	return normalized, nil
}
