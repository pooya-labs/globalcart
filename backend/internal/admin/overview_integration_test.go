package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pooya16pg-spec/globalcart/backend/internal/auth"
	"github.com/pooya16pg-spec/globalcart/backend/internal/catalog"
	"github.com/pooya16pg-spec/globalcart/backend/internal/database"
)

const overviewIntegrationPasswordHash = "$argon2id$v=19$m=19456,t=2,p=1$AAECAwQFBgcICQoLDA0ODw$AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8"

func openOverviewIntegrationPool(t *testing.T) (*pgxpool.Pool, context.Context) {
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

func createOverviewUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, suffix string) string {
	t.Helper()
	var userID string
	email := fmt.Sprintf("admin-overview-%s@example.invalid", suffix)
	if err := pool.QueryRow(ctx, `
		INSERT INTO users(email_normalized, password_hash)
		VALUES($1, $2)
		RETURNING id::text
	`, email, overviewIntegrationPasswordHash).Scan(&userID); err != nil {
		t.Fatalf("create integration user: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		UPDATE users
		SET status = 'active', email_verified_at = now(), updated_at = now()
		WHERE id = $1
	`, userID); err != nil {
		t.Fatalf("activate integration user: %v", err)
	}
	return userID
}

func createOverviewSession(t *testing.T, ctx context.Context, pool *pgxpool.Pool, userID string) auth.IssuedSession {
	t.Helper()
	session, err := auth.CreateSession(ctx, pool, userID, overviewIntegrationPasswordHash)
	if err != nil {
		t.Fatalf("create integration session: %v", err)
	}
	return session
}

func TestReadOverviewRequiresLiveAdminSessionAndReturnsAggregates(t *testing.T) {
	pool, ctx := openOverviewIntegrationPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	customerID := createOverviewUser(t, ctx, pool, suffix+"-customer")
	customerSession := createOverviewSession(t, ctx, pool, customerID)
	if _, err := ReadOverview(ctx, pool, customerSession.SessionToken); !errors.Is(err, auth.ErrForbidden) {
		t.Fatalf("customer overview access error = %v, want auth.ErrForbidden", err)
	}

	sellerID := createOverviewUser(t, ctx, pool, suffix+"-seller")
	var vendorID int64
	if err := pool.QueryRow(ctx, `
		INSERT INTO vendors(slug, display_name, status)
		VALUES($1, 'Overview Integration Seller', 'active')
		RETURNING id
	`, "overview-"+suffix).Scan(&vendorID); err != nil {
		t.Fatalf("create active seller vendor: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO user_roles(user_id, role_key) VALUES($1, 'seller')
	`, sellerID); err != nil {
		t.Fatalf("grant seller fixture role: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO seller_memberships(user_id, vendor_id, status, activated_at)
		VALUES($1, $2, 'active', now())
	`, sellerID, vendorID); err != nil {
		t.Fatalf("activate seller fixture membership: %v", err)
	}
	sellerSession := createOverviewSession(t, ctx, pool, sellerID)
	if _, err := ReadOverview(ctx, pool, sellerSession.SessionToken); !errors.Is(err, auth.ErrForbidden) {
		t.Fatalf("seller overview access error = %v, want auth.ErrForbidden", err)
	}

	adminID := createOverviewUser(t, ctx, pool, suffix+"-admin")
	// A NULL grant actor represents the documented out-of-band bootstrap
	// performed by an approved database operator in the integration fixture.
	if _, err := pool.Exec(ctx, `
		INSERT INTO user_roles(user_id, role_key, granted_by_user_id)
		VALUES($1, 'admin', NULL)
	`, adminID); err != nil {
		t.Fatalf("grant bootstrap admin fixture role: %v", err)
	}
	adminSession := createOverviewSession(t, ctx, pool, adminID)

	productID := time.Now().UnixNano()
	source := "admin-overview-integration"
	sourceID := suffix
	fixture := catalog.Product{
		ID: productID, Title: "Admin overview diagnostic fixture",
		Description: "", Category: "admin-overview-" + suffix,
		Price: 0, DiscountPercentage: 0, Rating: 0, Stock: 0,
		Brand: "GlobalCart Test", Thumbnail: "https://example.invalid/admin-overview.webp",
		Images: []string{}, Source: source, SourceID: sourceID,
		ImportedAt: time.Now().UTC(),
		Variants: []catalog.Variant{{
			SKU: "ADM-" + suffix, Title: "Diagnostic variant",
			Price: 1.25, Stock: 1, Attributes: map[string]any{"fixture": true},
			Source: source, SourceID: "variant-" + sourceID, IsActive: true,
		}},
	}
	if err := database.UpsertProducts(ctx, pool, []catalog.Product{fixture}); err != nil {
		t.Fatalf("upsert diagnostic catalog fixture: %v", err)
	}

	overview, err := ReadOverview(ctx, pool, adminSession.SessionToken)
	if err != nil {
		t.Fatalf("ReadOverview(admin): %v", err)
	}
	if overview.GeneratedAt.IsZero() {
		t.Fatal("overview generated timestamp is missing")
	}
	if overview.Catalog.ProductCount < 1 ||
		overview.Catalog.CategoryCount < 1 ||
		overview.Catalog.VariantCount < 1 ||
		overview.Catalog.ActiveVariantCount < 1 ||
		overview.Catalog.MissingDescriptionCount < 1 ||
		overview.Catalog.ZeroPriceProductCount < 1 {
		t.Fatalf("catalog diagnostics did not include fixture rows: %+v", overview.Catalog)
	}
	if overview.Catalog.LastImportedAt == nil || overview.Pricing.LastPriceChangeAt == nil {
		t.Fatalf("expected import and price timestamps, got catalog=%+v pricing=%+v", overview.Catalog, overview.Pricing)
	}
	if overview.Vendors.ActiveOfferCount < 1 || overview.Pricing.PriceHistoryEntryCount < 1 {
		t.Fatalf("vendor/price diagnostics did not include imported offers: vendors=%+v pricing=%+v", overview.Vendors, overview.Pricing)
	}
	if overview.Identity.ActiveUserCount < 3 {
		t.Fatalf("active user count = %d, want at least three integration accounts", overview.Identity.ActiveUserCount)
	}
	if overview.Activity.AuditEventsLast24Hours < 1 || overview.Activity.LastAuditEventAt == nil {
		t.Fatalf("session activity diagnostics missing: %+v", overview.Activity)
	}
	if overview.Translation.Status != "paused" || !strings.Contains(overview.Translation.Detail, "no translation worker") {
		t.Fatalf("translation state must disclose paused pipeline: %+v", overview.Translation)
	}
	foundSource := false
	for _, item := range overview.Catalog.TopSources {
		if item.Source == source && item.ProductCount >= 1 {
			foundSource = true
			break
		}
	}
	if !foundSource {
		t.Fatalf("top source summary omitted fixture source %q: %+v", source, overview.Catalog.TopSources)
	}

	encoded, err := json.Marshal(overview)
	if err != nil {
		t.Fatalf("marshal overview: %v", err)
	}
	if strings.Contains(string(encoded), "example.invalid") ||
		strings.Contains(string(encoded), "password_hash") ||
		strings.Contains(string(encoded), "session_token") {
		t.Fatalf("overview leaked private account or token fields: %s", encoded)
	}
}

func TestReadOverviewRejectsMissingAndRevokedSessions(t *testing.T) {
	if _, err := ReadOverview(context.Background(), nil, ""); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Fatalf("empty token error = %v, want auth.ErrUnauthenticated", err)
	}
	if _, err := ReadOverview(context.Background(), nil, "opaque-token"); !errors.Is(err, ErrDatabaseUnavailable) {
		t.Fatalf("nil database error = %v, want ErrDatabaseUnavailable", err)
	}

	pool, ctx := openOverviewIntegrationPool(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	adminID := createOverviewUser(t, ctx, pool, suffix+"-revoked")
	if _, err := pool.Exec(ctx, `
		INSERT INTO user_roles(user_id, role_key, granted_by_user_id)
		VALUES($1, 'admin', NULL)
	`, adminID); err != nil {
		t.Fatalf("grant bootstrap admin fixture role: %v", err)
	}
	session := createOverviewSession(t, ctx, pool, adminID)
	if err := auth.RevokeSession(ctx, pool, session.SessionToken); err != nil {
		t.Fatalf("revoke admin fixture session: %v", err)
	}
	if overview, err := ReadOverview(ctx, pool, session.SessionToken); !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("revoked session overview = %+v, error %v; want auth.ErrInvalidSession", overview, err)
	}
}
