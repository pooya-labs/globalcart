package commerce

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pooya-labs/globalcart/backend/internal/auth"
	"unicode/utf8"
)

type vendorProductQuery interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func vendorIDForOffer(ctx context.Context, tx pgx.Tx, offerID int64) (int64, error) {
	var vendorID int64
	err := tx.QueryRow(ctx, "SELECT vendor_id FROM vendor_products WHERE id = $1", offerID).Scan(&vendorID)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrVendorProductNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("resolve canonical vendor for offer: %w", err)
	}
	return vendorID, nil
}

func updateVendorPriceTx(ctx context.Context, tx pgx.Tx, vendorID, offerID int64, newPrice string) error {
	var oldPrice, currency, status string
	err := tx.QueryRow(ctx,
		"SELECT vp.price::text, vp.currency, v.status\n"+
			"FROM vendor_products vp JOIN vendors v ON v.id = vp.vendor_id\n"+
			"WHERE vp.id = $1 AND vp.vendor_id = $2\n"+
			"FOR UPDATE OF vp, v",
		offerID, vendorID,
	).Scan(&oldPrice, &currency, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrVendorProductNotFound
	}
	if err != nil {
		return fmt.Errorf("lock offer price: %w", err)
	}
	if status != "active" {
		return ErrVendorInactive
	}
	var updated string
	if err := tx.QueryRow(ctx,
		"UPDATE vendor_products SET price = $3::numeric, updated_at = now() "+
			"WHERE id = $1 AND vendor_id = $2 RETURNING price::text",
		offerID, vendorID, newPrice,
	).Scan(&updated); err != nil {
		return fmt.Errorf("update offer price: %w", err)
	}
	if oldPrice != updated {
		if _, err := tx.Exec(ctx,
			"INSERT INTO vendor_price_history(vendor_product_id,old_price,new_price,currency,change_source) "+
				"VALUES($1,$2::numeric,$3::numeric,$4,'merchant_update')",
			offerID, oldPrice, updated, currency,
		); err != nil {
			return fmt.Errorf("append price history: %w", err)
		}
	}
	return nil
}

func adjustVendorInventoryTx(
	ctx context.Context,
	tx pgx.Tx,
	vendorID, offerID int64,
	delta int32,
	reason, reference string,
) (int32, error) {
	var stock int32
	var status string
	err := tx.QueryRow(ctx,
		"SELECT vp.stock, v.status FROM vendor_products vp "+
			"JOIN vendors v ON v.id = vp.vendor_id "+
			"WHERE vp.id = $1 AND vp.vendor_id = $2 "+
			"FOR UPDATE OF vp, v",
		offerID, vendorID,
	).Scan(&stock, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrVendorProductNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("lock vendor stock: %w", err)
	}
	if status != "active" {
		return 0, ErrVendorInactive
	}
	next := int64(stock) + int64(delta)
	if next < 0 || next > 2147483647 {
		return 0, ErrInsufficientStock
	}
	var result int32
	if err := tx.QueryRow(ctx,
		"UPDATE vendor_products SET stock = $3, updated_at = now() "+
			"WHERE id = $1 AND vendor_id = $2 RETURNING stock",
		offerID, vendorID, next,
	).Scan(&result); err != nil {
		return 0, fmt.Errorf("update vendor stock: %w", err)
	}
	if _, err := tx.Exec(ctx,
		"INSERT INTO vendor_inventory_events(vendor_product_id,quantity_delta,resulting_stock,reason,reference) "+
			"VALUES($1,$2,$3,$4,$5)",
		offerID, delta, result, reason, nullIfEmpty(reference),
	); err != nil {
		return 0, fmt.Errorf("append inventory event: %w", err)
	}
	return result, nil
}

func getVendorProduct(ctx context.Context, q vendorProductQuery, vendorID, offerID int64, lockForShare bool) (VendorProduct, error) {
	var out VendorProduct
	var variant pgtype.Int8
	var sellerSKU pgtype.Text
	query := "SELECT id,vendor_id,product_id,variant_id,seller_sku,price::text,currency,stock,status " +
		"FROM vendor_products WHERE vendor_id=$1 AND id=$2"
	if lockForShare {
		query += " FOR SHARE"
	}
	err := q.QueryRow(ctx, query, vendorID, offerID).Scan(
		&out.ID, &out.VendorID, &out.ProductID, &variant, &sellerSKU,
		&out.Price, &out.Currency, &out.Stock, &out.Status,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return VendorProduct{}, ErrVendorProductNotFound
	}
	if err != nil {
		return VendorProduct{}, fmt.Errorf("read vendor offer: %w", err)
	}
	if variant.Valid {
		id := variant.Int64
		out.VariantID = &id
	}
	if sellerSKU.Valid {
		out.SellerSKU = sellerSKU.String
	}
	return out, nil
}

// UpdateVendorPriceForSeller derives the vendor from the canonical offer row,
// revalidates the current session and membership under row locks, then updates
// the offer and price history in the same transaction.
func (s *Service) UpdateVendorPriceForSeller(
	ctx context.Context,
	principal auth.SessionPrincipal,
	offerID int64,
	newPrice string,
) error {
	if offerID <= 0 {
		return ErrInvalidVendorInput
	}
	normalized, err := normalizeSellerOfferPrice(newPrice)
	if err != nil {
		return ErrInvalidVendorInput
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("begin authorized price update: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	vendorID, err := vendorIDForOffer(ctx, tx, offerID)
	if err != nil {
		return err
	}
	if err := auth.AuthorizeSellerVendorTx(ctx, tx, principal, auth.AuthorizationSellerWriteVendor, vendorID); err != nil {
		return err
	}
	before, err := readSellerOfferTx(ctx, tx, vendorID, offerID, true)
	if err != nil {
		return err
	}
	if before.Status == "archived" {
		return ErrSellerOfferArchived
	}
	if before.Price == normalized {
		return tx.Commit(ctx)
	}
	if err := updateVendorPriceTx(ctx, tx, vendorID, offerID, normalized); err != nil {
		return err
	}
	after, err := readSellerOfferTx(ctx, tx, vendorID, offerID, true)
	if err != nil {
		return err
	}
	if err := appendSellerOfferHistoryTx(ctx, tx, principal.UserID, "updated", &before, after, []string{"price"}); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit authorized price update: %w", err)
	}
	return nil
}

// AdjustVendorInventoryForSeller derives the vendor scope from the offer,
// revalidates it under the same transaction as the stock mutation and audit.
func (s *Service) AdjustVendorInventoryForSeller(
	ctx context.Context,
	principal auth.SessionPrincipal,
	offerID int64,
	delta int32,
	reason, reference string,
) (int32, error) {
	if offerID <= 0 || delta == 0 ||
		(reason != "stock_adjustment" && reason != "restock" && reason != "return" && reason != "correction") {
		return 0, ErrInvalidVendorInput
	}
	reference = strings.TrimSpace(reference)
	if !utf8.ValidString(reference) || strings.ContainsRune(reference, 0) || utf8.RuneCountInString(reference) > 160 {
		return 0, ErrInvalidVendorInput
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return 0, fmt.Errorf("begin authorized stock adjustment: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	vendorID, err := vendorIDForOffer(ctx, tx, offerID)
	if err != nil {
		return 0, err
	}
	if err := auth.AuthorizeSellerVendorTx(ctx, tx, principal, auth.AuthorizationSellerWriteVendor, vendorID); err != nil {
		return 0, err
	}
	before, err := readSellerOfferTx(ctx, tx, vendorID, offerID, true)
	if err != nil {
		return 0, err
	}
	if before.Status == "archived" {
		return 0, ErrSellerOfferArchived
	}
	stock, err := adjustVendorInventoryTx(ctx, tx, vendorID, offerID, delta, reason, reference)
	if err != nil {
		return 0, err
	}
	after, err := readSellerOfferTx(ctx, tx, vendorID, offerID, true)
	if err != nil {
		return 0, err
	}
	if err := appendSellerOfferHistoryTx(ctx, tx, principal.UserID, "updated", &before, after, []string{"stock"}); err != nil {
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit authorized stock adjustment: %w", err)
	}
	return stock, nil
}

// GetVendorProductForSeller never accepts a client-selected vendor ID. It
// resolves vendor ownership from the offer, revalidates session/membership,
// and reads the offer in that same transaction.
func (s *Service) GetVendorProductForSeller(
	ctx context.Context,
	principal auth.SessionPrincipal,
	offerID int64,
) (VendorProduct, error) {
	if offerID <= 0 {
		return VendorProduct{}, ErrInvalidVendorInput
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return VendorProduct{}, fmt.Errorf("begin authorized vendor offer read: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	vendorID, err := vendorIDForOffer(ctx, tx, offerID)
	if err != nil {
		return VendorProduct{}, err
	}
	if err := auth.AuthorizeSellerVendorTx(
		ctx, tx, principal, auth.AuthorizationSellerReadVendor, vendorID,
	); err != nil {
		return VendorProduct{}, err
	}
	offer, err := getVendorProduct(ctx, tx, vendorID, offerID, true)
	if err != nil {
		return VendorProduct{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return VendorProduct{}, fmt.Errorf("commit authorized vendor offer read: %w", err)
	}
	return offer, nil
}
