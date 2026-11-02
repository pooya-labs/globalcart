package database

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"

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
		applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`); err != nil {
		return fmt.Errorf("create migration ledger: %w", err)
	}
	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	for _, name := range names {
		var exists bool
		if err := lockConn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)", name).Scan(&exists); err != nil {
			return err
		}
		if exists {
			continue
		}
		sqlBytes, err := migrationFiles.ReadFile("migrations/" + name)
		if err != nil {
			return err
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
			_, err = tx.Exec(ctx, "INSERT INTO schema_migrations(version) VALUES($1)", name)
		}
		if err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply %s: %w", name, err)
		}
		if err = tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit %s: %w", name, err)
		}
	}
	return nil
}
