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
	ErrInvalidQuantity     = errors.New("quantity must be between 1 and 99")
	ErrInvalidProduct      = errors.New("product or variant is unavailable")
	ErrUnsupportedCurrency = errors.New("only USD catalog carts are supported")
)

type CartItem struct {
	ProductID    int64
	VariantID    *int64
	Quantity     int32
	UnitPrice    string
	Currency     string
	ProductTitle string
	VariantTitle string
}

type Cart struct {
	ID        string
	Status    string
	Currency  string
	ExpiresAt time.Time
	Items     []CartItem
}

func (s *Service) CreateCart(ctx context.Context, ownerTokenHash []byte, currency string, ttl time.Duration) (Cart, error) {
	if len(ownerTokenHash) != 32 {
		return Cart{}, errors.New("cart capability digest must be 32 bytes")
	}
	if currency != "USD" {
		return Cart{}, ErrUnsupportedCurrency
	}
	if ttl <= 0 {
		return Cart{}, ErrInvalidLifetime
	}
	var cart Cart
	err := s.pool.QueryRow(ctx, `INSERT INTO carts(owner_token_hash,currency,expires_at)
		VALUES($1,$2,now()+make_interval(secs => $3::double precision))
		RETURNING id::text,status,currency,expires_at`, ownerTokenHash, currency, ttl.Seconds()).
		Scan(&cart.ID, &cart.Status, &cart.Currency, &cart.ExpiresAt)
	if err != nil {
		return Cart{}, fmt.Errorf("create cart: %w", err)
	}
	return cart, nil
}

// AddItem increments the requested line. The cart row serializes concurrent
// mutations of the same guest cart; price/title are always read from catalog.
func (s *Service) AddItem(ctx context.Context, ownerTokenHash []byte, productID int64, variantID *int64, quantity int32) (Cart, error) {
	if len(ownerTokenHash) != 32 {
		return Cart{}, errors.New("cart capability digest must be 32 bytes")
	}
	if productID <= 0 || (variantID != nil && *variantID <= 0) {
		return Cart{}, ErrInvalidProduct
	}
	if quantity < 1 || quantity > 99 {
		return Cart{}, ErrInvalidQuantity
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return Cart{}, fmt.Errorf("begin cart update: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	var cartID, status, currency string
	var expired bool
	err = tx.QueryRow(ctx, `SELECT id::text,status,currency,expires_at<=now()
		FROM carts WHERE owner_token_hash=$1 FOR UPDATE`, ownerTokenHash).Scan(&cartID, &status, &currency, &expired)
	if errors.Is(err, pgx.ErrNoRows) {
		return Cart{}, ErrCartNotFound
	}
	if err != nil {
		return Cart{}, fmt.Errorf("lock cart: %w", err)
	}
	if status != "open" {
		return Cart{}, ErrCartClosed
	}
	if expired {
		return Cart{}, ErrCartExpired
	}
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
		return Cart{}, fmt.Errorf("read current catalog item: %w", err)
	}
	var existing int32
	if variantID == nil {
		err = tx.QueryRow(ctx, `SELECT COALESCE(max(quantity),0) FROM cart_items
			WHERE cart_id=$1 AND product_id=$2 AND variant_id IS NULL`, cartID, productID).Scan(&existing)
	} else {
		err = tx.QueryRow(ctx, `SELECT COALESCE(max(quantity),0) FROM cart_items
			WHERE cart_id=$1 AND variant_id=$2`, cartID, *variantID).Scan(&existing)
	}
	if err != nil {
		return Cart{}, fmt.Errorf("read existing cart quantity: %w", err)
	}
	if existing+quantity > 99 {
		return Cart{}, ErrInvalidQuantity
	}
	if existing+quantity > stock {
		return Cart{}, ErrInsufficientStock
	}
	var variantArg any
	if variantID != nil {
		variantArg = *variantID
	}
	if variantID == nil {
		_, err = tx.Exec(ctx, `INSERT INTO cart_items(cart_id,product_id,quantity,unit_price_snapshot,currency,product_title_snapshot)
			VALUES($1,$2,$3,$4::numeric,$5,$6)
			ON CONFLICT (cart_id,product_id) WHERE variant_id IS NULL
			DO UPDATE SET quantity=cart_items.quantity+EXCLUDED.quantity,unit_price_snapshot=EXCLUDED.unit_price_snapshot,
			currency=EXCLUDED.currency,product_title_snapshot=EXCLUDED.product_title_snapshot,updated_at=now()`,
			cartID, productID, quantity, price, currency, title)
	} else {
		_, err = tx.Exec(ctx, `INSERT INTO cart_items(cart_id,product_id,variant_id,quantity,unit_price_snapshot,currency,product_title_snapshot,variant_title_snapshot)
			VALUES($1,$2,$3,$4,$5::numeric,$6,$7,$8)
			ON CONFLICT (cart_id,variant_id) WHERE variant_id IS NOT NULL
			DO UPDATE SET quantity=cart_items.quantity+EXCLUDED.quantity,unit_price_snapshot=EXCLUDED.unit_price_snapshot,
			currency=EXCLUDED.currency,product_title_snapshot=EXCLUDED.product_title_snapshot,
			variant_title_snapshot=EXCLUDED.variant_title_snapshot,updated_at=now()`,
			cartID, productID, variantArg, quantity, price, currency, title, variantTitle)
	}
	if err != nil {
		return Cart{}, fmt.Errorf("upsert cart item: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return Cart{}, fmt.Errorf("commit cart update: %w", err)
	}
	return s.GetCart(ctx, ownerTokenHash)
}

// SetItemQuantity sets an existing line to a quantity from 1 through 99.
func (s *Service) SetItemQuantity(ctx context.Context, ownerTokenHash []byte, productID int64, variantID *int64, quantity int32) (Cart, error) {
	if quantity < 1 || quantity > 99 {
		return Cart{}, ErrInvalidQuantity
	}
	if len(ownerTokenHash) != 32 {
		return Cart{}, errors.New("cart capability digest must be 32 bytes")
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return Cart{}, fmt.Errorf("begin cart quantity update: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	var cartID, status string
	var expired bool
	err = tx.QueryRow(ctx, `SELECT id::text,status,expires_at<=now() FROM carts WHERE owner_token_hash=$1 FOR UPDATE`, ownerTokenHash).Scan(&cartID, &status, &expired)
	if errors.Is(err, pgx.ErrNoRows) {
		return Cart{}, ErrCartNotFound
	}
	if err != nil {
		return Cart{}, fmt.Errorf("lock cart: %w", err)
	}
	if status != "open" {
		return Cart{}, ErrCartClosed
	}
	if expired {
		return Cart{}, ErrCartExpired
	}
	var stock int32
	var price string
	var title, variantTitle string
	if variantID == nil {
		err = tx.QueryRow(ctx, `SELECT title,price::text,stock FROM products WHERE id=$1 FOR UPDATE`, productID).Scan(&title, &price, &stock)
	} else {
		err = tx.QueryRow(ctx, `SELECT p.title,v.title,v.price::text,v.stock FROM product_variants v
			JOIN products p ON p.id=v.product_id WHERE v.id=$1 AND v.product_id=$2 AND v.is_active FOR UPDATE OF v`, *variantID, productID).
			Scan(&title, &variantTitle, &price, &stock)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return Cart{}, ErrInvalidProduct
	}
	if err != nil {
		return Cart{}, fmt.Errorf("read current catalog item: %w", err)
	}
	if quantity > stock {
		return Cart{}, ErrInsufficientStock
	}
	var affected int64
	if variantID == nil {
		ct, execErr := tx.Exec(ctx, `UPDATE cart_items SET quantity=$3,unit_price_snapshot=$4::numeric,
			product_title_snapshot=$5,updated_at=now() WHERE cart_id=$1 AND product_id=$2 AND variant_id IS NULL`,
			cartID, productID, quantity, price, title)
		err = execErr
		affected = ct.RowsAffected()
	} else {
		ct, execErr := tx.Exec(ctx, `UPDATE cart_items SET quantity=$4,unit_price_snapshot=$5::numeric,
			product_title_snapshot=$6,variant_title_snapshot=$7,updated_at=now()
			WHERE cart_id=$1 AND product_id=$2 AND variant_id=$3`, cartID, productID, *variantID, quantity, price, title, variantTitle)
		err = execErr
		affected = ct.RowsAffected()
	}
	if err != nil {
		return Cart{}, fmt.Errorf("set cart quantity: %w", err)
	}
	if affected == 0 {
		return Cart{}, ErrInvalidProduct
	}
	if err = tx.Commit(ctx); err != nil {
		return Cart{}, fmt.Errorf("commit cart quantity update: %w", err)
	}
	return s.GetCart(ctx, ownerTokenHash)
}

// RemoveItem removes a line from an open guest cart.
func (s *Service) RemoveItem(ctx context.Context, ownerTokenHash []byte, productID int64, variantID *int64) (Cart, error) {
	if len(ownerTokenHash) != 32 {
		return Cart{}, errors.New("cart capability digest must be 32 bytes")
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return Cart{}, fmt.Errorf("begin cart removal: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	var cartID, status string
	var expired bool
	err = tx.QueryRow(ctx, `SELECT id::text,status,expires_at<=now() FROM carts WHERE owner_token_hash=$1 FOR UPDATE`, ownerTokenHash).Scan(&cartID, &status, &expired)
	if errors.Is(err, pgx.ErrNoRows) {
		return Cart{}, ErrCartNotFound
	}
	if err != nil {
		return Cart{}, fmt.Errorf("lock cart: %w", err)
	}
	if status != "open" {
		return Cart{}, ErrCartClosed
	}
	if expired {
		return Cart{}, ErrCartExpired
	}
	var affected int64
	if variantID == nil {
		tag, e := tx.Exec(ctx, `DELETE FROM cart_items WHERE cart_id=$1 AND product_id=$2 AND variant_id IS NULL`, cartID, productID)
		err = e
		affected = tag.RowsAffected()
	} else {
		tag, e := tx.Exec(ctx, `DELETE FROM cart_items WHERE cart_id=$1 AND product_id=$2 AND variant_id=$3`, cartID, productID, *variantID)
		err = e
		affected = tag.RowsAffected()
	}
	if err != nil {
		return Cart{}, fmt.Errorf("remove cart item: %w", err)
	}
	if affected == 0 {
		return Cart{}, ErrInvalidProduct
	}
	if err = tx.Commit(ctx); err != nil {
		return Cart{}, fmt.Errorf("commit cart removal: %w", err)
	}
	return s.GetCart(ctx, ownerTokenHash)
}

func (s *Service) GetCart(ctx context.Context, ownerTokenHash []byte) (Cart, error) {
	if len(ownerTokenHash) != 32 {
		return Cart{}, errors.New("cart capability digest must be 32 bytes")
	}
	var cart Cart
	err := s.pool.QueryRow(ctx, `SELECT id::text,status,currency,expires_at FROM carts WHERE owner_token_hash=$1`, ownerTokenHash).
		Scan(&cart.ID, &cart.Status, &cart.Currency, &cart.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Cart{}, ErrCartNotFound
	}
	if err != nil {
		return Cart{}, fmt.Errorf("read cart: %w", err)
	}
	rows, err := s.pool.Query(ctx, `SELECT product_id,variant_id,quantity,unit_price_snapshot::text,currency,product_title_snapshot,variant_title_snapshot
		FROM cart_items WHERE cart_id=$1 ORDER BY id`, cart.ID)
	if err != nil {
		return Cart{}, fmt.Errorf("read cart items: %w", err)
	}
	defer rows.Close()
	cart.Items = make([]CartItem, 0)
	for rows.Next() {
		var item CartItem
		var variant pgtype.Int8
		if err := rows.Scan(&item.ProductID, &variant, &item.Quantity, &item.UnitPrice, &item.Currency, &item.ProductTitle, &item.VariantTitle); err != nil {
			return Cart{}, fmt.Errorf("scan cart item: %w", err)
		}
		if variant.Valid {
			v := variant.Int64
			item.VariantID = &v
		}
		cart.Items = append(cart.Items, item)
	}
	if err := rows.Err(); err != nil {
		return Cart{}, fmt.Errorf("iterate cart items: %w", err)
	}
	return cart, nil
}
