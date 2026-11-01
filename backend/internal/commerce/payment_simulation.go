package commerce

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var ErrDemoPaymentWindowExpired = errors.New("demo payment reservation window expired")

// CompleteDemoPayment simulates a successful payment. It never contacts a
// payment provider, handles no payment credentials, and cannot move real money.
// The owner capability is required; completion and reservation release are
// committed atomically. A repeated call after success is idempotent.
func (s *Service) CompleteDemoPayment(ctx context.Context, ownerTokenHash []byte, orderID string) (Order, error) {
	if len(ownerTokenHash) != 32 {
		return Order{}, errors.New("cart capability digest must be 32 bytes")
	}
	if orderID == "" {
		return Order{}, ErrOrderNotFound
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return Order{}, fmt.Errorf("begin simulated payment completion: %w", err)
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
		return Order{}, fmt.Errorf("lock owned order for simulated payment: %w", err)
	}
	if order.Status == "paid_simulated" {
		if err := tx.Commit(ctx); err != nil {
			return Order{}, fmt.Errorf("commit simulated payment replay: %w", err)
		}
		return order, nil
	}
	if order.Status != "awaiting_demo_payment" {
		return Order{}, ErrOrderNotCancellable
	}

	var itemCount, liveReservationCount int64
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM demo_order_items WHERE order_id=$1`, orderID).Scan(&itemCount); err != nil {
		return Order{}, fmt.Errorf("count simulated order items: %w", err)
	}
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM inventory_reservations
		WHERE order_id=$1 AND status='active' AND expires_at > now()`, orderID).Scan(&liveReservationCount); err != nil {
		return Order{}, fmt.Errorf("count live simulated reservations: %w", err)
	}
	// Every order line must still have its own live hold. A single live hold
	// must never allow completion when another line's reservation has expired.
	if itemCount == 0 || liveReservationCount != itemCount {
		return Order{}, ErrDemoPaymentWindowExpired
	}
	if _, err := tx.Exec(ctx, `UPDATE inventory_reservations
		SET status='released',updated_at=now()
		WHERE order_id=$1 AND status='active'`, orderID); err != nil {
		return Order{}, fmt.Errorf("release simulated reservation: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE demo_orders
		SET status='paid_simulated',updated_at=now() WHERE id=$1`, orderID); err != nil {
		return Order{}, fmt.Errorf("mark demo order paid: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO demo_order_events(order_id,event_type,payload)
		VALUES($1,'demo_payment_simulated','{"provider":"none","mode":"simulation"}'::jsonb)`, orderID); err != nil {
		return Order{}, fmt.Errorf("append simulated payment event: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Order{}, fmt.Errorf("commit simulated payment completion: %w", err)
	}
	order.Status = "paid_simulated"
	return order, nil
}
