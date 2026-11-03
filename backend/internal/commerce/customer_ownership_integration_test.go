package commerce_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pooya-labs/globalcart/backend/internal/auth"
	"github.com/pooya-labs/globalcart/backend/internal/catalog"
	"github.com/pooya-labs/globalcart/backend/internal/commerce"
	"github.com/pooya-labs/globalcart/backend/internal/database"
)

const customerOwnershipPasswordHash = "$argon2id$v=19$m=19456,t=2,p=1$AAECAwQFBgcICQoLDA0ODw$AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8"

type customerOwnershipActor struct {
	userID    string
	session   auth.IssuedSession
	principal auth.SessionPrincipal
}

func openCustomerOwnershipPool(t *testing.T) (*pgxpool.Pool, context.Context) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set; PostgreSQL integration test skipped")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	t.Cleanup(cancel)
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("create PostgreSQL pool: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping PostgreSQL: %v", err)
	}
	if err := database.Migrate(ctx, pool); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	return pool, ctx
}

func createCustomerOwnershipActor(t *testing.T, ctx context.Context, pool *pgxpool.Pool, suffix string) customerOwnershipActor {
	t.Helper()
	var actor customerOwnershipActor
	email := fmt.Sprintf("customer-ownership-%s@example.invalid", suffix)
	if err := pool.QueryRow(ctx, `
		INSERT INTO users(email_normalized,password_hash) VALUES($1,$2) RETURNING id::text
	`, email, customerOwnershipPasswordHash).Scan(&actor.userID); err != nil {
		t.Fatalf("create customer account: %v", err)
	}
	if _, err := pool.Exec(ctx, "UPDATE users SET status='active',email_verified_at=now(),updated_at=now() WHERE id=$1", actor.userID); err != nil {
		t.Fatalf("activate customer account: %v", err)
	}
	issued, err := auth.CreateSession(ctx, pool, actor.userID, customerOwnershipPasswordHash)
	if err != nil {
		t.Fatalf("create customer session: %v", err)
	}
	actor.session = issued
	actor.principal, err = auth.ResolveSession(ctx, pool, issued.SessionToken)
	if err != nil {
		t.Fatalf("resolve customer session: %v", err)
	}
	return actor
}

func createCustomerOwnershipProduct(t *testing.T, ctx context.Context, pool *pgxpool.Pool, suffix string) int64 {
	t.Helper()
	id := int64(10_000_000_000_000 + time.Now().UnixNano()%1_000_000_000_000)
	product := catalog.Product{
		ID: id, Title: "Customer ownership fixture", Description: "Customer cart integration fixture",
		Category: "customer-ownership-test", Price: 18.75, Rating: 4.5, Stock: 20,
		Images: []string{"https://example.invalid/customer-ownership.webp"},
		Source: "customer-ownership-" + suffix, SourceID: "product-" + suffix, ImportedAt: time.Now().UTC(),
	}
	if err := database.UpsertProducts(ctx, pool, []catalog.Product{product}); err != nil {
		t.Fatalf("upsert customer ownership product: %v", err)
	}
	return id
}

func TestCustomerCartAndOrderOwnership(t *testing.T) {
	pool, ctx := openCustomerOwnershipPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	owner := createCustomerOwnershipActor(t, ctx, pool, suffix+"-owner")
	other := createCustomerOwnershipActor(t, ctx, pool, suffix+"-other")
	productID := createCustomerOwnershipProduct(t, ctx, pool, suffix)
	svc := commerce.New(pool)
	digest := sha256.Sum256([]byte("customer-cart-owner-" + suffix))
	cart, err := svc.CreateCart(ctx, digest[:], "USD", time.Hour)
	if err != nil {
		t.Fatalf("create guest cart: %v", err)
	}
	if _, err := svc.GetCartForCustomer(ctx, owner.principal, cart.ID); !errors.Is(err, commerce.ErrCartNotFound) {
		t.Fatalf("unbound cart read error=%v, want ErrCartNotFound", err)
	}
	if err := svc.BindCartToCustomer(ctx, owner.principal, digest[:]); err != nil {
		t.Fatalf("bind cart: %v", err)
	}
	if _, err := svc.GetCartForCustomer(ctx, other.principal, cart.ID); !errors.Is(err, commerce.ErrCartNotFound) {
		t.Fatalf("foreign cart read error=%v, want ErrCartNotFound", err)
	}
	if _, err := svc.AddItemForCustomer(ctx, other.principal, cart.ID, productID, nil, 1); !errors.Is(err, commerce.ErrCartNotFound) {
		t.Fatalf("foreign cart mutation error=%v, want ErrCartNotFound", err)
	}
	got, err := svc.AddItemForCustomer(ctx, owner.principal, cart.ID, productID, nil, 1)
	if err != nil || len(got.Items) != 1 || got.Items[0].Quantity != 1 {
		t.Fatalf("add cart item=%+v err=%v", got, err)
	}
	got, err = svc.SetItemQuantityForCustomer(ctx, owner.principal, cart.ID, productID, nil, 2)
	if err != nil || len(got.Items) != 1 || got.Items[0].Quantity != 2 {
		t.Fatalf("set cart quantity=%+v err=%v", got, err)
	}
	got, err = svc.RemoveItemForCustomer(ctx, owner.principal, cart.ID, productID, nil)
	if err != nil || len(got.Items) != 0 {
		t.Fatalf("remove cart item=%+v err=%v", got, err)
	}
	if _, err := svc.RemoveItemForCustomer(ctx, owner.principal, cart.ID, productID, nil); !errors.Is(err, commerce.ErrInvalidProduct) {
		t.Fatalf("remove missing item error=%v, want ErrInvalidProduct", err)
	}
	if _, err := svc.AddItemForCustomer(ctx, owner.principal, cart.ID, productID, nil, 1); err != nil {
		t.Fatalf("re-add cart item: %v", err)
	}
	key := sha256.Sum256([]byte("customer-checkout-" + suffix))
	order, err := svc.Checkout(ctx, digest[:], key[:], 20*time.Minute)
	if err != nil {
		t.Fatalf("checkout bound cart: %v", err)
	}
	detail, err := svc.GetOrderForCustomer(ctx, owner.principal, order.ID)
	if err != nil || detail.ID != order.ID || len(detail.Items) != 1 {
		t.Fatalf("read owned order=%+v err=%v", detail, err)
	}
	if _, err := svc.GetOrderForCustomer(ctx, other.principal, order.ID); !errors.Is(err, commerce.ErrOrderNotFound) {
		t.Fatalf("foreign order read error=%v, want ErrOrderNotFound", err)
	}
	cancelled, err := svc.CancelOrderForCustomer(ctx, owner.principal, order.ID)
	if err != nil || cancelled.Status != "cancelled" {
		t.Fatalf("cancel owned order=%+v err=%v", cancelled, err)
	}
	if _, err := svc.CancelOrderForCustomer(ctx, other.principal, order.ID); !errors.Is(err, commerce.ErrOrderNotFound) {
		t.Fatalf("foreign order cancellation error=%v, want ErrOrderNotFound", err)
	}
	replay, err := svc.CancelOrderForCustomer(ctx, owner.principal, order.ID)
	if err != nil || replay.Status != "cancelled" {
		t.Fatalf("cancellation replay=%+v err=%v", replay, err)
	}
}

func TestCustomerVendorOrderOwnership(t *testing.T) {
	pool, ctx := openCustomerOwnershipPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	owner := createCustomerOwnershipActor(t, ctx, pool, suffix+"-owner")
	other := createCustomerOwnershipActor(t, ctx, pool, suffix+"-other")
	productID := createCustomerOwnershipProduct(t, ctx, pool, suffix)
	svc := commerce.New(pool)
	vendor, err := svc.CreateVendor(ctx, "customer-owner-"+suffix, "Customer Ownership Vendor")
	if err != nil {
		t.Fatalf("create vendor: %v", err)
	}
	if _, err := pool.Exec(ctx, "UPDATE vendors SET status='active' WHERE id=$1", vendor.ID); err != nil {
		t.Fatalf("activate vendor: %v", err)
	}
	offer, err := svc.UpsertVendorProduct(ctx, commerce.VendorProduct{VendorID: vendor.ID, ProductID: productID, Price: "19.50", Currency: "USD", Stock: 5, Status: "active"})
	if err != nil {
		t.Fatalf("create vendor offer: %v", err)
	}
	digest := sha256.Sum256([]byte("customer-vendor-cart-" + suffix))
	cart, err := svc.CreateCart(ctx, digest[:], "USD", time.Hour)
	if err != nil {
		t.Fatalf("create vendor cart: %v", err)
	}
	if err := svc.BindCartToCustomer(ctx, owner.principal, digest[:]); err != nil {
		t.Fatalf("bind vendor cart: %v", err)
	}
	if _, err := svc.AddVendorItem(ctx, digest[:], offer.ID, 1); err != nil {
		t.Fatalf("add vendor item: %v", err)
	}
	key := sha256.Sum256([]byte("customer-vendor-checkout-" + suffix))
	order, err := svc.CheckoutVendorCart(ctx, digest[:], key[:], 20*time.Minute)
	if err != nil {
		t.Fatalf("vendor split checkout: %v", err)
	}
	var childID string
	if err := pool.QueryRow(ctx, "SELECT id::text FROM demo_vendor_orders WHERE order_id=$1", order.ID).Scan(&childID); err != nil {
		t.Fatalf("read child ID: %v", err)
	}
	detail, err := svc.GetVendorOrderForCustomer(ctx, owner.principal, childID)
	if err != nil || detail.ID != childID || detail.ParentOrderID != order.ID || len(detail.Items) != 1 {
		t.Fatalf("read customer child order=%+v err=%v", detail, err)
	}
	if _, err := svc.GetVendorOrderForCustomer(ctx, other.principal, childID); !errors.Is(err, commerce.ErrOrderNotFound) {
		t.Fatalf("foreign child read error=%v, want ErrOrderNotFound", err)
	}
	_ = cart
}

func TestCustomerCartBindingIsFirstWriterWins(t *testing.T) {
	pool, ctx := openCustomerOwnershipPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	first := createCustomerOwnershipActor(t, ctx, pool, suffix+"-first")
	second := createCustomerOwnershipActor(t, ctx, pool, suffix+"-second")
	svc := commerce.New(pool)
	digest := sha256.Sum256([]byte("customer-binding-race-" + suffix))
	cart, err := svc.CreateCart(ctx, digest[:], "USD", time.Hour)
	if err != nil {
		t.Fatalf("create race cart: %v", err)
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for _, actor := range []customerOwnershipActor{first, second} {
		actor := actor
		wg.Add(1)
		go func() { defer wg.Done(); <-start; results <- svc.BindCartToCustomer(ctx, actor.principal, digest[:]) }()
	}
	close(start)
	wg.Wait()
	close(results)
	successes, denials := 0, 0
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, commerce.ErrCartNotFound):
			denials++
		default:
			t.Fatalf("unexpected binding race error: %v", err)
		}
	}
	if successes != 1 || denials != 1 {
		t.Fatalf("binding race success=%d denial=%d; want one each", successes, denials)
	}
	var winningOwner string
	if err := pool.QueryRow(ctx, "SELECT owner_user_id::text FROM carts WHERE id=$1", cart.ID).Scan(&winningOwner); err != nil {
		t.Fatalf("read winning owner: %v", err)
	}
	if winningOwner != first.userID && winningOwner != second.userID {
		t.Fatalf("unexpected winning owner %s", winningOwner)
	}
}

func TestCustomerCartRejectsExpiredCartAndRevokedSession(t *testing.T) {
	pool, ctx := openCustomerOwnershipPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	expiredActor := createCustomerOwnershipActor(t, ctx, pool, suffix+"-expired")
	revokedActor := createCustomerOwnershipActor(t, ctx, pool, suffix+"-revoked")
	svc := commerce.New(pool)
	expiredDigest := sha256.Sum256([]byte("customer-expired-cart-" + suffix))
	expiredCart, err := svc.CreateCart(ctx, expiredDigest[:], "USD", time.Hour)
	if err != nil {
		t.Fatalf("create expiring cart: %v", err)
	}
	if err := svc.BindCartToCustomer(ctx, expiredActor.principal, expiredDigest[:]); err != nil {
		t.Fatalf("bind expiring cart: %v", err)
	}
	if _, err := pool.Exec(ctx, "UPDATE carts SET expires_at=now()-interval '1 second' WHERE id=$1", expiredCart.ID); err != nil {
		t.Fatalf("expire cart fixture: %v", err)
	}
	if _, err := svc.GetCartForCustomer(ctx, expiredActor.principal, expiredCart.ID); !errors.Is(err, commerce.ErrCartExpired) {
		t.Fatalf("expired cart read error=%v, want ErrCartExpired", err)
	}
	revokedDigest := sha256.Sum256([]byte("customer-revoked-session-cart-" + suffix))
	revokedCart, err := svc.CreateCart(ctx, revokedDigest[:], "USD", time.Hour)
	if err != nil {
		t.Fatalf("create revoked-session cart: %v", err)
	}
	if err := svc.BindCartToCustomer(ctx, revokedActor.principal, revokedDigest[:]); err != nil {
		t.Fatalf("bind revoked-session cart: %v", err)
	}
	if err := auth.RevokeSession(ctx, pool, revokedActor.session.SessionToken); err != nil {
		t.Fatalf("revoke customer session: %v", err)
	}
	if _, err := svc.GetCartForCustomer(ctx, revokedActor.principal, revokedCart.ID); !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("revoked-session cart read error=%v, want ErrInvalidSession", err)
	}
}
