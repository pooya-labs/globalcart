package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const manualProductSource = "admin_manual"
const manualProductDecimalPattern = `^(0|[1-9][0-9]{0,9})(\.[0-9]{1,2})?$`

var manualProductDecimalRegex = regexp.MustCompile(manualProductDecimalPattern)

type ProductCreateInput struct {
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	Category           string   `json:"category"`
	Brand              string   `json:"brand"`
	Price              string   `json:"price"`
	DiscountPercentage string   `json:"discountPercentage"`
	Rating             string   `json:"rating"`
	Stock              int32    `json:"stock"`
	Thumbnail          string   `json:"thumbnail"`
	Images             []string `json:"images"`
}

type normalizedProductCreateInput struct {
	Title              string
	Description        string
	Category           string
	Brand              string
	Price              string
	DiscountPercentage string
	Rating             string
	Stock              int32
	Thumbnail          string
	ImagesJSON         []byte
}

// CreateManagedProduct creates a product with explicit admin-owned source
// provenance and a GlobalCart Direct demo offer. Product, category, offer,
// opening price/stock history, revision and audit event commit atomically.
func CreateManagedProduct(ctx context.Context, pool *pgxpool.Pool, sessionToken string, input ProductCreateInput) (ProductManagementRecord, error) {
	normalized, err := normalizeProductCreate(input)
	if err != nil {
		return ProductManagementRecord{}, err
	}
	queryCtx, cancel, tx, actorID, err := beginAdminProductTransaction(ctx, pool, sessionToken)
	if err != nil {
		return ProductManagementRecord{}, err
	}
	defer cancel()
	defer func() { _ = tx.Rollback(queryCtx) }()

	var categoryID int64
	err = tx.QueryRow(queryCtx, `
        INSERT INTO categories(name) VALUES($1)
        ON CONFLICT (lower(btrim(name))) DO NOTHING
        RETURNING id
    `, normalized.Category).Scan(&categoryID)
	if errors.Is(err, pgx.ErrNoRows) {
		err = tx.QueryRow(queryCtx, "SELECT id FROM categories WHERE lower(btrim(name)) = lower(btrim($1))", normalized.Category).Scan(&categoryID)
	}
	if err != nil {
		return ProductManagementRecord{}, fmt.Errorf("create or resolve product category: %w", err)
	}

	var productID int64
	err = tx.QueryRow(queryCtx, `
        INSERT INTO products(
            id, title, description, category, category_id, price, discount_percentage, rating,
            stock, brand, thumbnail, images, source, source_id, source_url, imported_at
        )
        VALUES(
            nextval('admin_manual_product_id_seq'), $1, $2, $3, $4, $5::numeric,
            $6::numeric, $7::numeric, $8, $9, $10, $11::jsonb, $12, gen_random_uuid()::text, '', now()
        )
        RETURNING id
    `, normalized.Title, normalized.Description, normalized.Category, categoryID, normalized.Price,
		normalized.DiscountPercentage, normalized.Rating, normalized.Stock, normalized.Brand,
		normalized.Thumbnail, normalized.ImagesJSON, manualProductSource).Scan(&productID)
	if err != nil {
		return ProductManagementRecord{}, fmt.Errorf("insert manual product: %w", err)
	}

	var platformVendorID int64
	err = tx.QueryRow(queryCtx, `
        SELECT id FROM vendors
        WHERE slug = 'globalcart-direct' AND vendor_type = 'platform' AND status = 'active'
        FOR SHARE
    `).Scan(&platformVendorID)
	if err != nil {
		return ProductManagementRecord{}, fmt.Errorf("resolve active platform vendor for manual product: %w", err)
	}
	var offerID int64
	err = tx.QueryRow(queryCtx, `
        INSERT INTO vendor_products(vendor_id, product_id, price, currency, stock, status)
        VALUES($1, $2, $3::numeric, 'USD', $4, 'active') RETURNING id
    `, platformVendorID, productID, normalized.Price, normalized.Stock).Scan(&offerID)
	if err != nil {
		return ProductManagementRecord{}, fmt.Errorf("create platform demo offer for manual product: %w", err)
	}
	if _, err := tx.Exec(queryCtx, `
        INSERT INTO vendor_price_history(vendor_product_id, old_price, new_price, currency, change_source)
        VALUES($1, NULL, $2::numeric, 'USD', 'admin_adjustment')
    `, offerID, normalized.Price); err != nil {
		return ProductManagementRecord{}, fmt.Errorf("record manual product opening price: %w", err)
	}
	if _, err := tx.Exec(queryCtx, `
        INSERT INTO vendor_inventory_events(vendor_product_id, quantity_delta, resulting_stock, reason, reference)
        VALUES($1, $2, $2, 'opening_balance', 'admin_product_created')
    `, offerID, normalized.Stock); err != nil {
		return ProductManagementRecord{}, fmt.Errorf("record manual product opening stock: %w", err)
	}

	state := productOverrideState{
		SourceTitle:       normalized.Title,
		SourceDescription: normalized.Description,
		SourceBrand:       normalized.Brand,
		VisibilityStatus:  "active",
		Revision:          1,
	}
	if err := persistProductOverrideState(queryCtx, tx, productID, actorID, state); err != nil {
		return ProductManagementRecord{}, err
	}
	afterJSON, err := json.Marshal(struct {
		Title              string `json:"title"`
		Description        string `json:"description"`
		Category           string `json:"category"`
		Brand              string `json:"brand"`
		Price              string `json:"price"`
		DiscountPercentage string `json:"discountPercentage"`
		Rating             string `json:"rating"`
		Stock              int32  `json:"stock"`
		Source             string `json:"source"`
		VisibilityStatus   string `json:"visibilityStatus"`
	}{normalized.Title, normalized.Description, normalized.Category, normalized.Brand, normalized.Price,
		normalized.DiscountPercentage, normalized.Rating, normalized.Stock, manualProductSource, "active"})
	if err != nil {
		return ProductManagementRecord{}, fmt.Errorf("encode manual product initial snapshot: %w", err)
	}
	if err := appendProductRevision(queryCtx, tx, productID, actorID, "created", 1, []byte(`{}`), afterJSON,
		[]string{"title", "description", "category", "brand", "price", "discountPercentage", "rating", "stock", "images"}, "active"); err != nil {
		return ProductManagementRecord{}, err
	}
	record, err := readManagedProductTx(queryCtx, tx, productID, false)
	if err != nil {
		return ProductManagementRecord{}, err
	}
	if err := tx.Commit(queryCtx); err != nil {
		return ProductManagementRecord{}, fmt.Errorf("commit manual product creation: %w", err)
	}
	return record, nil
}

func normalizeProductCreate(input ProductCreateInput) (normalizedProductCreateInput, error) {
	out := normalizedProductCreateInput{
		Title:       strings.TrimSpace(input.Title),
		Description: strings.TrimSpace(input.Description),
		Category:    strings.TrimSpace(input.Category),
		Brand:       strings.TrimSpace(input.Brand),
		Stock:       input.Stock,
		Thumbnail:   strings.TrimSpace(input.Thumbnail),
	}
	if out.Title == "" || utf8RuneCount(out.Title) > 300 || utf8RuneCount(out.Description) > 10000 ||
		out.Category == "" || utf8RuneCount(out.Category) > 120 || utf8RuneCount(out.Brand) > 120 || out.Stock < 0 {
		return normalizedProductCreateInput{}, ErrInvalidProductInput
	}
	var err error
	if out.Price, err = normalizeManualProductDecimal(input.Price, "9999999999.99"); err != nil {
		return normalizedProductCreateInput{}, err
	}
	discount := strings.TrimSpace(input.DiscountPercentage)
	if discount == "" {
		discount = "0"
	}
	if out.DiscountPercentage, err = normalizeManualProductDecimal(discount, "100"); err != nil {
		return normalizedProductCreateInput{}, err
	}
	rating := strings.TrimSpace(input.Rating)
	if rating == "" {
		rating = "0"
	}
	if out.Rating, err = normalizeManualProductDecimal(rating, "5"); err != nil {
		return normalizedProductCreateInput{}, err
	}
	if out.Thumbnail != "" {
		if err := validateManualProductImageURL(out.Thumbnail); err != nil {
			return normalizedProductCreateInput{}, err
		}
	}
	if len(input.Images) > 12 {
		return normalizedProductCreateInput{}, ErrInvalidProductInput
	}
	images := make([]string, 0, len(input.Images))
	for _, image := range input.Images {
		image = strings.TrimSpace(image)
		if err := validateManualProductImageURL(image); err != nil {
			return normalizedProductCreateInput{}, err
		}
		images = append(images, image)
	}
	if out.Thumbnail == "" && len(images) > 0 {
		out.Thumbnail = images[0]
	}
	out.ImagesJSON, err = json.Marshal(images)
	if err != nil {
		return normalizedProductCreateInput{}, fmt.Errorf("encode manual product images: %w", err)
	}
	return out, nil
}

func normalizeManualProductDecimal(raw, maximum string) (string, error) {
	raw = strings.TrimSpace(raw)
	if !manualProductDecimalRegex.MatchString(raw) {
		return "", ErrInvalidProductInput
	}
	value, ok := new(big.Rat).SetString(raw)
	if !ok {
		return "", ErrInvalidProductInput
	}
	maxValue, ok := new(big.Rat).SetString(maximum)
	if !ok || value.Sign() < 0 || value.Cmp(maxValue) > 0 {
		return "", ErrInvalidProductInput
	}
	return raw, nil
}

func validateManualProductImageURL(raw string) error {
	if raw == "" || len(raw) > 2048 {
		return ErrInvalidProductInput
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil || parsed.Opaque != "" {
		return ErrInvalidProductInput
	}
	return nil
}

func utf8RuneCount(value string) int {
	return len([]rune(value))
}
