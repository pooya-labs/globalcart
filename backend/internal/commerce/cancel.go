package commerce

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var ErrOrderNotCancellable = errors.New("order is not cancellable")

// CancelOrder atomically cancels an unpaid demo order and releases its active
// reservations. The owner capability is checked in the locking query; knowing
// an order UUID alone never grants access. Repeating a successful cancellation
// is idempotent and returns the already-cancelled order.
func (s *Service) CancelOrder(ctx context.Context, ownerTokenHash []byte, orderID string) (Order, error) {
	if len(ownerTokenHash) != 32 {
		return Order{}, errors.New("cart capability digest must be 32 bytes")
	}
	if orderID == "" {
		return Order{}, ErrOrderNotFound
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return Order{}, fmt.Errorf("begin order cancellation: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	var order Order
	err = tx.QueryRow(ctx, `SELECT o.id::text,o.status,o.currency,o.total::text
		FROM demo_orders o
		JOIN carts c ON c.id=o.cart_id
		WHERE o.id=$1 AND c.owner_token_hash=$2
		FOR UPDATE OF o`, orderID, ownerTokenHash).
		Scan(&order.ID, &order.Status, &order.Currency, &order.Total)
	if errors.Is(err, pgx.ErrNoRows) {
		return Order{}, ErrOrderNotFound
	}
	if err != nil {
		return Order{}, fmt.Errorf("lock owned order: %w", err)
	}
	if order.Status == "cancelled" {
		if err := tx.Commit(ctx); err != nil {
			return Order{}, fmt.Errorf("commit cancellation replay: %w", err)
		}
		return order, nil
	}
	if order.Status != "awaiting_demo_payment" {
		return Order{}, ErrOrderNotCancellable
	}
	if _, err := tx.Exec(ctx, `UPDATE inventory_reservations SET status='released',updated_at=now()
		WHERE order_id=$1 AND status='active'`, orderID); err != nil {
		return Order{}, fmt.Errorf("release order reservations: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE demo_orders SET status='cancelled',updated_at=now() WHERE id=$1`, orderID); err != nil {
		return Order{}, fmt.Errorf("cancel demo order: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO demo_order_events(order_id,event_type,payload)
		VALUES($1,'demo_order_cancelled','{"reason":"owner_requested"}'::jsonb)`, orderID); err != nil {
		return Order{}, fmt.Errorf("append cancellation event: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Order{}, fmt.Errorf("commit order cancellation: %w", err)
	}
	order.Status = "cancelled"
	return order, nil
}
