package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrSellerMembershipNotFound      = errors.New("seller membership not found")
	ErrSellerMembershipAlreadyExists = errors.New("seller membership already exists")
	ErrSellerMembershipNotPending    = errors.New("seller membership is not pending")
	ErrSellerMembershipNotActive     = errors.New("seller membership is not active")
	ErrSellerVendorNotFound          = errors.New("seller vendor not found")
	ErrSellerVendorInactive          = errors.New("seller vendor is not active")
)

type SellerMembership struct {
	ID       string
	UserID   string
	VendorID int64
	Status   string
}

// RequestSellerMembership creates a pending request for an active, verified
// account. requesterUserID must come from a verified server-side session.
// No seller role or vendor scope is granted until an administrator activates it.
func RequestSellerMembership(ctx context.Context, pool *pgxpool.Pool, requesterUserID string, vendorID int64) (SellerMembership, error) {
	var membership SellerMembership
	if vendorID <= 0 {
		return membership, ErrSellerVendorNotFound
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return membership, fmt.Errorf("begin seller membership request: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var userStatus string
	var verified bool
	if err := tx.QueryRow(ctx, "SELECT status, email_verified_at IS NOT NULL FROM users WHERE id = $1 FOR UPDATE", requesterUserID).Scan(&userStatus, &verified); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return membership, ErrUserNotFound
		}
		return membership, fmt.Errorf("lock membership requester: %w", err)
	}
	if userStatus != "active" || !verified {
		return membership, ErrAccountNotActive
	}

	var vendorStatus string
	if err := tx.QueryRow(ctx, "SELECT status FROM vendors WHERE id = $1 FOR UPDATE", vendorID).Scan(&vendorStatus); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return membership, ErrSellerVendorNotFound
		}
		return membership, fmt.Errorf("lock membership vendor: %w", err)
	}
	if vendorStatus != "active" {
		return membership, ErrSellerVendorInactive
	}

	var existingID, existingStatus string
	err = tx.QueryRow(ctx, "SELECT id::text, status FROM seller_memberships WHERE user_id = $1 AND vendor_id = $2 FOR UPDATE", requesterUserID, vendorID).Scan(&existingID, &existingStatus)
	if err == nil {
		if existingStatus != "revoked" {
			return membership, ErrSellerMembershipAlreadyExists
		}
		err = tx.QueryRow(ctx, "UPDATE seller_memberships SET status = 'pending', reviewed_by_user_id = NULL, activated_at = NULL, updated_at = now() WHERE id = $1 RETURNING id::text, user_id::text, vendor_id, status", existingID).Scan(&membership.ID, &membership.UserID, &membership.VendorID, &membership.Status)
	} else if errors.Is(err, pgx.ErrNoRows) {
		err = tx.QueryRow(ctx, "INSERT INTO seller_memberships(user_id, vendor_id, status) VALUES ($1, $2, 'pending') RETURNING id::text, user_id::text, vendor_id, status", requesterUserID, vendorID).Scan(&membership.ID, &membership.UserID, &membership.VendorID, &membership.Status)
	}
	if err != nil {
		return SellerMembership{}, fmt.Errorf("create seller membership request: %w", err)
	}

	if _, err := tx.Exec(ctx, "INSERT INTO audit_events(event_type, actor_user_id, subject_user_id, resource_type, resource_id, metadata) VALUES ('auth.seller_membership.requested', $1, $1, 'seller_membership', $2, jsonb_build_object('vendor_id', $3::bigint, 'status', 'pending', 'origin', 'internal'))", requesterUserID, membership.ID, vendorID); err != nil {
		return SellerMembership{}, fmt.Errorf("audit seller membership request: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return SellerMembership{}, fmt.Errorf("commit seller membership request: %w", err)
	}
	return membership, nil
}

func findSellerMembershipOwner(ctx context.Context, pool *pgxpool.Pool, membershipID string) (string, error) {
	var userID string
	if err := pool.QueryRow(ctx, "SELECT user_id::text FROM seller_memberships WHERE id = $1", membershipID).Scan(&userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrSellerMembershipNotFound
		}
		return "", fmt.Errorf("find seller membership owner: %w", err)
	}
	return userID, nil
}

func lockSellerMembership(ctx context.Context, tx pgx.Tx, membershipID, userID string) (SellerMembership, error) {
	var m SellerMembership
	err := tx.QueryRow(ctx, "SELECT id::text, user_id::text, vendor_id, status FROM seller_memberships WHERE id = $1 AND user_id = $2 FOR UPDATE", membershipID, userID).Scan(&m.ID, &m.UserID, &m.VendorID, &m.Status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SellerMembership{}, ErrSellerMembershipNotFound
		}
		return SellerMembership{}, fmt.Errorf("lock seller membership: %w", err)
	}
	return m, nil
}

func checkSellerVendorActive(ctx context.Context, tx pgx.Tx, vendorID int64) error {
	var status string
	if err := tx.QueryRow(ctx, "SELECT status FROM vendors WHERE id = $1 FOR UPDATE", vendorID).Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrSellerVendorNotFound
		}
		return fmt.Errorf("lock seller vendor: %w", err)
	}
	if status != "active" {
		return ErrSellerVendorInactive
	}
	return nil
}

func grantSellerRoleIfMissing(ctx context.Context, tx pgx.Tx, actorID, targetID string) error {
	var roleID string
	err := tx.QueryRow(ctx, "INSERT INTO user_roles(user_id, role_key, granted_by_user_id) VALUES ($1, 'seller', $2) ON CONFLICT (user_id, role_key) WHERE revoked_at IS NULL DO NOTHING RETURNING id::text", targetID, actorID).Scan(&roleID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("grant seller role: %w", err)
	}
	if _, err := tx.Exec(ctx, "INSERT INTO audit_events(event_type, actor_user_id, subject_user_id, resource_type, resource_id, metadata) VALUES ('auth.role.granted', $1, $2, 'user_role', $3, '{\"role\":\"seller\",\"origin\":\"internal\"}'::jsonb)", actorID, targetID, roleID); err != nil {
		return fmt.Errorf("audit seller role grant: %w", err)
	}
	return nil
}

// revokeSellerRoleIfUnscoped retains seller when any other active vendor scope
// remains; otherwise it revokes the role and records a separate audit event.
func revokeSellerRoleIfUnscoped(ctx context.Context, tx pgx.Tx, actorID, targetID string) error {
	var hasScope bool
	if err := tx.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM seller_memberships sm JOIN vendors v ON v.id = sm.vendor_id WHERE sm.user_id = $1 AND sm.status = 'active' AND v.status = 'active')", targetID).Scan(&hasScope); err != nil {
		return fmt.Errorf("check remaining seller scopes: %w", err)
	}
	if hasScope {
		return nil
	}
	var roleID string
	err := tx.QueryRow(ctx, "UPDATE user_roles SET revoked_at = now() WHERE user_id = $1 AND role_key = 'seller' AND revoked_at IS NULL RETURNING id::text", targetID).Scan(&roleID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("revoke seller role: %w", err)
	}
	if _, err := tx.Exec(ctx, "INSERT INTO audit_events(event_type, actor_user_id, subject_user_id, resource_type, resource_id, metadata) VALUES ('auth.role.revoked', $1, $2, 'user_role', $3, '{\"role\":\"seller\",\"reason\":\"no_active_membership\",\"origin\":\"internal\"}'::jsonb)", actorID, targetID, roleID); err != nil {
		return fmt.Errorf("audit seller role revocation: %w", err)
	}
	return nil
}

func auditSellerMembershipChange(ctx context.Context, tx pgx.Tx, eventType, actorID string, m SellerMembership) error {
	_, err := tx.Exec(ctx, "INSERT INTO audit_events(event_type, actor_user_id, subject_user_id, resource_type, resource_id, metadata) VALUES ($1, $2, $3, 'seller_membership', $4, jsonb_build_object('vendor_id', $5::bigint, 'status', $6::text, 'origin', 'internal'))", eventType, actorID, m.UserID, m.ID, m.VendorID, m.Status)
	if err != nil {
		return fmt.Errorf("audit seller membership change: %w", err)
	}
	return nil
}

func beginSellerMembershipAdminMutation(ctx context.Context, pool *pgxpool.Pool, actorID, membershipID string, requireActiveTarget bool) (pgx.Tx, SellerMembership, error) {
	targetID, err := findSellerMembershipOwner(ctx, pool, membershipID)
	if err != nil {
		return nil, SellerMembership{}, err
	}
	tx, _, target, err := beginAdminRoleChange(ctx, pool, actorID, targetID)
	if err != nil {
		return nil, SellerMembership{}, err
	}
	if requireActiveTarget && (target.status != "active" || !target.emailVerified) {
		_ = tx.Rollback(ctx)
		return nil, SellerMembership{}, ErrAccountNotActive
	}
	m, err := lockSellerMembership(ctx, tx, membershipID, targetID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return nil, SellerMembership{}, err
	}
	return tx, m, nil
}

// ActivateSellerMembership approves a pending request. Seller role grant,
// session invalidation and both audit records commit atomically.
func ActivateSellerMembership(ctx context.Context, pool *pgxpool.Pool, actorID, membershipID string) (int64, error) {
	tx, m, err := beginSellerMembershipAdminMutation(ctx, pool, actorID, membershipID, true)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if m.Status != "pending" {
		return 0, ErrSellerMembershipNotPending
	}
	if err := checkSellerVendorActive(ctx, tx, m.VendorID); err != nil {
		return 0, err
	}
	if _, err := tx.Exec(ctx, "UPDATE seller_memberships SET status = 'active', reviewed_by_user_id = $1, activated_at = now(), updated_at = now() WHERE id = $2", actorID, m.ID); err != nil {
		return 0, fmt.Errorf("activate seller membership: %w", err)
	}
	m.Status = "active"
	if err := grantSellerRoleIfMissing(ctx, tx, actorID, m.UserID); err != nil {
		return 0, err
	}
	if err := auditSellerMembershipChange(ctx, tx, "auth.seller_membership.activated", actorID, m); err != nil {
		return 0, err
	}
	actor := actorID
	revoked, err := RevokeUserSessions(ctx, tx, m.UserID, &actor, SessionRevocationPrivilegeChanged)
	if err != nil {
		return 0, fmt.Errorf("revoke sessions after seller activation: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit seller membership activation: %w", err)
	}
	return revoked, nil
}

// SuspendSellerMembership removes one active vendor scope. The seller role is
// kept only when another active membership on an active vendor remains.
func SuspendSellerMembership(ctx context.Context, pool *pgxpool.Pool, actorID, membershipID string) (int64, error) {
	tx, m, err := beginSellerMembershipAdminMutation(ctx, pool, actorID, membershipID, false)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if m.Status != "active" {
		return 0, ErrSellerMembershipNotActive
	}
	// A vendor may already be suspended; that must not prevent its membership
	// from being suspended. The row is still locked to serialize vendor state.
	var vendorStatus string
	if err := tx.QueryRow(ctx, "SELECT status FROM vendors WHERE id = $1 FOR UPDATE", m.VendorID).Scan(&vendorStatus); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrSellerVendorNotFound
		}
		return 0, fmt.Errorf("lock seller vendor for suspension: %w", err)
	}
	if _, err := tx.Exec(ctx, "UPDATE seller_memberships SET status = 'suspended', reviewed_by_user_id = $1, updated_at = now() WHERE id = $2", actorID, m.ID); err != nil {
		return 0, fmt.Errorf("suspend seller membership: %w", err)
	}
	m.Status = "suspended"
	if err := revokeSellerRoleIfUnscoped(ctx, tx, actorID, m.UserID); err != nil {
		return 0, err
	}
	if err := auditSellerMembershipChange(ctx, tx, "auth.seller_membership.suspended", actorID, m); err != nil {
		return 0, err
	}
	actor := actorID
	revoked, err := RevokeUserSessions(ctx, tx, m.UserID, &actor, SessionRevocationPrivilegeChanged)
	if err != nil {
		return 0, fmt.Errorf("revoke sessions after seller suspension: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit seller membership suspension: %w", err)
	}
	return revoked, nil
}

// RevokeSellerMembership terminates a pending, active or suspended membership.
// Seller role is revoked only when no other active vendor scope remains.
func RevokeSellerMembership(ctx context.Context, pool *pgxpool.Pool, actorID, membershipID string) (int64, error) {
	tx, m, err := beginSellerMembershipAdminMutation(ctx, pool, actorID, membershipID, false)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if m.Status == "revoked" {
		return 0, ErrSellerMembershipNotActive
	}
	if _, err := tx.Exec(ctx, "UPDATE seller_memberships SET status = 'revoked', reviewed_by_user_id = $1, updated_at = now() WHERE id = $2", actorID, m.ID); err != nil {
		return 0, fmt.Errorf("revoke seller membership: %w", err)
	}
	m.Status = "revoked"
	if err := revokeSellerRoleIfUnscoped(ctx, tx, actorID, m.UserID); err != nil {
		return 0, err
	}
	if err := auditSellerMembershipChange(ctx, tx, "auth.seller_membership.revoked", actorID, m); err != nil {
		return 0, err
	}
	actor := actorID
	revoked, err := RevokeUserSessions(ctx, tx, m.UserID, &actor, SessionRevocationPrivilegeChanged)
	if err != nil {
		return 0, fmt.Errorf("revoke sessions after seller membership revocation: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit seller membership revocation: %w", err)
	}
	return revoked, nil
}
