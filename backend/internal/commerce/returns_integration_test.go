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

func newAwaitingDemoReturnOrder(t *testing.T, ctx context.Context, pool *pgxpool.Pool, label string) (*commerce.Service, []byte, commerce.Order, int64) {
	t.Helper()
	productID := int64(6000000000000 + time.Now().UnixNano()%1000000000000)
	product := catalog.Product{
		ID: productID, Title: "Demo return fixture", Description: "Internal return integration fixture",
		Category: "demo-return-test", Price: 20, Rating: 4.5, Stock: 8,
		Images: []string{"https://example.invalid/demo-return.webp"},
		Source: "demo-return-integration", SourceID: fmt.Sprintf("return-%s-%d", label, productID),
		ImportedAt: time.Now().UTC(),
	}
	if err := database.UpsertProducts(ctx, pool, []catalog.Product{product}); err != nil {
		t.Fatalf("upsert product: %v", err)
	}
	var offerID int64
	if err := pool.QueryRow(ctx, `SELECT vp.id FROM vendor_products vp
		JOIN vendors v ON v.id=vp.vendor_id
		WHERE vp.product_id=$1 AND vp.variant_id IS NULL
		  AND v.slug='globalcart-direct' AND vp.status='active'`, productID).Scan(&offerID); err != nil {
		t.Fatalf("read platform offer created by catalog import: %v", err)
	}

	svc := commerce.New(pool)
	ownerDigest := sha256.Sum256([]byte(fmt.Sprintf("demo-return-owner-%s-%d", label, productID)))
	owner := append([]byte(nil), ownerDigest[:]...)
	if _, err := svc.CreateCart(ctx, owner, "USD", time.Hour); err != nil {
		t.Fatalf("create cart: %v", err)
	}
	if _, err := svc.AddVendorItem(ctx, owner, offerID, 2); err != nil {
		t.Fatalf("add vendor item: %v", err)
	}
	idempotency := sha256.Sum256([]byte(fmt.Sprintf("demo-return-order-%s-%d", label, productID)))
	order, err := svc.CheckoutVendorCart(ctx, owner, idempotency[:], 20*time.Minute)
	if err != nil {
		t.Fatalf("checkout vendor cart: %v", err)
	}
	return svc, owner, order, offerID
}

func TestDemoReturnApprovalAndSimulatedRefundLifecycle(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set; PostgreSQL integration test skipped")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	defer pool.Close()
	if err := database.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	svc, owner, order, offerID := newAwaitingDemoReturnOrder(t, ctx, pool, "approved")
	if _, err := svc.RequestDemoReturn(ctx, owner, order.ID, commerce.DemoReturnReasonChangedMind); !errors.Is(err, commerce.ErrDemoReturnNotAllowed) {
		t.Fatalf("return request before simulated payment error=%v, want ErrDemoReturnNotAllowed", err)
	}
	if _, err := svc.CompleteDemoPayment(ctx, owner, order.ID); err != nil {
		t.Fatalf("complete simulated payment: %v", err)
	}

	otherDigest := sha256.Sum256([]byte("different-demo-return-owner"))
	if _, err := svc.RequestDemoReturn(ctx, otherDigest[:], order.ID, commerce.DemoReturnReasonChangedMind); !errors.Is(err, commerce.ErrOrderNotFound) {
		t.Fatalf("cross-owner request error=%v, want ErrOrderNotFound", err)
	}
	if _, err := svc.RequestDemoReturn(ctx, owner, order.ID, "free-text"); !errors.Is(err, commerce.ErrInvalidDemoReturnReason) {
		t.Fatalf("invalid reason error=%v, want ErrInvalidDemoReturnReason", err)
	}

	requested, err := svc.RequestDemoReturn(ctx, owner, order.ID, commerce.DemoReturnReasonChangedMind)
	if err != nil {
		t.Fatalf("request demo return: %v", err)
	}
	if requested.Status != "requested" || requested.RefundAmount != "0.00" || requested.ReviewedAt != nil || requested.RefundedAt != nil {
		t.Fatalf("unexpected new return state: %+v", requested)
	}
	replay, err := svc.RequestDemoReturn(ctx, owner, order.ID, commerce.DemoReturnReasonChangedMind)
	if err != nil || !replay.Replayed || replay.ID != requested.ID {
		t.Fatalf("idempotent return request=%+v err=%v", replay, err)
	}
	if _, err := svc.RequestDemoReturn(ctx, owner, order.ID, commerce.DemoReturnReasonWrongItem); !errors.Is(err, commerce.ErrDemoReturnAlreadyExists) {
		t.Fatalf("second reason error=%v, want ErrDemoReturnAlreadyExists", err)
	}
	if _, err := svc.CompleteDemoRefund(ctx, owner, order.ID); !errors.Is(err, commerce.ErrDemoReturnNotAllowed) {
		t.Fatalf("refund before review error=%v, want ErrDemoReturnNotAllowed", err)
	}

	approved, err := svc.ReviewDemoReturn(ctx, order.ID, true)
	if err != nil {
		t.Fatalf("approve return: %v", err)
	}
	if approved.Status != "approved" || approved.RefundAmount != order.Total || approved.ReviewedAt == nil || approved.RefundedAt != nil {
		t.Fatalf("unexpected approved return state: %+v (order total %s)", approved, order.Total)
	}
	approvalReplay, err := svc.ReviewDemoReturn(ctx, order.ID, true)
	if err != nil || !approvalReplay.Replayed {
		t.Fatalf("approval replay=%+v err=%v", approvalReplay, err)
	}
	if _, err := svc.ReviewDemoReturn(ctx, order.ID, false); !errors.Is(err, commerce.ErrDemoReturnNotAllowed) {
		t.Fatalf("conflicting review error=%v, want ErrDemoReturnNotAllowed", err)
	}
	if _, err := svc.CompleteDemoRefund(ctx, otherDigest[:], order.ID); !errors.Is(err, commerce.ErrOrderNotFound) {
		t.Fatalf("cross-owner refund error=%v, want ErrOrderNotFound", err)
	}

	refunded, err := svc.CompleteDemoRefund(ctx, owner, order.ID)
	if err != nil {
		t.Fatalf("complete simulated refund: %v", err)
	}
	if refunded.Status != "refunded_simulated" || refunded.RefundAmount != order.Total || refunded.RefundedAt == nil {
		t.Fatalf("unexpected completed refund state: %+v", refunded)
	}
	refundReplay, err := svc.CompleteDemoRefund(ctx, owner, order.ID)
	if err != nil || !refundReplay.Replayed || refundReplay.ID != refunded.ID {
		t.Fatalf("refund replay=%+v err=%v", refundReplay, err)
	}

	var orderStatus, returnStatus, refundAmount string
	var childRefunded, eventCount int
	if err := pool.QueryRow(ctx, `SELECT o.status,r.status,r.refund_amount::text,
		(SELECT count(*)::int FROM demo_vendor_orders vo WHERE vo.order_id=o.id AND vo.status='refunded_simulated'),
		(SELECT count(*)::int FROM demo_order_events e WHERE e.order_id=o.id AND e.event_type IN
		 ('demo_return_requested','demo_return_approved','demo_refund_simulated'))
		FROM demo_orders o JOIN demo_returns r ON r.order_id=o.id WHERE o.id=$1`,
		order.ID).Scan(&orderStatus, &returnStatus, &refundAmount, &childRefunded, &eventCount); err != nil {
		t.Fatalf("read completed return state: %v", err)
	}
	if orderStatus != "refunded_simulated" || returnStatus != "refunded_simulated" ||
		refundAmount != order.Total || childRefunded != 1 || eventCount != 3 {
		t.Fatalf("persisted return order=%s return=%s amount=%s childRefunded=%d events=%d",
			orderStatus, returnStatus, refundAmount, childRefunded, eventCount)
	}
	var stock int32
	if err := pool.QueryRow(ctx, "SELECT stock FROM vendor_products WHERE id=$1", offerID).Scan(&stock); err != nil {
		t.Fatalf("read vendor offer stock: %v", err)
	}
	if stock != 8 {
		t.Fatalf("simulated refund changed vendor stock to %d; want unchanged stock 8", stock)
	}
}

func TestRejectedDemoReturnCannotBeRefunded(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set; PostgreSQL integration test skipped")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	defer pool.Close()
	if err := database.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	svc, owner, order, _ := newAwaitingDemoReturnOrder(t, ctx, pool, "rejected")
	if _, err := svc.CompleteDemoPayment(ctx, owner, order.ID); err != nil {
		t.Fatalf("complete simulated payment: %v", err)
	}
	if _, err := svc.RequestDemoReturn(ctx, owner, order.ID, commerce.DemoReturnReasonDamagedDemo); err != nil {
		t.Fatalf("request demo return: %v", err)
	}
	rejected, err := svc.ReviewDemoReturn(ctx, order.ID, false)
	if err != nil {
		t.Fatalf("reject demo return: %v", err)
	}
	if rejected.Status != "rejected" || rejected.RefundAmount != "0.00" || rejected.ReviewedAt == nil || rejected.RefundedAt != nil {
		t.Fatalf("unexpected rejected return: %+v", rejected)
	}
	replay, err := svc.ReviewDemoReturn(ctx, order.ID, false)
	if err != nil || !replay.Replayed {
		t.Fatalf("rejection replay=%+v err=%v", replay, err)
	}
	if _, err := svc.CompleteDemoRefund(ctx, owner, order.ID); !errors.Is(err, commerce.ErrDemoReturnNotAllowed) {
		t.Fatalf("refund for rejected return error=%v, want ErrDemoReturnNotAllowed", err)
	}
	var orderStatus string
	if err := pool.QueryRow(ctx, "SELECT status FROM demo_orders WHERE id=$1", order.ID).Scan(&orderStatus); err != nil {
		t.Fatalf("read order after rejected return: %v", err)
	}
	if orderStatus != "paid_simulated" {
		t.Fatalf("rejected return changed order status to %q, want paid_simulated", orderStatus)
	}
}
