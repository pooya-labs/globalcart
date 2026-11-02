package auth

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func createActiveSellerVendorFixture(t *testing.T, pool *pgxpool.Pool, ctx context.Context, suffix string) int64 {
	t.Helper()
	var vendorID int64
	err := pool.QueryRow(ctx, "INSERT INTO vendors(slug, display_name, status) VALUES ($1, 'Seller Membership Fixture', 'active') RETURNING id", "seller-membership-"+suffix).Scan(&vendorID)
	if err != nil {
		t.Fatalf("create active seller vendor: %v", err)
	}
	return vendorID
}

func TestSellerMembershipLifecycleMaintainsLiveVendorScopes(t *testing.T) {
	pool, ctx := openAuthIntegrationPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	adminID := createActiveAuthUser(t, ctx, pool, suffix+"-admin")
	userID := createActiveAuthUser(t, ctx, pool, suffix+"-member")
	grantAdminRoleFixture(t, ctx, pool, adminID)
	vendorOne := createActiveSellerVendorFixture(t, pool, ctx, suffix+"-one")
	vendorTwo := createActiveSellerVendorFixture(t, pool, ctx, suffix+"-two")

	first, err := RequestSellerMembership(ctx, pool, userID, vendorOne)
	if err != nil {
		t.Fatalf("request first seller membership: %v", err)
	}
	second, err := RequestSellerMembership(ctx, pool, userID, vendorTwo)
	if err != nil {
		t.Fatalf("request second seller membership: %v", err)
	}
	if first.Status != "pending" || second.Status != "pending" || first.ID == second.ID {
		t.Fatalf("membership requests = %+v / %+v, want distinct pending rows", first, second)
	}
	if _, err := RequestSellerMembership(ctx, pool, userID, vendorOne); !errors.Is(err, ErrSellerMembershipAlreadyExists) {
		t.Fatalf("duplicate membership request error = %v, want ErrSellerMembershipAlreadyExists", err)
	}

	var sellerRoleBeforeApproval bool
	if err := pool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM user_roles WHERE user_id = $1 AND role_key = 'seller' AND revoked_at IS NULL)", userID).Scan(&sellerRoleBeforeApproval); err != nil {
		t.Fatalf("check seller role before approval: %v", err)
	}
	if sellerRoleBeforeApproval {
		t.Fatal("pending membership request granted seller role")
	}

	beforeApproval, err := CreateSession(ctx, pool, userID, integrationPasswordHash)
	if err != nil {
		t.Fatalf("create session before membership approval: %v", err)
	}
	revoked, err := ActivateSellerMembership(ctx, pool, adminID, first.ID)
	if err != nil {
		t.Fatalf("activate first seller membership: %v", err)
	}
	if revoked != 1 {
		t.Fatalf("first activation revoked %d sessions, want 1", revoked)
	}
	if _, err := ResolveSession(ctx, pool, beforeApproval.SessionToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("pre-approval session error = %v, want ErrInvalidSession", err)
	}

	firstScopeSession, err := CreateSession(ctx, pool, userID, integrationPasswordHash)
	if err != nil {
		t.Fatalf("create session after first membership activation: %v", err)
	}
	principal, err := ResolveSession(ctx, pool, firstScopeSession.SessionToken)
	if err != nil {
		t.Fatalf("resolve first seller scope: %v", err)
	}
	if !slices.Contains(principal.Roles, "seller") || !slices.Equal(principal.SellerVendorIDs, []int64{vendorOne}) {
		t.Fatalf("first seller principal roles=%v vendors=%v, want seller/%v",
			principal.Roles, principal.SellerVendorIDs, []int64{vendorOne})
	}

	revoked, err = ActivateSellerMembership(ctx, pool, adminID, second.ID)
	if err != nil {
		t.Fatalf("activate second seller membership: %v", err)
	}
	if revoked != 1 {
		t.Fatalf("second activation revoked %d sessions, want 1", revoked)
	}
	if _, err := ResolveSession(ctx, pool, firstScopeSession.SessionToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("session from first membership scope error = %v, want ErrInvalidSession", err)
	}

	bothScopesSession, err := CreateSession(ctx, pool, userID, integrationPasswordHash)
	if err != nil {
		t.Fatalf("create session after second membership activation: %v", err)
	}
	principal, err = ResolveSession(ctx, pool, bothScopesSession.SessionToken)
	if err != nil {
		t.Fatalf("resolve both seller scopes: %v", err)
	}
	if !slices.Contains(principal.Roles, "seller") ||
		!slices.Equal(principal.SellerVendorIDs, []int64{vendorOne, vendorTwo}) {
		t.Fatalf("two-vendor principal roles=%v vendors=%v, want seller/%v",
			principal.Roles, principal.SellerVendorIDs, []int64{vendorOne, vendorTwo})
	}

	revoked, err = SuspendSellerMembership(ctx, pool, adminID, first.ID)
	if err != nil {
		t.Fatalf("suspend first seller membership: %v", err)
	}
	if revoked != 1 {
		t.Fatalf("membership suspension revoked %d sessions, want 1", revoked)
	}
	if _, err := ResolveSession(ctx, pool, bothScopesSession.SessionToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("pre-suspension session error = %v, want ErrInvalidSession", err)
	}

	remainingScopeSession, err := CreateSession(ctx, pool, userID, integrationPasswordHash)
	if err != nil {
		t.Fatalf("create session with remaining seller scope: %v", err)
	}
	principal, err = ResolveSession(ctx, pool, remainingScopeSession.SessionToken)
	if err != nil {
		t.Fatalf("resolve remaining seller scope: %v", err)
	}
	if !slices.Contains(principal.Roles, "seller") || !slices.Equal(principal.SellerVendorIDs, []int64{vendorTwo}) {
		t.Fatalf("remaining seller principal roles=%v vendors=%v, want seller/%v",
			principal.Roles, principal.SellerVendorIDs, []int64{vendorTwo})
	}

	revoked, err = RevokeSellerMembership(ctx, pool, adminID, second.ID)
	if err != nil {
		t.Fatalf("revoke second seller membership: %v", err)
	}
	if revoked != 1 {
		t.Fatalf("membership revocation revoked %d sessions, want 1", revoked)
	}
	if _, err := ResolveSession(ctx, pool, remainingScopeSession.SessionToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("pre-revocation session error = %v, want ErrInvalidSession", err)
	}

	// Once no active vendor scope remains, the seller role itself is revoked.
	var activeSellerRole bool
	if err := pool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM user_roles WHERE user_id = $1 AND role_key = 'seller' AND revoked_at IS NULL)", userID).Scan(&activeSellerRole); err != nil {
		t.Fatalf("check seller role after final membership removal: %v", err)
	}
	if activeSellerRole {
		t.Fatal("seller role remained active after the final vendor scope was removed")
	}
	finalSession, err := CreateSession(ctx, pool, userID, integrationPasswordHash)
	if err != nil {
		t.Fatalf("create customer-only session after membership removal: %v", err)
	}
	principal, err = ResolveSession(ctx, pool, finalSession.SessionToken)
	if err != nil {
		t.Fatalf("resolve customer-only session: %v", err)
	}
	if slices.Contains(principal.Roles, "seller") || len(principal.SellerVendorIDs) != 0 ||
		!slices.Equal(principal.Roles, []string{"customer"}) {
		t.Fatalf("final principal roles=%v vendors=%v, want customer only/no vendor scopes",
			principal.Roles, principal.SellerVendorIDs)
	}

	var requested, activated, suspended, membershipRevoked, roleGranted, roleRevoked, sessionRevoked int
	if err := pool.QueryRow(ctx, "SELECT count(*) FILTER (WHERE event_type = 'auth.seller_membership.requested')::int, count(*) FILTER (WHERE event_type = 'auth.seller_membership.activated')::int, count(*) FILTER (WHERE event_type = 'auth.seller_membership.suspended')::int, count(*) FILTER (WHERE event_type = 'auth.seller_membership.revoked')::int, count(*) FILTER (WHERE event_type = 'auth.role.granted')::int, count(*) FILTER (WHERE event_type = 'auth.role.revoked')::int, count(*) FILTER (WHERE event_type = 'auth.session.revoked')::int FROM audit_events WHERE subject_user_id = $1", userID).Scan(&requested, &activated, &suspended, &membershipRevoked, &roleGranted, &roleRevoked, &sessionRevoked); err != nil {
		t.Fatalf("read seller membership audit events: %v", err)
	}
	if requested != 2 || activated != 2 || suspended != 1 || membershipRevoked != 1 ||
		roleGranted != 1 || roleRevoked != 1 || sessionRevoked != 4 {
		t.Fatalf("audit counts request/activate/suspend/revoke/role+/role-/session=%d/%d/%d/%d/%d/%d/%d, want 2/2/1/1/1/1/4",
			requested, activated, suspended, membershipRevoked, roleGranted, roleRevoked, sessionRevoked)
	}
}

func TestSellerMembershipAuthorizationAndDisabledTargetRevocation(t *testing.T) {
	pool, ctx := openAuthIntegrationPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	adminID := createActiveAuthUser(t, ctx, pool, suffix+"-admin")
	ordinaryAdminCandidate := createActiveAuthUser(t, ctx, pool, suffix+"-ordinary")
	targetID := createActiveAuthUser(t, ctx, pool, suffix+"-target")
	grantAdminRoleFixture(t, ctx, pool, adminID)
	vendorID := createActiveSellerVendorFixture(t, pool, ctx, suffix+"-auth")
	membership, err := RequestSellerMembership(ctx, pool, targetID, vendorID)
	if err != nil {
		t.Fatalf("request membership: %v", err)
	}
	if _, err := ActivateSellerMembership(ctx, pool, ordinaryAdminCandidate, membership.ID); !errors.Is(err, ErrAdminAuthorizationRequired) {
		t.Fatalf("non-admin membership activation error = %v, want ErrAdminAuthorizationRequired", err)
	}
	if _, err := ActivateSellerMembership(ctx, pool, adminID, membership.ID); err != nil {
		t.Fatalf("activate membership for disabled-target fixture: %v", err)
	}
	if _, err := pool.Exec(ctx, "UPDATE users SET status = 'disabled', disabled_at = now(), updated_at = now() WHERE id = $1", targetID); err != nil {
		t.Fatalf("disable seller account: %v", err)
	}
	if _, err := RevokeSellerMembership(ctx, pool, adminID, membership.ID); err != nil {
		t.Fatalf("admin must be able to revoke membership for a disabled account: %v", err)
	}
	var status string
	if err := pool.QueryRow(ctx, "SELECT status FROM seller_memberships WHERE id = $1", membership.ID).Scan(&status); err != nil {
		t.Fatalf("read disabled seller membership status: %v", err)
	}
	if status != "revoked" {
		t.Fatalf("disabled seller membership status = %q, want revoked", status)
	}
}

func TestConcurrentSellerMembershipActivationIsSingleWinner(t *testing.T) {
	pool, ctx := openAuthIntegrationPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	firstAdmin := createActiveAuthUser(t, ctx, pool, suffix+"-admin-one")
	secondAdmin := createActiveAuthUser(t, ctx, pool, suffix+"-admin-two")
	userID := createActiveAuthUser(t, ctx, pool, suffix+"-member")
	grantAdminRoleFixture(t, ctx, pool, firstAdmin)
	grantAdminRoleFixture(t, ctx, pool, secondAdmin)
	vendorID := createActiveSellerVendorFixture(t, pool, ctx, suffix+"-race")
	membership, err := RequestSellerMembership(ctx, pool, userID, vendorID)
	if err != nil {
		t.Fatalf("request membership for concurrent activation: %v", err)
	}

	start := make(chan struct{})
	type result struct {
		actorID string
		err     error
	}
	results := make(chan result, 2)
	for _, actorID := range []string{firstAdmin, secondAdmin} {
		actorID := actorID
		go func() {
			<-start
			_, activateErr := ActivateSellerMembership(ctx, pool, actorID, membership.ID)
			results <- result{actorID: actorID, err: activateErr}
		}()
	}
	close(start)
	var successes, alreadyActivated int
	for i := 0; i < 2; i++ {
		got := <-results
		switch {
		case got.err == nil:
			successes++
		case errors.Is(got.err, ErrSellerMembershipNotPending):
			alreadyActivated++
		default:
			t.Fatalf("concurrent membership activation by %s returned unexpected error: %v", got.actorID, got.err)
		}
	}
	if successes != 1 || alreadyActivated != 1 {
		t.Fatalf("concurrent activation outcomes successes=%d already-activated=%d, want 1/1", successes, alreadyActivated)
	}

	var activeMemberships, activationEvents int
	if err := pool.QueryRow(ctx, "SELECT (SELECT count(*)::int FROM seller_memberships WHERE id = $1 AND status = 'active'), (SELECT count(*)::int FROM audit_events WHERE subject_user_id = $2 AND event_type = 'auth.seller_membership.activated')", membership.ID, userID).Scan(&activeMemberships, &activationEvents); err != nil {
		t.Fatalf("read concurrent seller membership result: %v", err)
	}
	if activeMemberships != 1 || activationEvents != 1 {
		t.Fatalf("concurrent activation state membership=%d audit=%d, want 1/1", activeMemberships, activationEvents)
	}
}

func TestSellerMembershipRejectsInactiveVendorAndRequester(t *testing.T) {
	pool, ctx := openAuthIntegrationPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	userID := createActiveAuthUser(t, ctx, pool, suffix+"-requester")
	var pendingVendorID int64
	if err := pool.QueryRow(ctx, "INSERT INTO vendors(slug, display_name, status) VALUES ($1, 'Pending Seller Vendor', 'pending') RETURNING id", "seller-membership-pending-"+suffix).Scan(&pendingVendorID); err != nil {
		t.Fatalf("create pending vendor: %v", err)
	}
	if _, err := RequestSellerMembership(ctx, pool, userID, pendingVendorID); !errors.Is(err, ErrSellerVendorInactive) {
		t.Fatalf("pending vendor request error = %v, want ErrSellerVendorInactive", err)
	}
	if _, err := pool.Exec(ctx, "UPDATE users SET status = 'disabled', disabled_at = now(), updated_at = now() WHERE id = $1", userID); err != nil {
		t.Fatalf("disable requester: %v", err)
	}
	var activeVendorID int64
	if err := pool.QueryRow(ctx, "INSERT INTO vendors(slug, display_name, status) VALUES ($1, 'Active Seller Vendor', 'active') RETURNING id", "seller-membership-active-"+suffix).Scan(&activeVendorID); err != nil {
		t.Fatalf("create active vendor: %v", err)
	}
	if _, err := RequestSellerMembership(ctx, pool, userID, activeVendorID); !errors.Is(err, ErrAccountNotActive) {
		t.Fatalf("disabled requester membership error = %v, want ErrAccountNotActive", err)
	}
}
