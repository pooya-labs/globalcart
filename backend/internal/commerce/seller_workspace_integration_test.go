package commerce_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/pooya16pg-spec/globalcart/backend/internal/auth"
	"github.com/pooya16pg-spec/globalcart/backend/internal/catalog"
	"github.com/pooya16pg-spec/globalcart/backend/internal/commerce"
	"github.com/pooya16pg-spec/globalcart/backend/internal/database"
)

func TestSellerWorkspaceProfileAndDemoAnalyticsAreVendorScoped(t *testing.T) {
	pool, ctx := openAuthorizedVendorIntegrationPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	productID := time.Now().UnixNano()%8000000000000000000 + 1000000000000000000
	now := time.Now().UTC()
	product := catalog.Product{
		ID: productID, Title: "Seller workspace analytics fixture",
		Description: "Internal seller workspace integration fixture",
		Category: "seller-workspace-fixture", Price: 12.50, Rating: 4.6, Stock: 10,
		Images: []string{"https://example.invalid/seller-workspace.webp"},
		Source: "seller-workspace-integration",
		SourceID: fmt.Sprintf("seller-workspace-%d", productID),
		ImportedAt: now,
	}
	if err := database.UpsertProducts(ctx, pool, []catalog.Product{product}); err != nil {
		t.Fatalf("insert seller workspace product: %v", err)
	}

	svc := commerce.New(pool)
	adminID := createAuthorizedVendorUser(t, ctx, pool, suffix+"-admin")
	sellerID := createAuthorizedVendorUser(t, ctx, pool, suffix+"-seller")
	customerID := createAuthorizedVendorUser(t, ctx, pool, suffix+"-customer")
	if _, err := pool.Exec(ctx,
		"INSERT INTO user_roles(user_id,role_key) VALUES($1,'admin')", adminID,
	); err != nil {
		t.Fatalf("grant seller workspace fixture admin: %v", err)
	}

	vendor, offer := createAuthorizedVendorOffer(t, ctx, pool, svc, productID, suffix+"-owned")
	foreignVendor, _ := createAuthorizedVendorOffer(t, ctx, pool, svc, productID, suffix+"-foreign")
	membership, err := auth.RequestSellerMembership(ctx, pool, sellerID, vendor.ID)
	if err != nil {
		t.Fatalf("request seller workspace membership: %v", err)
	}
	if _, err := auth.ActivateSellerMembership(ctx, pool, adminID, membership.ID); err != nil {
		t.Fatalf("activate seller workspace membership: %v", err)
	}
	issued, err := auth.CreateSession(ctx, pool, sellerID, sellerIntegrationPasswordHash)
	if err != nil {
		t.Fatalf("create seller workspace session: %v", err)
	}
	principal, err := auth.ResolveSession(ctx, pool, issued.SessionToken)
	if err != nil {
		t.Fatalf("resolve seller workspace session: %v", err)
	}

	workspace, err := svc.GetSellerWorkspaceForSeller(ctx, principal, vendor.ID)
	if err != nil {
		t.Fatalf("read seller workspace: %v", err)
	}
	if workspace.Profile.VendorID != vendor.ID || workspace.Profile.DisplayName != vendor.DisplayName ||
		workspace.Profile.Slug != vendor.Slug || workspace.Profile.Status != "active" {
		t.Fatalf("seller workspace profile = %+v", workspace.Profile)
	}
	if workspace.Offers.Total != 1 || workspace.Offers.Active != 1 || workspace.Offers.ActiveStock != 8 {
		t.Fatalf("initial seller offer aggregates = %+v", workspace.Offers)
	}
	if workspace.Orders.Total != 0 || workspace.Orders.PaidSimulated != 0 ||
		workspace.Orders.Last30Days != 0 || len(workspace.Orders.PaidTotals) != 0 {
		t.Fatalf("initial seller order aggregates = %+v", workspace.Orders)
	}

	if _, err := svc.GetSellerWorkspaceForSeller(ctx, principal, foreignVendor.ID); !errors.Is(err, auth.ErrForbidden) {
		t.Fatalf("foreign vendor workspace error = %v, want auth.ErrForbidden", err)
	}
	customerSession, err := auth.CreateSession(ctx, pool, customerID, sellerIntegrationPasswordHash)
	if err != nil {
		t.Fatalf("create customer workspace session: %v", err)
	}
	customerPrincipal, err := auth.ResolveSession(ctx, pool, customerSession.SessionToken)
	if err != nil {
		t.Fatalf("resolve customer workspace session: %v", err)
	}
	if _, err := svc.GetSellerWorkspaceForSeller(ctx, customerPrincipal, vendor.ID); !errors.Is(err, auth.ErrForbidden) {
		t.Fatalf("customer workspace error = %v, want auth.ErrForbidden", err)
	}

	updatedName := "Seller Studio " + suffix
	updatedProfile, err := svc.UpdateSellerVendorProfileForSeller(ctx, principal, vendor.ID, commerce.SellerProfileUpdateInput{
		ExpectedUpdatedAt: workspace.Profile.UpdatedAt,
		DisplayName: updatedName,
	})
	if err != nil {
		t.Fatalf("update seller display name: %v", err)
	}
	if updatedProfile.DisplayName != updatedName || updatedProfile.Slug != vendor.Slug ||
		updatedProfile.Type != vendor.Type || updatedProfile.Status != vendor.Status {
		t.Fatalf("updated seller profile = %+v", updatedProfile)
	}
	if _, err := svc.UpdateSellerVendorProfileForSeller(ctx, principal, vendor.ID, commerce.SellerProfileUpdateInput{
		ExpectedUpdatedAt: workspace.Profile.UpdatedAt,
		DisplayName: "Stale name " + suffix,
	}); !errors.Is(err, commerce.ErrSellerProfileRevisionConflict) {
		t.Fatalf("stale seller profile update error = %v, want revision conflict", err)
	}
	if _, err := svc.UpdateSellerVendorProfileForSeller(ctx, principal, vendor.ID, commerce.SellerProfileUpdateInput{
		ExpectedUpdatedAt: updatedProfile.UpdatedAt,
		DisplayName: updatedName,
	}); !errors.Is(err, commerce.ErrSellerProfileNoChanges) {
		t.Fatalf("unchanged seller profile update error = %v, want no changes", err)
	}
	var profileAuditCount int
	if err := pool.QueryRow(ctx,
		"SELECT count(*)::int FROM audit_events WHERE actor_user_id=$1 AND event_type='seller.vendor_profile.updated' AND resource_type='vendor' AND resource_id=$2",
		sellerID, fmt.Sprint(vendor.ID),
	).Scan(&profileAuditCount); err != nil {
		t.Fatalf("count seller profile audit events: %v", err)
	}
	if profileAuditCount != 1 {
		t.Fatalf("seller profile audit count = %d, want 1", profileAuditCount)
	}

	owner := sha256.Sum256([]byte("seller-workspace-owner-" + suffix))
	if _, err := svc.CreateCart(ctx, owner[:], "USD", time.Hour); err != nil {
		t.Fatalf("create seller workspace demo cart: %v", err)
	}
	if _, err := svc.AddVendorItem(ctx, owner[:], offer.ID, 2); err != nil {
		t.Fatalf("add seller workspace vendor offer: %v", err)
	}
	idempotency := sha256.Sum256([]byte("seller-workspace-order-" + suffix))
	order, err := svc.CheckoutVendorCart(ctx, owner[:], idempotency[:], 20*time.Minute)
	if err != nil {
		t.Fatalf("checkout seller workspace demo cart: %v", err)
	}
	if order.Subtotal != "25.00" || order.Total != "25.00" {
		t.Fatalf("seller workspace demo order = %+v, want 25.00 subtotal and total", order)
	}
	if _, err := svc.CompleteDemoPayment(ctx, owner[:], order.ID); err != nil {
		t.Fatalf("complete simulated seller workspace payment: %v", err)
	}

	workspace, err = svc.GetSellerWorkspaceForSeller(ctx, principal, vendor.ID)
	if err != nil {
		t.Fatalf("read seller workspace after demo order: %v", err)
	}
	if workspace.Orders.Total != 1 || workspace.Orders.PaidSimulated != 1 ||
		workspace.Orders.AwaitingDemoPayment != 0 || workspace.Orders.Last30Days != 1 {
		t.Fatalf("seller order state aggregates = %+v", workspace.Orders)
	}
	if len(workspace.Orders.PaidTotals) != 1 ||
		workspace.Orders.PaidTotals[0].Currency != "USD" ||
		workspace.Orders.PaidTotals[0].Orders != 1 ||
		workspace.Orders.PaidTotals[0].SimulatedAmount != "25.00" {
		t.Fatalf("seller simulated paid totals = %+v", workspace.Orders.PaidTotals)
	}

	if _, err := auth.RevokeSellerMembership(ctx, pool, adminID, membership.ID); err != nil {
		t.Fatalf("revoke seller workspace membership: %v", err)
	}
	if _, err := svc.GetSellerWorkspaceForSeller(ctx, principal, vendor.ID); !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("revoked membership workspace error = %v, want auth.ErrInvalidSession", err)
	}
	if _, err := svc.UpdateSellerVendorProfileForSeller(ctx, principal, vendor.ID, commerce.SellerProfileUpdateInput{
		ExpectedUpdatedAt: updatedProfile.UpdatedAt,
		DisplayName: "After revocation",
	}); !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("revoked membership profile update error = %v, want auth.ErrInvalidSession", err)
	}
}
