package commerce

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// VendorOrderFullDetail is the owner-scoped view of one child order in a
// split checkout. The parent capability is required; a child UUID is only a
// selector and never an authorization credential.
type VendorOrderFullDetail struct {
	VendorOrderDetail
	ParentOrderID string
	ParentStatus  string
	Items         []OrderItem
}

// GetVendorOrder returns one vendor's child order and immutable item snapshots.
// Unknown child IDs and child IDs owned by another cart intentionally share
// ErrOrderNotFound to avoid leaking cross-customer order existence.
func (s *Service) GetVendorOrder(ctx context.Context, ownerTokenHash []byte, vendorOrderID string) (VendorOrderFullDetail, error) {
	if len(ownerTokenHash) != 32 {
		return VendorOrderFullDetail{}, errors.New("cart capability digest must be 32 bytes")
	}
	if vendorOrderID == "" {
		return VendorOrderFullDetail{}, ErrOrderNotFound
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return VendorOrderFullDetail{}, fmt.Errorf("begin vendor order read: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	var detail VendorOrderFullDetail
	err = tx.QueryRow(ctx, `SELECT vo.id::text,vo.vendor_id,v.display_name,vo.status,vo.currency,
		vo.subtotal::text,vo.discount_amount::text,vo.total::text,vo.created_at,
		o.id::text,o.status
		FROM demo_vendor_orders vo
		JOIN vendors v ON v.id=vo.vendor_id
		JOIN demo_orders o ON o.id=vo.order_id
		JOIN carts c ON c.id=o.cart_id
		WHERE vo.id=$1 AND c.owner_token_hash=$2`, vendorOrderID, ownerTokenHash).
		Scan(&detail.ID, &detail.VendorID, &detail.VendorName, &detail.Status, &detail.Currency,
			&detail.Subtotal, &detail.DiscountAmount, &detail.Total, &detail.CreatedAt,
			&detail.ParentOrderID, &detail.ParentStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return VendorOrderFullDetail{}, ErrOrderNotFound
	}
	if err != nil {
		return VendorOrderFullDetail{}, fmt.Errorf("read owned vendor order: %w", err)
	}

	rows, err := tx.Query(ctx, `SELECT oi.product_id,oi.variant_id,oi.quantity,
		oi.unit_price_snapshot::text,oi.currency,oi.product_title_snapshot,
		oi.variant_title_snapshot,oi.vendor_product_id,oi.vendor_order_id,
		vo.vendor_id,v.display_name
		FROM demo_order_items oi
		JOIN demo_vendor_orders vo ON vo.id=oi.vendor_order_id AND vo.order_id=oi.order_id
		JOIN vendors v ON v.id=vo.vendor_id
		WHERE oi.vendor_order_id=$1 AND oi.order_id=$2
		ORDER BY oi.id`, vendorOrderID, detail.ParentOrderID)
	if err != nil {
		return VendorOrderFullDetail{}, fmt.Errorf("read vendor order items: %w", err)
	}
	detail.Items = make([]OrderItem, 0)
	for rows.Next() {
		var item OrderItem
		var variant, vendorProduct, vendorID pgtype.Int8
		var childID pgtype.Text
		if err := rows.Scan(&item.ProductID, &variant, &item.Quantity, &item.UnitPrice,
			&item.Currency, &item.ProductTitle, &item.VariantTitle, &vendorProduct,
			&childID, &vendorID, &item.VendorName); err != nil {
			rows.Close()
			return VendorOrderFullDetail{}, fmt.Errorf("scan vendor order item: %w", err)
		}
		if variant.Valid {
			id := variant.Int64
			item.VariantID = &id
		}
		if vendorProduct.Valid {
			id := vendorProduct.Int64
			item.VendorProductID = &id
		}
		if childID.Valid {
			id := childID.String
			item.VendorOrderID = &id
		}
		if vendorID.Valid {
			id := vendorID.Int64
			item.VendorID = &id
		}
		detail.Items = append(detail.Items, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return VendorOrderFullDetail{}, fmt.Errorf("iterate vendor order items: %w", err)
	}
	rows.Close()

	if err := tx.Commit(ctx); err != nil {
		return VendorOrderFullDetail{}, fmt.Errorf("commit vendor order read: %w", err)
	}
	return detail, nil
}
