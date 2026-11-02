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
	if errors.Is(err, pgx.ErrNoRows) || !ownerUserID.Valid {
		return Cart{}, ErrCartNotFound
	}
	if err != nil {
		return Cart{}, fmt.Errorf("read customer cart owner: %w", err)
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
