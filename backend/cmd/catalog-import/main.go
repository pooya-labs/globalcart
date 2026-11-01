package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pooya16pg-spec/globalcart/backend/internal/catalog"
	"github.com/pooya16pg-spec/globalcart/backend/internal/config"
	"github.com/pooya16pg-spec/globalcart/backend/internal/database"
)

type snapshot struct {
	Products []catalog.Product `json:"products"`
}

func main() {
	file := flag.String("file", "../data/catalog.json", "path to normalized catalog snapshot JSON")
	flag.Parse()
	cfg, err := config.Load()
	if err != nil { log.Fatal(err) }
	raw, err := os.ReadFile(*file)
	if err != nil { log.Fatalf("read snapshot: %v", err) }
	var data snapshot
	if err := json.Unmarshal(raw, &data); err != nil { log.Fatalf("decode snapshot: %v", err) }
	if len(data.Products) == 0 { log.Fatal("refusing to import an empty catalog") }
	for i := range data.Products {
		p := &data.Products[i]
		if p.ID <= 0 || p.Title == "" || p.Category == "" || p.Source == "" || p.SourceID == "" || p.ImportedAt.IsZero() {
			log.Fatalf("invalid product at index %d (id=%d): required fields missing", i, p.ID)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil { log.Fatal(err) }
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil { log.Fatal(err) }
	if err := database.Migrate(ctx, pool); err != nil { log.Fatal(err) }
	if err := database.UpsertProducts(ctx, pool, data.Products); err != nil { log.Fatal(err) }
	fmt.Printf("Imported %d catalog products successfully.\n", len(data.Products))
}
