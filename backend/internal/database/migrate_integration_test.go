package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

	// Migration 005 backfills the legacy catalog into an explicit platform vendor.
	var platformVendorID int64
	if err := pool.QueryRow(ctx, "SELECT id FROM vendors WHERE slug='globalcart-direct' AND vendor_type='platform' AND status='active'").Scan(&platformVendorID); err != nil {
		t.Fatalf("read seeded platform vendor: %v", err)
	}
	var platformOfferID int64
	if err := pool.QueryRow(ctx, `INSERT INTO vendor_products
		(vendor_id,product_id,variant_id,price,currency,stock,status)
		VALUES($1,$2,$3,13.99,'USD',2,'active') RETURNING id`,
		platformVendorID, product.ID, variantID).Scan(&platformOfferID); err != nil {
		t.Fatalf("create platform variant offer: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO vendor_price_history
		(vendor_product_id,old_price,new_price,currency,change_source)
		VALUES($1,NULL,13.99,'USD','initial_import')`, platformOfferID); err != nil {
		t.Fatalf("record initial vendor price: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO vendor_inventory_events
		(vendor_product_id,quantity_delta,resulting_stock,reason,reference)
		VALUES($1,2,2,'opening_balance','integration-test')`, platformOfferID); err != nil {
		t.Fatalf("record initial vendor stock: %v", err)
	}
	var secondVendorID int64
	if err := pool.QueryRow(ctx, `INSERT INTO vendors(slug,display_name,status)
		VALUES('integration-seller','Integration Seller','active') RETURNING id`).Scan(&secondVendorID); err != nil {
		t.Fatalf("create second vendor: %v", err)
	}
	var secondOfferID int64
	if err := pool.QueryRow(ctx, `INSERT INTO vendor_products
		(vendor_id,product_id,variant_id,price,currency,stock,status)
		VALUES($1,$2,$3,14.49,'USD',7,'active') RETURNING id`,
		secondVendorID, product.ID, variantID).Scan(&secondOfferID); err != nil {
		t.Fatalf("create second vendor offer for same variant: %v", err)
	}
	var firstStock, secondStock int
	if err := pool.QueryRow(ctx, "SELECT stock FROM vendor_products WHERE id=$1", platformOfferID).Scan(&firstStock); err != nil {
		t.Fatalf("read first vendor stock: %v", err)
	}
	if err := pool.QueryRow(ctx, "SELECT stock FROM vendor_products WHERE id=$1", secondOfferID).Scan(&secondStock); err != nil {
		t.Fatalf("read second vendor stock: %v", err)
	}
	if firstStock != 2 || secondStock != 7 {
		t.Fatalf("vendor stocks are not independent: first=%d second=%d", firstStock, secondStock)
	}
	var priceHistoryCount, inventoryEventCount int
	if err := pool.QueryRow(ctx, "SELECT count(*)::int FROM vendor_price_history WHERE vendor_product_id=$1", platformOfferID).Scan(&priceHistoryCount); err != nil {
		t.Fatalf("read vendor price history: %v", err)
	}
	if err := pool.QueryRow(ctx, "SELECT count(*)::int FROM vendor_inventory_events WHERE vendor_product_id=$1", platformOfferID).Scan(&inventoryEventCount); err != nil {
		t.Fatalf("read vendor inventory events: %v", err)
	}
	if priceHistoryCount != 1 || inventoryEventCount != 1 {
		t.Fatalf("vendor audit rows: price=%d inventory=%d, want one each", priceHistoryCount, inventoryEventCount)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO vendor_products
		(vendor_id,product_id,variant_id,price,currency,stock,status)
		VALUES($1,$2,$3,1.00,'USD',1,'active')`, platformVendorID, product.ID, variantID); err == nil {
		t.Fatal("duplicate vendor offer for the same variant should be rejected")
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
		key   [32]byte
		order commerce.Order
		err   error
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
	// Exercise the internal guest cart lifecycle using only a SHA-256 capability digest.
	cartOwner := make([]byte, 32)
	cartOwner[0] = 77
	cartService := commerce.New(pool)

	// Coupon values are recalculated from canonical checkout prices and snapshotted.
	var couponOwner [32]byte
	couponOwner[0] = 70
	if _, err := cartService.CreateCart(ctx, couponOwner[:], "USD", time.Hour); err != nil {
		t.Fatalf("create coupon cart: %v", err)
	}
	if _, err := cartService.AddItem(ctx, couponOwner[:], product.ID, nil, 1); err != nil {
		t.Fatalf("add item to coupon cart: %v", err)
	}
	maxDiscount := "5.00"
	globalCouponUses := int64(1)
	if err := cartService.CreateCoupon(ctx, "SAVE10", commerce.CouponSpec{
		DiscountType: "percentage", DiscountValue: "10", MinimumSubtotal: "10.00",
		MaximumDiscount: &maxDiscount, StartsAt: time.Now().UTC().Add(-time.Hour),
		EndsAt: time.Now().UTC().Add(time.Hour), UsageLimit: &globalCouponUses,
		PerCustomerLimit: 1,
	}); err != nil {
		t.Fatalf("create percentage coupon: %v", err)
	}
	if err := cartService.ApplyCoupon(ctx, couponOwner[:], " save10 "); err != nil {
		t.Fatalf("apply normalized coupon code: %v", err)
	}
	couponCart, err := cartService.GetCart(ctx, couponOwner[:])
	if err != nil || !couponCart.CouponApplied {
		t.Fatalf("cart coupon state = %+v, err=%v; want coupon applied", couponCart, err)
	}
	couponKey := make([]byte, 32)
	couponKey[0] = 71
	couponOrder, err := cartService.Checkout(ctx, couponOwner[:], couponKey, 15*time.Minute)
	if err != nil {
		t.Fatalf("checkout with coupon: %v", err)
	}
	if couponOrder.Subtotal != "12.34" || couponOrder.DiscountAmount != "1.23" || couponOrder.Total != "11.11" {
		t.Fatalf("coupon order totals: subtotal=%s discount=%s total=%s, want 12.34 / 1.23 / 11.11",
			couponOrder.Subtotal, couponOrder.DiscountAmount, couponOrder.Total)
	}
	couponDetail, err := cartService.GetOrder(ctx, couponOwner[:], couponOrder.ID)
	if err != nil || couponDetail.Subtotal != couponOrder.Subtotal ||
		couponDetail.DiscountAmount != couponOrder.DiscountAmount || couponDetail.Total != couponOrder.Total {
		t.Fatalf("coupon order detail totals = %+v, err=%v; want checkout totals", couponDetail.Order, err)
	}
	var couponRedemptions int
	if err := pool.QueryRow(ctx, "SELECT count(*)::int FROM coupon_redemptions WHERE order_id=$1", couponOrder.ID).Scan(&couponRedemptions); err != nil {
		t.Fatalf("read coupon redemption: %v", err)
	}
	if couponRedemptions != 1 {
		t.Fatalf("coupon redemption rows = %d, want 1", couponRedemptions)
	}
	var secondCouponOwner [32]byte
	secondCouponOwner[0] = 72
	if _, err := cartService.CreateCart(ctx, secondCouponOwner[:], "USD", time.Hour); err != nil {
		t.Fatalf("create second coupon cart: %v", err)
	}
	if err := cartService.ApplyCoupon(ctx, secondCouponOwner[:], "SAVE10"); !errors.Is(err, commerce.ErrCouponUnavailable) {
		t.Fatalf("apply globally exhausted coupon error = %v, want unavailable", err)
	}

	// Minimum subtotal is enforced again against canonical checkout prices.
	var minimumOwner [32]byte
	minimumOwner[0] = 73
	if _, err := cartService.CreateCart(ctx, minimumOwner[:], "USD", time.Hour); err != nil {
		t.Fatalf("create minimum-subtotal cart: %v", err)
	}
	if _, err := cartService.AddItem(ctx, minimumOwner[:], product.ID, nil, 1); err != nil {
		t.Fatalf("add item to minimum-subtotal cart: %v", err)
	}
	if err := cartService.CreateCoupon(ctx, "MIN20", commerce.CouponSpec{
		DiscountType: "fixed", DiscountValue: "3.00", MinimumSubtotal: "20.00",
		StartsAt: time.Now().UTC().Add(-time.Hour), EndsAt: time.Now().UTC().Add(time.Hour),
		PerCustomerLimit: 1,
	}); err != nil {
		t.Fatalf("create minimum-subtotal coupon: %v", err)
	}
	if err := cartService.ApplyCoupon(ctx, minimumOwner[:], "MIN20"); err != nil {
		t.Fatalf("apply minimum-subtotal coupon: %v", err)
	}
	minimumKey := make([]byte, 32)
	minimumKey[0] = 74
	if _, err := cartService.Checkout(ctx, minimumOwner[:], minimumKey, 15*time.Minute); !errors.Is(err, commerce.ErrCouponMinimum) {
		t.Fatalf("checkout below coupon minimum error = %v, want minimum-subtotal error", err)
	}
	minimumCart, err := cartService.GetCart(ctx, minimumOwner[:])
	if err != nil || minimumCart.Status != "open" {
		t.Fatalf("cart after rejected minimum coupon checkout = %+v, err=%v; want open", minimumCart, err)
	}
	if err := cartService.RemoveCoupon(ctx, minimumOwner[:]); err != nil {
		t.Fatalf("remove coupon from open cart: %v", err)
	}
	removedCouponCart, err := cartService.GetCart(ctx, minimumOwner[:])
	if err != nil || removedCouponCart.CouponApplied {
		t.Fatalf("cart coupon state after removal = %+v, err=%v; want no coupon", removedCouponCart, err)
	}
	createdCart, err := cartService.CreateCart(ctx, cartOwner, "USD", time.Hour)
	if err != nil {
		t.Fatalf("create guest cart through service: %v", err)
	}
	if createdCart.ID == "" || createdCart.Status != "open" || createdCart.Currency != "USD" {
		t.Fatalf("unexpected created cart: %+v", createdCart)
	}
	cartAfterAdd, err := cartService.AddItem(ctx, cartOwner, product.ID, nil, 1)
	if err != nil {
		t.Fatalf("add base product to guest cart: %v", err)
	}
	if len(cartAfterAdd.Items) != 1 || cartAfterAdd.Items[0].Quantity != 1 || cartAfterAdd.Items[0].UnitPrice != "12.34" {
		t.Fatalf("unexpected cart after first add: %+v", cartAfterAdd)
	}
	cartAfterIncrement, err := cartService.AddItem(ctx, cartOwner, product.ID, nil, 1)
	if err != nil {
		t.Fatalf("increment guest cart line: %v", err)
	}
	if len(cartAfterIncrement.Items) != 1 || cartAfterIncrement.Items[0].Quantity != 2 {
		t.Fatalf("cart line was not incremented: %+v", cartAfterIncrement)
	}
	cartAfterSet, err := cartService.SetItemQuantity(ctx, cartOwner, product.ID, nil, 3)
	if err != nil {
		t.Fatalf("set guest cart quantity: %v", err)
	}
	if len(cartAfterSet.Items) != 1 || cartAfterSet.Items[0].Quantity != 3 {
		t.Fatalf("cart quantity was not set: %+v", cartAfterSet)
	}
	if _, err := cartService.SetItemQuantity(ctx, cartOwner, product.ID, nil, 4); !errors.Is(err, commerce.ErrInsufficientStock) {
		t.Fatalf("setting quantity beyond stock error = %v, want insufficient stock", err)
	}
	cartAfterRemove, err := cartService.RemoveItem(ctx, cartOwner, product.ID, nil)
	if err != nil {
		t.Fatalf("remove guest cart line: %v", err)
	}
	if len(cartAfterRemove.Items) != 0 {
		t.Fatalf("cart should be empty after removal: %+v", cartAfterRemove.Items)
	}
	if _, err := cartService.AddItem(ctx, cartOwner, product.ID, &variantID, 1); err != nil {
		t.Fatalf("add active variant to guest cart: %v", err)
	}
	if _, err := cartService.RemoveItem(ctx, cartOwner, product.ID, &variantID); err != nil {
		t.Fatalf("remove variant line before cancellation fixture: %v", err)
	}
	if _, err := cartService.AddItem(ctx, cartOwner, product.ID, nil, 1); err != nil {
		t.Fatalf("add base product for cancellation fixture: %v", err)
	}

	// A cart owner may cancel an unpaid demo order; active holds are released.
	cancelKey := make([]byte, 32)
	cancelKey[0] = 79
	cancelledOrder, err := cartService.Checkout(ctx, cartOwner, cancelKey, 15*time.Minute)
	if err != nil {
		t.Fatalf("checkout variant cart for cancellation test: %v", err)
	}
	cancelledOrder, err = cartService.CancelOrder(ctx, cartOwner, cancelledOrder.ID)
	if err != nil {
		t.Fatalf("cancel owned demo order: %v", err)
	}
	if cancelledOrder.Status != "cancelled" {
		t.Fatalf("cancelled order status = %q, want cancelled", cancelledOrder.Status)
	}
	cancelledReplay, err := cartService.CancelOrder(ctx, cartOwner, cancelledOrder.ID)
	if err != nil || cancelledReplay.Status != "cancelled" {
		t.Fatalf("repeat cancellation = %+v, err=%v; want idempotent cancelled result", cancelledReplay, err)
	}
	cancelledDetail, err := cartService.GetOrder(ctx, cartOwner, cancelledOrder.ID)
	if err != nil {
		t.Fatalf("read cancelled order: %v", err)
	}
	if len(cancelledDetail.Events) != 2 || cancelledDetail.Events[1].Type != "demo_order_cancelled" {
		t.Fatalf("cancelled order event history = %+v", cancelledDetail.Events)
	}
	var releasedReservations int
	if err := pool.QueryRow(ctx, `SELECT count(*)::int FROM inventory_reservations
		WHERE order_id=$1 AND status='released'`, cancelledOrder.ID).Scan(&releasedReservations); err != nil {
		t.Fatalf("read released reservation: %v", err)
	}
	if releasedReservations != 1 {
		t.Fatalf("released reservations = %d, want 1", releasedReservations)
	}
	var unrelatedOwner [32]byte
	unrelatedOwner[0] = 80
	if _, err := cartService.CancelOrder(ctx, unrelatedOwner[:], cancelledOrder.ID); !errors.Is(err, commerce.ErrOrderNotFound) {
		t.Fatalf("cancelling another cart's order error = %v, want not found", err)
	}

	// Simulated payment is owner-scoped, idempotent and never invokes a provider.
	var paymentOwner [32]byte
	paymentOwner[0] = 81
	if _, err := cartService.CreateCart(ctx, paymentOwner[:], "USD", time.Hour); err != nil {
		t.Fatalf("create simulated payment cart: %v", err)
	}
	if _, err := cartService.AddItem(ctx, paymentOwner[:], product.ID, nil, 1); err != nil {
		t.Fatalf("add item for simulated payment: %v", err)
	}
	paymentKey := make([]byte, 32)
	paymentKey[0] = 82
	paymentOrder, err := cartService.Checkout(ctx, paymentOwner[:], paymentKey, 15*time.Minute)
	if err != nil {
		t.Fatalf("checkout simulated payment cart: %v", err)
	}
	var wrongPaymentOwner [32]byte
	wrongPaymentOwner[0] = 83
	if _, err := cartService.CompleteDemoPayment(ctx, wrongPaymentOwner[:], paymentOrder.ID); !errors.Is(err, commerce.ErrOrderNotFound) {
		t.Fatalf("simulated payment by another cart owner error = %v, want not found", err)
	}
	paidOrder, err := cartService.CompleteDemoPayment(ctx, paymentOwner[:], paymentOrder.ID)
	if err != nil {
		t.Fatalf("complete simulated payment: %v", err)
	}
	if paidOrder.Status != "paid_simulated" {
		t.Fatalf("simulated payment status = %q, want paid_simulated", paidOrder.Status)
	}
	paidReplay, err := cartService.CompleteDemoPayment(ctx, paymentOwner[:], paymentOrder.ID)
	if err != nil || paidReplay.Status != "paid_simulated" {
		t.Fatalf("repeat simulated payment = %+v, err=%v; want idempotent paid result", paidReplay, err)
	}
	if _, err := cartService.CancelOrder(ctx, paymentOwner[:], paymentOrder.ID); !errors.Is(err, commerce.ErrOrderNotCancellable) {
		t.Fatalf("cancelling simulated-paid order error = %v, want not cancellable", err)
	}
	var paidReleased, paymentEvents int
	if err := pool.QueryRow(ctx, `SELECT count(*)::int FROM inventory_reservations
		WHERE order_id=$1 AND status='released'`, paymentOrder.ID).Scan(&paidReleased); err != nil {
		t.Fatalf("read released simulated-payment reservation: %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*)::int FROM demo_order_events
		WHERE order_id=$1 AND event_type='demo_payment_simulated'`, paymentOrder.ID).Scan(&paymentEvents); err != nil {
		t.Fatalf("read simulated-payment event: %v", err)
	}
	if paidReleased != 1 || paymentEvents != 1 {
		t.Fatalf("simulated payment side effects: released=%d events=%d, want 1 each", paidReleased, paymentEvents)
	}

	// Payment completion must reject an order whose reservation has expired.
	var partialExpiryOwner [32]byte
	partialExpiryOwner[0] = 84
	if _, err := cartService.CreateCart(ctx, partialExpiryOwner[:], "USD", time.Hour); err != nil {
		t.Fatalf("create partial-expiry cart: %v", err)
	}
	if _, err := cartService.AddItem(ctx, partialExpiryOwner[:], product.ID, nil, 1); err != nil {
		t.Fatalf("add base product to partial-expiry cart: %v", err)
	}
	partialExpiryKey := make([]byte, 32)
	partialExpiryKey[0] = 85
	partialExpiryOrder, err := cartService.Checkout(ctx, partialExpiryOwner[:], partialExpiryKey, 15*time.Minute)
	if err != nil {
		t.Fatalf("checkout partial-expiry cart: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE inventory_reservations SET expires_at=now()-interval '1 second'
		WHERE order_id=$1 AND product_id=$2 AND variant_id IS NULL`, partialExpiryOrder.ID, product.ID); err != nil {
		t.Fatalf("expire one order-line reservation: %v", err)
	}
	if _, err := cartService.CompleteDemoPayment(ctx, partialExpiryOwner[:], partialExpiryOrder.ID); !errors.Is(err, commerce.ErrDemoPaymentWindowExpired) {
		t.Fatalf("complete partially expired order error = %v, want expired payment window", err)
	}
	var expiredCartOwner [32]byte
	expiredCartOwner[0] = 78
	if _, err := pool.Exec(ctx, `INSERT INTO carts(owner_token_hash,currency,expires_at)
		VALUES($1,'USD',now()-interval '1 second')`, expiredCartOwner[:]); err != nil {
		t.Fatalf("create expired guest cart fixture: %v", err)
	}
	if _, err := cartService.AddItem(ctx, expiredCartOwner[:], product.ID, nil, 1); !errors.Is(err, commerce.ErrCartExpired) {
		t.Fatalf("adding to expired cart error = %v, want cart expired", err)
	}
	if _, err := cartService.CreateCart(ctx, make([]byte, 32), "EUR", time.Hour); !errors.Is(err, commerce.ErrUnsupportedCurrency) {
		t.Fatalf("unsupported currency error = %v, want unsupported currency", err)
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

	// Order reads are scoped to the owning guest capability, not merely the UUID.
	orderDetail, err := service.GetOrder(ctx, attempts[winner].owner[:], attempts[winner].order.ID)
	if err != nil {
		t.Fatalf("read order through owning cart capability: %v", err)
	}
	if orderDetail.ID != attempts[winner].order.ID || orderDetail.Total != attempts[winner].order.Total {
		t.Fatalf("order detail does not match checkout result: %+v", orderDetail)
	}
	if len(orderDetail.Items) != 1 || orderDetail.Items[0].Quantity != 2 || orderDetail.Items[0].UnitPrice != "12.34" {
		t.Fatalf("unexpected immutable order item snapshots: %+v", orderDetail.Items)
	}
	if len(orderDetail.Events) != 1 || orderDetail.Events[0].Type != "demo_order_created" {
		t.Fatalf("unexpected initial order event history: %+v", orderDetail.Events)
	}
	var wrongOrderOwner [32]byte
	wrongOrderOwner[0] = 99
	if _, err := service.GetOrder(ctx, wrongOrderOwner[:], attempts[winner].order.ID); !errors.Is(err, commerce.ErrOrderNotFound) {
		t.Fatalf("reading another cart's order error = %v, want not found", err)
	}
	var winnerReservations int
	if err := pool.QueryRow(ctx, `SELECT COALESCE(sum(quantity),0)::int
		FROM inventory_reservations WHERE order_id=$1 AND status='active'`, attempts[winner].order.ID).Scan(&winnerReservations); err != nil {
		t.Fatalf("read winner reservations: %v", err)
	}
	if winnerReservations != 2 {
		t.Fatalf("winner reserved %d units, want 2", winnerReservations)
	}

	if _, err := pool.Exec(ctx, `UPDATE inventory_reservations
		SET expires_at=now()-interval '1 second'
		WHERE order_id=$1 AND status='active'`, attempts[winner].order.ID); err != nil {
		t.Fatalf("force demo reservation expiry: %v", err)
	}
	expired, err := service.ExpirePendingOrders(ctx)
	if err != nil {
		t.Fatalf("expire pending demo order: %v", err)
	}
	if expired != 2 {
		t.Fatalf("expired orders = %d, want 2 (checkout winner and expired-payment fixture)", expired)
	}
	var expiredReservationCount int
	if err := pool.QueryRow(ctx, `SELECT count(*)::int FROM inventory_reservations
		WHERE order_id=$1 AND status='expired'`, attempts[winner].order.ID).Scan(&expiredReservationCount); err != nil {
		t.Fatalf("read expired reservations: %v", err)
	}
	if expiredReservationCount != 1 {
		t.Fatalf("expired reservation rows = %d, want 1", expiredReservationCount)
	}
	if _, err := service.CancelOrder(ctx, attempts[winner].owner[:], attempts[winner].order.ID); !errors.Is(err, commerce.ErrOrderNotCancellable) {
		t.Fatalf("cancelling expired order error = %v, want not cancellable", err)
	}
	expiredAgain, err := service.ExpirePendingOrders(ctx)
	if err != nil {
		t.Fatalf("repeat expiration cleanup: %v", err)
	}
	if expiredAgain != 0 {
		t.Fatalf("repeat cleanup expired %d orders, want 0", expiredAgain)
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

func TestCatalogImportSynchronizesPlatformOfferIdempotently(t *testing.T) {
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
	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	product := catalog.Product{
		ID: time.Now().UnixNano()%7000000000000000000 + 2000000000000000000,
		Title: "Platform import sync fixture", Description: "Importer integration",
		Category: "platform-import-test", Price: 21.45, Stock: 8,
		Images: []string{}, Source: "platform-import-test",
		SourceID: fmt.Sprintf("platform-%d", time.Now().UnixNano()), ImportedAt: time.Now().UTC(),
	}
	if err := UpsertProducts(ctx, pool, []catalog.Product{product}); err != nil {
		t.Fatalf("initial catalog import: %v", err)
	}
	var offerID int64
	var vendorSlug string
	var price string
	var stock int32
	err = pool.QueryRow(ctx, `SELECT vp.id,v.slug,vp.price::text,vp.stock
		FROM vendor_products vp JOIN vendors v ON v.id=vp.vendor_id
		WHERE vp.product_id=$1 AND vp.variant_id IS NULL`, product.ID).
		Scan(&offerID, &vendorSlug, &price, &stock)
	if err != nil {
		t.Fatalf("read imported platform offer: %v", err)
	}
	if vendorSlug != "globalcart-direct" || price != "21.45" || stock != 8 {
		t.Fatalf("platform offer = vendor:%s price:%s stock:%d", vendorSlug, price, stock)
	}
	if err := UpsertProducts(ctx, pool, []catalog.Product{product}); err != nil {
		t.Fatalf("repeat catalog import: %v", err)
	}
	var priceEvents, inventoryEvents int
	if err := pool.QueryRow(ctx, "SELECT count(*)::int FROM vendor_price_history WHERE vendor_product_id=$1", offerID).Scan(&priceEvents); err != nil {
		t.Fatalf("count price events: %v", err)
	}
	if err := pool.QueryRow(ctx, "SELECT count(*)::int FROM vendor_inventory_events WHERE vendor_product_id=$1", offerID).Scan(&inventoryEvents); err != nil {
		t.Fatalf("count inventory events: %v", err)
	}
	if priceEvents != 1 || inventoryEvents != 1 {
		t.Fatalf("idempotent import duplicated audit events: price=%d inventory=%d", priceEvents, inventoryEvents)
	}
	product.Price = 24.90
	product.Stock = 5
	product.ImportedAt = time.Now().UTC()
	if err := UpsertProducts(ctx, pool, []catalog.Product{product}); err != nil {
		t.Fatalf("updated catalog import: %v", err)
	}
	if err := pool.QueryRow(ctx, "SELECT price::text,stock FROM vendor_products WHERE id=$1", offerID).Scan(&price, &stock); err != nil {
		t.Fatalf("read updated platform offer: %v", err)
	}
	if price != "24.90" || stock != 5 {
		t.Fatalf("updated platform offer price=%s stock=%d, want 24.90 / 5", price, stock)
	}
	if err := pool.QueryRow(ctx, "SELECT count(*)::int FROM vendor_price_history WHERE vendor_product_id=$1", offerID).Scan(&priceEvents); err != nil {
		t.Fatalf("count updated price events: %v", err)
	}
	if err := pool.QueryRow(ctx, "SELECT count(*)::int FROM vendor_inventory_events WHERE vendor_product_id=$1", offerID).Scan(&inventoryEvents); err != nil {
		t.Fatalf("count updated inventory events: %v", err)
	}
	if priceEvents != 2 || inventoryEvents != 2 {
		t.Fatalf("changed import audit counts: price=%d inventory=%d, want 2 each", priceEvents, inventoryEvents)
	}
}
