package database

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestIdentitySchemaConstraints(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set; PostgreSQL integration test skipped")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("create PostgreSQL pool: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping PostgreSQL: %v", err)
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	var roleCount int
	if err := pool.QueryRow(ctx, `SELECT count(*)::int FROM roles WHERE role_key = ANY($1::text[])`,
		[]string{"customer", "seller", "admin"}).Scan(&roleCount); err != nil {
		t.Fatalf("read controlled roles: %v", err)
	}
	if roleCount != 3 {
		t.Fatalf("seeded roles = %d, want customer/seller/admin", roleCount)
	}

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := "identity-" + suffix + "@example.invalid"
	passwordHash := "$argon2id$v=19$m=65536,t=3,p=1$c2FsdA$aGFzaA"
	var userID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO users(email_normalized, password_hash)
		VALUES($1, $2) RETURNING id::text
	`, email, passwordHash).Scan(&userID); err != nil {
		t.Fatalf("create pending user: %v", err)
	}

	var defaultCustomerRole bool
	if err := pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM user_roles WHERE user_id=$1 AND role_key='customer' AND revoked_at IS NULL)
	`, userID).Scan(&defaultCustomerRole); err != nil {
		t.Fatalf("read default customer role: %v", err)
	}
	if !defaultCustomerRole {
		t.Fatal("new account did not receive only the default customer role")
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO users(email_normalized, password_hash, status, email_verified_at)
		VALUES($1, $2, 'active', NULL)
	`, "unverified-"+suffix+"@example.invalid", passwordHash); err == nil {
		t.Fatal("active account without verified email was accepted")
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO users(email_normalized, password_hash) VALUES($1, $2)
	`, "MixedCase-"+suffix+"@example.invalid", passwordHash); err == nil {
		t.Fatal("non-normalized email was accepted")
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO users(email_normalized, password_hash) VALUES($1, $2)
	`, email, passwordHash); err == nil {
		t.Fatal("duplicate normalized email was accepted")
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO users(email_normalized, password_hash) VALUES($1, $2)
	`, "bad-hash-"+suffix+"@example.invalid", "plaintext-password"); err == nil {
		t.Fatal("non-Argon2id password hash was accepted")
	}

	var vendorID int64
	if err := pool.QueryRow(ctx, `
		INSERT INTO vendors(slug, display_name, status)
		VALUES($1, 'Identity Fixture Vendor', 'pending') RETURNING id
	`, "identity-"+suffix).Scan(&vendorID); err != nil {
		t.Fatalf("create seller vendor fixture: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO user_roles(user_id, role_key) VALUES($1, 'seller')
	`, userID); err != nil {
		t.Fatalf("grant test seller role: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO seller_memberships(user_id, vendor_id, status) VALUES($1, $2, 'pending')
	`, userID, vendorID); err != nil {
		t.Fatalf("create pending seller membership: %v", err)
	}
	var sellerScopeActive bool
	sellerScopeQuery := `
		SELECT EXISTS (
			SELECT 1 FROM user_roles ur
			JOIN seller_memberships sm ON sm.user_id = ur.user_id
			WHERE ur.user_id = $1 AND ur.role_key = 'seller'
				AND ur.revoked_at IS NULL AND sm.vendor_id = $2 AND sm.status = 'active'
		)
	`
	if err := pool.QueryRow(ctx, sellerScopeQuery, userID, vendorID).Scan(&sellerScopeActive); err != nil {
		t.Fatalf("check pending seller authorization: %v", err)
	}
	if sellerScopeActive {
		t.Fatal("pending seller membership granted an active seller scope")
	}
	if _, err := pool.Exec(ctx, `
		UPDATE seller_memberships SET status='active', activated_at=now(), updated_at=now()
		WHERE user_id=$1 AND vendor_id=$2
	`, userID, vendorID); err != nil {
		t.Fatalf("activate seller membership fixture: %v", err)
	}
	if err := pool.QueryRow(ctx, sellerScopeQuery, userID, vendorID).Scan(&sellerScopeActive); err != nil {
		t.Fatalf("check active seller authorization: %v", err)
	}
	if !sellerScopeActive {
		t.Fatal("active seller membership did not produce the expected scoped role join")
	}
	if _, err := pool.Exec(ctx, `INSERT INTO user_roles(user_id, role_key) VALUES($1, 'seller')`, userID); err == nil {
		t.Fatal("duplicate active role grant was accepted")
	}

	now := time.Now().UTC()
	sessionDigest := bytes.Repeat([]byte{0x31}, 32)
	csrfDigest := bytes.Repeat([]byte{0x42}, 32)
	if _, err := pool.Exec(ctx, `
		INSERT INTO user_sessions(user_id, token_digest, csrf_token_digest, idle_expires_at, expires_at)
		VALUES($1, $2, $3, $4, $5)
	`, userID, sessionDigest, csrfDigest, now.Add(15*time.Minute), now.Add(24*time.Hour)); err != nil {
		t.Fatalf("insert opaque session digests: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO user_sessions(user_id, token_digest, csrf_token_digest, idle_expires_at, expires_at)
		VALUES($1, $2, $3, $4, $5)
	`, userID, bytes.Repeat([]byte{0x51}, 31), csrfDigest, now.Add(15*time.Minute), now.Add(24*time.Hour)); err == nil {
		t.Fatal("session token digest with a non-SHA-256 length was accepted")
	}

	actionDigest := bytes.Repeat([]byte{0x63}, 32)
	if _, err := pool.Exec(ctx, `
		INSERT INTO auth_action_tokens(user_id, purpose, token_digest, expires_at)
		VALUES($1, 'password_reset', $2, $3)
	`, userID, actionDigest, now.Add(time.Hour)); err != nil {
		t.Fatalf("insert password-reset digest: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO auth_action_tokens(user_id, purpose, token_digest, expires_at)
		VALUES($1, 'password_reset', $2, $3)
	`, userID, bytes.Repeat([]byte{0x64}, 32), now.Add(time.Hour)); err == nil {
		t.Fatal("second active token for the same user and purpose was accepted")
	}
	if _, err := pool.Exec(ctx, `
		UPDATE auth_action_tokens SET consumed_at=now()
		WHERE user_id=$1 AND purpose='password_reset'
	`, userID); err != nil {
		t.Fatalf("consume password-reset token: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO auth_action_tokens(user_id, purpose, token_digest, expires_at)
		VALUES($1, 'password_reset', $2, $3)
	`, userID, bytes.Repeat([]byte{0x65}, 32), now.Add(time.Hour)); err != nil {
		t.Fatalf("issue replacement token after consumption: %v", err)
	}

	var auditID int64
	if err := pool.QueryRow(ctx, `
		INSERT INTO audit_events(event_type, actor_user_id, subject_user_id, resource_type, resource_id, metadata)
		VALUES('auth.login.success', $1, $1, 'session', 'fixture', '{"method":"password"}'::jsonb)
		RETURNING id
	`, userID).Scan(&auditID); err != nil {
		t.Fatalf("append audit event: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO audit_events(event_type, metadata)
		VALUES('auth.login.failed', '{"password_hash":"must-not-be-audited"}'::jsonb)
	`); err == nil {
		t.Fatal("audit metadata containing a credential field was accepted")
	}
	if _, err := pool.Exec(ctx, "UPDATE audit_events SET metadata='{}'::jsonb WHERE id=$1", auditID); err == nil {
		t.Fatal("audit event UPDATE was not rejected")
	}
	if _, err := pool.Exec(ctx, "DELETE FROM audit_events WHERE id=$1", auditID); err == nil {
		t.Fatal("audit event DELETE was not rejected")
	}
	if _, err := pool.Exec(ctx, "TRUNCATE TABLE audit_events"); err == nil {
		t.Fatal("audit event TRUNCATE was not rejected")
	}
}
