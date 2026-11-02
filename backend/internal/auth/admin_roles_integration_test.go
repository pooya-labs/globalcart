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

func grantAdminRoleFixture(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	userID string,
) {
	t.Helper()
	if _, err := pool.Exec(ctx, `
		INSERT INTO user_roles(user_id, role_key)
		VALUES ($1, 'admin')
	`, userID); err != nil {
		t.Fatalf("grant administrator fixture role: %v", err)
	}
}

func TestAdminRoleGrantAndRevokeInvalidateSessionsAndAudit(t *testing.T) {
	pool, ctx := openAuthIntegrationPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	actorID := createActiveAuthUser(t, ctx, pool, suffix+"-admin-actor")
	targetID := createActiveAuthUser(t, ctx, pool, suffix+"-admin-target")
	grantAdminRoleFixture(t, ctx, pool, actorID)

	beforeGrant := make([]IssuedSession, 0, 2)
	for i := 0; i < 2; i++ {
		issued, err := CreateSession(ctx, pool, targetID, integrationPasswordHash)
		if err != nil {
			t.Fatalf("create target session before admin grant %d: %v", i, err)
		}
		beforeGrant = append(beforeGrant, issued)
	}

	revoked, err := GrantAdminRole(ctx, pool, actorID, targetID)
	if err != nil {
		t.Fatalf("GrantAdminRole: %v", err)
	}
	if revoked != int64(len(beforeGrant)) {
		t.Fatalf("GrantAdminRole revoked %d sessions, want %d", revoked, len(beforeGrant))
	}
	for i, issued := range beforeGrant {
		if _, err := ResolveSession(ctx, pool, issued.SessionToken); !errors.Is(err, ErrInvalidSession) {
			t.Fatalf("pre-grant session %d resolve error = %v, want ErrInvalidSession", i, err)
		}
	}

	var roles []string
	if err := pool.QueryRow(ctx, `
		SELECT array_agg(role_key ORDER BY role_key)
		FROM user_roles
		WHERE user_id = $1 AND revoked_at IS NULL
	`, targetID).Scan(&roles); err != nil {
		t.Fatalf("read target roles after admin grant: %v", err)
	}
	if !slices.Equal(roles, []string{"admin", "customer"}) {
		t.Fatalf("target roles after admin grant = %v, want [admin customer]", roles)
	}

	var grantActor string
	if err := pool.QueryRow(ctx, `
		SELECT granted_by_user_id::text
		FROM user_roles
		WHERE user_id = $1 AND role_key = 'admin' AND revoked_at IS NULL
	`, targetID).Scan(&grantActor); err != nil {
		t.Fatalf("read administrator grant actor: %v", err)
	}
	if grantActor != actorID {
		t.Fatalf("administrator grant actor = %s, want %s", grantActor, actorID)
	}

	if _, err := GrantAdminRole(ctx, pool, actorID, targetID); !errors.Is(err, ErrAdminRoleAlreadyGranted) {
		t.Fatalf("duplicate GrantAdminRole error = %v, want ErrAdminRoleAlreadyGranted", err)
	}

	afterGrant := make([]IssuedSession, 0, 2)
	for i := 0; i < 2; i++ {
		issued, err := CreateSession(ctx, pool, targetID, integrationPasswordHash)
		if err != nil {
			t.Fatalf("create target session after admin grant %d: %v", i, err)
		}
		afterGrant = append(afterGrant, issued)
	}

	revoked, err = RevokeAdminRole(ctx, pool, actorID, targetID)
	if err != nil {
		t.Fatalf("RevokeAdminRole: %v", err)
	}
	if revoked != int64(len(afterGrant)) {
		t.Fatalf("RevokeAdminRole revoked %d sessions, want %d", revoked, len(afterGrant))
	}
	for i, issued := range afterGrant {
		if _, err := ResolveSession(ctx, pool, issued.SessionToken); !errors.Is(err, ErrInvalidSession) {
			t.Fatalf("post-grant session %d resolve error = %v, want ErrInvalidSession", i, err)
		}
	}

	if err := pool.QueryRow(ctx, `
		SELECT array_agg(role_key ORDER BY role_key)
		FROM user_roles
		WHERE user_id = $1 AND revoked_at IS NULL
	`, targetID).Scan(&roles); err != nil {
		t.Fatalf("read target roles after admin revoke: %v", err)
	}
	if !slices.Equal(roles, []string{"customer"}) {
		t.Fatalf("target roles after admin revoke = %v, want [customer]", roles)
	}

	var grantEvents, revokeEvents, sessionRevocationEvents, attributedEvents int
	if err := pool.QueryRow(ctx, `
		SELECT
			count(*) FILTER (WHERE event_type = 'auth.role.granted')::int,
			count(*) FILTER (WHERE event_type = 'auth.role.revoked')::int,
			count(*) FILTER (WHERE event_type = 'auth.session.revoked')::int,
			count(*) FILTER (
				WHERE event_type IN ('auth.role.granted', 'auth.role.revoked')
					AND actor_user_id = $2
			)::int
		FROM audit_events
		WHERE subject_user_id = $1
	`, targetID, actorID).Scan(
		&grantEvents, &revokeEvents, &sessionRevocationEvents, &attributedEvents,
	); err != nil {
		t.Fatalf("read admin role audit events: %v", err)
	}
	if grantEvents != 1 || revokeEvents != 1 ||
		sessionRevocationEvents != len(beforeGrant)+len(afterGrant) || attributedEvents != 2 {
		t.Fatalf("audit counts grant=%d revoke=%d sessions=%d attributed=%d, want 1/1/%d/2",
			grantEvents, revokeEvents, sessionRevocationEvents,
			attributedEvents, len(beforeGrant)+len(afterGrant))
	}

	if _, err := RevokeAdminRole(ctx, pool, actorID, targetID); !errors.Is(err, ErrAdminRoleNotGranted) {
		t.Fatalf("duplicate RevokeAdminRole error = %v, want ErrAdminRoleNotGranted", err)
	}
}

func TestAdminRoleMutationRejectsUnauthorizedSelfAndInactiveTargets(t *testing.T) {
	pool, ctx := openAuthIntegrationPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	ordinaryID := createActiveAuthUser(t, ctx, pool, suffix+"-ordinary-actor")
	targetID := createActiveAuthUser(t, ctx, pool, suffix+"-ordinary-target")

	if _, err := GrantAdminRole(ctx, pool, ordinaryID, targetID); !errors.Is(err, ErrAdminAuthorizationRequired) {
		t.Fatalf("non-admin grant error = %v, want ErrAdminAuthorizationRequired", err)
	}
	if _, err := GrantAdminRole(ctx, pool, ordinaryID, ordinaryID); !errors.Is(err, ErrSelfRoleChange) {
		t.Fatalf("self grant error = %v, want ErrSelfRoleChange", err)
	}

	grantAdminRoleFixture(t, ctx, pool, ordinaryID)
	if _, err := GrantAdminRole(ctx, pool, ordinaryID, ordinaryID); !errors.Is(err, ErrSelfRoleChange) {
		t.Fatalf("admin self grant error = %v, want ErrSelfRoleChange", err)
	}
	if _, err := RevokeAdminRole(ctx, pool, ordinaryID, ordinaryID); !errors.Is(err, ErrSelfRoleChange) {
		t.Fatalf("admin self revoke error = %v, want ErrSelfRoleChange", err)
	}

	pendingID := createPendingAuthUser(t, ctx, pool, suffix+"-pending-target")
	if _, err := GrantAdminRole(ctx, pool, ordinaryID, pendingID); !errors.Is(err, ErrAccountNotActive) {
		t.Fatalf("pending target grant error = %v, want ErrAccountNotActive", err)
	}
	if _, err := RevokeAdminRole(ctx, pool, ordinaryID, targetID); !errors.Is(err, ErrAdminRoleNotGranted) {
		t.Fatalf("revoke non-admin target error = %v, want ErrAdminRoleNotGranted", err)
	}

	var activeAdminGrants int
	if err := pool.QueryRow(ctx, `
		SELECT count(*)::int
		FROM user_roles
		WHERE user_id IN ($1, $2, $3)
			AND role_key = 'admin' AND revoked_at IS NULL
	`, ordinaryID, targetID, pendingID).Scan(&activeAdminGrants); err != nil {
		t.Fatalf("count role grants after rejected mutations: %v", err)
	}
	if activeAdminGrants != 1 {
		t.Fatalf("active admin grants after rejected mutations = %d, want only actor fixture", activeAdminGrants)
	}
}

func TestAdminRoleMutationRequiresActiveVerifiedActor(t *testing.T) {
	pool, ctx := openAuthIntegrationPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	actorID := createActiveAuthUser(t, ctx, pool, suffix+"-disabled-admin")
	targetID := createActiveAuthUser(t, ctx, pool, suffix+"-disabled-admin-target")
	grantAdminRoleFixture(t, ctx, pool, actorID)

	if _, err := pool.Exec(ctx, `
		UPDATE users
		SET status = 'disabled', disabled_at = now(), updated_at = now()
		WHERE id = $1
	`, actorID); err != nil {
		t.Fatalf("disable admin actor: %v", err)
	}
	if _, err := GrantAdminRole(ctx, pool, actorID, targetID); !errors.Is(err, ErrAdminAuthorizationRequired) {
		t.Fatalf("disabled admin grant error = %v, want ErrAdminAuthorizationRequired", err)
	}
}

func TestConcurrentAdminRoleGrantsSerialize(t *testing.T) {
	pool, ctx := openAuthIntegrationPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	firstAdmin := createActiveAuthUser(t, ctx, pool, suffix+"-admin-one")
	secondAdmin := createActiveAuthUser(t, ctx, pool, suffix+"-admin-two")
	targetID := createActiveAuthUser(t, ctx, pool, suffix+"-admin-race-target")
	grantAdminRoleFixture(t, ctx, pool, firstAdmin)
	grantAdminRoleFixture(t, ctx, pool, secondAdmin)

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
			_, err := GrantAdminRole(ctx, pool, actorID, targetID)
			results <- result{actorID: actorID, err: err}
		}()
	}
	close(start)

	var successes, duplicates int
	for i := 0; i < 2; i++ {
		got := <-results
		switch {
		case got.err == nil:
			successes++
		case errors.Is(got.err, ErrAdminRoleAlreadyGranted):
			duplicates++
		default:
			t.Fatalf("concurrent admin grant by %s returned unexpected error: %v", got.actorID, got.err)
		}
	}
	if successes != 1 || duplicates != 1 {
		t.Fatalf("concurrent admin grant outcomes successes=%d duplicates=%d, want 1/1", successes, duplicates)
	}

	var activeAdminCount, grantAuditCount int
	if err := pool.QueryRow(ctx, `
		SELECT
			(SELECT count(*)::int FROM user_roles
			 WHERE user_id = $1 AND role_key = 'admin' AND revoked_at IS NULL),
			(SELECT count(*)::int FROM audit_events
			 WHERE subject_user_id = $1 AND event_type = 'auth.role.granted')
	`, targetID).Scan(&activeAdminCount, &grantAuditCount); err != nil {
		t.Fatalf("read concurrent admin grant result: %v", err)
	}
	if activeAdminCount != 1 || grantAuditCount != 1 {
		t.Fatalf("concurrent admin grants left role=%d audit=%d, want 1/1",
			activeAdminCount, grantAuditCount)
	}
}
