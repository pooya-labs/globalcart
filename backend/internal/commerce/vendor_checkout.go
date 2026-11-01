package commerce

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// CheckoutVendorCart atomically creates one parent demo order and one child
// order per vendor. It reserves only vendor_products stock; it never touches
// legacy products.stock. No payment provider or real money is involved.
func (s *Service) CheckoutVendorCart(ctx context.Context, ownerTokenHash, idempotencyHash []byte, reservationTTL time.Duration) (Order, error) {
	if len(ownerTokenHash) != 32 || len(idempotencyHash) != 32 {
		return Order{}, errors.New("capability and idempotency digests must be 32 bytes")
	}
	if reservationTTL <= 0 {
		return Order{}, ErrInvalidLifetime
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return Order{}, fmt.Errorf("begin vendor checkout: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	var cartID, currency, status string
	var expired bool
	var couponID pgtype.Int8
	err = tx.QueryRow(ctx, `SELECT id::text,currency,status,expires_at<=now(),coupon_id FROM carts WHERE owner_token_hash=$1 FOR UPDATE`, ownerTokenHash).Scan(&cartID, &currency, &status, &expired, &couponID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Order{}, ErrCartNotFound
	}
	if err != nil {
		return Order{}, fmt.Errorf("lock cart: %w", err)
	}
	var replay Order
	err = tx.QueryRow(ctx, `SELECT id::text,status,currency,subtotal::text,discount_amount::text,total::text FROM demo_orders WHERE cart_id=$1 AND idempotency_key_hash=$2`, cartID, idempotencyHash).Scan(&replay.ID, &replay.Status, &replay.Currency, &replay.Subtotal, &replay.DiscountAmount, &replay.Total)
	if err == nil {
		replay.Replayed = true
		if err = tx.Commit(ctx); err != nil {
			return Order{}, fmt.Errorf("commit vendor checkout replay: %w", err)
		}
		return replay, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Order{}, fmt.Errorf("lookup vendor checkout replay: %w", err)
	}
	if status != "open" {
		return Order{}, ErrCartClosed
	}
	if expired {
		return Order{}, ErrCartExpired
	}
	var legacyLines int64
	if err = tx.QueryRow(ctx, "SELECT count(*) FROM cart_items WHERE cart_id=$1 AND vendor_product_id IS NULL", cartID).Scan(&legacyLines); err != nil {
		return Order{}, fmt.Errorf("check legacy cart lines: %w", err)
	}
	if legacyLines > 0 {
		return Order{}, ErrMixedCartNotSupported
	}
	type line struct {
		offerID, vendorID, productID              int64
		variant                                   pgtype.Int8
		qty                                       int32
		price, title, variantTitle, offerCurrency string
		stock                                     int32
	}
	rows, err := tx.Query(ctx, `SELECT vp.id,vp.vendor_id,vp.product_id,vp.variant_id,ci.quantity,
 vp.price::text,vp.currency,vp.stock,p.title,COALESCE(pv.title,'')
 FROM cart_items ci JOIN vendor_products vp ON vp.id=ci.vendor_product_id
 JOIN vendors v ON v.id=vp.vendor_id JOIN products p ON p.id=vp.product_id
 LEFT JOIN product_variants pv ON pv.id=vp.variant_id
 WHERE ci.cart_id=$1 ORDER BY vp.vendor_id,vp.id FOR UPDATE OF vp,v`, cartID)
	if err != nil {
		return Order{}, fmt.Errorf("lock vendor cart offers: %w", err)
	}
	lines := make([]line, 0)
	for rows.Next() {
		var x line
		if err = rows.Scan(&x.offerID, &x.vendorID, &x.productID, &x.variant, &x.qty, &x.price, &x.offerCurrency, &x.stock, &x.title, &x.variantTitle); err != nil {
			rows.Close()
			return Order{}, fmt.Errorf("scan vendor offer: %w", err)
		}
		lines = append(lines, x)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return Order{}, fmt.Errorf("iterate vendor offers: %w", err)
	}
	rows.Close()
	if len(lines) == 0 {
		return Order{}, ErrEmptyCart
	}
	for _, x := range lines {
		if x.offerCurrency != currency {
			return Order{}, ErrUnsupportedCurrency
		}
		if x.qty < 1 || x.qty > 99 {
			return Order{}, ErrInvalidQuantity
		}
		if x.variant.Valid {
			var active bool
			if err = tx.QueryRow(ctx, "SELECT is_active FROM product_variants WHERE id=$1 AND product_id=$2", x.variant.Int64, x.productID).Scan(&active); err != nil {
				return Order{}, fmt.Errorf("validate vendor variant: %w", err)
			}
			if !active {
				return Order{}, ErrUnavailable
			}
		}
		var reserved int64
		if err = tx.QueryRow(ctx, `SELECT COALESCE(sum(quantity),0)::bigint FROM inventory_reservations WHERE vendor_product_id=$1 AND status='active' AND expires_at>now()`, x.offerID).Scan(&reserved); err != nil {
			return Order{}, fmt.Errorf("sum vendor reservations: %w", err)
		}
		if int64(x.stock)-reserved < int64(x.qty) {
			return Order{}, ErrInsufficientStock
		}
	}
	var orderID string
	if err = tx.QueryRow(ctx, `INSERT INTO demo_orders(cart_id,idempotency_key_hash,currency,subtotal,total) VALUES($1,$2,$3,0,0) RETURNING id::text`, cartID, idempotencyHash, currency).Scan(&orderID); err != nil {
		return Order{}, fmt.Errorf("create parent demo order: %w", err)
	}
	vendorOrders := map[int64]string{}
	for _, x := range lines {
		childID, ok := vendorOrders[x.vendorID]
		if !ok {
			if err = tx.QueryRow(ctx, `INSERT INTO demo_vendor_orders(order_id,vendor_id,currency) VALUES($1,$2,$3) RETURNING id::text`, orderID, x.vendorID, currency).Scan(&childID); err != nil {
				return Order{}, fmt.Errorf("create vendor child order: %w", err)
			}
			vendorOrders[x.vendorID] = childID
		}
		var variantArg any
		if x.variant.Valid {
			variantArg = x.variant.Int64
		}
		if _, err = tx.Exec(ctx, `INSERT INTO demo_order_items(order_id,product_id,variant_id,quantity,unit_price_snapshot,currency,product_title_snapshot,variant_title_snapshot,vendor_product_id,vendor_order_id)
  VALUES($1,$2,$3,$4,$5::numeric,$6,$7,$8,$9,$10)`, orderID, x.productID, variantArg, x.qty, x.price, currency, x.title, x.variantTitle, x.offerID, childID); err != nil {
			return Order{}, fmt.Errorf("snapshot vendor order line: %w", err)
		}
		if _, err = tx.Exec(ctx, `INSERT INTO inventory_reservations(order_id,product_id,variant_id,quantity,expires_at,vendor_product_id,vendor_order_id)
  VALUES($1,$2,$3,$4,now()+make_interval(secs => $5::double precision),$6,$7)`, orderID, x.productID, variantArg, x.qty, reservationTTL.Seconds(), x.offerID, childID); err != nil {
			return Order{}, fmt.Errorf("reserve vendor inventory: %w", err)
		}
	}
	var subtotal string
	if err = tx.QueryRow(ctx, `SELECT COALESCE(sum(unit_price_snapshot*quantity),0)::numeric(12,2)::text FROM demo_order_items WHERE order_id=$1`, orderID).Scan(&subtotal); err != nil {
		return Order{}, fmt.Errorf("calculate parent subtotal: %w", err)
	}
	discount := "0.00"
	if couponID.Valid {
		var typ, value, minimum string
		var maximum pgtype.Text
		var starts, ends, now time.Time
		var limit pgtype.Int8
		var per int32
		var active bool
		err = tx.QueryRow(ctx, `SELECT discount_type,discount_value::text,minimum_subtotal::text,maximum_discount::text,starts_at,ends_at,usage_limit,per_customer_limit,is_active,now() FROM coupons WHERE id=$1 FOR UPDATE`, couponID.Int64).Scan(&typ, &value, &minimum, &maximum, &starts, &ends, &limit, &per, &active, &now)
		if errors.Is(err, pgx.ErrNoRows) {
			return Order{}, ErrCouponUnavailable
		}
		if err != nil {
			return Order{}, fmt.Errorf("lock vendor checkout coupon: %w", err)
		}
		if !active || now.Before(starts) || !now.Before(ends) {
			return Order{}, ErrCouponUnavailable
		}
		var meets bool
		if err = tx.QueryRow(ctx, "SELECT $1::numeric >= $2::numeric", subtotal, minimum).Scan(&meets); err != nil {
			return Order{}, fmt.Errorf("validate coupon minimum: %w", err)
		}
		if !meets {
			return Order{}, ErrCouponMinimum
		}
		var globalUses, ownerUses int64
		if err = tx.QueryRow(ctx, "SELECT count(*) FROM coupon_redemptions WHERE coupon_id=$1", couponID.Int64).Scan(&globalUses); err != nil {
			return Order{}, fmt.Errorf("count coupon uses: %w", err)
		}
		if err = tx.QueryRow(ctx, "SELECT count(*) FROM coupon_redemptions WHERE coupon_id=$1 AND owner_token_hash=$2", couponID.Int64, ownerTokenHash).Scan(&ownerUses); err != nil {
			return Order{}, fmt.Errorf("count owner coupon uses: %w", err)
		}
		if (limit.Valid && globalUses >= limit.Int64) || ownerUses >= int64(per) {
			return Order{}, ErrCouponUnavailable
		}
		var cap any
		if maximum.Valid {
			cap = maximum.String
		}
		if err = tx.QueryRow(ctx, `SELECT LEAST($1::numeric,CASE WHEN $2='percentage' THEN round($1::numeric*$3::numeric/100,2) ELSE $3::numeric END,COALESCE($4::numeric,$1::numeric))::numeric(12,2)::text`, subtotal, typ, value, cap).Scan(&discount); err != nil {
			return Order{}, fmt.Errorf("calculate coupon discount: %w", err)
		}
		if _, err = tx.Exec(ctx, `INSERT INTO coupon_redemptions(coupon_id,order_id,owner_token_hash,discount_amount) VALUES($1,$2,$3,$4::numeric)`, couponID.Int64, orderID, ownerTokenHash, discount); err != nil {
			return Order{}, fmt.Errorf("record coupon redemption: %w", err)
		}
	}
	var total string
	if err = tx.QueryRow(ctx, "SELECT ($1::numeric-$2::numeric)::numeric(12,2)::text", subtotal, discount).Scan(&total); err != nil {
		return Order{}, fmt.Errorf("calculate parent total: %w", err)
	}
	if _, err = tx.Exec(ctx, `UPDATE demo_orders SET coupon_id=$2,subtotal=$3::numeric,discount_amount=$4::numeric,total=$5::numeric,updated_at=now() WHERE id=$1`, orderID, nullableCouponID(couponID), subtotal, discount, total); err != nil {
		return Order{}, fmt.Errorf("finalize parent order totals: %w", err)
	}
	if _, err = tx.Exec(ctx, `UPDATE demo_vendor_orders d SET subtotal=x.subtotal,total=x.subtotal-d.discount_amount,updated_at=now()
 FROM (SELECT vendor_order_id,sum(unit_price_snapshot*quantity)::numeric(12,2) AS subtotal FROM demo_order_items WHERE order_id=$1 GROUP BY vendor_order_id) x
 WHERE d.id=x.vendor_order_id`, orderID); err != nil {
		return Order{}, fmt.Errorf("finalize vendor subtotals: %w", err)
	}
	// Reapply allocation after exact child subtotals have been materialized.
	_, err = tx.Exec(ctx, `WITH amounts AS (
  SELECT id,subtotal,row_number() OVER(ORDER BY vendor_id,id) rn,count(*) OVER() n,sum(subtotal) OVER() all_subtotal
  FROM demo_vendor_orders WHERE order_id=$1
 ), rounded AS (SELECT id,rn,n,round($2::numeric*subtotal/NULLIF(all_subtotal,0),2) share FROM amounts),
 allocated AS (SELECT id,CASE WHEN rn=n THEN $2::numeric-COALESCE(sum(share) OVER(ORDER BY rn ROWS BETWEEN UNBOUNDED PRECEDING AND 1 PRECEDING),0) ELSE share END discount FROM rounded)
 UPDATE demo_vendor_orders d SET discount_amount=a.discount,total=d.subtotal-a.discount,updated_at=now() FROM allocated a WHERE d.id=a.id`, orderID, discount)
	if err != nil {
		return Order{}, fmt.Errorf("finalize vendor discounts: %w", err)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO demo_order_events(order_id,event_type,payload) VALUES($1,'demo_order_created','{"mode":"simulation","split_by_vendor":true}'::jsonb)`, orderID); err != nil {
		return Order{}, fmt.Errorf("append parent order event: %w", err)
	}
	if _, err = tx.Exec(ctx, `UPDATE carts SET status='checked_out',updated_at=now() WHERE id=$1`, cartID); err != nil {
		return Order{}, fmt.Errorf("close vendor cart: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return Order{}, fmt.Errorf("commit vendor checkout: %w", err)
	}
	return Order{ID: orderID, Status: "awaiting_demo_payment", Currency: currency, Subtotal: subtotal, DiscountAmount: discount, Total: total}, nil
}
