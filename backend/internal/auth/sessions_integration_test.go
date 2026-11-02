package auth

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pooya16pg-spec/globalcart/backend/internal/database"
)

const integrationPasswordHash = "$argon2id$v=19$m=19456,t=2,p=1$AAECAwQFBgcICQoLDA0ODw$AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8"

func openAuthIntegrationPool(t *testing.T) (*pgxpool.Pool, context.Context) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set; PostgreSQL integration test skipped")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("create PostgreSQL pool: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping PostgreSQL: %v", err)
	}
	if err := database.Migrate(ctx, pool); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	return pool, ctx
}

func createActiveAuthUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, suffix string) string {
	t.Helper()
	var userID string
	email := fmt.Sprintf("session-%s@example.invalid", suffix)
	if err := pool.QueryRow(ctx, `
		INSERT INTO users(email_normalized, password_hash)
		VALUES($1, $2) RETURNING id::text
	`, email, integrationPasswordHash).Scan(&userID); err != nil {
		t.Fatalf("create pending user: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		UPDATE users SET status='active', email_verified_at=now(), updated_at=now()
		WHERE id=$1
	`, userID); err != nil {
		t.Fatalf("activate verified integration user: %v", err)
	}
	return userID
}

func TestSessionLifecycleAndLiveSellerScope(t *testing.T) {
	pool, ctx := openAuthIntegrationPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	userID := createActiveAuthUser(t, ctx, pool, suffix)

	var vendorID int64
	if err := pool.QueryRow(ctx, `
		INSERT INTO vendors(slug, display_name, status)
		VALUES($1, 'Session Integration Vendor', 'active') RETURNING id
	`, "session-"+suffix).Scan(&vendorID); err != nil {
		t.Fatalf("create active vendor: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO user_roles(user_id, role_key) VALUES($1, 'seller')
	`, userID); err != nil {
		t.Fatalf("grant seller fixture role: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO seller_memberships(user_id, vendor_id, status)
		VALUES($1, $2, 'pending')
	`, userID, vendorID); err != nil {
		t.Fatalf("create pending seller membership: %v", err)
	}

	issued, err := CreateSession(ctx, pool, userID, integrationPasswordHash)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if issued.SessionToken == "" || issued.CSRFToken == "" || issued.ID == "" {
		t.Fatal("session creation returned an empty identifier or token")
	}
	if !issued.IdleExpiresAt.After(time.Now()) ||
		!issued.AbsoluteExpiry.After(issued.IdleExpiresAt) ||
		issued.AbsoluteExpiry.Sub(issued.IdleExpiresAt) < 23*time.Hour {
		t.Fatalf("unexpected session expiry window: idle=%s absolute=%s",
			issued.IdleExpiresAt, issued.AbsoluteExpiry)
	}

	sessionDigest, err := DigestOpaqueToken(issued.SessionToken)
	if err != nil {
		t.Fatalf("digest issued session token: %v", err)
	}
	csrfDigest, err := DigestOpaqueToken(issued.CSRFToken)
	if err != nil {
		t.Fatalf("digest issued CSRF token: %v", err)
	}
	var storedSessionDigest, storedCSRFDigest []byte
	if err := pool.QueryRow(ctx, `
		SELECT token_digest, csrf_token_digest FROM user_sessions WHERE id=$1
	`, issued.ID).Scan(&storedSessionDigest, &storedCSRFDigest); err != nil {
		t.Fatalf("read stored session digests: %v", err)
	}
	if !bytes.Equal(storedSessionDigest, sessionDigest[:]) ||
		!bytes.Equal(storedCSRFDigest, csrfDigest[:]) {
		t.Fatal("database did not store the exact SHA-256 token digests")
	}
	if bytes.Equal(storedSessionDigest, []byte(issued.SessionToken)) ||
		bytes.Equal(storedCSRFDigest, []byte(issued.CSRFToken)) {
		t.Fatal("database stored a raw bearer token")
	}

	principal, err := ResolveSession(ctx, pool, issued.SessionToken)
	if err != nil {
		t.Fatalf("ResolveSession: %v", err)
	}
	if principal.ID != issued.ID || principal.UserID != userID {
		t.Fatalf("resolved identity = %+v, want session %s user %s", principal, issued.ID, userID)
	}
	if !slices.Equal(principal.Roles, []string{"customer"}) {
		t.Fatalf("effective roles for pending membership = %v, want [customer]", principal.Roles)
	}
	if len(principal.SellerVendorIDs) != 0 {
		t.Fatalf("pending seller membership granted vendor scopes: %v", principal.SellerVendorIDs)
	}
	if !principal.VerifyCSRF(issued.CSRFToken) {
		t.Fatal("session did not accept its CSRF token")
	}
	if principal.VerifyCSRF(issued.SessionToken) || principal.VerifyCSRF("malformed") {
		t.Fatal("session accepted a wrong or malformed CSRF token")
	}

	if _, err := pool.Exec(ctx, `
		UPDATE seller_memberships
		SET status='active', activated_at=now(), updated_at=now()
		WHERE user_id=$1 AND vendor_id=$2
	`, userID, vendorID); err != nil {
		t.Fatalf("activate seller membership: %v", err)
	}
	principal, err = ResolveSession(ctx, pool, issued.SessionToken)
	if err != nil {
		t.Fatalf("resolve session after membership activation: %v", err)
	}
	if !slices.Contains(principal.Roles, "seller") ||
		!slices.Equal(principal.SellerVendorIDs, []int64{vendorID}) {
		t.Fatalf("active seller membership not reflected in live scope: roles=%v vendors=%v",
			principal.Roles, principal.SellerVendorIDs)
	}

	if err := RevokeSession(ctx, pool, issued.SessionToken); err != nil {
		t.Fatalf("RevokeSession: %v", err)
	}
	if _, err := ResolveSession(ctx, pool, issued.SessionToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("ResolveSession after logout error=%v, want ErrInvalidSession", err)
	}
	var createdEvents, revokedEvents int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FILTER (WHERE event_type='auth.session.created')::int,
		       count(*) FILTER (WHERE event_type='auth.session.revoked')::int
		FROM audit_events WHERE subject_user_id=$1
	`, userID).Scan(&createdEvents, &revokedEvents); err != nil {
		t.Fatalf("read session audit events: %v", err)
	}
	if createdEvents != 1 || revokedEvents != 1 {
		t.Fatalf("session audit counts created=%d revoked=%d, want one each", createdEvents, revokedEvents)
	}
	if err := RevokeSession(ctx, pool, issued.SessionToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("duplicate logout error=%v, want ErrInvalidSession", err)
	}
}

func TestSessionRejectsExpiredAndInactiveAccounts(t *testing.T) {
	pool, ctx := openAuthIntegrationPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	userID := createActiveAuthUser(t, ctx, pool, suffix)

	expired, err := CreateSession(ctx, pool, userID, integrationPasswordHash)
	if err != nil {
		t.Fatalf("create expiring session: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		UPDATE user_sessions
		SET created_at=now()-INTERVAL '2 hours',
		    last_seen_at=now()-INTERVAL '2 hours',
		    idle_expires_at=now()-INTERVAL '1 hour',
		    expires_at=now()-INTERVAL '1 hour'
		WHERE id=$1
	`, expired.ID); err != nil {
		t.Fatalf("expire session fixture: %v", err)
	}
	if _, err := ResolveSession(ctx, pool, expired.SessionToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("expired session error=%v, want ErrInvalidSession", err)
	}

	disabled, err := CreateSession(ctx, pool, userID, integrationPasswordHash)
	if err != nil {
		t.Fatalf("create session before disabling account: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		UPDATE users SET status='disabled', disabled_at=now(), updated_at=now()
		WHERE id=$1
	`, userID); err != nil {
		t.Fatalf("disable account fixture: %v", err)
	}
	if _, err := ResolveSession(ctx, pool, disabled.SessionToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("disabled account session error=%v, want ErrInvalidSession", err)
	}
	if _, err := CreateSession(ctx, pool, userID, integrationPasswordHash); !errors.Is(err, ErrAccountNotActive) {
		t.Fatalf("CreateSession(disabled) error=%v, want ErrAccountNotActive", err)
	}

	pendingID := createPendingAuthUser(t, ctx, pool, suffix+"-pending")
	if _, err := CreateSession(ctx, pool, pendingID, integrationPasswordHash); !errors.Is(err, ErrAccountNotActive) {
		t.Fatalf("CreateSession(unverified) error=%v, want ErrAccountNotActive", err)
	}
}

func createPendingAuthUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, suffix string) string {
	t.Helper()
	var userID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO users(email_normalized, password_hash)
		VALUES($1, $2) RETURNING id::text
	`, "pending-"+suffix+"@example.invalid", integrationPasswordHash).Scan(&userID); err != nil {
		t.Fatalf("create pending account: %v", err)
	}
	return userID
}

func TestRevokeUserSessionsIsTransactionalScopedAndAudited(t *testing.T) {
	pool, ctx := openAuthIntegrationPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	targetID := createActiveAuthUser(t, ctx, pool, suffix+"-bulk-target")
	otherID := createActiveAuthUser(t, ctx, pool, suffix+"-bulk-other")

	targetSessions := make([]IssuedSession, 0, 3)
	for i := 0; i < 3; i++ {
		issued, err := CreateSession(ctx, pool, targetID, integrationPasswordHash)
		if err != nil {
			t.Fatalf("create target session %d: %v", i, err)
		}
		targetSessions = append(targetSessions, issued)
	}
	otherSession, err := CreateSession(ctx, pool, otherID, integrationPasswordHash)
	if err != nil {
		t.Fatalf("create unrelated user's session: %v", err)
	}

	actorID := targetID
	rollbackTx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin rollback transaction: %v", err)
	}
	rolledBackCount, err := RevokeUserSessions(
		ctx, rollbackTx, targetID, &actorID, SessionRevocationPasswordChanged,
	)
	if err != nil {
		_ = rollbackTx.Rollback(ctx)
		t.Fatalf("RevokeUserSessions before rollback: %v", err)
	}
	if rolledBackCount != int64(len(targetSessions)) {
		_ = rollbackTx.Rollback(ctx)
		t.Fatalf("revoked %d sessions before rollback, want %d", rolledBackCount, len(targetSessions))
	}
	if err := rollbackTx.Rollback(ctx); err != nil {
		t.Fatalf("rollback revocation transaction: %v", err)
	}
	for i, issued := range targetSessions {
		if _, err := ResolveSession(ctx, pool, issued.SessionToken); err != nil {
			t.Fatalf("session %d became invalid after caller rollback: %v", i, err)
		}
	}

	invalidTx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin invalid-reason transaction: %v", err)
	}
	if _, err := RevokeUserSessions(
		ctx, invalidTx, targetID, &actorID, SessionRevocationReason("unapproved_reason"),
	); !errors.Is(err, ErrInvalidSessionRevocationReason) {
		_ = invalidTx.Rollback(ctx)
		t.Fatalf("invalid reason error = %v, want ErrInvalidSessionRevocationReason", err)
	}
	if err := invalidTx.Rollback(ctx); err != nil {
		t.Fatalf("rollback invalid-reason transaction: %v", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin committed revocation transaction: %v", err)
	}
	revokedCount, err := RevokeUserSessions(
		ctx, tx, targetID, &actorID, SessionRevocationPasswordChanged,
	)
	if err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("RevokeUserSessions: %v", err)
	}
	if revokedCount != int64(len(targetSessions)) {
		_ = tx.Rollback(ctx)
		t.Fatalf("revoked %d sessions, want %d", revokedCount, len(targetSessions))
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit bulk revocation: %v", err)
	}

	for i, issued := range targetSessions {
		if _, err := ResolveSession(ctx, pool, issued.SessionToken); !errors.Is(err, ErrInvalidSession) {
			t.Fatalf("target session %d resolve error = %v, want ErrInvalidSession", i, err)
		}
	}
	if _, err := ResolveSession(ctx, pool, otherSession.SessionToken); err != nil {
		t.Fatalf("bulk revocation affected another user's session: %v", err)
	}

	var revokedRows, passwordChangedRows int
	if err := pool.QueryRow(ctx, `
		SELECT
			count(*) FILTER (WHERE revoked_at IS NOT NULL)::int,
			count(*) FILTER (WHERE revocation_reason = 'password_changed')::int
		FROM user_sessions
		WHERE user_id = $1
	`, targetID).Scan(&revokedRows, &passwordChangedRows); err != nil {
		t.Fatalf("read bulk-revoked sessions: %v", err)
	}
	if revokedRows != len(targetSessions) || passwordChangedRows != len(targetSessions) {
		t.Fatalf("revoked rows=%d password_changed rows=%d, want %d each",
			revokedRows, passwordChangedRows, len(targetSessions))
	}

	var auditRows, reasonRows, actorRows int
	if err := pool.QueryRow(ctx, `
		SELECT
			count(*) FILTER (WHERE event_type = 'auth.session.revoked')::int,
			count(*) FILTER (WHERE metadata->>'reason' = 'password_changed')::int,
			count(*) FILTER (WHERE event_type = 'auth.session.revoked' AND actor_user_id = $1)::int
		FROM audit_events
		WHERE subject_user_id = $1
	`, targetID).Scan(&auditRows, &reasonRows, &actorRows); err != nil {
		t.Fatalf("read bulk-revocation audit events: %v", err)
	}
	if auditRows != len(targetSessions) || reasonRows != len(targetSessions) ||
		actorRows != len(targetSessions) {
		t.Fatalf("audit rows=%d reason rows=%d actor rows=%d, want %d each",
			auditRows, reasonRows, actorRows, len(targetSessions))
	}

	idempotentTx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin repeated revocation transaction: %v", err)
	}
	repeatedCount, err := RevokeUserSessions(
		ctx, idempotentTx, targetID, &actorID, SessionRevocationPrivilegeChanged,
	)
	if err != nil {
		_ = idempotentTx.Rollback(ctx)
		t.Fatalf("repeat RevokeUserSessions: %v", err)
	}
	if repeatedCount != 0 {
		_ = idempotentTx.Rollback(ctx)
		t.Fatalf("repeat revocation affected %d sessions, want 0", repeatedCount)
	}
	if err := idempotentTx.Commit(ctx); err != nil {
		t.Fatalf("commit repeated revocation: %v", err)
	}

	if err := pool.QueryRow(ctx, `
		SELECT count(*)::int
		FROM audit_events
		WHERE subject_user_id = $1 AND event_type = 'auth.session.revoked'
	`, targetID).Scan(&auditRows); err != nil {
		t.Fatalf("count audit events after idempotent retry: %v", err)
	}
	if auditRows != len(targetSessions) {
		t.Fatalf("idempotent retry added audit events: got %d, want %d", auditRows, len(targetSessions))
	}
}

func TestSessionCreationSerializesWithAccountDisableAndBulkRevocation(t *testing.T) {
	pool, ctx := openAuthIntegrationPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	userID := createActiveAuthUser(t, ctx, pool, suffix+"-disable-race")

	existing, err := CreateSession(ctx, pool, userID, integrationPasswordHash)
	if err != nil {
		t.Fatalf("create existing session: %v", err)
	}

	// Keep the account row locked while disabling the account and revoking
	// its existing sessions in one transaction.
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin account-disable transaction: %v", err)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE users
		SET status = 'disabled', disabled_at = now(), updated_at = now()
		WHERE id = $1
	`, userID); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("disable account inside transaction: %v", err)
	}

	revoked, err := RevokeUserSessions(
		ctx, tx, userID, &userID, SessionRevocationAccountDisabled,
	)
	if err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("revoke sessions inside account-disable transaction: %v", err)
	}
	if revoked != 1 {
		_ = tx.Rollback(ctx)
		t.Fatalf("revoked %d sessions, want 1", revoked)
	}

	// CreateSession must wait on the same user-row lock. Its bounded context
	// proves it cannot slip through while the account mutation is uncommitted.
	attemptCtx, cancelAttempt := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancelAttempt()
	started := make(chan struct{})
	createResult := make(chan error, 1)
	go func() {
		close(started)
		_, createErr := CreateSession(attemptCtx, pool, userID, integrationPasswordHash)
		createResult <- createErr
	}()
	<-started

	select {
	case createErr := <-createResult:
		if !errors.Is(createErr, context.DeadlineExceeded) {
			_ = tx.Rollback(ctx)
			t.Fatalf("concurrent CreateSession error = %v, want lock-wait deadline exceeded", createErr)
		}
	case <-time.After(2 * time.Second):
		cancelAttempt()
		_ = tx.Rollback(ctx)
		t.Fatal("concurrent CreateSession did not finish after its bounded lock-wait context")
	}

	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit account disable and session revocation: %v", err)
	}

	if _, err := ResolveSession(ctx, pool, existing.SessionToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("pre-existing session resolve error = %v, want ErrInvalidSession", err)
	}
	if _, err := CreateSession(ctx, pool, userID, integrationPasswordHash); !errors.Is(err, ErrAccountNotActive) {
		t.Fatalf("CreateSession after account disable error = %v, want ErrAccountNotActive", err)
	}

	var disabledRevocations int
	if err := pool.QueryRow(ctx, `
		SELECT count(*)::int
		FROM user_sessions
		WHERE user_id = $1
			AND revoked_at IS NOT NULL
			AND revocation_reason = 'account_disabled'
	`, userID).Scan(&disabledRevocations); err != nil {
		t.Fatalf("count account-disable revocations: %v", err)
	}
	if disabledRevocations != 1 {
		t.Fatalf("account-disable revocations = %d, want 1", disabledRevocations)
	}
}

func TestChangePasswordRevokesSessionsAndRejectsStaleCredential(t *testing.T) {
	pool, ctx := openAuthIntegrationPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	userID := createActiveAuthUser(t, ctx, pool, suffix+"-password-change")

	oldPassword := "old-password-value"
	oldPasswordHash, err := HashPassword(oldPassword)
	if err != nil {
		t.Fatalf("hash old password fixture: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		UPDATE users
		SET password_hash = $1, password_changed_at = now(), updated_at = now()
		WHERE id = $2
	`, oldPasswordHash, userID); err != nil {
		t.Fatalf("install old password fixture: %v", err)
	}

	resetToken, err := IssuePasswordResetToken(ctx, pool, userID, 0)
	if err != nil {
		t.Fatalf("issue reset token before password change: %v", err)
	}

	oldSessions := make([]IssuedSession, 0, 2)
	for i := 0; i < 2; i++ {
		issued, err := CreateSession(ctx, pool, userID, oldPasswordHash)
		if err != nil {
			t.Fatalf("create old-credential session %d: %v", i, err)
		}
		oldSessions = append(oldSessions, issued)
	}

	actorID := userID
	revoked, err := ChangePassword(
		ctx, pool, userID, oldPasswordHash, "new-password-value", &actorID,
	)
	if err != nil {
		t.Fatalf("ChangePassword: %v", err)
	}
	if revoked != int64(len(oldSessions)) {
		t.Fatalf("ChangePassword revoked %d sessions, want %d", revoked, len(oldSessions))
	}

	var storedHash string
	if err := pool.QueryRow(ctx, `
		SELECT password_hash FROM users WHERE id = $1
	`, userID).Scan(&storedHash); err != nil {
		t.Fatalf("read updated password hash: %v", err)
	}
	if storedHash == oldPasswordHash {
		t.Fatal("password change retained the previous credential hash")
	}
	if ok, err := VerifyPassword(storedHash, "new-password-value"); err != nil || !ok {
		t.Fatalf("VerifyPassword(new password) = (%t, %v), want (true, nil)", ok, err)
	}
	if ok, err := VerifyPassword(storedHash, oldPassword); err != nil || ok {
		t.Fatalf("VerifyPassword(wrong old password) = (%t, %v), want (false, nil)", ok, err)
	}

	// This simulates a login that verified the old password before the
	// password-change transaction committed, but tried to issue its session
	// only afterwards. The credential snapshot must now be rejected.
	if _, err := CreateSession(ctx, pool, userID, oldPasswordHash); !errors.Is(err, ErrCredentialsChanged) {
		t.Fatalf("CreateSession with stale credential error = %v, want ErrCredentialsChanged", err)
	}
	if _, err := CompletePasswordReset(ctx, pool, resetToken.Token, "must-not-replace-new-password"); !errors.Is(err, ErrInvalidActionToken) {
		t.Fatalf("reset token issued before password change error = %v, want ErrInvalidActionToken", err)
	}

	current, err := CreateSession(ctx, pool, userID, storedHash)
	if err != nil {
		t.Fatalf("CreateSession with current credential: %v", err)
	}
	if _, err := ResolveSession(ctx, pool, current.SessionToken); err != nil {
		t.Fatalf("resolve session issued with current credential: %v", err)
	}
	for i, issued := range oldSessions {
		if _, err := ResolveSession(ctx, pool, issued.SessionToken); !errors.Is(err, ErrInvalidSession) {
			t.Fatalf("old session %d resolve error = %v, want ErrInvalidSession", i, err)
		}
	}

	var passwordEvents, revokedEvents, actorEvents int
	if err := pool.QueryRow(ctx, `
		SELECT
			count(*) FILTER (WHERE event_type = 'auth.password.changed')::int,
			count(*) FILTER (WHERE event_type = 'auth.session.revoked')::int,
			count(*) FILTER (WHERE event_type = 'auth.password.changed' AND actor_user_id = $1)::int
		FROM audit_events
		WHERE subject_user_id = $1
	`, userID).Scan(&passwordEvents, &revokedEvents, &actorEvents); err != nil {
		t.Fatalf("read password-change audit events: %v", err)
	}
	if passwordEvents != 1 || revokedEvents != len(oldSessions) || actorEvents != 1 {
		t.Fatalf("password events=%d revoked events=%d actor events=%d, want 1/%d/1",
			passwordEvents, revokedEvents, actorEvents, len(oldSessions))
	}
}

func TestPasswordResetTokenIsDigestOnlyRotatesAndConsumesOnce(t *testing.T) {
	pool, ctx := openAuthIntegrationPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	userID := createActiveAuthUser(t, ctx, pool, suffix+"-reset-flow")
	oldSession, err := CreateSession(ctx, pool, userID, integrationPasswordHash)
	if err != nil {
		t.Fatalf("create pre-reset session: %v", err)
	}

	first, err := IssuePasswordResetToken(ctx, pool, userID, 0)
	if err != nil {
		t.Fatalf("issue first reset token: %v", err)
	}
	if first.Token == "" || first.ID == "" || !first.ExpiresAt.After(time.Now()) ||
		first.ExpiresAt.After(time.Now().Add(DefaultPasswordResetTTL+time.Minute)) {
		t.Fatalf("invalid first reset token metadata: %+v", first)
	}
	firstDigest, err := DigestOpaqueToken(first.Token)
	if err != nil {
		t.Fatalf("digest first reset token: %v", err)
	}
	var storedDigest []byte
	if err := pool.QueryRow(ctx, `
		SELECT token_digest FROM auth_action_tokens WHERE id = $1
	`, first.ID).Scan(&storedDigest); err != nil {
		t.Fatalf("read stored reset-token digest: %v", err)
	}
	if !bytes.Equal(storedDigest, firstDigest[:]) || bytes.Equal(storedDigest, []byte(first.Token)) {
		t.Fatal("password-reset token was not stored exclusively as its SHA-256 digest")
	}

	second, err := IssuePasswordResetToken(ctx, pool, userID, 10*time.Minute)
	if err != nil {
		t.Fatalf("rotate password-reset token: %v", err)
	}
	var firstRevoked, secondRevoked bool
	if err := pool.QueryRow(ctx, `
		SELECT revoked_at IS NOT NULL FROM auth_action_tokens WHERE id = $1
	`, first.ID).Scan(&firstRevoked); err != nil {
		t.Fatalf("read first token state: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		SELECT revoked_at IS NOT NULL FROM auth_action_tokens WHERE id = $1
	`, second.ID).Scan(&secondRevoked); err != nil {
		t.Fatalf("read second token state: %v", err)
	}
	if !firstRevoked || secondRevoked {
		t.Fatalf("token rotation states first revoked=%t second revoked=%t, want true/false",
			firstRevoked, secondRevoked)
	}
	if _, err := CompletePasswordReset(ctx, pool, first.Token, "first-token-password"); !errors.Is(err, ErrInvalidActionToken) {
		t.Fatalf("rotated token completion error = %v, want ErrInvalidActionToken", err)
	}

	revokedSessions, err := CompletePasswordReset(ctx, pool, second.Token, "password-reset-success")
	if err != nil {
		t.Fatalf("complete password reset: %v", err)
	}
	if revokedSessions != 1 {
		t.Fatalf("password reset revoked %d sessions, want 1", revokedSessions)
	}
	if _, err := CompletePasswordReset(ctx, pool, second.Token, "second-use-password"); !errors.Is(err, ErrInvalidActionToken) {
		t.Fatalf("second token use error = %v, want ErrInvalidActionToken", err)
	}

	var newHash string
	if err := pool.QueryRow(ctx, `
		SELECT password_hash FROM users WHERE id = $1
	`, userID).Scan(&newHash); err != nil {
		t.Fatalf("read reset password hash: %v", err)
	}
	if ok, err := VerifyPassword(newHash, "password-reset-success"); err != nil || !ok {
		t.Fatalf("VerifyPassword(reset password) = (%t, %v), want (true, nil)", ok, err)
	}
	if _, err := CreateSession(ctx, pool, userID, integrationPasswordHash); !errors.Is(err, ErrCredentialsChanged) {
		t.Fatalf("old credential after reset error = %v, want ErrCredentialsChanged", err)
	}
	if _, err := ResolveSession(ctx, pool, oldSession.SessionToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("pre-reset session resolve error = %v, want ErrInvalidSession", err)
	}
	current, err := CreateSession(ctx, pool, userID, newHash)
	if err != nil {
		t.Fatalf("create session after reset with current password hash: %v", err)
	}
	if _, err := ResolveSession(ctx, pool, current.SessionToken); err != nil {
		t.Fatalf("resolve session after reset: %v", err)
	}

	var issuedEvents, completedEvents int
	if err := pool.QueryRow(ctx, `
		SELECT
			count(*) FILTER (WHERE event_type = 'auth.password_reset.issued')::int,
			count(*) FILTER (WHERE event_type = 'auth.password_reset.completed')::int
		FROM audit_events
		WHERE subject_user_id = $1
	`, userID).Scan(&issuedEvents, &completedEvents); err != nil {
		t.Fatalf("read reset audit events: %v", err)
	}
	if issuedEvents != 2 || completedEvents != 1 {
		t.Fatalf("reset audit counts issued=%d completed=%d, want 2/1", issuedEvents, completedEvents)
	}
}

func TestPasswordResetRejectsExpiredToken(t *testing.T) {
	pool, ctx := openAuthIntegrationPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	userID := createActiveAuthUser(t, ctx, pool, suffix+"-reset-expired")
	issued, err := IssuePasswordResetToken(ctx, pool, userID, 5*time.Minute)
	if err != nil {
		t.Fatalf("issue expiring reset token: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		UPDATE auth_action_tokens
		SET expires_at = now() - INTERVAL '1 minute'
		WHERE id = $1
	`, issued.ID); err != nil {
		t.Fatalf("expire reset-token fixture: %v", err)
	}
	if _, err := CompletePasswordReset(ctx, pool, issued.Token, "expired-token-password"); !errors.Is(err, ErrInvalidActionToken) {
		t.Fatalf("expired token completion error = %v, want ErrInvalidActionToken", err)
	}
}

func TestConcurrentPasswordResetTokenConsumptionIsSingleUse(t *testing.T) {
	pool, ctx := openAuthIntegrationPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	userID := createActiveAuthUser(t, ctx, pool, suffix+"-reset-concurrent")
	issued, err := IssuePasswordResetToken(ctx, pool, userID, 10*time.Minute)
	if err != nil {
		t.Fatalf("issue concurrent reset token: %v", err)
	}

	type result struct {
		password string
		revoked  int64
		err      error
	}
	results := make(chan result, 2)
	for _, password := range []string{"concurrent-reset-a", "concurrent-reset-b"} {
		password := password
		go func() {
			count, resetErr := CompletePasswordReset(ctx, pool, issued.Token, password)
			results <- result{password: password, revoked: count, err: resetErr}
		}()
	}

	var successes, invalids int
	var successfulPassword string
	for i := 0; i < 2; i++ {
		got := <-results
		switch {
		case got.err == nil:
			successes++
			successfulPassword = got.password
		case errors.Is(got.err, ErrInvalidActionToken):
			invalids++
		default:
			t.Fatalf("concurrent reset returned unexpected error: %v", got.err)
		}
	}
	if successes != 1 || invalids != 1 {
		t.Fatalf("concurrent reset outcomes successes=%d invalid=%d, want 1/1", successes, invalids)
	}
	var storedHash string
	if err := pool.QueryRow(ctx, `
		SELECT password_hash FROM users WHERE id = $1
	`, userID).Scan(&storedHash); err != nil {
		t.Fatalf("read concurrently reset password: %v", err)
	}
	if ok, err := VerifyPassword(storedHash, successfulPassword); err != nil || !ok {
		t.Fatalf("stored password does not match the sole successful reset: (%t, %v)", ok, err)
	}
}

func TestPasswordResetTTLBounds(t *testing.T) {
	cases := []struct {
		name string
		in   time.Duration
		want time.Duration
		err  error
	}{
		{name: "default", in: 0, want: DefaultPasswordResetTTL},
		{name: "minimum", in: MinimumPasswordResetTTL, want: MinimumPasswordResetTTL},
		{name: "maximum", in: MaximumPasswordResetTTL, want: MaximumPasswordResetTTL},
		{name: "negative", in: -time.Minute, err: ErrInvalidActionTokenTTL},
		{name: "too short", in: time.Minute, err: ErrInvalidActionTokenTTL},
		{name: "too long", in: MaximumPasswordResetTTL + time.Second, err: ErrInvalidActionTokenTTL},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := normalizePasswordResetTTL(tc.in)
			if !errors.Is(err, tc.err) {
				t.Fatalf("normalizePasswordResetTTL error = %v, want %v", err, tc.err)
			}
			if err == nil && got != tc.want {
				t.Fatalf("normalizePasswordResetTTL = %s, want %s", got, tc.want)
			}
		})
	}
}
