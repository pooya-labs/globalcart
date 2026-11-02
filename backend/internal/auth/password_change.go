package auth

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ChangePassword updates an active verified account's password and revokes
// every existing session atomically. expectedCurrentPasswordHash is the exact
// credential snapshot whose password the caller has already verified; the
// transaction re-checks it while holding the account row lock so an outdated
// password check cannot overwrite a newer password.
func ChangePassword(
	ctx context.Context,
	pool *pgxpool.Pool,
	userID string,
	expectedCurrentPasswordHash string,
	newPassword string,
	actorUserID *string,
) (int64, error) {
	newPasswordHash, err := HashPassword(newPassword)
	if err != nil {
		return 0, err
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin password change: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var lockedUserID, currentPasswordHash string
	if err := tx.QueryRow(ctx, `
		SELECT id::text, password_hash
		FROM users
		WHERE id = $1 AND status = 'active' AND email_verified_at IS NOT NULL
		FOR UPDATE
	`, userID).Scan(&lockedUserID, &currentPasswordHash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrAccountNotActive
		}
		return 0, fmt.Errorf("lock password owner: %w", err)
	}
	if expectedCurrentPasswordHash == "" ||
		subtle.ConstantTimeCompare([]byte(currentPasswordHash), []byte(expectedCurrentPasswordHash)) != 1 {
		return 0, ErrCredentialsChanged
	}

	if _, err := tx.Exec(ctx, `
		UPDATE users
		SET password_hash = $1,
			password_changed_at = now(),
			updated_at = now()
		WHERE id = $2
	`, newPasswordHash, lockedUserID); err != nil {
		return 0, fmt.Errorf("update account password: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE auth_action_tokens
		SET revoked_at = now()
		WHERE user_id = $1
			AND purpose = 'password_reset'
			AND consumed_at IS NULL
			AND revoked_at IS NULL
	`, lockedUserID); err != nil {
		return 0, fmt.Errorf("revoke stale password-reset tokens: %w", err)
	}

	revokedCount, err := RevokeUserSessions(
		ctx, tx, lockedUserID, actorUserID, SessionRevocationPasswordChanged,
	)
	if err != nil {
		return 0, fmt.Errorf("revoke sessions after password change: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_events(
			event_type, actor_user_id, subject_user_id,
			resource_type, resource_id, metadata
		)
		VALUES (
			'auth.password.changed', $1, $2, 'account', $3,
			'{"origin":"internal"}'::jsonb
		)
	`, actorUserID, lockedUserID, lockedUserID); err != nil {
		return 0, fmt.Errorf("audit password change: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit password change: %w", err)
	}
	return revokedCount, nil
}
