package commerce_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pooya16pg-spec/globalcart/backend/internal/catalog"
	"github.com/pooya16pg-spec/globalcart/backend/internal/commerce"
	"github.com/pooya16pg-spec/globalcart/backend/internal/database"
)

func TestGetVendorOrderIsOwnerScopedAndReturnsOnlyChildItems(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set; PostgreSQL integration test skipped")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	defer pool.Close()
	if err := database.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	productID := int64(6000000000000 + time.Now().UnixNano()%1000000000000)
	fixture := catalog.Product{
		ID: productID, Title: "Vendor detail fixture", Description: "Integration fixture",
		Category: "vendor-detail-test", Price: 5, Rating: 4, Stock: 12,
		Images: []string{"https://example.invalid/vendor-detail.webp"},
		Source: "vendor-detail-integration", SourceID: fmt.Sprintf("vendor-detail-%d", productID),
		ImportedAt: time.Now().UTC(),
	}
	if err := database.UpsertProducts(ctx, pool, []catalog.Product{fixture}); err != nil {
		t.Fatalf("upsert product: %v", err)
	}

	svc := commerce.New(pool)
	vendor, err := svc.CreateVendor(ctx, fmt.Sprintf("vendor-detail-%d", productID), "Vendor Detail Test")
	if err != nil {
		t.Fatalf("create vendor: %v", err)
	}
	if _, err := pool.Exec(ctx, "UPDATE vendors SET status='active' WHERE id=$1", vendor.ID); err != nil {
		t.Fatalf("activate vendor: %v", err)
	}
	offer, err := svc.UpsertVendorProduct(ctx, commerce.VendorProduct{
		VendorID: vendor.ID, ProductID: productID, Price: "7.25", Currency: "USD", Stock: 6, Status: "active",
	})
	if err != nil {
		t.Fatalf("create vendor offer: %v", err)
	}

	owner := sha256.Sum256([]byte(fmt.Sprintf("vendor-detail-owner-%d", productID)))
	otherOwner := sha256.Sum256([]byte(fmt.Sprintf("vendor-detail-other-%d", productID)))
	if _, err := svc.CreateCart(ctx, owner[:], "USD", time.Hour); err != nil {
		t.Fatalf("create cart: %v", err)
	}
	if _, err := svc.AddVendorItem(ctx, owner[:], offer.ID, 2); err != nil {
		t.Fatalf("add vendor item: %v", err)
	}
	idempotency := sha256.Sum256([]byte(fmt.Sprintf("vendor-detail-key-%d", productID)))
	parent, err := svc.CheckoutVendorCart(ctx, owner[:], idempotency[:], 20*time.Minute)
	if err != nil {
		t.Fatalf("checkout vendor cart: %v", err)
	}

	var childID string
	if err := pool.QueryRow(ctx, "SELECT id::text FROM demo_vendor_orders WHERE order_id=$1 AND vendor_id=$2", parent.ID, vendor.ID).Scan(&childID); err != nil {
		t.Fatalf("read child ID: %v", err)
	}

	detail, err := svc.GetVendorOrder(ctx, owner[:], childID)
	if err != nil {
		t.Fatalf("get owned vendor order: %v", err)
	}
	if detail.ID != childID || detail.ParentOrderID != parent.ID || detail.ParentStatus != "awaiting_demo_payment" {
		t.Fatalf("unexpected parent/child identity or state: %+v", detail)
	}
	if detail.VendorID != vendor.ID || detail.VendorName != "Vendor Detail Test" || detail.Status != "awaiting_demo_payment" {
		t.Fatalf("unexpected vendor order metadata: %+v", detail.VendorOrderDetail)
	}
	if detail.Subtotal != "14.50" || detail.DiscountAmount != "0.00" || detail.Total != "14.50" {
		t.Fatalf("unexpected vendor order totals: %+v", detail.VendorOrderDetail)
	}
	if len(detail.Items) != 1 {
		t.Fatalf("vendor order item count=%d, want 1", len(detail.Items))
	}
	item := detail.Items[0]
	if item.ProductID != productID || item.Quantity != 2 || item.UnitPrice != "7.25" ||
		item.VendorProductID == nil || *item.VendorProductID != offer.ID ||
		item.VendorID == nil || *item.VendorID != vendor.ID ||
		item.VendorOrderID == nil || *item.VendorOrderID != childID {
		t.Fatalf("vendor item snapshot/ownership mismatch: %+v", item)
	}

	if _, err := svc.GetVendorOrder(ctx, otherOwner[:], childID); !errors.Is(err, commerce.ErrOrderNotFound) {
		t.Fatalf("cross-owner read error=%v, want ErrOrderNotFound", err)
	}
	if _, err := svc.GetVendorOrder(ctx, owner[:], ""); !errors.Is(err, commerce.ErrOrderNotFound) {
		t.Fatalf("empty child ID error=%v, want ErrOrderNotFound", err)
	}
	if _, err := svc.GetVendorOrder(ctx, []byte("short"), childID); err == nil {
		t.Fatal("invalid capability digest unexpectedly succeeded")
	}
}
