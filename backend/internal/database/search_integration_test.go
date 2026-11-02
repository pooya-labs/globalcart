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

func TestCatalogSearchTrigramFallbackAndExactMatchRanking(t *testing.T) {
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

	baseID := time.Now().UnixNano()%2_000_000_000_000_000 + 3_500_000_000_000_000
	importedAt := time.Now().UTC()
	products := []catalog.Product{
		{
			ID: baseID + 1, Title: "Wireless Headphones", Description: "Over-ear audio with clear sound",
			Category: "Audio", Price: 79.99, Rating: 4.8, Stock: 8, Brand: "Auralab",
			Thumbnail: "https://example.invalid/headphones.webp", Images: []string{},
			Source: "trigram-integration", SourceID: fmt.Sprintf("trigram-%d-a", baseID), ImportedAt: importedAt,
		},
		{
			ID: baseID + 2, Title: "Headphone Case", Description: "Protective travel case for electronics",
			Category: "Accessories", Price: 19.99, Rating: 4.1, Stock: 5, Brand: "CaseCo",
			Thumbnail: "https://example.invalid/hedphones.webp", Images: []string{},
			Source: "trigram-integration", SourceID: fmt.Sprintf("trigram-%d-b", baseID), ImportedAt: importedAt,
		},
		{
			ID: baseID + 3, Title: "Portable Speaker", Description: "Compact wireless speaker",
			Category: "Audio", Price: 39.99, Rating: 4.3, Stock: 12, Brand: "SoundWorks",
			Thumbnail: "https://example.invalid/speaker.webp", Images: []string{},
			Source: "trigram-integration", SourceID: fmt.Sprintf("trigram-%d-c", baseID), ImportedAt: importedAt,
		},
	}
	if err := UpsertProducts(ctx, pool, products); err != nil {
		t.Fatalf("import search fixtures: %v", err)
	}

	queries := dbgen.New(pool)
	exactRows, err := queries.ListProducts(ctx, dbgen.ListProductsParams{
		Query: "headphones", PageLimit: 10, PageOffset: 0,
	})
	if err != nil {
		t.Fatalf("list exact full-text matches: %v", err)
	}
	if len(exactRows) < 2 || exactRows[0].ID != baseID+1 {
		t.Fatalf("exact search ranking = %v; expected exact full-text product %d first", searchProductIDs(exactRows), baseID+1)
	}

	fuzzyRows, err := queries.ListProducts(ctx, dbgen.ListProductsParams{
		Query: "headphons", PageLimit: 10, PageOffset: 0,
	})
	if err != nil {
		t.Fatalf("list typo-tolerant matches: %v", err)
	}
	if !containsProductID(fuzzyRows, baseID+1) {
		t.Fatalf("fuzzy search results = %v; expected typo-tolerant match %d", searchProductIDs(fuzzyRows), baseID+1)
	}

	fuzzyTotal, err := queries.CountProducts(ctx, dbgen.CountProductsParams{Query: "headphons"})
	if err != nil {
		t.Fatalf("count typo-tolerant matches: %v", err)
	}
	if fuzzyTotal != int64(len(fuzzyRows)) {
		t.Fatalf("fuzzy total = %d, page result count = %d", fuzzyTotal, len(fuzzyRows))
	}

	brandTypoRows, err := queries.ListProducts(ctx, dbgen.ListProductsParams{
		Query: "auralba", PageLimit: 10, PageOffset: 0,
	})
	if err != nil {
		t.Fatalf("list fuzzy brand matches: %v", err)
	}
	if len(brandTypoRows) != 1 || brandTypoRows[0].ID != baseID+1 {
		t.Fatalf("fuzzy brand results = %v; expected only %d", searchProductIDs(brandTypoRows), baseID+1)
	}
}

func containsProductID(rows []dbgen.ListProductsRow, wanted int64) bool {
	for _, row := range rows {
		if row.ID == wanted {
			return true
		}
	}
	return false
}

func searchProductIDs(rows []dbgen.ListProductsRow) []int64 {
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids
}
