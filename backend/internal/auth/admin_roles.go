package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const adminRoleMutationLock int64 = 0x474C43415254

var (
	ErrAdminAuthorizationRequired = errors.New("active administrator authorization required")
	ErrAdminRoleAlreadyGranted    = errors.New("administrator role is already active")
	ErrAdminRoleNotGranted        = errors.New("administrator role is not active")
	ErrSelfRoleChange             = errors.New("administrator role cannot be granted to or revoked by the same user")
)

type roleChangeUserState struct {
	id            string
	status        string
	emailVerified bool
}

// beginAdminRoleChange serializes privileged-role mutations and locks the
// actor/target account rows in stable UUID order. This prevents competing
// admin grant/revoke requests from deadlocking or both using stale actor roles.
func beginAdminRoleChange(
	ctx context.Context,
	pool *pgxpool.Pool,
	actorUserID string,
	targetUserID string,
) (pgx.Tx, roleChangeUserState, roleChangeUserState, error) {
	var actor, target roleChangeUserState
	if actorUserID == targetUserID {
		return nil, actor, target, ErrSelfRoleChange
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, actor, target, fmt.Errorf("begin administrator role change: %w", err)
	}
	fail := func(err error) (pgx.Tx, roleChangeUserState, roleChangeUserState, error) {
		_ = tx.Rollback(ctx)
		return nil, roleChangeUserState{}, roleChangeUserState{}, err
	}

	if _, err := tx.Exec(ctx, `
		SELECT pg_advisory_xact_lock($1::bigint)
	`, adminRoleMutationLock); err != nil {
		return fail(fmt.Errorf("serialize administrator role changes: %w", err))
	}

	ids := []string{actorUserID, targetUserID}
	if ids[1] < ids[0] {
		ids[0], ids[1] = ids[1], ids[0]
	}
	states := make(map[string]roleChangeUserState, 2)
	for _, id := range ids {
		var state roleChangeUserState
		err := tx.QueryRow(ctx, `
			SELECT id::text, status, email_verified_at IS NOT NULL
			FROM users
			WHERE id = $1
			FOR UPDATE
		`, id).Scan(&state.id, &state.status, &state.emailVerified)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				if id == actorUserID {
					return fail(ErrAdminAuthorizationRequired)
				}
				return fail(ErrUserNotFound)
			}
			return fail(fmt.Errorf("lock role-change account: %w", err))
		}
		states[id] = state
	}
	actor = states[actorUserID]
	target = states[targetUserID]

	var actorIsAdmin bool
	if actor.status == "active" && actor.emailVerified {
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM user_roles
				WHERE user_id = $1
					AND role_key = 'admin'
					AND revoked_at IS NULL
			)
		`, actorUserID).Scan(&actorIsAdmin); err != nil {
			return fail(fmt.Errorf("check administrator actor role: %w", err))
		}
	}
	if !actorIsAdmin {
		return fail(ErrAdminAuthorizationRequired)
	}
	return tx, actor, target, nil
}

// GrantAdminRole grants the privileged administrator role to another active,
// email-verified account. Only an already active administrator may perform the
// operation. The target's existing sessions are revoked in the same transaction
// so the privilege change cannot leave an old session alive.
func GrantAdminRole(
	ctx context.Context,
	pool *pgxpool.Pool,
	actorUserID string,
	targetUserID string,
) (int64, error) {
	tx, _, target, err := beginAdminRoleChange(ctx, pool, actorUserID, targetUserID)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if target.status != "active" || !target.emailVerified {
		return 0, ErrAccountNotActive
	}

	var alreadyGranted bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM user_roles
			WHERE user_id = $1 AND role_key = 'admin' AND revoked_at IS NULL
		)
	`, targetUserID).Scan(&alreadyGranted); err != nil {
		return 0, fmt.Errorf("check existing administrator grant: %w", err)
	}
	if alreadyGranted {
		return 0, ErrAdminRoleAlreadyGranted
	}

	var grantID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO user_roles(user_id, role_key, granted_by_user_id)
		VALUES ($1, 'admin', $2)
		RETURNING id::text
	`, targetUserID, actorUserID).Scan(&grantID); err != nil {
		return 0, fmt.Errorf("grant administrator role: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_events(
			event_type, actor_user_id, subject_user_id,
			resource_type, resource_id, metadata
		)
		VALUES (
			'auth.role.granted', $1, $2, 'user_role', $3,
			'{"role":"admin","origin":"internal"}'::jsonb
		)
	`, actorUserID, targetUserID, grantID); err != nil {
		return 0, fmt.Errorf("audit administrator role grant: %w", err)
	}

	actor := actorUserID
	revokedSessions, err := RevokeUserSessions(
		ctx, tx, targetUserID, &actor, SessionRevocationPrivilegeChanged,
	)
	if err != nil {
		return 0, fmt.Errorf("revoke target sessions after admin grant: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit administrator role grant: %w", err)
	}
	return revokedSessions, nil
}

// RevokeAdminRole revokes a target administrator's active grant. A different
// active administrator must authorize the operation; self-revocation is
// rejected, ensuring an application-level role change cannot remove its own
// authorizing principal. The target's sessions are invalidated atomically.
func RevokeAdminRole(
	ctx context.Context,
	pool *pgxpool.Pool,
	actorUserID string,
	targetUserID string,
) (int64, error) {
	tx, _, _, err := beginAdminRoleChange(ctx, pool, actorUserID, targetUserID)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var revokedGrantID string
	err = tx.QueryRow(ctx, `
		UPDATE user_roles
		SET revoked_at = now()
		WHERE user_id = $1
			AND role_key = 'admin'
			AND revoked_at IS NULL
		RETURNING id::text
	`, targetUserID).Scan(&revokedGrantID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrAdminRoleNotGranted
		}
		return 0, fmt.Errorf("revoke administrator role: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_events(
			event_type, actor_user_id, subject_user_id,
			resource_type, resource_id, metadata
		)
		VALUES (
			'auth.role.revoked', $1, $2, 'user_role', $3,
			'{"role":"admin","origin":"internal"}'::jsonb
		)
	`, actorUserID, targetUserID, revokedGrantID); err != nil {
		return 0, fmt.Errorf("audit administrator role revocation: %w", err)
	}

	actor := actorUserID
	revokedSessions, err := RevokeUserSessions(
		ctx, tx, targetUserID, &actor, SessionRevocationPrivilegeChanged,
	)
	if err != nil {
		return 0, fmt.Errorf("revoke target sessions after admin revocation: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit administrator role revocation: %w", err)
	}
	return revokedSessions, nil
}
