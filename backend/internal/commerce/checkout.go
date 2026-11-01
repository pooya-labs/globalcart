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
	ErrCartNotFound = errors.New("cart not found")
	ErrCartClosed = errors.New("cart is not open")
	ErrCartExpired = errors.New("cart expired")
	ErrEmptyCart = errors.New("cart is empty")
	ErrUnavailable = errors.New("item is unavailable")
	ErrInsufficientStock = errors.New("insufficient demo stock")
	ErrInvalidLifetime = errors.New("reservation lifetime must be positive")
)

type Order struct {
	ID       string
	Status   string
	Currency string
	Total    string
	Replayed bool
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
	if err != nil { return Order{}, fmt.Errorf("begin checkout: %w", err) }
	defer func() { _ = tx.Rollback(context.Background()) }()

	var cartID, currency, cartStatus string
	var expiresAt time.Time
	err = tx.QueryRow(ctx, `SELECT id::text, currency, status, expires_at
		FROM carts WHERE owner_token_hash=$1 FOR UPDATE`, ownerTokenHash).
		Scan(&cartID, &currency, &cartStatus, &expiresAt)
	if errors.Is(err, pgx.ErrNoRows) { return Order{}, ErrCartNotFound }
	if err != nil { return Order{}, fmt.Errorf("lock cart: %w", err) }

	// Replay lookup precedes lifecycle checks so a retry after a successful
	// commit can recover the original order even though the cart is checked out.
	var replay Order
	err = tx.QueryRow(ctx, `SELECT id::text, status, currency, total::text
		FROM demo_orders WHERE cart_id=$1 AND idempotency_key_hash=$2`,
		cartID, idempotencyHash).Scan(&replay.ID, &replay.Status, &replay.Currency, &replay.Total)
	if err == nil {
		replay.Replayed = true
		if err := tx.Commit(ctx); err != nil { return Order{}, fmt.Errorf("commit checkout replay: %w", err) }
		return replay, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) { return Order{}, fmt.Errorf("lookup idempotent order: %w", err) }
	if cartStatus != "open" { return Order{}, ErrCartClosed }
	if !expiresAt.After(time.Now()) { return Order{}, ErrCartExpired }

	rows, err := tx.Query(ctx, `SELECT product_id, variant_id, quantity
		FROM cart_items WHERE cart_id=$1 ORDER BY product_id, variant_id NULLS FIRST, id FOR UPDATE`, cartID)
	if err != nil { return Order{}, fmt.Errorf("read cart items: %w", err) }
	type item struct { productID int64; variantID pgtype.Int8; quantity int32 }
	items := make([]item, 0)
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.productID, &it.variantID, &it.quantity); err != nil { rows.Close(); return Order{}, fmt.Errorf("scan cart item: %w", err) }
		items = append(items, it)
	}
	if err := rows.Err(); err != nil { rows.Close(); return Order{}, fmt.Errorf("iterate cart items: %w", err) }
	rows.Close()
	if len(items) == 0 { return Order{}, ErrEmptyCart }

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
		if errors.Is(err, pgx.ErrNoRows) { return Order{}, ErrUnavailable }
		if err != nil { return Order{}, fmt.Errorf("lock catalog stock: %w", err) }

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
		if err != nil { return Order{}, fmt.Errorf("calculate reserved stock: %w", err) }
		if int64(stock)-reserved < int64(it.quantity) { return Order{}, ErrInsufficientStock }

		var variantArg any
		if it.variantID.Valid { variantArg = it.variantID.Int64 }
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

	var total string
	if err := tx.QueryRow(ctx, `SELECT COALESCE(sum(unit_price_snapshot * quantity),0)::numeric(12,2)::text
		FROM demo_order_items WHERE order_id=$1`, orderID).Scan(&total); err != nil {
		return Order{}, fmt.Errorf("calculate order total: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE demo_orders SET subtotal=$2::numeric,total=$2::numeric,updated_at=now() WHERE id=$1`, orderID, total); err != nil {
		return Order{}, fmt.Errorf("finalize order total: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO demo_order_events(order_id,event_type,payload)
		VALUES($1,'demo_order_created','{"mode":"simulation"}'::jsonb)`, orderID); err != nil {
		return Order{}, fmt.Errorf("append order event: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE carts SET status='checked_out',updated_at=now() WHERE id=$1`, cartID); err != nil {
		return Order{}, fmt.Errorf("close checked-out cart: %w", err)
	}
	if err := tx.Commit(ctx); err != nil { return Order{}, fmt.Errorf("commit checkout: %w", err) }
	return Order{ID: orderID, Status: "awaiting_demo_payment", Currency: currency, Total: total}, nil
}
