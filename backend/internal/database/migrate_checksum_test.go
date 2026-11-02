package database

import (
	"context"
	"encoding/hex"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestMigrationChecksumUsesExactEmbeddedFileBytes(t *testing.T) {
	contents := []byte("-- migration fixture\nSELECT 1;\n")
	checksum := migrationChecksum(contents)
	if len(checksum) != 64 {
		t.Fatalf("checksum length = %d, want 64 hexadecimal characters", len(checksum))
	}
	decoded, err := hex.DecodeString(checksum)
	if err != nil || len(decoded) != 32 {
		t.Fatalf("checksum is not a SHA-256 hex digest: %q (%v)", checksum, err)
	}
	if got := migrationChecksum(append([]byte(nil), contents...)); got != checksum {
		t.Fatalf("same migration bytes produced different checksums: %q != %q", got, checksum)
	}
	if got := migrationChecksum(append(contents, '\n')); got == checksum {
		t.Fatal("changing the exact migration bytes did not change its checksum")
	}
}

func TestMigrateBackfillsLegacyChecksumsAndRejectsAppliedFileDrift(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set; PostgreSQL integration test skipped")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("create database pool: %v", err)
	}
	var migrationName, expectedChecksum string
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if migrationName != "" && expectedChecksum != "" {
			_, _ = pool.Exec(cleanupCtx, "UPDATE schema_migrations SET checksum=$2 WHERE version=$1", migrationName, expectedChecksum)
		}
		pool.Close()
	})
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("connect to test PostgreSQL: %v", err)
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("apply current migrations: %v", err)
	}

	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		t.Fatalf("read embedded migrations: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("no embedded migration files found")
	}
	migrationName = entries[0].Name()
	contents, err := migrationFiles.ReadFile("migrations/" + migrationName)
	if err != nil {
		t.Fatalf("read migration %s: %v", migrationName, err)
	}
	expectedChecksum = migrationChecksum(contents)

	// Simulate a database whose ledger predates checksum tracking.
	if _, err := pool.Exec(ctx, "UPDATE schema_migrations SET checksum=NULL WHERE version=$1", migrationName); err != nil {
		t.Fatalf("clear checksum to simulate legacy ledger: %v", err)
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("backfill legacy migration checksum: %v", err)
	}
	var recorded string
	if err := pool.QueryRow(ctx, "SELECT checksum FROM schema_migrations WHERE version=$1", migrationName).Scan(&recorded); err != nil {
		t.Fatalf("read backfilled migration checksum: %v", err)
	}
	if recorded != expectedChecksum {
		t.Fatalf("backfilled checksum = %q, want %q", recorded, expectedChecksum)
	}

	wrongChecksum := strings.Repeat("0", 64)
	if wrongChecksum == expectedChecksum {
		wrongChecksum = strings.Repeat("f", 64)
	}
	if _, err := pool.Exec(ctx, "UPDATE schema_migrations SET checksum=$2 WHERE version=$1", migrationName, wrongChecksum); err != nil {
		t.Fatalf("tamper with migration checksum for regression test: %v", err)
	}
	if err := Migrate(ctx, pool); err == nil || !strings.Contains(err.Error(), "migration checksum mismatch") {
		t.Fatalf("Migrate error = %v, want an applied migration checksum mismatch", err)
	}
}
