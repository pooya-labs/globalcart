package commerce

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrInvalidDemoReturnReason = errors.New("unsupported demo return reason")
	ErrDemoReturnNotFound      = errors.New("demo return not found")
	ErrDemoReturnNotAllowed    = errors.New("demo return transition is not allowed")
	ErrDemoReturnAlreadyExists = errors.New("a return request already exists for this order")
)

const (
	DemoReturnReasonChangedMind    = "changed_mind"
	DemoReturnReasonDamagedDemo    = "damaged_demo"
	DemoReturnReasonNotAsDescribed = "not_as_described"
	DemoReturnReasonWrongItem      = "wrong_item"
)

type DemoReturn struct {
	ID           string
	OrderID      string
	Reason       string
	Status       string
	Currency     string
	RefundAmount string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ReviewedAt   *time.Time
	RefundedAt   *time.Time
	Replayed     bool
}

type demoReturnScanner interface {
	Scan(dest ...any) error
}

func scanDemoReturn(row demoReturnScanner, extra ...any) (DemoReturn, error) {
	var result DemoReturn
	var reviewedAt, refundedAt pgtype.Timestamptz
	dest := []any{
		&result.ID, &result.OrderID, &result.Reason, &result.Status,
		&result.Currency, &result.RefundAmount, &result.CreatedAt, &result.UpdatedAt,
		&reviewedAt, &refundedAt,
	}
	dest = append(dest, extra...)
	if err := row.Scan(dest...); err != nil {
		return DemoReturn{}, err
	}
	if reviewedAt.Valid {
		value := reviewedAt.Time
		result.ReviewedAt = &value
	}
	if refundedAt.Valid {
		value := refundedAt.Time
		result.RefundedAt = &value
	}
	return result, nil
}

func validDemoReturnReason(reason string) bool {
	switch reason {
	case DemoReturnReasonChangedMind, DemoReturnReasonDamagedDemo,
		DemoReturnReasonNotAsDescribed, DemoReturnReasonWrongItem:
		return true
	default:
		return false
	}
}

// RequestDemoReturn creates one full-order return request for a paid simulated
// order. Ownership is proven through the cart capability; the order UUID is
// only a selector. Repeating the same reason is idempotent. This method is
// internal-only and is deliberately not registered with the HTTP router.
func (s *Service) RequestDemoReturn(ctx context.Context, ownerTokenHash []byte, orderID, reason string) (DemoReturn, error) {
	if len(ownerTokenHash) != 32 {
		return DemoReturn{}, errors.New("cart capability digest must be 32 bytes")
	}
	if orderID == "" {
		return DemoReturn{}, ErrOrderNotFound
	}
	if !validDemoReturnReason(reason) {
		return DemoReturn{}, ErrInvalidDemoReturnReason
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return DemoReturn{}, fmt.Errorf("begin demo return request: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	var orderStatus string
	err = tx.QueryRow(ctx, `SELECT o.status FROM demo_orders o
		JOIN carts c ON c.id=o.cart_id
		WHERE o.id=$1 AND c.owner_token_hash=$2
		FOR UPDATE OF o`, orderID, ownerTokenHash).Scan(&orderStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return DemoReturn{}, ErrOrderNotFound
	}
	if err != nil {
		return DemoReturn{}, fmt.Errorf("lock owned order for return: %w", err)
	}

	existing, err := scanDemoReturn(tx.QueryRow(ctx, `SELECT id::text,order_id::text,reason,status,currency,
		refund_amount::text,created_at,updated_at,reviewed_at,refunded_at
		FROM demo_returns WHERE order_id=$1 FOR UPDATE`, orderID))
	if err == nil {
		if existing.Reason != reason {
			return DemoReturn{}, ErrDemoReturnAlreadyExists
		}
		existing.Replayed = true
		if err := tx.Commit(ctx); err != nil {
			return DemoReturn{}, fmt.Errorf("commit demo return replay: %w", err)
		}
		return existing, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return DemoReturn{}, fmt.Errorf("check existing demo return: %w", err)
	}
	if orderStatus != "paid_simulated" {
		return DemoReturn{}, ErrDemoReturnNotAllowed
	}

	result, err := scanDemoReturn(tx.QueryRow(ctx, `INSERT INTO demo_returns(order_id,reason,currency)
		SELECT o.id,$2,o.currency FROM demo_orders o WHERE o.id=$1
		RETURNING id::text,order_id::text,reason,status,currency,refund_amount::text,
		          created_at,updated_at,reviewed_at,refunded_at`, orderID, reason))
	if err != nil {
		return DemoReturn{}, fmt.Errorf("insert demo return: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO demo_order_events(order_id,event_type,payload)
		VALUES($1,'demo_return_requested',jsonb_build_object('return_id',$2::text,'reason',$3::text))`,
		orderID, result.ID, reason); err != nil {
		return DemoReturn{}, fmt.Errorf("append demo return request event: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return DemoReturn{}, fmt.Errorf("commit demo return request: %w", err)
	}
	return result, nil
}

// ReviewDemoReturn is an internal operator decision, not a customer or seller
// endpoint. It must remain behind a verified admin authorization boundary
// before any HTTP route is added. Approval snapshots the full persisted net
// order total (including coupon discounts); rejection records no refund.
func (s *Service) ReviewDemoReturn(ctx context.Context, orderID string, approve bool) (DemoReturn, error) {
	if orderID == "" {
		return DemoReturn{}, ErrDemoReturnNotFound
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return DemoReturn{}, fmt.Errorf("begin demo return review: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	var orderStatus, orderTotal string
	result, err := scanDemoReturn(tx.QueryRow(ctx, `SELECT r.id::text,r.order_id::text,r.reason,r.status,r.currency,
		r.refund_amount::text,r.created_at,r.updated_at,r.reviewed_at,r.refunded_at,
		o.status,o.total::text
		FROM demo_orders o JOIN demo_returns r ON r.order_id=o.id
		WHERE o.id=$1 FOR UPDATE OF o,r`, orderID), &orderStatus, &orderTotal)
	if errors.Is(err, pgx.ErrNoRows) {
		return DemoReturn{}, ErrDemoReturnNotFound
	}
	if err != nil {
		return DemoReturn{}, fmt.Errorf("lock demo return for review: %w", err)
	}

	targetStatus, eventType, refundAmount := "rejected", "demo_return_rejected", "0.00"
	if approve {
		targetStatus, eventType, refundAmount = "approved", "demo_return_approved", orderTotal
	}
	if result.Status == targetStatus {
		result.Replayed = true
		if err := tx.Commit(ctx); err != nil {
			return DemoReturn{}, fmt.Errorf("commit demo return review replay: %w", err)
		}
		return result, nil
	}
	if result.Status != "requested" || orderStatus != "paid_simulated" {
		return DemoReturn{}, ErrDemoReturnNotAllowed
	}
	if _, err := tx.Exec(ctx, `UPDATE demo_returns
		SET status=$2,refund_amount=$3::numeric,reviewed_at=now(),updated_at=now()
		WHERE order_id=$1`, orderID, targetStatus, refundAmount); err != nil {
		return DemoReturn{}, fmt.Errorf("record demo return decision: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO demo_order_events(order_id,event_type,payload)
		VALUES($1,$2,jsonb_build_object('return_id',$3::text,'refund_amount',$4::numeric,'currency',$5::text))`,
		orderID, eventType, result.ID, refundAmount, result.Currency); err != nil {
		return DemoReturn{}, fmt.Errorf("append demo return decision event: %w", err)
	}
	result, err = scanDemoReturn(tx.QueryRow(ctx, `SELECT id::text,order_id::text,reason,status,currency,
		refund_amount::text,created_at,updated_at,reviewed_at,refunded_at
		FROM demo_returns WHERE order_id=$1`, orderID))
	if err != nil {
		return DemoReturn{}, fmt.Errorf("read reviewed demo return: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return DemoReturn{}, fmt.Errorf("commit demo return review: %w", err)
	}
	return result, nil
}

// CompleteDemoRefund records a full-order simulated refund after internal
// approval. It never calls a payment provider, changes inventory or moves
// money. Parent and vendor child states plus the order event are updated
// atomically. The owner capability is required and repeated completion is
// idempotent. This method is not exposed over HTTP.
func (s *Service) CompleteDemoRefund(ctx context.Context, ownerTokenHash []byte, orderID string) (DemoReturn, error) {
	if len(ownerTokenHash) != 32 {
		return DemoReturn{}, errors.New("cart capability digest must be 32 bytes")
	}
	if orderID == "" {
		return DemoReturn{}, ErrOrderNotFound
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return DemoReturn{}, fmt.Errorf("begin simulated demo refund: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	var orderStatus string
	err = tx.QueryRow(ctx, `SELECT o.status FROM demo_orders o
		JOIN carts c ON c.id=o.cart_id
		WHERE o.id=$1 AND c.owner_token_hash=$2
		FOR UPDATE OF o`, orderID, ownerTokenHash).Scan(&orderStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return DemoReturn{}, ErrOrderNotFound
	}
	if err != nil {
		return DemoReturn{}, fmt.Errorf("lock owned order for simulated refund: %w", err)
	}

	result, err := scanDemoReturn(tx.QueryRow(ctx, `SELECT id::text,order_id::text,reason,status,currency,
		refund_amount::text,created_at,updated_at,reviewed_at,refunded_at
		FROM demo_returns WHERE order_id=$1 FOR UPDATE`, orderID))
	if errors.Is(err, pgx.ErrNoRows) {
		return DemoReturn{}, ErrDemoReturnNotFound
	}
	if err != nil {
		return DemoReturn{}, fmt.Errorf("lock approved demo return: %w", err)
	}
	if orderStatus == "refunded_simulated" && result.Status == "refunded_simulated" {
		result.Replayed = true
		if err := tx.Commit(ctx); err != nil {
			return DemoReturn{}, fmt.Errorf("commit simulated refund replay: %w", err)
		}
		return result, nil
	}
	if orderStatus != "paid_simulated" || result.Status != "approved" {
		return DemoReturn{}, ErrDemoReturnNotAllowed
	}

	if _, err := tx.Exec(ctx, `UPDATE demo_orders
		SET status='refunded_simulated',updated_at=now()
		WHERE id=$1 AND status='paid_simulated'`, orderID); err != nil {
		return DemoReturn{}, fmt.Errorf("mark parent order refunded in simulation: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE demo_vendor_orders
		SET status='refunded_simulated',updated_at=now()
		WHERE order_id=$1 AND status='paid_simulated'`, orderID); err != nil {
		return DemoReturn{}, fmt.Errorf("mark vendor child orders refunded in simulation: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE demo_returns
		SET status='refunded_simulated',refunded_at=now(),updated_at=now()
		WHERE order_id=$1 AND status='approved'`, orderID); err != nil {
		return DemoReturn{}, fmt.Errorf("complete simulated demo refund: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO demo_order_events(order_id,event_type,payload)
		VALUES($1,'demo_refund_simulated',
			jsonb_build_object('provider','none','mode','simulation','return_id',$2::text,
			                   'amount',$3::numeric,'currency',$4::text))`,
		orderID, result.ID, result.RefundAmount, result.Currency); err != nil {
		return DemoReturn{}, fmt.Errorf("append simulated refund event: %w", err)
	}
	result, err = scanDemoReturn(tx.QueryRow(ctx, `SELECT id::text,order_id::text,reason,status,currency,
		refund_amount::text,created_at,updated_at,reviewed_at,refunded_at
		FROM demo_returns WHERE order_id=$1`, orderID))
	if err != nil {
		return DemoReturn{}, fmt.Errorf("read completed simulated refund: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return DemoReturn{}, fmt.Errorf("commit simulated demo refund: %w", err)
	}
	return result, nil
}
