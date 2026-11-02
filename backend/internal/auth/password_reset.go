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
	DefaultPasswordResetTTL = 30 * time.Minute
	MinimumPasswordResetTTL = 5 * time.Minute
	MaximumPasswordResetTTL = 1 * time.Hour
)

var (
	ErrInvalidActionToken    = errors.New("invalid or expired action token")
	ErrInvalidActionTokenTTL = errors.New("password-reset token TTL must be between 5 minutes and 1 hour")
)

type IssuedPasswordResetToken struct {
	ID        string
	Token     string
	ExpiresAt time.Time
}

func normalizePasswordResetTTL(ttl time.Duration) (time.Duration, error) {
	if ttl == 0 {
		return DefaultPasswordResetTTL, nil
	}
	if ttl < MinimumPasswordResetTTL || ttl > MaximumPasswordResetTTL {
		return 0, ErrInvalidActionTokenTTL
	}
	return ttl, nil
}

// IssuePasswordResetToken creates a short-lived one-time reset token for an
// active, verified account. Only its SHA-256 digest is stored. Issuing a new
// token revokes every prior unconsumed password-reset token for that user.
// Delivery is intentionally left to an approved email worker.
func IssuePasswordResetToken(
	ctx context.Context,
	pool *pgxpool.Pool,
	userID string,
	ttl time.Duration,
) (IssuedPasswordResetToken, error) {
	var issued IssuedPasswordResetToken
	ttl, err := normalizePasswordResetTTL(ttl)
	if err != nil {
		return issued, err
	}

	rawToken, err := NewOpaqueToken()
	if err != nil {
		return issued, fmt.Errorf("generate password-reset token: %w", err)
	}
	digest, err := DigestOpaqueToken(rawToken)
	if err != nil {
		return issued, fmt.Errorf("digest password-reset token: %w", err)
	}
	defer clear(digest[:])

	tx, err := pool.Begin(ctx)
	if err != nil {
		return issued, fmt.Errorf("begin password-reset token issue: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var lockedUserID string
	if err := tx.QueryRow(ctx, `
		SELECT id::text
		FROM users
		WHERE id = $1 AND status = 'active' AND email_verified_at IS NOT NULL
		FOR UPDATE
	`, userID).Scan(&lockedUserID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return issued, ErrAccountNotActive
		}
		return issued, fmt.Errorf("lock password-reset account: %w", err)
	}

	// Expired but unconsumed tokens also occupy the partial unique index.
	// Revoke all previous rows before inserting the replacement.
	if _, err := tx.Exec(ctx, `
		UPDATE auth_action_tokens
		SET revoked_at = now()
		WHERE user_id = $1
			AND purpose = 'password_reset'
			AND consumed_at IS NULL
			AND revoked_at IS NULL
	`, lockedUserID); err != nil {
		return issued, fmt.Errorf("revoke previous password-reset tokens: %w", err)
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO auth_action_tokens(user_id, purpose, token_digest, expires_at)
		VALUES ($1, 'password_reset', $2, now() + ($3::bigint * INTERVAL '1 second'))
		RETURNING id::text, expires_at
	`, lockedUserID, digest[:], int64(ttl/time.Second)).Scan(&issued.ID, &issued.ExpiresAt)
	if err != nil {
		return IssuedPasswordResetToken{}, fmt.Errorf("insert password-reset token: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_events(
			event_type, subject_user_id, resource_type, resource_id, metadata
		)
		VALUES (
			'auth.password_reset.issued', $1, 'action_token', $2,
			'{"purpose":"password_reset","origin":"internal"}'::jsonb
		)
	`, lockedUserID, issued.ID); err != nil {
		return IssuedPasswordResetToken{}, fmt.Errorf("audit password-reset issue: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return IssuedPasswordResetToken{}, fmt.Errorf("commit password-reset token issue: %w", err)
	}
	issued.Token = rawToken
	return issued, nil
}

// CompletePasswordReset consumes a one-time token, changes the password,
// revokes outstanding reset tokens and active sessions, and audits the event
// in one transaction. It always locks the user row before updating token rows,
// matching IssuePasswordResetToken and preventing a lock-order deadlock.
func CompletePasswordReset(
	ctx context.Context,
	pool *pgxpool.Pool,
	rawToken string,
	newPassword string,
) (int64, error) {
	digest, err := DigestOpaqueToken(rawToken)
	if err != nil {
		return 0, ErrInvalidActionToken
	}
	defer clear(digest[:])

	var tokenID, userID string
	err = pool.QueryRow(ctx, `
		SELECT id::text, user_id::text
		FROM auth_action_tokens
		WHERE token_digest = $1
			AND purpose = 'password_reset'
			AND consumed_at IS NULL
			AND revoked_at IS NULL
			AND expires_at > now()
	`, digest[:]).Scan(&tokenID, &userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrInvalidActionToken
		}
		return 0, fmt.Errorf("find password-reset token: %w", err)
	}

	// Hash only after the cheap token preflight, then revalidate token state
	// under locks before consuming it.
	newPasswordHash, err := HashPassword(newPassword)
	if err != nil {
		return 0, err
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin password reset: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var lockedUserID string
	var status string
	var emailVerified bool
	if err := tx.QueryRow(ctx, `
		SELECT id::text, status, email_verified_at IS NOT NULL
		FROM users
		WHERE id = $1
		FOR UPDATE
	`, userID).Scan(&lockedUserID, &status, &emailVerified); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrInvalidActionToken
		}
		return 0, fmt.Errorf("lock password-reset user: %w", err)
	}
	if status != "active" || !emailVerified {
		return 0, ErrAccountNotActive
	}

	var consumedTokenID string
	err = tx.QueryRow(ctx, `
		UPDATE auth_action_tokens
		SET consumed_at = now()
		WHERE id = $1
			AND user_id = $2
			AND token_digest = $3
			AND purpose = 'password_reset'
			AND consumed_at IS NULL
			AND revoked_at IS NULL
			AND expires_at > now()
		RETURNING id::text
	`, tokenID, lockedUserID, digest[:]).Scan(&consumedTokenID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrInvalidActionToken
		}
		return 0, fmt.Errorf("consume password-reset token: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE users
		SET password_hash = $1,
			password_changed_at = now(),
			updated_at = now()
		WHERE id = $2
	`, newPasswordHash, lockedUserID); err != nil {
		return 0, fmt.Errorf("store reset password: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE auth_action_tokens
		SET revoked_at = now()
		WHERE user_id = $1
			AND purpose = 'password_reset'
			AND id <> $2
			AND consumed_at IS NULL
			AND revoked_at IS NULL
	`, lockedUserID, consumedTokenID); err != nil {
		return 0, fmt.Errorf("revoke sibling password-reset tokens: %w", err)
	}

	revokedSessions, err := RevokeUserSessions(
		ctx, tx, lockedUserID, nil, SessionRevocationPasswordChanged,
	)
	if err != nil {
		return 0, fmt.Errorf("revoke sessions after password reset: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_events(
			event_type, subject_user_id, resource_type, resource_id, metadata
		)
		VALUES (
			'auth.password_reset.completed', $1, 'action_token', $2,
			'{"purpose":"password_reset","origin":"internal"}'::jsonb
		)
	`, lockedUserID, consumedTokenID); err != nil {
		return 0, fmt.Errorf("audit password-reset completion: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit password reset: %w", err)
	}
	return revokedSessions, nil
}
