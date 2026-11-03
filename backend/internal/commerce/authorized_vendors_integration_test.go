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


func TestSellerProductListIsBoundedFilteredAndVendorScoped(t *testing.T) {
	pool, ctx := openAuthorizedVendorIntegrationPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	baseID := time.Now().UnixNano()%8000000000000000000 + 1000000000000000000
	now := time.Now().UTC()
	products := []catalog.Product{
		{
			ID: baseID, Title: "Seller Product Alpha", Description: "Alpha catalog description",
			Category: "seller-workqueue-fixture", Price: 12.50, Rating: 4.4, Stock: 8,
			Brand: "Source Brand", Images: []string{"https://example.invalid/seller-alpha.webp"},
			Source: "seller-workqueue-fixture", SourceID: fmt.Sprintf("seller-alpha-%d", baseID), ImportedAt: now,
		},
		{
			ID: baseID + 1, Title: "Seller Product Beta", Description: "Beta catalog description",
			Category: "seller-workqueue-fixture", Price: 9.99, Rating: 4.6, Stock: 5,
			Brand: "Source Brand", Images: []string{"https://example.invalid/seller-beta.webp"},
			Source: "seller-workqueue-fixture", SourceID: fmt.Sprintf("seller-beta-%d", baseID), ImportedAt: now.Add(time.Second),
		},
		{
			ID: baseID + 2, Title: "Foreign Seller Product", Description: "Foreign catalog description",
			Category: "seller-workqueue-fixture", Price: 19.99, Rating: 4.2, Stock: 3,
			Brand: "Foreign Brand", Images: []string{"https://example.invalid/seller-foreign.webp"},
			Source: "seller-workqueue-fixture", SourceID: fmt.Sprintf("seller-foreign-%d", baseID), ImportedAt: now.Add(2 * time.Second),
		},
	}
	if err := database.UpsertProducts(ctx, pool, products); err != nil {
		t.Fatalf("insert seller product fixtures: %v", err)
	}

	svc := commerce.New(pool)
	adminID := createAuthorizedVendorUser(t, ctx, pool, suffix+"-admin")
	sellerID := createAuthorizedVendorUser(t, ctx, pool, suffix+"-seller")
	customerID := createAuthorizedVendorUser(t, ctx, pool, suffix+"-customer")
	if _, err := pool.Exec(ctx,
		"INSERT INTO user_roles(user_id,role_key) VALUES($1,'admin')", adminID,
	); err != nil {
		t.Fatalf("grant admin fixture role: %v", err)
	}

	vendor, alphaOffer := createAuthorizedVendorOffer(t, ctx, pool, svc, products[0].ID, suffix+"-owned")
	betaOffer, err := svc.UpsertVendorProduct(ctx, commerce.VendorProduct{
		VendorID: vendor.ID, ProductID: products[1].ID,
		Price: "9.99", Currency: "USD", Stock: 5, Status: "draft",
	})
	if err != nil {
		t.Fatalf("create second seller offer: %v", err)
	}
	foreignVendor, _ := createAuthorizedVendorOffer(t, ctx, pool, svc, products[2].ID, suffix+"-foreign")

	editorialTitle := "Curated Seller Product Beta"
	if _, err := pool.Exec(ctx,
		"INSERT INTO admin_product_overrides(product_id,title_override,created_by_user_id,updated_by_user_id) VALUES($1,$2,$3,$3)",
		products[1].ID, editorialTitle, adminID,
	); err != nil {
		t.Fatalf("add fixture editorial title: %v", err)
	}

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

	all, err := svc.ListSellerProductsForSeller(ctx, principal, vendor.ID, commerce.SellerProductListInput{
		SortBy: "price", SortOrder: "asc",
	})
	if err != nil {
		t.Fatalf("list seller products: %v", err)
	}
	if all.Total != 2 || len(all.Items) != 2 || all.Limit != 20 || all.Offset != 0 {
		t.Fatalf("default seller page total/items/limit/offset = %d/%d/%d/%d, want 2/2/20/0",
			all.Total, len(all.Items), all.Limit, all.Offset)
	}
	if all.Items[0].OfferID != betaOffer.ID || all.Items[0].Title != editorialTitle ||
		all.Items[0].CatalogVisibility != "active" || all.Items[0].VendorID != vendor.ID {
		t.Fatalf("first price-sorted seller offer = %+v", all.Items[0])
	}
	if all.Items[1].OfferID != alphaOffer.ID || all.Items[1].VendorID != vendor.ID {
		t.Fatalf("second seller offer = %+v", all.Items[1])
	}

	search, err := svc.ListSellerProductsForSeller(ctx, principal, vendor.ID, commerce.SellerProductListInput{
		Query: "  curated seller product beta  ", Status: "draft", Category: "seller-workqueue",
		Source: "SELLER-WORKQUEUE-FIXTURE",
	})
	if err != nil {
		t.Fatalf("search effective seller product values: %v", err)
	}
	if search.Total != 1 || len(search.Items) != 1 ||
		search.Items[0].OfferID != betaOffer.ID || search.Items[0].Title != editorialTitle {
		t.Fatalf("effective-title filtered seller page = %+v", search)
	}

	active, err := svc.ListSellerProductsForSeller(ctx, principal, vendor.ID, commerce.SellerProductListInput{Status: "active"})
	if err != nil {
		t.Fatalf("filter active offers: %v", err)
	}
	draft, err := svc.ListSellerProductsForSeller(ctx, principal, vendor.ID, commerce.SellerProductListInput{Status: "draft"})
	if err != nil {
		t.Fatalf("filter draft offers: %v", err)
	}
	if active.Total != 1 || active.Items[0].OfferID != alphaOffer.ID ||
		draft.Total != 1 || draft.Items[0].OfferID != betaOffer.ID {
		t.Fatalf("active/draft totals = %d/%d, want 1/1", active.Total, draft.Total)
	}

	secondPage, err := svc.ListSellerProductsForSeller(ctx, principal, vendor.ID, commerce.SellerProductListInput{
		SortBy: "price", SortOrder: "asc", Limit: 1, Offset: 1,
	})
	if err != nil {
		t.Fatalf("read second seller page: %v", err)
	}
	if secondPage.Total != 2 || len(secondPage.Items) != 1 || secondPage.Items[0].OfferID != alphaOffer.ID {
		t.Fatalf("second seller page = %+v", secondPage)
	}

	if _, err := svc.ListSellerProductsForSeller(ctx, principal, foreignVendor.ID, commerce.SellerProductListInput{}); !errors.Is(err, auth.ErrForbidden) {
		t.Fatalf("foreign vendor list error = %v, want auth.ErrForbidden", err)
	}
	customerSession, err := auth.CreateSession(ctx, pool, customerID, sellerIntegrationPasswordHash)
	if err != nil {
		t.Fatalf("create customer session: %v", err)
	}
	customerPrincipal, err := auth.ResolveSession(ctx, pool, customerSession.SessionToken)
	if err != nil {
		t.Fatalf("resolve customer session: %v", err)
	}
	if _, err := svc.ListSellerProductsForSeller(ctx, customerPrincipal, vendor.ID, commerce.SellerProductListInput{}); !errors.Is(err, auth.ErrForbidden) {
		t.Fatalf("customer seller-list error = %v, want auth.ErrForbidden", err)
	}

	if _, err := svc.ListSellerProductsForSeller(ctx, principal, vendor.ID, commerce.SellerProductListInput{Limit: 101}); !errors.Is(err, commerce.ErrInvalidSellerProductListInput) {
		t.Fatalf("oversized page error = %v, want ErrInvalidSellerProductListInput", err)
	}
	if _, err := svc.ListSellerProductsForSeller(ctx, principal, vendor.ID, commerce.SellerProductListInput{Query: "invalid\x00filter"}); !errors.Is(err, commerce.ErrInvalidSellerProductListInput) {
		t.Fatalf("NUL search error = %v, want ErrInvalidSellerProductListInput", err)
	}
	if _, err := svc.ListSellerProductsForSeller(ctx, principal, vendor.ID, commerce.SellerProductListInput{SortBy: "title; DROP TABLE products"}); !errors.Is(err, commerce.ErrInvalidSellerProductListInput) {
		t.Fatalf("unsupported sort error = %v, want ErrInvalidSellerProductListInput", err)
	}

	if _, err := auth.RevokeSellerMembership(ctx, pool, adminID, membership.ID); err != nil {
		t.Fatalf("revoke seller membership: %v", err)
	}
	if _, err := svc.ListSellerProductsForSeller(ctx, principal, vendor.ID, commerce.SellerProductListInput{}); !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("stale seller principal list error = %v, want auth.ErrInvalidSession", err)
	}
}
