package commerce

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// AddVendorItem adds one vendor-specific offer to an open cart. The offer row
// is locked and its price, currency, product and variant are read canonically.
func (s *Service) AddVendorItem(ctx context.Context, ownerTokenHash []byte, vendorProductID int64, quantity int32) (Cart, error) {
	if len(ownerTokenHash) != 32 {
		return Cart{}, errors.New("cart capability digest must be 32 bytes")
	}
	if vendorProductID <= 0 {
		return Cart{}, ErrInvalidVendorInput
	}
	if quantity < 1 || quantity > 99 {
		return Cart{}, ErrInvalidQuantity
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return Cart{}, fmt.Errorf("begin vendor cart add: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	var cartID, currency, status string
	var expired bool
	err = tx.QueryRow(ctx, `SELECT id::text,currency,status,expires_at<=now() FROM carts WHERE owner_token_hash=$1 FOR UPDATE`, ownerTokenHash).Scan(&cartID, &currency, &status, &expired)
	if errors.Is(err, pgx.ErrNoRows) {
		return Cart{}, ErrCartNotFound
	}
	if err != nil {
		return Cart{}, fmt.Errorf("lock cart: %w", err)
	}
	if status != "open" {
		return Cart{}, ErrCartClosed
	}
	if expired {
		return Cart{}, ErrCartExpired
	}
	var productID int64
	var variant pgtype.Int8
	var price, offerCurrency, title, variantTitle, vendorName, vendorStatus, offerStatus string
	var stock int32
	err = tx.QueryRow(ctx, `SELECT vp.product_id,vp.variant_id,vp.price::text,vp.currency,vp.stock,vp.status,
 p.title,COALESCE(pv.title,''),v.display_name,v.status
 FROM vendor_products vp JOIN vendors v ON v.id=vp.vendor_id
 JOIN products p ON p.id=vp.product_id LEFT JOIN product_variants pv ON pv.id=vp.variant_id
 WHERE vp.id=$1 FOR UPDATE OF vp,v`, vendorProductID).
		Scan(&productID, &variant, &price, &offerCurrency, &stock, &offerStatus, &title, &variantTitle, &vendorName, &vendorStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return Cart{}, ErrVendorProductNotFound
	}
	if err != nil {
		return Cart{}, fmt.Errorf("lock vendor offer: %w", err)
	}
	if vendorStatus != "active" {
		return Cart{}, ErrVendorInactive
	}
	if offerStatus != "active" {
		return Cart{}, ErrUnavailable
	}
	if offerCurrency != currency {
		return Cart{}, ErrUnsupportedCurrency
	}
	if variant.Valid {
		var active bool
		if err = tx.QueryRow(ctx, "SELECT is_active FROM product_variants WHERE id=$1 AND product_id=$2", variant.Int64, productID).Scan(&active); err != nil {
			return Cart{}, fmt.Errorf("validate offered variant: %w", err)
		}
		if !active {
			return Cart{}, ErrUnavailable
		}
	}
	var existing int32
	if err = tx.QueryRow(ctx, "SELECT COALESCE(max(quantity),0) FROM cart_items WHERE cart_id=$1 AND vendor_product_id=$2", cartID, vendorProductID).Scan(&existing); err != nil {
		return Cart{}, fmt.Errorf("read vendor cart quantity: %w", err)
	}
	if existing+quantity > 99 {
		return Cart{}, ErrInvalidQuantity
	}
	if existing+quantity > stock {
		return Cart{}, ErrInsufficientStock
	}
	var variantArg any
	if variant.Valid {
		variantArg = variant.Int64
	}
	_, err = tx.Exec(ctx, `INSERT INTO cart_items(cart_id,product_id,variant_id,vendor_product_id,quantity,unit_price_snapshot,currency,product_title_snapshot,variant_title_snapshot)
 VALUES($1,$2,$3,$4,$5,$6::numeric,$7,$8,$9)
 ON CONFLICT (cart_id,vendor_product_id) WHERE vendor_product_id IS NOT NULL
 DO UPDATE SET quantity=cart_items.quantity+EXCLUDED.quantity,unit_price_snapshot=EXCLUDED.unit_price_snapshot,
 currency=EXCLUDED.currency,product_title_snapshot=EXCLUDED.product_title_snapshot,
 variant_title_snapshot=EXCLUDED.variant_title_snapshot,updated_at=now()`,
		cartID, productID, nullableInt8(variant), vendorProductID, quantity, price, offerCurrency, title, variantTitle)
	if err != nil {
		return Cart{}, fmt.Errorf("upsert vendor cart item: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return Cart{}, fmt.Errorf("commit vendor cart add: %w", err)
	}
	_ = vendorName
	return s.GetCart(ctx, ownerTokenHash)
}

// SetVendorItemQuantity changes the quantity of a vendor-specific cart line.
func (s *Service) SetVendorItemQuantity(ctx context.Context, ownerTokenHash []byte, vendorProductID int64, quantity int32) (Cart, error) {
	if len(ownerTokenHash) != 32 {
		return Cart{}, errors.New("cart capability digest must be 32 bytes")
	}
	if vendorProductID <= 0 {
		return Cart{}, ErrInvalidVendorInput
	}
	if quantity < 1 || quantity > 99 {
		return Cart{}, ErrInvalidQuantity
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return Cart{}, fmt.Errorf("begin vendor quantity update: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	var cartID, status string
	var expired bool
	err = tx.QueryRow(ctx, "SELECT id::text,status,expires_at<=now() FROM carts WHERE owner_token_hash=$1 FOR UPDATE", ownerTokenHash).Scan(&cartID, &status, &expired)
	if errors.Is(err, pgx.ErrNoRows) {
		return Cart{}, ErrCartNotFound
	}
	if err != nil {
		return Cart{}, fmt.Errorf("lock cart: %w", err)
	}
	if status != "open" {
		return Cart{}, ErrCartClosed
	}
	if expired {
		return Cart{}, ErrCartExpired
	}
	var stock int32
	var offerStatus, vendorStatus string
	err = tx.QueryRow(ctx, `SELECT vp.stock,vp.status,v.status FROM vendor_products vp JOIN vendors v ON v.id=vp.vendor_id WHERE vp.id=$1 FOR UPDATE OF vp,v`, vendorProductID).Scan(&stock, &offerStatus, &vendorStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return Cart{}, ErrVendorProductNotFound
	}
	if err != nil {
		return Cart{}, fmt.Errorf("lock vendor offer: %w", err)
	}
	if vendorStatus != "active" {
		return Cart{}, ErrVendorInactive
	}
	if offerStatus != "active" {
		return Cart{}, ErrUnavailable
	}
	if quantity > stock {
		return Cart{}, ErrInsufficientStock
	}
	tag, err := tx.Exec(ctx, "UPDATE cart_items SET quantity=$3,updated_at=now() WHERE cart_id=$1 AND vendor_product_id=$2", cartID, vendorProductID, quantity)
	if err != nil {
		return Cart{}, fmt.Errorf("set vendor cart quantity: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return Cart{}, ErrVendorProductNotFound
	}
	if err = tx.Commit(ctx); err != nil {
		return Cart{}, fmt.Errorf("commit vendor quantity update: %w", err)
	}
	return s.GetCart(ctx, ownerTokenHash)
}

// RemoveVendorItem removes only the specified vendor offer from an open cart.
func (s *Service) RemoveVendorItem(ctx context.Context, ownerTokenHash []byte, vendorProductID int64) (Cart, error) {
	if len(ownerTokenHash) != 32 {
		return Cart{}, errors.New("cart capability digest must be 32 bytes")
	}
	if vendorProductID <= 0 {
		return Cart{}, ErrInvalidVendorInput
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return Cart{}, fmt.Errorf("begin vendor cart removal: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	var cartID, status string
	var expired bool
	err = tx.QueryRow(ctx, "SELECT id::text,status,expires_at<=now() FROM carts WHERE owner_token_hash=$1 FOR UPDATE", ownerTokenHash).Scan(&cartID, &status, &expired)
	if errors.Is(err, pgx.ErrNoRows) {
		return Cart{}, ErrCartNotFound
	}
	if err != nil {
		return Cart{}, fmt.Errorf("lock cart: %w", err)
	}
	if status != "open" {
		return Cart{}, ErrCartClosed
	}
	if expired {
		return Cart{}, ErrCartExpired
	}
	tag, err := tx.Exec(ctx, "DELETE FROM cart_items WHERE cart_id=$1 AND vendor_product_id=$2", cartID, vendorProductID)
	if err != nil {
		return Cart{}, fmt.Errorf("remove vendor cart item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return Cart{}, ErrVendorProductNotFound
	}
	if err = tx.Commit(ctx); err != nil {
		return Cart{}, fmt.Errorf("commit vendor cart removal: %w", err)
	}
	return s.GetCart(ctx, ownerTokenHash)
}

func nullableInt8(v pgtype.Int8) any {
	if !v.Valid {
		return nil
	}
	return v.Int64
}
