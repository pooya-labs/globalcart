package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	DefaultSessionIdleTTL     = 15 * time.Minute
	DefaultSessionAbsoluteTTL = 24 * time.Hour
)

var (
	ErrAccountNotActive = errors.New("account is not active and verified")
	ErrInvalidSession   = errors.New("invalid session")
)

type IssuedSession struct {
	ID             string
	SessionToken   string
	CSRFToken      string
	IdleExpiresAt  time.Time
	AbsoluteExpiry time.Time
}

// SessionPrincipal is resolved from current database state for each request.
// Its CSRF digest remains unexported so it cannot be accidentally serialized.
type SessionPrincipal struct {
	ID              string
	UserID          string
	Roles           []string
	SellerVendorIDs []int64
	IdleExpiresAt   time.Time
	AbsoluteExpiry  time.Time
	csrfDigest      [32]byte
}

// VerifyCSRF validates the session-bound synchronizer token in constant time.
func (p SessionPrincipal) VerifyCSRF(token string) bool {
	digest, err := DigestOpaqueToken(token)
	if err != nil {
		return false
	}
	defer clear(digest[:])
	return EqualTokenDigests(p.csrfDigest, digest)
}

// CreateSession inserts an opaque server-side session only for an active,
// email-verified account. Raw session and CSRF tokens are returned once.
func CreateSession(ctx context.Context, pool *pgxpool.Pool, userID string) (IssuedSession, error) {
	var issued IssuedSession
	sessionToken, err := NewOpaqueToken()
	if err != nil {
		return issued, fmt.Errorf("generate session token: %w", err)
	}
	csrfToken, err := NewOpaqueToken()
	if err != nil {
		return issued, fmt.Errorf("generate CSRF token: %w", err)
	}
	sessionDigest, err := DigestOpaqueToken(sessionToken)
	if err != nil {
		return issued, fmt.Errorf("digest session token: %w", err)
	}
	defer clear(sessionDigest[:])
	csrfDigest, err := DigestOpaqueToken(csrfToken)
	if err != nil {
		return issued, fmt.Errorf("digest CSRF token: %w", err)
	}
	defer clear(csrfDigest[:])

	tx, err := pool.Begin(ctx)
	if err != nil {
		return issued, fmt.Errorf("begin session creation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var verifiedUserID string
	if err := tx.QueryRow(ctx, `
		SELECT id::text
		FROM users
		WHERE id = $1 AND status = 'active' AND email_verified_at IS NOT NULL
		FOR UPDATE
	`, userID).Scan(&verifiedUserID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return issued, ErrAccountNotActive
		}
		return issued, fmt.Errorf("check session account: %w", err)
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO user_sessions(
			user_id, token_digest, csrf_token_digest, idle_expires_at, expires_at
		)
		VALUES (
			$1, $2, $3,
			now() + INTERVAL '15 minutes',
			now() + INTERVAL '24 hours'
		)
		RETURNING id::text, idle_expires_at, expires_at
	`, verifiedUserID, sessionDigest[:], csrfDigest[:]).Scan(
		&issued.ID, &issued.IdleExpiresAt, &issued.AbsoluteExpiry,
	)
	if err != nil {
		return IssuedSession{}, fmt.Errorf("insert session: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_events(
			event_type, actor_user_id, subject_user_id, resource_type, resource_id, metadata
		)
		VALUES (
			'auth.session.created', $1, $1, 'session', $2,
			'{"kind":"opaque","origin":"internal"}'::jsonb
		)
	`, verifiedUserID, issued.ID); err != nil {
		return IssuedSession{}, fmt.Errorf("audit session creation: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return IssuedSession{}, fmt.Errorf("commit session creation: %w", err)
	}
	issued.SessionToken = sessionToken
	issued.CSRFToken = csrfToken
	return issued, nil
}

// ResolveSession atomically refreshes the idle timeout and resolves current
// roles and active seller scopes. It rejects expired, revoked, disabled and
// unverified accounts with the same error.
func ResolveSession(ctx context.Context, pool *pgxpool.Pool, rawToken string) (SessionPrincipal, error) {
	var principal SessionPrincipal
	digest, err := DigestOpaqueToken(rawToken)
	if err != nil {
		return principal, ErrInvalidSession
	}
	defer clear(digest[:])

	var csrfDigest []byte
	err = pool.QueryRow(ctx, `
		WITH refreshed AS (
			UPDATE user_sessions AS s
			SET last_seen_at = GREATEST(s.last_seen_at, now()),
				idle_expires_at = LEAST(s.expires_at, now() + INTERVAL '15 minutes')
			FROM users AS u
			WHERE s.token_digest = $1
				AND u.id = s.user_id
				AND u.status = 'active'
				AND u.email_verified_at IS NOT NULL
				AND s.revoked_at IS NULL
				AND s.idle_expires_at > now()
				AND s.expires_at > now()
			RETURNING s.id, s.user_id, s.idle_expires_at, s.expires_at, s.csrf_token_digest
		)
		SELECT
			r.id::text,
			r.user_id::text,
			r.idle_expires_at,
			r.expires_at,
			r.csrf_token_digest,
			COALESCE((
				SELECT array_agg(ur.role_key ORDER BY ur.role_key)
				FROM user_roles AS ur
				WHERE ur.user_id = r.user_id
					AND ur.revoked_at IS NULL
					AND (
						ur.role_key <> 'seller'
						OR EXISTS (
							SELECT 1
							FROM seller_memberships AS sm
							JOIN vendors AS v ON v.id = sm.vendor_id
							WHERE sm.user_id = ur.user_id
								AND sm.status = 'active'
								AND v.status = 'active'
						)
					)
			), ARRAY[]::text[]),
			COALESCE((
				SELECT array_agg(DISTINCT sm.vendor_id ORDER BY sm.vendor_id)
				FROM seller_memberships AS sm
				JOIN vendors AS v ON v.id = sm.vendor_id
				JOIN user_roles AS ur
					ON ur.user_id = sm.user_id
					AND ur.role_key = 'seller'
					AND ur.revoked_at IS NULL
				WHERE sm.user_id = r.user_id
					AND sm.status = 'active'
					AND v.status = 'active'
			), ARRAY[]::bigint[])
		FROM refreshed AS r
	`, digest[:]).Scan(
		&principal.ID,
		&principal.UserID,
		&principal.IdleExpiresAt,
		&principal.AbsoluteExpiry,
		&csrfDigest,
		&principal.Roles,
		&principal.SellerVendorIDs,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SessionPrincipal{}, ErrInvalidSession
		}
		return SessionPrincipal{}, fmt.Errorf("resolve session: %w", err)
	}
	if len(csrfDigest) != len(principal.csrfDigest) {
		return SessionPrincipal{}, ErrInvalidSession
	}
	copy(principal.csrfDigest[:], csrfDigest)
	clear(csrfDigest)
	return principal, nil
}

// RevokeSession invalidates a session and records the logout event atomically.
func RevokeSession(ctx context.Context, pool *pgxpool.Pool, rawToken string) error {
	digest, err := DigestOpaqueToken(rawToken)
	if err != nil {
		return ErrInvalidSession
	}
	defer clear(digest[:])

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin session revocation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var sessionID, userID string
	err = tx.QueryRow(ctx, `
		UPDATE user_sessions
		SET revoked_at = now(), revocation_reason = 'logout'
		WHERE token_digest = $1 AND revoked_at IS NULL
		RETURNING id::text, user_id::text
	`, digest[:]).Scan(&sessionID, &userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidSession
		}
		return fmt.Errorf("revoke session: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_events(
			event_type, actor_user_id, subject_user_id, resource_type, resource_id, metadata
		)
		VALUES (
			'auth.session.revoked', $1, $1, 'session', $2,
			'{"reason":"logout"}'::jsonb
		)
	`, userID, sessionID); err != nil {
		return fmt.Errorf("audit session revocation: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit session revocation: %w", err)
	}
	return nil
}

// SessionRevocationReason is a database-approved reason for revoking every
// active session belonging to one account.
type SessionRevocationReason string

const (
	SessionRevocationPasswordChanged  SessionRevocationReason = "password_changed"
	SessionRevocationAccountDisabled  SessionRevocationReason = "account_disabled"
	SessionRevocationPrivilegeChanged SessionRevocationReason = "privilege_changed"
	SessionRevocationAdminRevocation  SessionRevocationReason = "admin_revocation"
	SessionRevocationSecurityRotation SessionRevocationReason = "security_rotation"
)

var (
	ErrInvalidSessionRevocationReason = errors.New("invalid session revocation reason")
	ErrUserNotFound                   = errors.New("user not found")
)

func validSessionRevocationReason(reason SessionRevocationReason) bool {
	switch reason {
	case SessionRevocationPasswordChanged,
		SessionRevocationAccountDisabled,
		SessionRevocationPrivilegeChanged,
		SessionRevocationAdminRevocation,
		SessionRevocationSecurityRotation:
		return true
	default:
		return false
	}
}

// RevokeUserSessions revokes all active sessions for userID and appends one
// redacted audit event per revoked session. It deliberately accepts the
// caller's transaction so password, role or account-state changes can commit
// atomically with revocation. actorUserID may be nil for a system action.
//
// The user row is locked before touching sessions. CreateSession takes the
// same lock, so a concurrently created session is either included in this
// revocation or is created after the caller commits its account change.
func RevokeUserSessions(
	ctx context.Context,
	tx pgx.Tx,
	userID string,
	actorUserID *string,
	reason SessionRevocationReason,
) (int64, error) {
	if !validSessionRevocationReason(reason) {
		return 0, ErrInvalidSessionRevocationReason
	}

	var lockedUserID string
	if err := tx.QueryRow(ctx, `
		SELECT id::text
		FROM users
		WHERE id = $1
		FOR UPDATE
	`, userID).Scan(&lockedUserID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrUserNotFound
		}
		return 0, fmt.Errorf("lock session owner: %w", err)
	}

	var revokedCount int64
	err := tx.QueryRow(ctx, `
		WITH revoked AS (
			UPDATE user_sessions
			SET revoked_at = now(), revocation_reason = $2
			WHERE user_id = $1 AND revoked_at IS NULL
			RETURNING id, user_id
		),
		audited AS (
			INSERT INTO audit_events(
				event_type, actor_user_id, subject_user_id,
				resource_type, resource_id, metadata
			)
			SELECT
				'auth.session.revoked', $3, revoked.user_id,
				'session', revoked.id::text,
				jsonb_build_object('reason', $2::text, 'origin', 'internal')
			FROM revoked
			RETURNING id
		)
		SELECT count(*)::bigint FROM audited
	`, lockedUserID, string(reason), actorUserID).Scan(&revokedCount)
	if err != nil {
		return 0, fmt.Errorf("revoke user sessions: %w", err)
	}
	return revokedCount, nil
}
