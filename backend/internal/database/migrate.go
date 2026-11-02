package database

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	// Serialize startup migrations across API replicas and one-shot import jobs.
	lockConn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire migration lock connection: %w", err)
	}
	defer lockConn.Release()
	const migrationLockID int64 = 7142032601
	if _, err := lockConn.Exec(ctx, "SELECT pg_advisory_lock($1)", migrationLockID); err != nil {
		return fmt.Errorf("acquire migration advisory lock: %w", err)
	}
	defer func() {
		_, _ = lockConn.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", migrationLockID)
	}()
	if _, err := lockConn.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		checksum TEXT
	)`); err != nil {
		return fmt.Errorf("create migration ledger: %w", err)
	}
	// Older databases have the original two-column ledger. Add the checksum
	// column in place; NULL entries are adopted once from this immutable source
	// tree, then every subsequent startup verifies the exact SQL file bytes.
	if _, err := lockConn.Exec(ctx, "ALTER TABLE schema_migrations ADD COLUMN IF NOT EXISTS checksum TEXT"); err != nil {
		return fmt.Errorf("add migration checksum ledger column: %w", err)
	}

	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	known := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			names = append(names, entry.Name())
			known[entry.Name()] = struct{}{}
		}
	}
	sort.Strings(names)

	// An applied ledger entry without a corresponding embedded file means an
	// already-applied migration was renamed or removed. Fail closed rather
	// than silently continuing with an incomplete migration history.
	ledgerRows, err := lockConn.Query(ctx, "SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return fmt.Errorf("read migration ledger: %w", err)
	}
	for ledgerRows.Next() {
		var version string
		if err := ledgerRows.Scan(&version); err != nil {
			ledgerRows.Close()
			return fmt.Errorf("read migration ledger version: %w", err)
		}
		if _, ok := known[version]; !ok {
			ledgerRows.Close()
			return fmt.Errorf("migration ledger references missing embedded file %q; restore the original migration file instead of deleting or renaming applied migrations", version)
		}
	}
	if err := ledgerRows.Err(); err != nil {
		ledgerRows.Close()
		return fmt.Errorf("iterate migration ledger: %w", err)
	}
	ledgerRows.Close()

	for _, name := range names {
		sqlBytes, err := migrationFiles.ReadFile("migrations/" + name)
		if err != nil {
			return err
		}
		checksum := migrationChecksum(sqlBytes)

		var recordedChecksum pgtype.Text
		err = lockConn.QueryRow(ctx, "SELECT checksum FROM schema_migrations WHERE version=$1", name).Scan(&recordedChecksum)
		if err == nil {
			if recordedChecksum.Valid {
				if recordedChecksum.String != checksum {
					return fmt.Errorf("migration checksum mismatch for %s: embedded SQL does not match the applied file; restore the original migration or add a new forward-fix migration", name)
				}
			} else {
				// One-time compatibility path for ledger rows written before
				// checksums were introduced. The advisory lock serializes this
				// adoption with every other migration runner.
				tag, updateErr := lockConn.Exec(ctx, "UPDATE schema_migrations SET checksum=$2 WHERE version=$1 AND checksum IS NULL", name, checksum)
				if updateErr != nil {
					return fmt.Errorf("backfill checksum for legacy migration %s: %w", name, updateErr)
				}
				if tag.RowsAffected() != 1 {
					return fmt.Errorf("backfill checksum for legacy migration %s affected %d rows; refusing ambiguous ledger state", name, tag.RowsAffected())
				}
			}
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("read migration ledger entry %s: %w", name, err)
		}

		tx, err := lockConn.Begin(ctx)
		if err != nil {
			return err
		}
		// Execute statements individually because pgx extended-query mode does not
		// accept multi-command statements. splitSQLStatements preserves quoted SQL
		// and dollar-quoted PostgreSQL function bodies.
		statements, splitErr := splitSQLStatements(string(sqlBytes))
		if splitErr != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("split %s: %w", name, splitErr)
		}
		for _, statement := range statements {
			if _, err = tx.Exec(ctx, statement); err != nil {
				break
			}
		}
		if err == nil {
			_, err = tx.Exec(ctx, "INSERT INTO schema_migrations(version, checksum) VALUES($1, $2)", name, checksum)
		}
		if err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply %s: %w", name, err)
		}
		if err = tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit %s: %w", name, err)
		}
	}
	if _, err := lockConn.Exec(ctx, "ALTER TABLE schema_migrations ALTER COLUMN checksum SET NOT NULL"); err != nil {
		return fmt.Errorf("enforce migration checksum ledger: %w", err)
	}
	return nil
}
