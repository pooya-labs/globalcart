package database

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pooya16pg-spec/globalcart/backend/internal/catalog"
)

func TestMigrationsAndCatalogUpsert(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set; PostgreSQL integration test skipped")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("create database pool: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("connect to test PostgreSQL: %v", err)
	}

	// Concurrent startup attempts must serialize and remain idempotent.
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- Migrate(ctx, pool)
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent migration: %v", err)
		}
	}

	product := catalog.Product{
		ID: 987654321, Title: "Integration fixture", Description: "Database test",
		Category: "test", Price: 12.34, Rating: 4.5, Stock: 3,
		Images: []string{"https://example.invalid/image.webp"},
		Source: "integration-test", SourceID: "fixture-1",
		ImportedAt: time.Now().UTC(),
	}
	if err := UpsertProducts(ctx, pool, []catalog.Product{product}); err != nil {
		t.Fatalf("upsert product: %v", err)
	}
	var title string
	if err := pool.QueryRow(ctx, "SELECT title FROM products WHERE id=$1", product.ID).Scan(&title); err != nil {
		t.Fatalf("read imported product: %v", err)
	}
	if title != product.Title {
		t.Fatalf("title = %q, want %q", title, product.Title)
	}

	// A failed item must roll back earlier writes in the same import batch.
	valid := product
	valid.ID = 987654322
	valid.SourceID = "fixture-2"
	invalid := product
	invalid.ID = 987654323
	invalid.SourceID = "fixture-3"
	invalid.Price = -1
	if err := UpsertProducts(ctx, pool, []catalog.Product{valid, invalid}); err == nil {
		t.Fatal("expected database price constraint to reject invalid product")
	}
	var exists bool
	if err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM products WHERE id=$1)", valid.ID).Scan(&exists); err != nil {
		t.Fatalf("check transaction rollback: %v", err)
	}
	if exists {
		t.Fatal("valid row from failed batch should have been rolled back")
	}
}
