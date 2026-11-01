package commerce

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrCartNotFound      = errors.New("cart not found")
	ErrCartClosed        = errors.New("cart is not open")
	ErrCartExpired       = errors.New("cart expired")
	ErrEmptyCart         = errors.New("cart is empty")
	ErrUnavailable       = errors.New("item is unavailable")
	ErrInsufficientStock = errors.New("insufficient demo stock")
	ErrInvalidLifetime   = errors.New("reservation lifetime must be positive")
	ErrVendorCheckoutNotReady = errors.New("vendor cart checkout is not available yet")
)

type Order struct {
	ID             string
	Status         string
	Currency       string
	Subtotal       string
	DiscountAmount string
	Total          string
	Replayed       bool
}

type Service struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Service { return &Service{pool: pool} }

// Checkout atomically snapshots current catalog prices, reserves demo stock and
// creates an order. Callers must pass SHA-256 digests, never raw bearer tokens
// or idempotency keys. This service is deliberately not wired to an HTTP route.
func (s *Service) Checkout(ctx context.Context, ownerTokenHash, idempotencyHash []byte, reservationTTL time.Duration) (Order, error) {
	if len(ownerTokenHash) != 32 || len(idempotencyHash) != 32 {
		return Order{}, errors.New("capability and idempotency digests must be 32 bytes")
	}
	if reservationTTL <= 0 {
		return Order{}, ErrInvalidLifetime
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return Order{}, fmt.Errorf("begin checkout: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	var cartID, currency, cartStatus string
	var cartExpired bool
	var couponID pgtype.Int8
	err = tx.QueryRow(ctx, `SELECT id::text, currency, status, expires_at <= now(), coupon_id
		FROM carts WHERE owner_token_hash=$1 FOR UPDATE`, ownerTokenHash).
		Scan(&cartID, &currency, &cartStatus, &cartExpired, &couponID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Order{}, ErrCartNotFound
	}
	if err != nil {
		return Order{}, fmt.Errorf("lock cart: %w", err)
	}

	// Replay lookup precedes lifecycle checks so a retry after a successful
	// commit can recover the original order even though the cart is checked out.
	var replay Order
	err = tx.QueryRow(ctx, `SELECT id::text,status,currency,subtotal::text,discount_amount::text,total::text
		FROM demo_orders WHERE cart_id=$1 AND idempotency_key_hash=$2`,
		cartID, idempotencyHash).Scan(&replay.ID, &replay.Status, &replay.Currency,
		&replay.Subtotal, &replay.DiscountAmount, &replay.Total)
	if err == nil {
		replay.Replayed = true
		if err := tx.Commit(ctx); err != nil {
			return Order{}, fmt.Errorf("commit checkout replay: %w", err)
		}
		return replay, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Order{}, fmt.Errorf("lookup idempotent order: %w", err)
	}
	if cartStatus != "open" {
		return Order{}, ErrCartClosed
	}
	if cartExpired {
		return Order{}, ErrCartExpired
	}

	// Never process vendor offers through the legacy single-catalog stock path.
	// Vendor-aware split checkout is deliberately introduced as a separate step.
	var vendorLines int64
	if err := tx.QueryRow(ctx, "SELECT count(*) FROM cart_items WHERE cart_id=$1 AND vendor_product_id IS NOT NULL", cartID).Scan(&vendorLines); err != nil {
		return Order{}, fmt.Errorf("check vendor cart lines: %w", err)
	}
	if vendorLines > 0 {
		return Order{}, ErrVendorCheckoutNotReady
	}

	rows, err := tx.Query(ctx, `SELECT product_id, variant_id, quantity
		FROM cart_items WHERE cart_id=$1 ORDER BY product_id, variant_id NULLS FIRST, id FOR UPDATE`, cartID)
	if err != nil {
		return Order{}, fmt.Errorf("read cart items: %w", err)
	}
	type item struct {
		productID int64
		variantID pgtype.Int8
		quantity  int32
	}
	items := make([]item, 0)
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.productID, &it.variantID, &it.quantity); err != nil {
			rows.Close()
			return Order{}, fmt.Errorf("scan cart item: %w", err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return Order{}, fmt.Errorf("iterate cart items: %w", err)
	}
	rows.Close()
	if len(items) == 0 {
		return Order{}, ErrEmptyCart
	}

	var orderID string
	if err := tx.QueryRow(ctx, `INSERT INTO demo_orders
		(cart_id, idempotency_key_hash, currency, subtotal, total)
		VALUES($1,$2,$3,0,0) RETURNING id::text`, cartID, idempotencyHash, currency).Scan(&orderID); err != nil {
		return Order{}, fmt.Errorf("create demo order: %w", err)
	}

	for _, it := range items {
		var title, variantTitle, price string
		var stock int32
		if it.variantID.Valid {
			err = tx.QueryRow(ctx, `SELECT v.price::text, v.stock, p.title, v.title
				FROM product_variants v JOIN products p ON p.id=v.product_id
				WHERE v.id=$1 AND v.product_id=$2 AND v.is_active
				FOR UPDATE OF v`, it.variantID.Int64, it.productID).
				Scan(&price, &stock, &title, &variantTitle)
		} else {
			err = tx.QueryRow(ctx, `SELECT price::text, stock, title, ''
				FROM products WHERE id=$1 FOR UPDATE`, it.productID).
				Scan(&price, &stock, &title, &variantTitle)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return Order{}, ErrUnavailable
		}
		if err != nil {
			return Order{}, fmt.Errorf("lock catalog stock: %w", err)
		}

		var reserved int64
		if it.variantID.Valid {
			err = tx.QueryRow(ctx, `SELECT COALESCE(sum(quantity),0)::bigint
				FROM inventory_reservations
				WHERE variant_id=$1 AND status='active' AND expires_at > now()`, it.variantID.Int64).Scan(&reserved)
		} else {
			err = tx.QueryRow(ctx, `SELECT COALESCE(sum(quantity),0)::bigint
				FROM inventory_reservations
				WHERE product_id=$1 AND variant_id IS NULL
				  AND status='active' AND expires_at > now()`, it.productID).Scan(&reserved)
		}
		if err != nil {
			return Order{}, fmt.Errorf("calculate reserved stock: %w", err)
		}
		if int64(stock)-reserved < int64(it.quantity) {
			return Order{}, ErrInsufficientStock
		}

		var variantArg any
		if it.variantID.Valid {
			variantArg = it.variantID.Int64
		}
		if _, err := tx.Exec(ctx, `INSERT INTO demo_order_items
			(order_id, product_id, variant_id, quantity, unit_price_snapshot, currency, product_title_snapshot, variant_title_snapshot)
			VALUES($1,$2,$3,$4,$5::numeric,$6,$7,$8)`,
			orderID, it.productID, variantArg, it.quantity, price, currency, title, variantTitle); err != nil {
			return Order{}, fmt.Errorf("snapshot order item: %w", err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO inventory_reservations
			(order_id, product_id, variant_id, quantity, expires_at)
			VALUES($1,$2,$3,$4,now()+make_interval(secs => $5::double precision))`,
			orderID, it.productID, variantArg, it.quantity, reservationTTL.Seconds()); err != nil {
			return Order{}, fmt.Errorf("reserve demo stock: %w", err)
		}
	}

	var subtotal string
	if err := tx.QueryRow(ctx, `SELECT COALESCE(sum(unit_price_snapshot * quantity),0)::numeric(12,2)::text
		FROM demo_order_items WHERE order_id=$1`, orderID).Scan(&subtotal); err != nil {
		return Order{}, fmt.Errorf("calculate order subtotal: %w", err)
	}
	discount := "0.00"
	if couponID.Valid {
		var discountType, discountValue, minimumSubtotal string
		var maximumDiscount pgtype.Text
		var startsAt, endsAt, dbNow time.Time
		var usageLimit pgtype.Int8
		var perCustomerLimit int32
		var active bool
		err = tx.QueryRow(ctx, `SELECT discount_type,discount_value::text,minimum_subtotal::text,
			maximum_discount::text,starts_at,ends_at,usage_limit,per_customer_limit,is_active,now()
			FROM coupons WHERE id=$1 FOR UPDATE`, couponID.Int64).
			Scan(&discountType, &discountValue, &minimumSubtotal, &maximumDiscount, &startsAt, &endsAt,
				&usageLimit, &perCustomerLimit, &active, &dbNow)
		if errors.Is(err, pgx.ErrNoRows) {
			return Order{}, ErrCouponUnavailable
		}
		if err != nil {
			return Order{}, fmt.Errorf("lock checkout coupon: %w", err)
		}
		if !active || dbNow.Before(startsAt) || !dbNow.Before(endsAt) {
			return Order{}, ErrCouponUnavailable
		}
		var meetsMinimum bool
		if err := tx.QueryRow(ctx, "SELECT $1::numeric >= $2::numeric", subtotal, minimumSubtotal).Scan(&meetsMinimum); err != nil {
			return Order{}, fmt.Errorf("validate coupon minimum subtotal: %w", err)
		}
		if !meetsMinimum {
			return Order{}, ErrCouponMinimum
		}
		var globalUses, ownerUses int64
		if err := tx.QueryRow(ctx, "SELECT count(*) FROM coupon_redemptions WHERE coupon_id=$1", couponID.Int64).Scan(&globalUses); err != nil {
			return Order{}, fmt.Errorf("count global coupon uses: %w", err)
		}
		if err := tx.QueryRow(ctx, "SELECT count(*) FROM coupon_redemptions WHERE coupon_id=$1 AND owner_token_hash=$2", couponID.Int64, ownerTokenHash).Scan(&ownerUses); err != nil {
			return Order{}, fmt.Errorf("count owner coupon uses: %w", err)
		}
		if (usageLimit.Valid && globalUses >= usageLimit.Int64) || ownerUses >= int64(perCustomerLimit) {
			return Order{}, ErrCouponUnavailable
		}
		var maxDiscount any
		if maximumDiscount.Valid {
			maxDiscount = maximumDiscount.String
		}
		if err := tx.QueryRow(ctx, `SELECT LEAST($1::numeric,
			CASE WHEN $2='percentage' THEN round($1::numeric*$3::numeric/100,2) ELSE $3::numeric END,
			COALESCE($4::numeric,$1::numeric))::numeric(12,2)::text`,
			subtotal, discountType, discountValue, maxDiscount).Scan(&discount); err != nil {
			return Order{}, fmt.Errorf("calculate coupon discount: %w", err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO coupon_redemptions(coupon_id,order_id,owner_token_hash,discount_amount)
			VALUES($1,$2,$3,$4::numeric)`, couponID.Int64, orderID, ownerTokenHash, discount); err != nil {
			return Order{}, fmt.Errorf("record coupon redemption: %w", err)
		}
	}
	var total string
	if err := tx.QueryRow(ctx, "SELECT ($1::numeric-$2::numeric)::numeric(12,2)::text", subtotal, discount).Scan(&total); err != nil {
		return Order{}, fmt.Errorf("calculate discounted order total: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE demo_orders SET coupon_id=$2,subtotal=$3::numeric,
		discount_amount=$4::numeric,total=$5::numeric,updated_at=now() WHERE id=$1`,
		orderID, nullableCouponID(couponID), subtotal, discount, total); err != nil {
		return Order{}, fmt.Errorf("finalize discounted order totals: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO demo_order_events(order_id,event_type,payload)
		VALUES($1,'demo_order_created','{"mode":"simulation"}'::jsonb)`, orderID); err != nil {
		return Order{}, fmt.Errorf("append order event: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE carts SET status='checked_out',updated_at=now() WHERE id=$1`, cartID); err != nil {
		return Order{}, fmt.Errorf("close checked-out cart: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Order{}, fmt.Errorf("commit checkout: %w", err)
	}
	return Order{ID: orderID, Status: "awaiting_demo_payment", Currency: currency, Subtotal: subtotal, DiscountAmount: discount, Total: total}, nil
}

// ExpirePendingOrders releases every active reservation belonging to an order
// whose reservation window has elapsed. The operation is safe to run repeatedly.
func (s *Service) ExpirePendingOrders(ctx context.Context) (int64, error) {
	var expired int64
	err := s.pool.QueryRow(ctx, `WITH expired_orders AS (
		UPDATE demo_orders AS o
		SET status='expired', updated_at=now()
		WHERE o.status='awaiting_demo_payment'
		  AND EXISTS (
			SELECT 1 FROM inventory_reservations AS r
			WHERE r.order_id=o.id AND r.status='active' AND r.expires_at <= now()
		  )
		RETURNING o.id
	), expired_reservations AS (
		UPDATE inventory_reservations AS r
		SET status='expired', updated_at=now()
		FROM expired_orders AS o
		WHERE r.order_id=o.id AND r.status='active'
		RETURNING r.order_id
	), events AS (
		INSERT INTO demo_order_events(order_id,event_type,payload)
		SELECT DISTINCT order_id,'demo_order_expired','{"reason":"reservation_timeout"}'::jsonb
		FROM expired_reservations
		RETURNING id
	)
	SELECT count(*)::bigint FROM expired_orders`).Scan(&expired)
	if err != nil {
		return 0, fmt.Errorf("expire demo orders and reservations: %w", err)
	}
	return expired, nil
}
