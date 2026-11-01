package commerce

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var ErrOrderNotFound = errors.New("order not found")

type OrderItem struct {
	ProductID    int64
	VariantID    *int64
	Quantity     int32
	UnitPrice    string
	Currency     string
	ProductTitle string
	VariantTitle string
}

type OrderEvent struct {
	Type      string
	Payload   json.RawMessage
	CreatedAt time.Time
}

type OrderDetail struct {
	Order
	CreatedAt time.Time
	Items     []OrderItem
	Events    []OrderEvent
}

// GetOrder reads an order only through the guest capability that owns its cart.
// It deliberately accepts an order identifier only as a selector, never as proof
// of ownership. A wrong capability and an unknown order return the same error.
func (s *Service) GetOrder(ctx context.Context, ownerTokenHash []byte, orderID string) (OrderDetail, error) {
	if len(ownerTokenHash) != 32 {
		return OrderDetail{}, errors.New("cart capability digest must be 32 bytes")
	}
	if orderID == "" {
		return OrderDetail{}, ErrOrderNotFound
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return OrderDetail{}, fmt.Errorf("begin order read: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	var detail OrderDetail
	err = tx.QueryRow(ctx, `SELECT o.id::text,o.status,o.currency,o.subtotal::text,o.discount_amount::text,o.total::text,o.created_at
		FROM demo_orders o
		JOIN carts c ON c.id=o.cart_id
		WHERE o.id=$1 AND c.owner_token_hash=$2`, orderID, ownerTokenHash).
		Scan(&detail.ID, &detail.Status, &detail.Currency, &detail.Subtotal, &detail.DiscountAmount, &detail.Total, &detail.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return OrderDetail{}, ErrOrderNotFound
	}
	if err != nil {
		return OrderDetail{}, fmt.Errorf("read owned order: %w", err)
	}

	rows, err := tx.Query(ctx, `SELECT product_id,variant_id,quantity,unit_price_snapshot::text,currency,
		product_title_snapshot,variant_title_snapshot
		FROM demo_order_items WHERE order_id=$1 ORDER BY id`, orderID)
	if err != nil {
		return OrderDetail{}, fmt.Errorf("read order items: %w", err)
	}
	detail.Items = make([]OrderItem, 0)
	for rows.Next() {
		var item OrderItem
		var variant pgtype.Int8
		if err := rows.Scan(&item.ProductID, &variant, &item.Quantity, &item.UnitPrice, &item.Currency, &item.ProductTitle, &item.VariantTitle); err != nil {
			rows.Close()
			return OrderDetail{}, fmt.Errorf("scan order item: %w", err)
		}
		if variant.Valid {
			id := variant.Int64
			item.VariantID = &id
		}
		detail.Items = append(detail.Items, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return OrderDetail{}, fmt.Errorf("iterate order items: %w", err)
	}
	rows.Close()

	eventRows, err := tx.Query(ctx, `SELECT event_type,payload,created_at FROM demo_order_events
		WHERE order_id=$1 ORDER BY created_at,id`, orderID)
	if err != nil {
		return OrderDetail{}, fmt.Errorf("read order events: %w", err)
	}
	detail.Events = make([]OrderEvent, 0)
	for eventRows.Next() {
		var event OrderEvent
		if err := eventRows.Scan(&event.Type, &event.Payload, &event.CreatedAt); err != nil {
			eventRows.Close()
			return OrderDetail{}, fmt.Errorf("scan order event: %w", err)
		}
		detail.Events = append(detail.Events, event)
	}
	if err := eventRows.Err(); err != nil {
		eventRows.Close()
		return OrderDetail{}, fmt.Errorf("iterate order events: %w", err)
	}
	eventRows.Close()
	if err := tx.Commit(ctx); err != nil {
		return OrderDetail{}, fmt.Errorf("commit order read: %w", err)
	}
	return detail, nil
}
