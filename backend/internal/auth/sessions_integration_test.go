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
	const passwordHash = "$argon2id$v=19$m=19456,t=2,p=1$AAECAwQFBgcICQoLDA0ODw$AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8"
	var userID string
	email := fmt.Sprintf("session-%s@example.invalid", suffix)
	if err := pool.QueryRow(ctx, `
		INSERT INTO users(email_normalized, password_hash)
		VALUES($1, $2) RETURNING id::text
	`, email, passwordHash).Scan(&userID); err != nil {
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

	issued, err := CreateSession(ctx, pool, userID)
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

	expired, err := CreateSession(ctx, pool, userID)
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

	disabled, err := CreateSession(ctx, pool, userID)
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
	if _, err := CreateSession(ctx, pool, userID); !errors.Is(err, ErrAccountNotActive) {
		t.Fatalf("CreateSession(disabled) error=%v, want ErrAccountNotActive", err)
	}

	pendingID := createPendingAuthUser(t, ctx, pool, suffix+"-pending")
	if _, err := CreateSession(ctx, pool, pendingID); !errors.Is(err, ErrAccountNotActive) {
		t.Fatalf("CreateSession(unverified) error=%v, want ErrAccountNotActive", err)
	}
}

func createPendingAuthUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, suffix string) string {
	t.Helper()
	const passwordHash = "$argon2id$v=19$m=19456,t=2,p=1$AAECAwQFBgcICQoLDA0ODw$AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8"
	var userID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO users(email_normalized, password_hash)
		VALUES($1, $2) RETURNING id::text
	`, "pending-"+suffix+"@example.invalid", passwordHash).Scan(&userID); err != nil {
		t.Fatalf("create pending account: %v", err)
	}
	return userID
}
