package commerce

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

var (
	ErrVendorNotFound        = errors.New("vendor not found")
	ErrVendorInactive        = errors.New("vendor is not active")
	ErrVendorProductNotFound = errors.New("vendor product not found")
	ErrInvalidVendorInput    = errors.New("invalid vendor input")
)

type Vendor struct {
	ID          int64
	Slug        string
	DisplayName string
	Type        string
	Status      string
}

type VendorProduct struct {
	ID        int64
	VendorID  int64
	ProductID int64
	VariantID *int64
	SellerSKU string
	Price     string
	Currency  string
	Stock     int32
	Status    string
}

func (s *Service) CreateVendor(ctx context.Context, slug, displayName string) (Vendor, error) {
	slug = strings.TrimSpace(strings.ToLower(slug))
	displayName = strings.TrimSpace(displayName)
	if slug == "" || displayName == "" || len(slug) > 120 || len(displayName) > 200 {
		return Vendor{}, ErrInvalidVendorInput
	}
	var v Vendor
	err := s.pool.QueryRow(ctx, `INSERT INTO vendors(slug,display_name,vendor_type,status)
 VALUES($1,$2,'merchant','pending') RETURNING id,slug,display_name,vendor_type,status`, slug, displayName).
		Scan(&v.ID, &v.Slug, &v.DisplayName, &v.Type, &v.Status)
	if err != nil {
		return Vendor{}, fmt.Errorf("create vendor: %w", err)
	}
	return v, nil
}

// UpsertVendorProduct records initial/current price and stock in the same transaction.
func (s *Service) UpsertVendorProduct(ctx context.Context, in VendorProduct) (VendorProduct, error) {
	if in.VendorID <= 0 || in.ProductID <= 0 || strings.TrimSpace(in.Price) == "" || in.Stock < 0 || (in.VariantID != nil && *in.VariantID <= 0) {
		return VendorProduct{}, ErrInvalidVendorInput
	}
	in.Currency = strings.ToUpper(strings.TrimSpace(in.Currency))
	if in.Currency == "" {
		in.Currency = "USD"
	}
	if in.Currency != "USD" {
		return VendorProduct{}, ErrUnsupportedCurrency
	}
	in.SellerSKU = strings.TrimSpace(in.SellerSKU)
	if in.Status == "" {
		in.Status = "draft"
	}
	if in.Status != "draft" && in.Status != "active" && in.Status != "paused" && in.Status != "archived" {
		return VendorProduct{}, ErrInvalidVendorInput
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return VendorProduct{}, fmt.Errorf("begin vendor offer upsert: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	var vendorStatus string
	err = tx.QueryRow(ctx, "SELECT status FROM vendors WHERE id=$1 FOR UPDATE", in.VendorID).Scan(&vendorStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return VendorProduct{}, ErrVendorNotFound
	}
	if err != nil {
		return VendorProduct{}, fmt.Errorf("lock vendor: %w", err)
	}
	if vendorStatus != "active" {
		return VendorProduct{}, ErrVendorInactive
	}
	var variantArg any
	if in.VariantID == nil {
		var ok bool
		if err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM products WHERE id=$1)", in.ProductID).Scan(&ok); err != nil {
			return VendorProduct{}, fmt.Errorf("validate product: %w", err)
		}
		if !ok {
			return VendorProduct{}, ErrInvalidProduct
		}
	} else {
		variantArg = *in.VariantID
		var ok bool
		if err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM product_variants WHERE id=$1 AND product_id=$2)", *in.VariantID, in.ProductID).Scan(&ok); err != nil {
			return VendorProduct{}, fmt.Errorf("validate variant: %w", err)
		}
		if !ok {
			return VendorProduct{}, ErrInvalidProduct
		}
	}
	var id int64
	var oldPrice, oldCurrency string
	var oldStock int32
	var oldStatus string
	if in.VariantID == nil {
		err = tx.QueryRow(ctx, `SELECT id,price::text,stock,currency,status FROM vendor_products WHERE vendor_id=$1 AND product_id=$2 AND variant_id IS NULL FOR UPDATE`, in.VendorID, in.ProductID).Scan(&id, &oldPrice, &oldStock, &oldCurrency, &oldStatus)
	} else {
		err = tx.QueryRow(ctx, `SELECT id,price::text,stock,currency,status FROM vendor_products WHERE vendor_id=$1 AND variant_id=$2 FOR UPDATE`, in.VendorID, *in.VariantID).Scan(&id, &oldPrice, &oldStock, &oldCurrency, &oldStatus)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		err = tx.QueryRow(ctx, `INSERT INTO vendor_products(vendor_id,product_id,variant_id,seller_sku,price,currency,stock,status)
  VALUES($1,$2,$3,$4,$5::numeric,$6,$7,$8) RETURNING id,price::text,stock,currency,status`,
			in.VendorID, in.ProductID, variantArg, nullIfEmpty(in.SellerSKU), in.Price, in.Currency, in.Stock, in.Status).Scan(&id, &oldPrice, &oldStock, &oldCurrency, &oldStatus)
		if err != nil {
			return VendorProduct{}, fmt.Errorf("insert vendor offer: %w", err)
		}
		if _, err = tx.Exec(ctx, `INSERT INTO vendor_price_history(vendor_product_id,old_price,new_price,currency,change_source) VALUES($1,NULL,$2::numeric,$3,'merchant_update')`, id, oldPrice, oldCurrency); err != nil {
			return VendorProduct{}, fmt.Errorf("record initial price: %w", err)
		}
		if _, err = tx.Exec(ctx, `INSERT INTO vendor_inventory_events(vendor_product_id,quantity_delta,resulting_stock,reason,reference) VALUES($1,$2,$2,'opening_balance','vendor_offer_created')`, id, oldStock); err != nil {
			return VendorProduct{}, fmt.Errorf("record opening stock: %w", err)
		}
	} else if err != nil {
		return VendorProduct{}, fmt.Errorf("read vendor offer: %w", err)
	} else {
		var updatedPrice string
		err = tx.QueryRow(ctx, `UPDATE vendor_products SET seller_sku=$2,price=$3::numeric,currency=$4,stock=$5,status=$6,updated_at=now() WHERE id=$1 RETURNING price::text`, id, nullIfEmpty(in.SellerSKU), in.Price, in.Currency, in.Stock, in.Status).Scan(&updatedPrice)
		if err != nil {
			return VendorProduct{}, fmt.Errorf("update vendor offer: %w", err)
		}
		if oldPrice != updatedPrice || oldCurrency != in.Currency {
			if _, err = tx.Exec(ctx, `INSERT INTO vendor_price_history(vendor_product_id,old_price,new_price,currency,change_source) VALUES($1,$2::numeric,$3::numeric,$4,'merchant_update')`, id, oldPrice, updatedPrice, in.Currency); err != nil {
				return VendorProduct{}, fmt.Errorf("record price history: %w", err)
			}
		}
		if oldStock != in.Stock {
			if _, err = tx.Exec(ctx, `INSERT INTO vendor_inventory_events(vendor_product_id,quantity_delta,resulting_stock,reason,reference) VALUES($1,$2,$3,'stock_adjustment','vendor_offer_upsert')`, id, int64(in.Stock)-int64(oldStock), in.Stock); err != nil {
				return VendorProduct{}, fmt.Errorf("record stock history: %w", err)
			}
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return VendorProduct{}, fmt.Errorf("commit vendor offer: %w", err)
	}
	in.ID = id
	return in, nil
}

func (s *Service) UpdateVendorPrice(ctx context.Context, vendorID, offerID int64, newPrice string) error {
	if vendorID <= 0 || offerID <= 0 || strings.TrimSpace(newPrice) == "" {
		return ErrInvalidVendorInput
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("begin price update: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if err := updateVendorPriceTx(ctx, tx, vendorID, offerID, newPrice); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit price update: %w", err)
	}
	return nil
}

func (s *Service) AdjustVendorInventory(ctx context.Context, vendorID, offerID int64, delta int32, reason, reference string) (int32, error) {
	if vendorID <= 0 || offerID <= 0 || delta == 0 ||
		(reason != "stock_adjustment" && reason != "restock" && reason != "return" && reason != "correction") {
		return 0, ErrInvalidVendorInput
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return 0, fmt.Errorf("begin stock adjustment: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	result, err := adjustVendorInventoryTx(ctx, tx, vendorID, offerID, delta, reason, reference)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit stock adjustment: %w", err)
	}
	return result, nil
}

func (s *Service) GetVendorProduct(ctx context.Context, vendorID, offerID int64) (VendorProduct, error) {
	if vendorID <= 0 || offerID <= 0 {
		return VendorProduct{}, ErrInvalidVendorInput
	}
	return getVendorProduct(ctx, s.pool, vendorID, offerID, false)
}

func nullIfEmpty(v string) any {
	if v == "" {
		return nil
	}
	return v
}
