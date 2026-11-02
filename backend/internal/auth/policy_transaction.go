package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// AuthorizeSellerVendorTx revalidates the resolved session, account state,
// seller grant, vendor lifecycle and exact membership while holding locks in
// the caller's transaction. The user row is locked first, matching membership
// mutation and bulk-session-revocation order. A seller operation must perform
// its protected read/write in this same transaction after this function
// succeeds; checking a principal before opening a separate transaction leaves
// a revocation race.
func AuthorizeSellerVendorTx(
	ctx context.Context,
	tx pgx.Tx,
	principal SessionPrincipal,
	action AuthorizationAction,
	vendorID int64,
) error {
	if action != AuthorizationSellerReadVendor &&
		action != AuthorizationSellerWriteVendor {
		return ErrForbidden
	}
	if err := Authorize(
		principal,
		action,
		AuthorizationResource{VendorID: vendorID},
	); err != nil {
		return err
	}

	var accountActive bool
	err := tx.QueryRow(ctx,
		"SELECT status = 'active' AND email_verified_at IS NOT NULL\n"+
			"FROM users WHERE id = $1 FOR UPDATE",
		principal.UserID,
	).Scan(&accountActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidSession
		}
		return fmt.Errorf("lock seller account for authorization: %w", err)
	}
	if !accountActive {
		return ErrInvalidSession
	}

	var lockedSessionID string
	err = tx.QueryRow(ctx,
		"SELECT id::text FROM user_sessions\n"+
			"WHERE id = $1 AND user_id = $2\n"+
			"AND revoked_at IS NULL AND idle_expires_at > now() AND expires_at > now()\n"+
			"FOR UPDATE",
		principal.ID, principal.UserID,
	).Scan(&lockedSessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidSession
		}
		return fmt.Errorf("lock seller session for authorization: %w", err)
	}

	var vendorStatus string
	err = tx.QueryRow(ctx,
		"SELECT status FROM vendors WHERE id = $1 FOR UPDATE",
		vendorID,
	).Scan(&vendorStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrForbidden
		}
		return fmt.Errorf("lock seller vendor for authorization: %w", err)
	}
	if vendorStatus != "active" {
		return ErrForbidden
	}

	var activeSellerRole string
	err = tx.QueryRow(ctx,
		"SELECT id::text FROM user_roles\n"+
			"WHERE user_id = $1 AND role_key = 'seller' AND revoked_at IS NULL\n"+
			"ORDER BY granted_at DESC, id DESC LIMIT 1 FOR UPDATE",
		principal.UserID,
	).Scan(&activeSellerRole)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrForbidden
		}
		return fmt.Errorf("lock seller role for authorization: %w", err)
	}

	var membershipID string
	err = tx.QueryRow(ctx,
		"SELECT id::text FROM seller_memberships\n"+
			"WHERE user_id = $1 AND vendor_id = $2 AND status = 'active'\n"+
			"FOR UPDATE",
		principal.UserID, vendorID,
	).Scan(&membershipID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrForbidden
		}
		return fmt.Errorf("lock seller membership for authorization: %w", err)
	}
	return nil
}


// AuthorizeCustomerSessionTx revalidates an account-backed customer session
// under row locks in the caller's transaction. Customer resource ownership
// must still be loaded from PostgreSQL and checked separately with Authorize;
// this helper intentionally does not accept an owner ID from the caller.
func AuthorizeCustomerSessionTx(ctx context.Context, tx pgx.Tx, principal SessionPrincipal) error {
	if err := Authorize(
		principal,
		AuthorizationCustomerReadOwnResource,
		AuthorizationResource{OwnerUserID: principal.UserID},
	); err != nil {
		return err
	}

	var accountActive bool
	err := tx.QueryRow(ctx,
		"SELECT status = 'active' AND email_verified_at IS NOT NULL\n"+
			"FROM users WHERE id = $1 FOR UPDATE",
		principal.UserID,
	).Scan(&accountActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidSession
		}
		return fmt.Errorf("lock customer account for authorization: %w", err)
	}
	if !accountActive {
		return ErrInvalidSession
	}

	var lockedSessionID string
	err = tx.QueryRow(ctx,
		"SELECT id::text FROM user_sessions\n"+
			"WHERE id = $1 AND user_id = $2\n"+
			"AND revoked_at IS NULL AND idle_expires_at > now() AND expires_at > now()\n"+
			"FOR UPDATE",
		principal.ID, principal.UserID,
	).Scan(&lockedSessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidSession
		}
		return fmt.Errorf("lock customer session for authorization: %w", err)
	}

	var activeCustomerRole string
	err = tx.QueryRow(ctx,
		"SELECT id::text FROM user_roles\n"+
			"WHERE user_id = $1 AND role_key = 'customer' AND revoked_at IS NULL\n"+
			"ORDER BY granted_at DESC, id DESC LIMIT 1 FOR UPDATE",
		principal.UserID,
	).Scan(&activeCustomerRole)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrForbidden
		}
		return fmt.Errorf("lock customer role for authorization: %w", err)
	}
	return nil
}
