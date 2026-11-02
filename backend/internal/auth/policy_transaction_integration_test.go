package auth

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func createSellerPrincipalFixture(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	userID string,
) SessionPrincipal {
	t.Helper()
	issued, err := CreateSession(ctx, pool, userID, integrationPasswordHash)
	if err != nil {
		t.Fatalf("create seller fixture session: %v", err)
	}
	principal, err := ResolveSession(ctx, pool, issued.SessionToken)
	if err != nil {
		t.Fatalf("resolve seller fixture session: %v", err)
	}
	return principal
}

func TestAuthorizeSellerVendorTxRevalidatesMembershipAndSession(t *testing.T) {
	pool, ctx := openAuthIntegrationPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	adminID := createActiveAuthUser(t, ctx, pool, suffix+"-policy-admin")
	sellerID := createActiveAuthUser(t, ctx, pool, suffix+"-policy-seller")
	grantAdminRoleFixture(t, ctx, pool, adminID)
	vendorID := createActiveSellerVendorFixture(t, pool, ctx, suffix+"-policy-vendor")

	request, err := RequestSellerMembership(ctx, pool, sellerID, vendorID)
	if err != nil {
		t.Fatalf("request seller membership: %v", err)
	}
	if _, err := ActivateSellerMembership(ctx, pool, adminID, request.ID); err != nil {
		t.Fatalf("activate seller membership: %v", err)
	}
	principal := createSellerPrincipalFixture(t, ctx, pool, sellerID)

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		t.Fatalf("begin active-scope authorization: %v", err)
	}
	if err := AuthorizeSellerVendorTx(ctx, tx, principal, AuthorizationSellerWriteVendor, vendorID); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("authorize current seller scope: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit active-scope authorization: %v", err)
	}

	// Simulate a state change after ResolveSession returned the in-memory
	// principal. Transaction-time checks must reject the suspended membership.
	if _, err := pool.Exec(ctx, `
		UPDATE seller_memberships
		SET status = 'suspended', reviewed_by_user_id = $1, updated_at = now()
		WHERE user_id = $2 AND vendor_id = $3
	`, adminID, sellerID, vendorID); err != nil {
		t.Fatalf("suspend membership without touching fixture session: %v", err)
	}
	tx, err = pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		t.Fatalf("begin stale-membership authorization: %v", err)
	}
	err = AuthorizeSellerVendorTx(ctx, tx, principal, AuthorizationSellerWriteVendor, vendorID)
	_ = tx.Rollback(ctx)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("stale membership authorization error = %v, want ErrForbidden", err)
	}

	// Restore the membership, then revoke the session but retain the old
	// principal. The guard must reject the revoked database session.
	if _, err := pool.Exec(ctx, `
		UPDATE seller_memberships
		SET status = 'active', activated_at = now(), updated_at = now()
		WHERE user_id = $1 AND vendor_id = $2
	`, sellerID, vendorID); err != nil {
		t.Fatalf("restore seller membership fixture: %v", err)
	}
	revokeTx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		t.Fatalf("begin seller fixture session revocation: %v", err)
	}
	if _, err := RevokeUserSessions(ctx, revokeTx, sellerID, &adminID, SessionRevocationPrivilegeChanged); err != nil {
		_ = revokeTx.Rollback(ctx)
		t.Fatalf("revoke seller fixture session: %v", err)
	}
	if err := revokeTx.Commit(ctx); err != nil {
		t.Fatalf("commit seller fixture session revocation: %v", err)
	}
	tx, err = pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		t.Fatalf("begin revoked-session authorization: %v", err)
	}
	err = AuthorizeSellerVendorTx(ctx, tx, principal, AuthorizationSellerWriteVendor, vendorID)
	_ = tx.Rollback(ctx)
	if !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("revoked session authorization error = %v, want ErrInvalidSession", err)
	}
}

func TestSellerAuthorizationSerializesWithMembershipRevocation(t *testing.T) {
	pool, ctx := openAuthIntegrationPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	adminID := createActiveAuthUser(t, ctx, pool, suffix+"-policy-race-admin")
	sellerID := createActiveAuthUser(t, ctx, pool, suffix+"-policy-race-seller")
	grantAdminRoleFixture(t, ctx, pool, adminID)
	vendorID := createActiveSellerVendorFixture(t, pool, ctx, suffix+"-policy-race-vendor")
	request, err := RequestSellerMembership(ctx, pool, sellerID, vendorID)
	if err != nil {
		t.Fatalf("request seller race membership: %v", err)
	}
	if _, err := ActivateSellerMembership(ctx, pool, adminID, request.ID); err != nil {
		t.Fatalf("activate seller race membership: %v", err)
	}
	principal := createSellerPrincipalFixture(t, ctx, pool, sellerID)

	// Acquire the same account/session/vendor/membership locks that a protected
	// seller service must hold through its mutation.
	protectedTx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		t.Fatalf("begin protected seller operation: %v", err)
	}
	if err := AuthorizeSellerVendorTx(
		ctx, protectedTx, principal, AuthorizationSellerWriteVendor, vendorID,
	); err != nil {
		_ = protectedTx.Rollback(ctx)
		t.Fatalf("authorize protected seller operation: %v", err)
	}
	if _, err := protectedTx.Exec(ctx, `
		INSERT INTO audit_events(
			event_type, actor_user_id, subject_user_id,
			resource_type, resource_id, metadata
		)
		VALUES (
			'auth.seller_scope.test_operation', $1, $1, 'vendor', $2::bigint::text,
			'{"origin":"integration_test"}'::jsonb
		)
	`, sellerID, vendorID); err != nil {
		_ = protectedTx.Rollback(ctx)
		t.Fatalf("record protected operation fixture: %v", err)
	}

	revokeCtx, cancel := context.WithTimeout(ctx, 400*time.Millisecond)
	defer cancel()
	revokeResult := make(chan error, 1)
	go func() {
		_, revokeErr := RevokeSellerMembership(revokeCtx, pool, adminID, request.ID)
		revokeResult <- revokeErr
	}()
	select {
	case revokeErr := <-revokeResult:
		if !errors.Is(revokeErr, context.DeadlineExceeded) {
			_ = protectedTx.Rollback(ctx)
			t.Fatalf("membership revocation while seller transaction held locks = %v, want deadline exceeded", revokeErr)
		}
	case <-time.After(2 * time.Second):
		_ = protectedTx.Rollback(ctx)
		t.Fatal("membership revocation did not finish within the bounded lock-wait test")
	}

	if err := protectedTx.Commit(ctx); err != nil {
		t.Fatalf("commit authorized seller operation before revocation: %v", err)
	}
	if _, err := RevokeSellerMembership(ctx, pool, adminID, request.ID); err != nil {
		t.Fatalf("revoke seller membership after protected operation committed: %v", err)
	}

	var operationEvents int
	if err := pool.QueryRow(ctx, `
		SELECT count(*)::int
		FROM audit_events
		WHERE subject_user_id = $1
			AND event_type = 'auth.seller_scope.test_operation'
	`, sellerID).Scan(&operationEvents); err != nil {
		t.Fatalf("count protected seller operation: %v", err)
	}
	if operationEvents != 1 {
		t.Fatalf("protected operation events = %d, want 1", operationEvents)
	}
	if err := authorizeSellerVendorWithPool(ctx, pool, principal, AuthorizationSellerWriteVendor, vendorID); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("stale session after membership revocation error = %v, want ErrInvalidSession", err)
	}
}

func authorizeSellerVendorWithPool(
	ctx context.Context,
	pool *pgxpool.Pool,
	principal SessionPrincipal,
	action AuthorizationAction,
	vendorID int64,
) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := AuthorizeSellerVendorTx(ctx, tx, principal, action, vendorID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func TestAuthorizeAdminTxRevalidatesRoleAndSessionRevocation(t *testing.T) {
	pool, ctx := openAuthIntegrationPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	adminID := createActiveAuthUser(t, ctx, pool, suffix+"-overview-admin")
	grantAdminRoleFixture(t, ctx, pool, adminID)

	issued, err := CreateSession(ctx, pool, adminID, integrationPasswordHash)
	if err != nil {
		t.Fatalf("create admin overview session: %v", err)
	}
	principal, err := ResolveSession(ctx, pool, issued.SessionToken)
	if err != nil {
		t.Fatalf("resolve admin overview session: %v", err)
	}

	// Keep the session active but revoke the role after principal resolution.
	if _, err := pool.Exec(ctx, `
		UPDATE user_roles
		SET revoked_at = now()
		WHERE user_id = $1 AND role_key = 'admin' AND revoked_at IS NULL
	`, adminID); err != nil {
		t.Fatalf("revoke admin role fixture: %v", err)
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin stale admin-role authorization: %v", err)
	}
	err = AuthorizeAdminTx(ctx, tx, principal, AuthorizationAdminReadOverview)
	_ = tx.Rollback(ctx)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("stale admin-role authorization error = %v, want ErrForbidden", err)
	}

	// Regrant the role, resolve a fresh principal, then revoke only its session.
	if _, err := pool.Exec(ctx, `
		INSERT INTO user_roles(user_id, role_key) VALUES($1, 'admin')
	`, adminID); err != nil {
		t.Fatalf("regrant admin fixture role: %v", err)
	}
	principal, err = ResolveSession(ctx, pool, issued.SessionToken)
	if err != nil {
		t.Fatalf("resolve admin overview session after regrant: %v", err)
	}
	if err := RevokeSession(ctx, pool, issued.SessionToken); err != nil {
		t.Fatalf("revoke admin overview session: %v", err)
	}
	tx, err = pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin stale admin-session authorization: %v", err)
	}
	err = AuthorizeAdminTx(ctx, tx, principal, AuthorizationAdminReadOverview)
	_ = tx.Rollback(ctx)
	if !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("stale admin-session authorization error = %v, want ErrInvalidSession", err)
	}
}
