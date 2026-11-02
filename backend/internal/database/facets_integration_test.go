package database

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pooya16pg-spec/globalcart/backend/internal/catalog"
	"github.com/pooya16pg-spec/globalcart/backend/internal/dbgen"
)

func TestCatalogFacetFiltersAndCountStayConsistent(t *testing.T) {
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
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("connect to test PostgreSQL: %v", err)
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	baseID := time.Now().UnixNano()%2_000_000_000_000_000 + 3_000_000_000_000_000
	importedAt := time.Now().UTC()
	products := []catalog.Product{
		{
			ID: baseID + 1, Title: "Aurora Headphones", Description: "Studio over-ear audio",
			Category: "Audio", Price: 49.99, Rating: 4.8, Stock: 7, Brand: "Auralab",
			Thumbnail: "https://example.invalid/aurora.webp", Images: []string{},
			Source: "facet-integration", SourceID: fmt.Sprintf("facet-%d-a", baseID), ImportedAt: importedAt,
		},
		{
			ID: baseID + 2, Title: "Auralab Speaker", Description: "Wireless audio speaker",
			Category: "Audio", Price: 129.99, Rating: 4.2, Stock: 4, Brand: "Auralab",
			Thumbnail: "https://example.invalid/speaker.webp", Images: []string{},
			Source: "facet-integration", SourceID: fmt.Sprintf("facet-%d-b", baseID), ImportedAt: importedAt,
		},
		{
			ID: baseID + 3, Title: "Budget Headphones", Description: "Entry audio product",
			Category: "Audio", Price: 29.99, Rating: 3.5, Stock: 12, Brand: "Auralab",
			Thumbnail: "https://example.invalid/budget.webp", Images: []string{},
			Source: "facet-integration", SourceID: fmt.Sprintf("facet-%d-c", baseID), ImportedAt: importedAt,
		},
		{
			ID: baseID + 4, Title: "Studio Keyboard", Description: "Mechanical typing accessory",
			Category: "Accessories", Price: 89.50, Rating: 4.9, Stock: 3, Brand: "Keyworks",
			Thumbnail: "https://example.invalid/keyboard.webp", Images: []string{},
			Source: "facet-integration", SourceID: fmt.Sprintf("facet-%d-d", baseID), ImportedAt: importedAt,
		},
	}
	if err := UpsertProducts(ctx, pool, products); err != nil {
		t.Fatalf("import facet fixtures: %v", err)
	}

	queries := dbgen.New(pool)
	filtered := dbgen.ListProductsParams{
		Category: " audio ", Query: "headphones", Brand: " AURALAB ",
		MinPrice: "30.00", MaxPrice: "100.00", MinRating: "4.00", MaxRating: "5.00",
		PageLimit: 10, PageOffset: 0,
	}
	rows, err := queries.ListProducts(ctx, filtered)
	if err != nil {
		t.Fatalf("list filtered products: %v", err)
	}
	if len(rows) != 1 || rows[0].ID != baseID+1 {
		t.Fatalf("filtered product IDs = %v, want only %d", productIDs(rows), baseID+1)
	}
	count, err := queries.CountProducts(ctx, dbgen.CountProductsParams{
		Category: filtered.Category, Query: filtered.Query, Brand: filtered.Brand,
		MinPrice: filtered.MinPrice, MaxPrice: filtered.MaxPrice,
		MinRating: filtered.MinRating, MaxRating: filtered.MaxRating,
	})
	if err != nil {
		t.Fatalf("count filtered products: %v", err)
	}
	if count != 1 {
		t.Fatalf("filtered total = %d, want 1", count)
	}

	// Total remains independent of the current page size.
	broad := dbgen.ListProductsParams{
		Brand: "auralab", MinPrice: "30", MaxPrice: "130", MinRating: "4",
		PageLimit: 1, PageOffset: 0,
	}
	page, err := queries.ListProducts(ctx, broad)
	if err != nil {
		t.Fatalf("list first facet page: %v", err)
	}
	if len(page) != 1 {
		t.Fatalf("page count = %d, want 1", len(page))
	}
	total, err := queries.CountProducts(ctx, dbgen.CountProductsParams{
		Category: broad.Category, Query: broad.Query, Brand: broad.Brand,
		MinPrice: broad.MinPrice, MaxPrice: broad.MaxPrice,
		MinRating: broad.MinRating, MaxRating: broad.MaxRating,
	})
	if err != nil {
		t.Fatalf("count all facet matches: %v", err)
	}
	if total != 2 {
		t.Fatalf("filtered total = %d, want 2 independent of page size", total)
	}
}

func productIDs(rows []dbgen.ListProductsRow) []int64 {
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids
}
