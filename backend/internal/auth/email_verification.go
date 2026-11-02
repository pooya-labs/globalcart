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
	DefaultEmailVerificationTTL = 24 * time.Hour
	MinimumEmailVerificationTTL = 5 * time.Minute
	MaximumEmailVerificationTTL = 72 * time.Hour
)

var (
	ErrAccountNotPendingVerification = errors.New("account is not pending email verification")
	ErrInvalidEmailVerificationTTL   = errors.New("email-verification token TTL must be between 5 minutes and 72 hours")
)

type IssuedEmailVerificationToken struct {
	ID        string
	Token     string
	ExpiresAt time.Time
}

func normalizeEmailVerificationTTL(ttl time.Duration) (time.Duration, error) {
	if ttl == 0 {
		return DefaultEmailVerificationTTL, nil
	}
	if ttl < MinimumEmailVerificationTTL || ttl > MaximumEmailVerificationTTL {
		return 0, ErrInvalidEmailVerificationTTL
	}
	return ttl, nil
}

// IssueEmailVerificationToken creates a short-lived one-time token for a
// pending account. Only its SHA-256 digest is stored. Issuing a replacement
// revokes all previous unconsumed verification tokens for this user.
// Delivery is intentionally left to a separately approved email worker.
func IssueEmailVerificationToken(
	ctx context.Context,
	pool *pgxpool.Pool,
	userID string,
	ttl time.Duration,
) (IssuedEmailVerificationToken, error) {
	var issued IssuedEmailVerificationToken
	ttl, err := normalizeEmailVerificationTTL(ttl)
	if err != nil {
		return issued, err
	}

	rawToken, err := NewOpaqueToken()
	if err != nil {
		return issued, fmt.Errorf("generate email-verification token: %w", err)
	}
	digest, err := DigestOpaqueToken(rawToken)
	if err != nil {
		return issued, fmt.Errorf("digest email-verification token: %w", err)
	}
	defer clear(digest[:])

	tx, err := pool.Begin(ctx)
	if err != nil {
		return issued, fmt.Errorf("begin email-verification token issue: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var lockedUserID string
	if err := tx.QueryRow(ctx, `
		SELECT id::text
		FROM users
		WHERE id = $1
			AND status = 'pending_verification'
			AND email_verified_at IS NULL
		FOR UPDATE
	`, userID).Scan(&lockedUserID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return issued, ErrAccountNotPendingVerification
		}
		return issued, fmt.Errorf("lock pending verification account: %w", err)
	}

	// Revoke expired as well as valid unconsumed tokens before inserting a
	// replacement; the partial unique index includes expired rows.
	if _, err := tx.Exec(ctx, `
		UPDATE auth_action_tokens
		SET revoked_at = now()
		WHERE user_id = $1
			AND purpose = 'email_verification'
			AND consumed_at IS NULL
			AND revoked_at IS NULL
	`, lockedUserID); err != nil {
		return issued, fmt.Errorf("revoke previous email-verification tokens: %w", err)
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO auth_action_tokens(user_id, purpose, token_digest, expires_at)
		VALUES ($1, 'email_verification', $2, now() + ($3::bigint * INTERVAL '1 second'))
		RETURNING id::text, expires_at
	`, lockedUserID, digest[:], int64(ttl/time.Second)).Scan(&issued.ID, &issued.ExpiresAt)
	if err != nil {
		return IssuedEmailVerificationToken{}, fmt.Errorf("insert email-verification token: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_events(
			event_type, subject_user_id, resource_type, resource_id, metadata
		)
		VALUES (
			'auth.email_verification.issued', $1, 'action_token', $2,
			'{"purpose":"email_verification","origin":"internal"}'::jsonb
		)
	`, lockedUserID, issued.ID); err != nil {
		return IssuedEmailVerificationToken{}, fmt.Errorf("audit email-verification issue: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return IssuedEmailVerificationToken{}, fmt.Errorf("commit email-verification token issue: %w", err)
	}
	issued.Token = rawToken
	return issued, nil
}

// CompleteEmailVerification consumes a one-time token and activates only its
// matching pending account. The account row is locked before token rows, which
// matches issuance/password mutation lock order and serializes concurrent
// verification or token rotation. New users retain only the database-triggered
// customer role; verification never grants seller/admin permissions.
func CompleteEmailVerification(
	ctx context.Context,
	pool *pgxpool.Pool,
	rawToken string,
) error {
	digest, err := DigestOpaqueToken(rawToken)
	if err != nil {
		return ErrInvalidActionToken
	}
	defer clear(digest[:])

	var tokenID, userID string
	err = pool.QueryRow(ctx, `
		SELECT id::text, user_id::text
		FROM auth_action_tokens
		WHERE token_digest = $1
			AND purpose = 'email_verification'
			AND consumed_at IS NULL
			AND revoked_at IS NULL
			AND expires_at > now()
	`, digest[:]).Scan(&tokenID, &userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidActionToken
		}
		return fmt.Errorf("find email-verification token: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin email verification: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var lockedUserID, status string
	var emailVerified bool
	if err := tx.QueryRow(ctx, `
		SELECT id::text, status, email_verified_at IS NOT NULL
		FROM users
		WHERE id = $1
		FOR UPDATE
	`, userID).Scan(&lockedUserID, &status, &emailVerified); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidActionToken
		}
		return fmt.Errorf("lock email-verification account: %w", err)
	}
	if status != "pending_verification" || emailVerified {
		return ErrInvalidActionToken
	}

	var consumedTokenID string
	err = tx.QueryRow(ctx, `
		UPDATE auth_action_tokens
		SET consumed_at = now()
		WHERE id = $1
			AND user_id = $2
			AND token_digest = $3
			AND purpose = 'email_verification'
			AND consumed_at IS NULL
			AND revoked_at IS NULL
			AND expires_at > now()
		RETURNING id::text
	`, tokenID, lockedUserID, digest[:]).Scan(&consumedTokenID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidActionToken
		}
		return fmt.Errorf("consume email-verification token: %w", err)
	}

	result, err := tx.Exec(ctx, `
		UPDATE users
		SET status = 'active',
			email_verified_at = now(),
			updated_at = now()
		WHERE id = $1
			AND status = 'pending_verification'
			AND email_verified_at IS NULL
	`, lockedUserID)
	if err != nil {
		return fmt.Errorf("activate verified account: %w", err)
	}
	if result.RowsAffected() != 1 {
		return ErrInvalidActionToken
	}

	if _, err := tx.Exec(ctx, `
		UPDATE auth_action_tokens
		SET revoked_at = now()
		WHERE user_id = $1
			AND purpose = 'email_verification'
			AND id <> $2
			AND consumed_at IS NULL
			AND revoked_at IS NULL
	`, lockedUserID, consumedTokenID); err != nil {
		return fmt.Errorf("revoke sibling email-verification tokens: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_events(
			event_type, subject_user_id, resource_type, resource_id, metadata
		)
		VALUES (
			'auth.email_verification.completed', $1, 'action_token', $2,
			'{"purpose":"email_verification","origin":"internal"}'::jsonb
		)
	`, lockedUserID, consumedTokenID); err != nil {
		return fmt.Errorf("audit email-verification completion: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit email verification: %w", err)
	}
	return nil
}
