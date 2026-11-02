package commerce

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pooya16pg-spec/globalcart/backend/internal/auth"
)

// BindCartToCustomer attaches a guest-capability cart to the currently
// authenticated customer account. Possession of the 32-byte digest proves
// control of the guest cart; a live customer session proves the destination
// account. The first account to bind wins, and a different account receives
// the same not-found result as an unknown capability.
func (s *Service) BindCartToCustomer(
	ctx context.Context,
	principal auth.SessionPrincipal,
	ownerTokenHash []byte,
) error {
	if len(ownerTokenHash) != 32 {
		return errors.New("cart capability digest must be 32 bytes")
	}
	if err := auth.Authorize(
		principal,
		auth.AuthorizationCustomerWriteOwnResource,
		auth.AuthorizationResource{OwnerUserID: principal.UserID},
	); err != nil {
		return err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("begin cart account binding: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	if err := auth.AuthorizeCustomerSessionTx(ctx, tx, principal); err != nil {
		return err
	}

	var cartID, status string
	var ownerUserID pgtype.Text
	var expired bool
	err = tx.QueryRow(ctx,
		"SELECT id::text,status,owner_user_id::text,expires_at <= now() "+
			"FROM carts WHERE owner_token_hash = $1 FOR UPDATE",
		ownerTokenHash,
	).Scan(&cartID, &status, &ownerUserID, &expired)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrCartNotFound
	}
	if err != nil {
		return fmt.Errorf("lock cart for customer binding: %w", err)
	}
	if expired {
		return ErrCartExpired
	}
	if status != "open" && status != "checked_out" {
		return ErrCartClosed
	}
	if ownerUserID.Valid {
		if ownerUserID.String != principal.UserID {
			return ErrCartNotFound
		}
	} else {
		if _, err := tx.Exec(ctx,
			"UPDATE carts SET owner_user_id = $1, updated_at = now() "+
				"WHERE id = $2 AND owner_user_id IS NULL",
			principal.UserID, cartID,
		); err != nil {
			return fmt.Errorf("bind cart to customer: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit cart account binding: %w", err)
	}
	return nil
}

// GetCartForCustomer reads a cart by its opaque ID only after resolving its
// canonical owner_user_id from PostgreSQL and checking it against the live
// customer session in the same transaction. Foreign and guest-only carts are
// intentionally indistinguishable.
func (s *Service) GetCartForCustomer(
	ctx context.Context,
	principal auth.SessionPrincipal,
	cartID string,
) (Cart, error) {
	if cartID == "" {
		return Cart{}, ErrCartNotFound
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return Cart{}, fmt.Errorf("begin customer cart read: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	if err := auth.AuthorizeCustomerSessionTx(ctx, tx, principal); err != nil {
		return Cart{}, err
	}

	var ownerUserID pgtype.Text
	var cart Cart
	var expired bool
	err = tx.QueryRow(ctx,
		"SELECT id::text,status,currency,expires_at,coupon_id IS NOT NULL,"+
			"owner_user_id::text,expires_at <= now() "+
			"FROM carts WHERE id = $1 FOR SHARE",
		cartID,
	).Scan(
		&cart.ID, &cart.Status, &cart.Currency, &cart.ExpiresAt,
		&cart.CouponApplied, &ownerUserID, &expired,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Cart{}, ErrCartNotFound
	}
	if err != nil {
		return Cart{}, fmt.Errorf("read customer cart owner: %w", err)
	}
	if !ownerUserID.Valid {
		return Cart{}, ErrCartNotFound
	}
	if err := auth.Authorize(
		principal,
		auth.AuthorizationCustomerReadOwnResource,
		auth.AuthorizationResource{OwnerUserID: ownerUserID.String},
	); err != nil {
		return Cart{}, ErrCartNotFound
	}
	if expired {
		return Cart{}, ErrCartExpired
	}

	cart.Items, err = loadCartItems(ctx, tx, cart.ID)
	if err != nil {
		return Cart{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Cart{}, fmt.Errorf("commit customer cart read: %w", err)
	}
	return cart, nil
}

type customerCartMutation uint8

const (
	customerCartAdd customerCartMutation = iota + 1
	customerCartSetQuantity
	customerCartRemove
)

// AddItemForCustomer adds a catalog item to a cart already bound to the
// authenticated customer. The cart UUID is only a selector: ownership and the
// live customer session are revalidated under locks in this transaction.
func (s *Service) AddItemForCustomer(ctx context.Context, principal auth.SessionPrincipal, cartID string, productID int64, variantID *int64, quantity int32) (Cart, error) {
	return s.mutateCustomerCartItem(ctx, principal, cartID, productID, variantID, quantity, customerCartAdd)
}

// SetItemQuantityForCustomer changes an existing customer-owned cart line.
func (s *Service) SetItemQuantityForCustomer(ctx context.Context, principal auth.SessionPrincipal, cartID string, productID int64, variantID *int64, quantity int32) (Cart, error) {
	return s.mutateCustomerCartItem(ctx, principal, cartID, productID, variantID, quantity, customerCartSetQuantity)
}

// RemoveItemForCustomer removes a line from a customer-owned cart.
func (s *Service) RemoveItemForCustomer(ctx context.Context, principal auth.SessionPrincipal, cartID string, productID int64, variantID *int64) (Cart, error) {
	return s.mutateCustomerCartItem(ctx, principal, cartID, productID, variantID, 0, customerCartRemove)
}

func (s *Service) mutateCustomerCartItem(ctx context.Context, principal auth.SessionPrincipal, cartID string, productID int64, variantID *int64, quantity int32, mutation customerCartMutation) (Cart, error) {
	if cartID == "" {
		return Cart{}, ErrCartNotFound
	}
	if productID <= 0 || (variantID != nil && *variantID <= 0) {
		return Cart{}, ErrInvalidProduct
	}
	if mutation != customerCartRemove && (quantity < 1 || quantity > 99) {
		return Cart{}, ErrInvalidQuantity
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return Cart{}, fmt.Errorf("begin customer cart mutation: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	if err := auth.AuthorizeCustomerSessionTx(ctx, tx, principal); err != nil {
		return Cart{}, err
	}

	var cart Cart
	var ownerUserID pgtype.Text
	var expired bool
	err = tx.QueryRow(ctx, `SELECT id::text,status,currency,expires_at,coupon_id IS NOT NULL,
		owner_user_id::text,expires_at <= now()
		FROM carts WHERE id=$1 FOR UPDATE`, cartID).Scan(
		&cart.ID, &cart.Status, &cart.Currency, &cart.ExpiresAt,
		&cart.CouponApplied, &ownerUserID, &expired,
	)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && !ownerUserID.Valid) {
		return Cart{}, ErrCartNotFound
	}
	if err != nil {
		return Cart{}, fmt.Errorf("lock customer cart: %w", err)
	}
	if err := auth.Authorize(principal, auth.AuthorizationCustomerWriteOwnResource, auth.AuthorizationResource{OwnerUserID: ownerUserID.String}); err != nil {
		return Cart{}, ErrCartNotFound
	}
	if cart.Status != "open" {
		return Cart{}, ErrCartClosed
	}
	if expired {
		return Cart{}, ErrCartExpired
	}

	if mutation == customerCartRemove {
		var tag interface{ RowsAffected() int64 }
		if variantID == nil {
			tag, err = tx.Exec(ctx, `DELETE FROM cart_items
				WHERE cart_id=$1 AND product_id=$2 AND variant_id IS NULL AND vendor_product_id IS NULL`, cart.ID, productID)
		} else {
			tag, err = tx.Exec(ctx, `DELETE FROM cart_items
				WHERE cart_id=$1 AND product_id=$2 AND variant_id=$3 AND vendor_product_id IS NULL`, cart.ID, productID, *variantID)
		}
		if err != nil {
			return Cart{}, fmt.Errorf("remove customer cart item: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return Cart{}, ErrInvalidProduct
		}
	} else {
		var title, variantTitle, price string
		var stock int32
		if variantID == nil {
			err = tx.QueryRow(ctx, `SELECT title,price::text,stock FROM products WHERE id=$1 FOR UPDATE`, productID).Scan(&title, &price, &stock)
		} else {
			err = tx.QueryRow(ctx, `SELECT p.title,v.title,v.price::text,v.stock
				FROM product_variants v JOIN products p ON p.id=v.product_id
				WHERE v.id=$1 AND v.product_id=$2 AND v.is_active FOR UPDATE OF v`, *variantID, productID).Scan(&title, &variantTitle, &price, &stock)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return Cart{}, ErrInvalidProduct
		}
		if err != nil {
			return Cart{}, fmt.Errorf("read customer cart catalog item: %w", err)
		}
		if quantity > stock {
			return Cart{}, ErrInsufficientStock
		}

		if mutation == customerCartAdd {
			var existing int32
			if variantID == nil {
				err = tx.QueryRow(ctx, `SELECT COALESCE(max(quantity),0) FROM cart_items
					WHERE cart_id=$1 AND product_id=$2 AND variant_id IS NULL AND vendor_product_id IS NULL`, cart.ID, productID).Scan(&existing)
			} else {
				err = tx.QueryRow(ctx, `SELECT COALESCE(max(quantity),0) FROM cart_items
					WHERE cart_id=$1 AND variant_id=$2 AND vendor_product_id IS NULL`, cart.ID, *variantID).Scan(&existing)
			}
			if err != nil {
				return Cart{}, fmt.Errorf("read existing customer cart quantity: %w", err)
			}
			if existing+quantity > 99 {
				return Cart{}, ErrInvalidQuantity
			}
			if existing+quantity > stock {
				return Cart{}, ErrInsufficientStock
			}

			if variantID == nil {
				_, err = tx.Exec(ctx, `INSERT INTO cart_items
					(cart_id,product_id,quantity,unit_price_snapshot,currency,product_title_snapshot)
					VALUES($1,$2,$3,$4::numeric,$5,$6)
					ON CONFLICT (cart_id,product_id)
					WHERE vendor_product_id IS NULL AND variant_id IS NULL
					DO UPDATE SET quantity=cart_items.quantity+EXCLUDED.quantity,
						unit_price_snapshot=EXCLUDED.unit_price_snapshot,currency=EXCLUDED.currency,
						product_title_snapshot=EXCLUDED.product_title_snapshot,updated_at=now()`,
					cart.ID, productID, quantity, price, cart.Currency, title)
			} else {
				_, err = tx.Exec(ctx, `INSERT INTO cart_items
					(cart_id,product_id,variant_id,quantity,unit_price_snapshot,currency,product_title_snapshot,variant_title_snapshot)
					VALUES($1,$2,$3,$4,$5::numeric,$6,$7,$8)
					ON CONFLICT (cart_id,variant_id)
					WHERE vendor_product_id IS NULL AND variant_id IS NOT NULL
					DO UPDATE SET quantity=cart_items.quantity+EXCLUDED.quantity,
						unit_price_snapshot=EXCLUDED.unit_price_snapshot,currency=EXCLUDED.currency,
						product_title_snapshot=EXCLUDED.product_title_snapshot,
						variant_title_snapshot=EXCLUDED.variant_title_snapshot,updated_at=now()`,
					cart.ID, productID, *variantID, quantity, price, cart.Currency, title, variantTitle)
			}
			if err != nil {
				return Cart{}, fmt.Errorf("add customer cart item: %w", err)
			}
		} else {
			var tag interface{ RowsAffected() int64 }
			if variantID == nil {
				tag, err = tx.Exec(ctx, `UPDATE cart_items SET quantity=$3,unit_price_snapshot=$4::numeric,
					currency=$5,product_title_snapshot=$6,updated_at=now()
					WHERE cart_id=$1 AND product_id=$2 AND variant_id IS NULL AND vendor_product_id IS NULL`,
					cart.ID, productID, quantity, price, cart.Currency, title)
			} else {
				tag, err = tx.Exec(ctx, `UPDATE cart_items SET quantity=$4,unit_price_snapshot=$5::numeric,
					currency=$6,product_title_snapshot=$7,variant_title_snapshot=$8,updated_at=now()
					WHERE cart_id=$1 AND product_id=$2 AND variant_id=$3 AND vendor_product_id IS NULL`,
					cart.ID, productID, *variantID, quantity, price, cart.Currency, title, variantTitle)
			}
			if err != nil {
				return Cart{}, fmt.Errorf("set customer cart quantity: %w", err)
			}
			if tag.RowsAffected() == 0 {
				return Cart{}, ErrInvalidProduct
			}
		}
	}

	cart.Items, err = loadCartItems(ctx, tx, cart.ID)
	if err != nil {
		return Cart{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Cart{}, fmt.Errorf("commit customer cart mutation: %w", err)
	}
	return cart, nil
}

// CancelOrderForCustomer atomically cancels an unpaid demo order after checking
// the current customer session and the cart's canonical database owner. An
// order UUID never acts as proof of ownership; foreign and guest-only orders
// deliberately return ErrOrderNotFound.
func (s *Service) CancelOrderForCustomer(ctx context.Context, principal auth.SessionPrincipal, orderID string) (Order, error) {
	if orderID == "" {
		return Order{}, ErrOrderNotFound
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return Order{}, fmt.Errorf("begin customer order cancellation: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	if err := auth.AuthorizeCustomerSessionTx(ctx, tx, principal); err != nil {
		return Order{}, err
	}

	var order Order
	var ownerUserID pgtype.Text
	err = tx.QueryRow(ctx, `SELECT o.id::text,o.status,o.currency,o.total::text,c.owner_user_id::text
		FROM demo_orders o JOIN carts c ON c.id=o.cart_id
		WHERE o.id=$1 FOR UPDATE OF c,o`, orderID).Scan(
		&order.ID, &order.Status, &order.Currency, &order.Total, &ownerUserID)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && !ownerUserID.Valid) {
		return Order{}, ErrOrderNotFound
	}
	if err != nil {
		return Order{}, fmt.Errorf("lock customer order owner: %w", err)
	}
	if err := auth.Authorize(principal, auth.AuthorizationCustomerWriteOwnResource, auth.AuthorizationResource{OwnerUserID: ownerUserID.String}); err != nil {
		return Order{}, ErrOrderNotFound
	}

	if order.Status == "cancelled" {
		if err := tx.Commit(ctx); err != nil {
			return Order{}, fmt.Errorf("commit customer cancellation replay: %w", err)
		}
		return order, nil
	}
	if order.Status != "awaiting_demo_payment" {
		return Order{}, ErrOrderNotCancellable
	}
	if _, err := tx.Exec(ctx, `UPDATE inventory_reservations SET status='released',updated_at=now()
		WHERE order_id=$1 AND status='active'`, orderID); err != nil {
		return Order{}, fmt.Errorf("release customer order reservations: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE demo_orders SET status='cancelled',updated_at=now() WHERE id=$1`, orderID); err != nil {
		return Order{}, fmt.Errorf("cancel customer demo order: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE demo_vendor_orders SET status='cancelled',updated_at=now()
		WHERE order_id=$1 AND status='awaiting_demo_payment'`, orderID); err != nil {
		return Order{}, fmt.Errorf("cancel customer vendor child orders: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO demo_order_events(order_id,event_type,payload)
		VALUES($1,'demo_order_cancelled','{"reason":"owner_requested","origin":"customer_account"}'::jsonb)`, orderID); err != nil {
		return Order{}, fmt.Errorf("append customer cancellation event: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Order{}, fmt.Errorf("commit customer order cancellation: %w", err)
	}
	order.Status = "cancelled"
	return order, nil
}

// GetVendorOrderForCustomer reads one vendor child order through its parent
// cart's canonical customer owner. Foreign and guest-only child IDs are
// indistinguishable from missing child orders.
func (s *Service) GetVendorOrderForCustomer(ctx context.Context, principal auth.SessionPrincipal, vendorOrderID string) (VendorOrderFullDetail, error) {
	if vendorOrderID == "" {
		return VendorOrderFullDetail{}, ErrOrderNotFound
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return VendorOrderFullDetail{}, fmt.Errorf("begin customer vendor-order read: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	if err := auth.AuthorizeCustomerSessionTx(ctx, tx, principal); err != nil {
		return VendorOrderFullDetail{}, err
	}

	var detail VendorOrderFullDetail
	var ownerUserID pgtype.Text
	err = tx.QueryRow(ctx, `SELECT vo.id::text,vo.vendor_id,v.display_name,vo.status,vo.currency,
		vo.subtotal::text,vo.discount_amount::text,vo.total::text,vo.created_at,
		o.id::text,o.status,c.owner_user_id::text
		FROM demo_vendor_orders vo JOIN vendors v ON v.id=vo.vendor_id
		JOIN demo_orders o ON o.id=vo.order_id JOIN carts c ON c.id=o.cart_id
		WHERE vo.id=$1`, vendorOrderID).Scan(
		&detail.ID, &detail.VendorID, &detail.VendorName, &detail.Status, &detail.Currency,
		&detail.Subtotal, &detail.DiscountAmount, &detail.Total, &detail.CreatedAt,
		&detail.ParentOrderID, &detail.ParentStatus, &ownerUserID)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && !ownerUserID.Valid) {
		return VendorOrderFullDetail{}, ErrOrderNotFound
	}
	if err != nil {
		return VendorOrderFullDetail{}, fmt.Errorf("read customer vendor-order owner: %w", err)
	}
	if err := auth.Authorize(principal, auth.AuthorizationCustomerReadOwnResource, auth.AuthorizationResource{OwnerUserID: ownerUserID.String}); err != nil {
		return VendorOrderFullDetail{}, ErrOrderNotFound
	}

	rows, err := tx.Query(ctx, `SELECT oi.product_id,oi.variant_id,oi.quantity,
		oi.unit_price_snapshot::text,oi.currency,oi.product_title_snapshot,
		oi.variant_title_snapshot,oi.vendor_product_id,oi.vendor_order_id,
		vo.vendor_id,v.display_name
		FROM demo_order_items oi
		JOIN demo_vendor_orders vo ON vo.id=oi.vendor_order_id AND vo.order_id=oi.order_id
		JOIN vendors v ON v.id=vo.vendor_id
		WHERE oi.vendor_order_id=$1 AND oi.order_id=$2 ORDER BY oi.id`, vendorOrderID, detail.ParentOrderID)
	if err != nil {
		return VendorOrderFullDetail{}, fmt.Errorf("read customer vendor-order items: %w", err)
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
			return VendorOrderFullDetail{}, fmt.Errorf("scan customer vendor-order item: %w", err)
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
		return VendorOrderFullDetail{}, fmt.Errorf("iterate customer vendor-order items: %w", err)
	}
	rows.Close()

	if err := tx.Commit(ctx); err != nil {
		return VendorOrderFullDetail{}, fmt.Errorf("commit customer vendor-order read: %w", err)
	}
	return detail, nil
}
