package commerce_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pooya16pg-spec/globalcart/backend/internal/auth"
	"github.com/pooya16pg-spec/globalcart/backend/internal/catalog"
	"github.com/pooya16pg-spec/globalcart/backend/internal/commerce"
	"github.com/pooya16pg-spec/globalcart/backend/internal/database"
)

const sellerIntegrationPasswordHash = "$argon2id$v=19$m=19456,t=2,p=1$AAECAwQFBgcICQoLDA0ODw$AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8"

func openAuthorizedVendorIntegrationPool(t *testing.T) (*pgxpool.Pool, context.Context) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set; PostgreSQL integration test skipped")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

func createAuthorizedVendorUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, suffix string) string {
	t.Helper()
	var userID string
	email := fmt.Sprintf("authorized-vendor-%s@example.invalid", suffix)
	if err := pool.QueryRow(ctx,
		"INSERT INTO users(email_normalized,password_hash) VALUES($1,$2) RETURNING id::text",
		email, sellerIntegrationPasswordHash,
	).Scan(&userID); err != nil {
		t.Fatalf("create authorized vendor user: %v", err)
	}
	if _, err := pool.Exec(ctx,
		"UPDATE users SET status='active', email_verified_at=now(), updated_at=now() WHERE id=$1",
		userID,
	); err != nil {
		t.Fatalf("activate authorized vendor user: %v", err)
	}
	return userID
}

func createAuthorizedVendorOffer(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	svc *commerce.Service,
	productID int64,
	suffix string,
) (commerce.Vendor, commerce.VendorProduct) {
	t.Helper()
	vendor, err := svc.CreateVendor(ctx, "authorized-vendor-"+suffix, "Authorized Vendor "+suffix)
	if err != nil {
		t.Fatalf("create authorized vendor: %v", err)
	}
	if _, err := pool.Exec(ctx, "UPDATE vendors SET status='active', updated_at=now() WHERE id=$1", vendor.ID); err != nil {
		t.Fatalf("activate authorized vendor: %v", err)
	}
	offer, err := svc.UpsertVendorProduct(ctx, commerce.VendorProduct{
		VendorID: vendor.ID, ProductID: productID,
		Price: "12.50", Currency: "USD", Stock: 8, Status: "active",
	})
	if err != nil {
		t.Fatalf("create authorized vendor offer: %v", err)
	}
	return vendor, offer
}

func TestSellerVendorOperationsUseCanonicalOfferScopeAndCurrentSession(t *testing.T) {
	pool, ctx := openAuthorizedVendorIntegrationPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	productID := time.Now().UnixNano()%8000000000000000000 + 1000000000000000000
	product := catalog.Product{
		ID: productID, Title: "Authorized seller vendor fixture",
		Description: "Integration fixture", Category: "authorized-vendor-test",
		Price: 12.50, Rating: 4.5, Stock: 8,
		Images: []string{"https://example.invalid/authorized-vendor.webp"},
		Source: "authorized-vendor-integration", SourceID: fmt.Sprintf("authorized-vendor-%d", productID),
		ImportedAt: time.Now().UTC(),
	}
	if err := database.UpsertProducts(ctx, pool, []catalog.Product{product}); err != nil {
		t.Fatalf("insert authorized vendor product: %v", err)
	}

	svc := commerce.New(pool)
	adminID := createAuthorizedVendorUser(t, ctx, pool, suffix+"-admin")
	sellerID := createAuthorizedVendorUser(t, ctx, pool, suffix+"-seller")
	if _, err := pool.Exec(ctx,
		"INSERT INTO user_roles(user_id,role_key) VALUES($1,'admin')",
		adminID,
	); err != nil {
		t.Fatalf("grant admin fixture role: %v", err)
	}

	vendor, offer := createAuthorizedVendorOffer(t, ctx, pool, svc, productID, suffix+"-owned")
	foreignVendor, foreignOffer := createAuthorizedVendorOffer(t, ctx, pool, svc, productID, suffix+"-foreign")

	membership, err := auth.RequestSellerMembership(ctx, pool, sellerID, vendor.ID)
	if err != nil {
		t.Fatalf("request seller membership: %v", err)
	}
	if _, err := auth.ActivateSellerMembership(ctx, pool, adminID, membership.ID); err != nil {
		t.Fatalf("activate seller membership: %v", err)
	}
	issued, err := auth.CreateSession(ctx, pool, sellerID, sellerIntegrationPasswordHash)
	if err != nil {
		t.Fatalf("create seller session: %v", err)
	}
	principal, err := auth.ResolveSession(ctx, pool, issued.SessionToken)
	if err != nil {
		t.Fatalf("resolve seller session: %v", err)
	}

	got, err := svc.GetVendorProductForSeller(ctx, principal, offer.ID)
	if err != nil {
		t.Fatalf("read own vendor offer: %v", err)
	}
	if got.VendorID != vendor.ID || got.ID != offer.ID {
		t.Fatalf("resolved offer vendor=%d id=%d, want vendor=%d id=%d",
			got.VendorID, got.ID, vendor.ID, offer.ID)
	}
	if err := svc.UpdateVendorPriceForSeller(ctx, principal, offer.ID, "17.25"); err != nil {
		t.Fatalf("update own vendor price: %v", err)
	}
	stock, err := svc.AdjustVendorInventoryForSeller(ctx, principal, offer.ID, 3, "restock", "AUTHORIZED-PO-1")
	if err != nil {
		t.Fatalf("adjust own vendor stock: %v", err)
	}
	if stock != 11 {
		t.Fatalf("authorized vendor stock = %d, want 11", stock)
	}
	got, err = svc.GetVendorProductForSeller(ctx, principal, offer.ID)
	if err != nil {
		t.Fatalf("read updated own vendor offer: %v", err)
	}
	if got.Price != "17.25" || got.Stock != 11 {
		t.Fatalf("updated own vendor offer price=%s stock=%d, want 17.25/11", got.Price, got.Stock)
	}

	// The API does not accept vendorID. It resolves the foreign vendor from
	// the offer row and denies it against the principal's database-backed scope.
	if _, err := svc.GetVendorProductForSeller(ctx, principal, foreignOffer.ID); !errors.Is(err, auth.ErrForbidden) {
		t.Fatalf("foreign vendor offer read error = %v, want auth.ErrForbidden", err)
	}
	if err := svc.UpdateVendorPriceForSeller(ctx, principal, foreignOffer.ID, "99.99"); !errors.Is(err, auth.ErrForbidden) {
		t.Fatalf("foreign vendor price update error = %v, want auth.ErrForbidden", err)
	}
	if _, err := svc.AdjustVendorInventoryForSeller(ctx, principal, foreignOffer.ID, 5, "restock", "FOREIGN-PO"); !errors.Is(err, auth.ErrForbidden) {
		t.Fatalf("foreign vendor stock adjustment error = %v, want auth.ErrForbidden", err)
	}

	// After membership revocation the resolved principal is stale, but the
	// wrapper revalidates its session and membership in the protected tx.
	if _, err := auth.RevokeSellerMembership(ctx, pool, adminID, membership.ID); err != nil {
		t.Fatalf("revoke seller membership: %v", err)
	}
	if err := svc.UpdateVendorPriceForSeller(ctx, principal, offer.ID, "18.00"); !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("stale seller session price update error = %v, want auth.ErrInvalidSession", err)
	}

	// A vendor ID hidden inside an offer cannot be selected through the caller's
	// request; the server derives it from the canonical offer row each time.
	if foreignVendor.ID == vendor.ID {
		t.Fatal("fixture vendors unexpectedly share an ID")
	}
}
