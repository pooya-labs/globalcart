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
	ProductID       int64
	VariantID       *int64
	VendorProductID *int64
	VendorOrderID   *string
	VendorID        *int64
	VendorName      string
	Quantity        int32
	UnitPrice       string
	Currency        string
	ProductTitle    string
	VariantTitle    string
}

type VendorOrderDetail struct {
	ID             string
	VendorID       int64
	VendorName     string
	Status         string
	Currency       string
	Subtotal       string
	DiscountAmount string
	Total          string
	CreatedAt      time.Time
}

type OrderEvent struct {
	Type      string
	Payload   json.RawMessage
	CreatedAt time.Time
}

type OrderDetail struct {
	Order
	CreatedAt   time.Time
	Items       []OrderItem
	VendorOrders []VendorOrderDetail
	Events      []OrderEvent
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

	rows, err := tx.Query(ctx, `SELECT oi.product_id,oi.variant_id,oi.quantity,oi.unit_price_snapshot::text,oi.currency,
		oi.product_title_snapshot,oi.variant_title_snapshot,oi.vendor_product_id,oi.vendor_order_id,
		vo.vendor_id,COALESCE(v.display_name,'')
		FROM demo_order_items oi
		LEFT JOIN demo_vendor_orders vo ON vo.id=oi.vendor_order_id
		LEFT JOIN vendors v ON v.id=vo.vendor_id
		WHERE oi.order_id=$1 ORDER BY oi.id`, orderID)
	if err != nil {
		return OrderDetail{}, fmt.Errorf("read order items: %w", err)
	}
	detail.Items = make([]OrderItem, 0)
	for rows.Next() {
		var item OrderItem
		var variant, vendorProduct, vendorID pgtype.Int8
		var vendorOrder pgtype.Text
		if err := rows.Scan(&item.ProductID, &variant, &item.Quantity, &item.UnitPrice, &item.Currency, &item.ProductTitle, &item.VariantTitle, &vendorProduct, &vendorOrder, &vendorID, &item.VendorName); err != nil {
			rows.Close()
			return OrderDetail{}, fmt.Errorf("scan order item: %w", err)
		}
		if variant.Valid { id := variant.Int64; item.VariantID = &id }
		if vendorProduct.Valid { id := vendorProduct.Int64; item.VendorProductID = &id }
		if vendorOrder.Valid { id := vendorOrder.String; item.VendorOrderID = &id }
		if vendorID.Valid { id := vendorID.Int64; item.VendorID = &id }
		detail.Items = append(detail.Items, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return OrderDetail{}, fmt.Errorf("iterate order items: %w", err)
	}
	rows.Close()

	vendorRows, err := tx.Query(ctx, `SELECT vo.id::text,vo.vendor_id,v.display_name,vo.status,vo.currency,
		vo.subtotal::text,vo.discount_amount::text,vo.total::text,vo.created_at
		FROM demo_vendor_orders vo JOIN vendors v ON v.id=vo.vendor_id
		WHERE vo.order_id=$1 ORDER BY vo.vendor_id,vo.id`, orderID)
	if err != nil { return OrderDetail{}, fmt.Errorf("read vendor child orders: %w", err) }
	detail.VendorOrders = make([]VendorOrderDetail, 0)
	for vendorRows.Next() {
		var child VendorOrderDetail
		if err := vendorRows.Scan(&child.ID,&child.VendorID,&child.VendorName,&child.Status,&child.Currency,&child.Subtotal,&child.DiscountAmount,&child.Total,&child.CreatedAt); err != nil {
			vendorRows.Close(); return OrderDetail{}, fmt.Errorf("scan vendor child order: %w", err)
		}
		detail.VendorOrders = append(detail.VendorOrders, child)
	}
	if err := vendorRows.Err(); err != nil { vendorRows.Close(); return OrderDetail{}, fmt.Errorf("iterate vendor child orders: %w", err) }
	vendorRows.Close()

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
