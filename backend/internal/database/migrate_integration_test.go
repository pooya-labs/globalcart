package database

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pooya16pg-spec/globalcart/backend/internal/catalog"
	"github.com/pooya16pg-spec/globalcart/backend/internal/commerce"
	"github.com/pooya16pg-spec/globalcart/backend/internal/dbgen"
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

	// The catalog extension backfills a normalized category and supports
	// constrained product variants without changing the source product record.
	var categoryID int64
	if err := pool.QueryRow(ctx, "SELECT category_id FROM products WHERE id=$1", product.ID).Scan(&categoryID); err != nil {
		t.Fatalf("read backfilled category: %v", err)
	}
	var categoryName string
	if err := pool.QueryRow(ctx, "SELECT name FROM categories WHERE id=$1", categoryID).Scan(&categoryName); err != nil {
		t.Fatalf("read category: %v", err)
	}
	if categoryName != product.Category {
		t.Fatalf("category = %q, want %q", categoryName, product.Category)
	}

	// Exercise the generated category query against PostgreSQL.
	categoryRows, err := dbgen.New(pool).ListCategories(ctx)
	if err != nil {
		t.Fatalf("list categories through sqlc: %v", err)
	}
	foundCategory := false
	for _, row := range categoryRows {
		if row.ID == categoryID {
			foundCategory = true
			if row.Name != product.Category || row.ProductCount < 1 {
				t.Fatalf("unexpected category row: %+v", row)
			}
		}
	}
	if !foundCategory {
		t.Fatalf("category %d missing from generated listing", categoryID)
	}
	// Category filtering is normalized and total counts are independent of page size.
	catalogQueries := dbgen.New(pool)
	filtered, err := catalogQueries.ListProducts(ctx, dbgen.ListProductsParams{
		Category: "  TEST ", Query: "", PageLimit: 1, PageOffset: 0,
	})
	if err != nil {
		t.Fatalf("list products by normalized category: %v", err)
	}
	if len(filtered) != 1 || filtered[0].ID != product.ID {
		t.Fatalf("normalized category filter returned unexpected products: %+v", filtered)
	}
	total, err := catalogQueries.CountProducts(ctx, dbgen.CountProductsParams{
		Category: "tEsT", Query: "",
	})
	if err != nil {
		t.Fatalf("count products by normalized category: %v", err)
	}
	if total < 1 {
		t.Fatalf("normalized category count = %d, want at least one", total)
	}
	variantAttributes := []byte(`{"color":"black","size":"M"}`)
	var variantID int64
	if err := pool.QueryRow(ctx, `INSERT INTO product_variants
		(product_id,sku,title,price,stock,attributes,source,source_id)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		product.ID, "GC-INT-987654321-BLK-M", "Black / Medium", 13.99, 2,
		variantAttributes, "integration-test", "variant-1").Scan(&variantID); err != nil {
		t.Fatalf("insert product variant: %v", err)
	}
	var storedAttributes []byte
	if err := pool.QueryRow(ctx, "SELECT attributes FROM product_variants WHERE id=$1", variantID).Scan(&storedAttributes); err != nil {
		t.Fatalf("read product variant attributes: %v", err)
	}
	var decodedAttributes map[string]string
	if err := json.Unmarshal(storedAttributes, &decodedAttributes); err != nil {
		t.Fatalf("decode stored variant attributes: %v", err)
	}
	if decodedAttributes["color"] != "black" || decodedAttributes["size"] != "M" {
		t.Fatalf("variant attributes = %v, want color=black and size=M", decodedAttributes)
	}

	// Exercise the demo commerce schema. Tokens and idempotency keys are stored
	// only as fixed-size digests, and order lines retain immutable price/title snapshots.
	ownerDigest := make([]byte, 32)
	ownerDigest[0] = 1
	var cartID string
	if err := pool.QueryRow(ctx, `INSERT INTO carts(owner_token_hash, currency, expires_at)
		VALUES($1, 'USD', now() + interval '1 hour') RETURNING id`, ownerDigest).Scan(&cartID); err != nil {
		t.Fatalf("create demo cart: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO cart_items
		(cart_id, product_id, variant_id, quantity, unit_price_snapshot, currency, product_title_snapshot, variant_title_snapshot)
		VALUES($1, $2, $3, 2, 13.99, 'USD', 'Integration fixture', 'Black / Medium')`,
		cartID, product.ID, variantID); err != nil {
		t.Fatalf("insert cart item: %v", err)
	}
	idempotencyDigest := make([]byte, 32)
	idempotencyDigest[0] = 2
	var orderID string
	if err := pool.QueryRow(ctx, `INSERT INTO demo_orders
		(cart_id, idempotency_key_hash, currency, subtotal, total)
		VALUES($1, $2, 'USD', 27.98, 27.98) RETURNING id`,
		cartID, idempotencyDigest).Scan(&orderID); err != nil {
		t.Fatalf("create demo order: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO demo_order_items
		(order_id, product_id, variant_id, quantity, unit_price_snapshot, currency, product_title_snapshot, variant_title_snapshot)
		VALUES($1, $2, $3, 2, 13.99, 'USD', 'Integration fixture', 'Black / Medium')`,
		orderID, product.ID, variantID); err != nil {
		t.Fatalf("insert immutable order item snapshot: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO inventory_reservations
		(order_id, product_id, variant_id, quantity, expires_at)
		VALUES($1, $2, $3, 2, now() + interval '15 minutes')`,
		orderID, product.ID, variantID); err != nil {
		t.Fatalf("create inventory reservation: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO demo_order_events(order_id, event_type, payload)
		VALUES($1, 'demo_order_created', '{"source":"integration-test"}'::jsonb)`, orderID); err != nil {
		t.Fatalf("append demo order event: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO demo_orders
		(cart_id, idempotency_key_hash, currency, subtotal, total)
		VALUES($1, $2, 'USD', 27.98, 27.98)`, cartID, idempotencyDigest); err == nil {
		t.Fatal("duplicate cart idempotency key should be rejected")
	}
	if _, err := pool.Exec(ctx, `INSERT INTO cart_items
		(cart_id, product_id, variant_id, quantity, unit_price_snapshot, currency, product_title_snapshot)
		VALUES($1, $2, $3, 0, 13.99, 'USD', 'Invalid quantity')`,
		cartID, product.ID, variantID); err == nil {
		t.Fatal("cart quantity zero should be rejected")
	}

	// Two guest carts compete for three units of the same base product.
	// Row locking plus active reservations must allow only one 2-unit checkout.
	type checkoutAttempt struct {
		owner [32]byte
		key [32]byte
		order commerce.Order
		err error
	}
	attempts := make([]checkoutAttempt, 2)
	for i := range attempts {
		attempts[i].owner[0] = byte(10 + i)
		attempts[i].key[0] = byte(20 + i)
		var newCartID string
		if err := pool.QueryRow(ctx, `INSERT INTO carts(owner_token_hash,currency,expires_at)
			VALUES($1,'USD',now()+interval '1 hour') RETURNING id::text`, attempts[i].owner[:]).Scan(&newCartID); err != nil {
			t.Fatalf("create competing cart %d: %v", i, err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO cart_items
			(cart_id,product_id,quantity,unit_price_snapshot,currency,product_title_snapshot)
			VALUES($1,$2,2,12.34,'USD','Integration fixture')`, newCartID, product.ID); err != nil {
			t.Fatalf("add competing cart item %d: %v", i, err)
		}
	}
	service := commerce.New(pool)
	var checkoutWG sync.WaitGroup
	for i := range attempts {
		checkoutWG.Add(1)
		go func(i int) {
			defer checkoutWG.Done()
			attempts[i].order, attempts[i].err = service.Checkout(ctx, attempts[i].owner[:], attempts[i].key[:], 15*time.Minute)
		}(i)
	}
	checkoutWG.Wait()
	successes, stockFailures := 0, 0
	winner := -1
	for i := range attempts {
		switch {
		case attempts[i].err == nil:
			successes++
			winner = i
		case errors.Is(attempts[i].err, commerce.ErrInsufficientStock):
			stockFailures++
		default:
			t.Fatalf("unexpected checkout result %d: order=%+v err=%v", i, attempts[i].order, attempts[i].err)
		}
	}
	if successes != 1 || stockFailures != 1 {
		t.Fatalf("concurrent checkout outcomes: successes=%d stock failures=%d, want 1 each", successes, stockFailures)
	}
	replayed, err := service.Checkout(ctx, attempts[winner].owner[:], attempts[winner].key[:], 15*time.Minute)
	if err != nil {
		t.Fatalf("replay successful checkout: %v", err)
	}
	if !replayed.Replayed || replayed.ID != attempts[winner].order.ID {
		t.Fatalf("idempotent replay = %+v, original = %+v", replayed, attempts[winner].order)
	}
	var winnerReservations int
	if err := pool.QueryRow(ctx, `SELECT COALESCE(sum(quantity),0)::int
		FROM inventory_reservations WHERE order_id=$1 AND status='active'`, attempts[winner].order.ID).Scan(&winnerReservations); err != nil {
		t.Fatalf("read winner reservations: %v", err)
	}
	if winnerReservations != 2 {
		t.Fatalf("winner reserved %d units, want 2", winnerReservations)
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
