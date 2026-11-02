package admin

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pooya16pg-spec/globalcart/backend/internal/auth"
)

const (
	overviewQueryTimeout = 4 * time.Second
	sourceSummaryLimit   = 10
)

var ErrDatabaseUnavailable = errors.New("admin overview database is unavailable")

// Overview contains aggregate operational diagnostics only. It intentionally
// excludes user emails, product payloads, audit metadata and secret material.
type Overview struct {
	GeneratedAt       time.Time              `json:"generatedAt"`
	Catalog           CatalogDiagnostics     `json:"catalog"`
	Vendors           VendorDiagnostics      `json:"vendors"`
	SellerMemberships MembershipDiagnostics  `json:"sellerMemberships"`
	Identity          IdentityDiagnostics    `json:"identity"`
	Pricing           PricingDiagnostics     `json:"pricing"`
	Translation       TranslationDiagnostics `json:"translation"`
	Activity          ActivityDiagnostics    `json:"activity"`
}

type CatalogDiagnostics struct {
	ProductCount              int64         `json:"productCount"`
	CategoryCount             int64         `json:"categoryCount"`
	VariantCount              int64         `json:"variantCount"`
	ActiveVariantCount        int64         `json:"activeVariantCount"`
	MissingDescriptionCount   int64         `json:"missingDescriptionCount"`
	UncategorizedProductCount int64         `json:"uncategorizedProductCount"`
	ZeroPriceProductCount     int64         `json:"zeroPriceProductCount"`
	LastImportedAt            *time.Time    `json:"lastImportedAt,omitempty"`
	TopSources                []SourceCount `json:"topSources"`
}

type SourceCount struct {
	Source       string `json:"source"`
	ProductCount int64  `json:"productCount"`
}

type VendorDiagnostics struct {
	ActiveVendorCount    int64 `json:"activeVendorCount"`
	PendingVendorCount   int64 `json:"pendingVendorCount"`
	SuspendedVendorCount int64 `json:"suspendedVendorCount"`
	ClosedVendorCount    int64 `json:"closedVendorCount"`
	ActiveOfferCount     int64 `json:"activeOfferCount"`
	DraftOfferCount      int64 `json:"draftOfferCount"`
	PausedOfferCount     int64 `json:"pausedOfferCount"`
	ArchivedOfferCount   int64 `json:"archivedOfferCount"`
}

type MembershipDiagnostics struct {
	PendingCount   int64 `json:"pendingCount"`
	ActiveCount    int64 `json:"activeCount"`
	SuspendedCount int64 `json:"suspendedCount"`
	RevokedCount   int64 `json:"revokedCount"`
}

type IdentityDiagnostics struct {
	ActiveUserCount          int64 `json:"activeUserCount"`
	PendingVerificationCount int64 `json:"pendingVerificationCount"`
	DisabledUserCount        int64 `json:"disabledUserCount"`
}

type PricingDiagnostics struct {
	PriceHistoryEntryCount    int64      `json:"priceHistoryEntryCount"`
	OffersWithoutPriceHistory int64      `json:"offersWithoutPriceHistory"`
	LastPriceChangeAt         *time.Time `json:"lastPriceChangeAt,omitempty"`
}

type TranslationDiagnostics struct {
	Status string `json:"status"`
	Detail string `json:"detail"`
}

type ActivityDiagnostics struct {
	AuditEventsLast24Hours int64      `json:"auditEventsLast24Hours"`
	LastAuditEventAt       *time.Time `json:"lastAuditEventAt,omitempty"`
}

// ReadOverview accepts only an opaque server-side session token. It resolves
// current account/role state from PostgreSQL and applies the dedicated admin
// read policy before querying aggregate diagnostics. It is not an HTTP handler.
func ReadOverview(ctx context.Context, pool *pgxpool.Pool, sessionToken string) (Overview, error) {
	var overview Overview
	if strings.TrimSpace(sessionToken) == "" {
		return overview, auth.ErrUnauthenticated
	}
	if pool == nil {
		return overview, ErrDatabaseUnavailable
	}

	queryCtx, cancel := context.WithTimeout(ctx, overviewQueryTimeout)
	defer cancel()

	principal, err := auth.ResolveSession(queryCtx, pool, sessionToken)
	if err != nil {
		return overview, err
	}
	if err := auth.Authorize(principal, auth.AuthorizationAdminReadOverview, auth.AuthorizationResource{}); err != nil {
		return overview, err
	}

	var lastImportedAt, lastPriceChangeAt, lastAuditEventAt pgtype.Timestamptz
	err = pool.QueryRow(queryCtx, `
		SELECT
			now(),
			(SELECT count(*)::bigint FROM products),
			(SELECT count(*)::bigint FROM categories),
			(SELECT count(*)::bigint FROM product_variants),
			(SELECT count(*)::bigint FROM product_variants WHERE is_active),
			(SELECT count(*)::bigint FROM products WHERE length(btrim(description)) = 0),
			(SELECT count(*)::bigint FROM products WHERE category_id IS NULL),
			(SELECT count(*)::bigint FROM products WHERE price = 0),
			(SELECT max(imported_at) FROM products),
			(SELECT count(*)::bigint FROM vendors WHERE status = 'active'),
			(SELECT count(*)::bigint FROM vendors WHERE status = 'pending'),
			(SELECT count(*)::bigint FROM vendors WHERE status = 'suspended'),
			(SELECT count(*)::bigint FROM vendors WHERE status = 'closed'),
			(SELECT count(*)::bigint FROM vendor_products WHERE status = 'active'),
			(SELECT count(*)::bigint FROM vendor_products WHERE status = 'draft'),
			(SELECT count(*)::bigint FROM vendor_products WHERE status = 'paused'),
			(SELECT count(*)::bigint FROM vendor_products WHERE status = 'archived'),
			(SELECT count(*)::bigint FROM seller_memberships WHERE status = 'pending'),
			(SELECT count(*)::bigint FROM seller_memberships WHERE status = 'active'),
			(SELECT count(*)::bigint FROM seller_memberships WHERE status = 'suspended'),
			(SELECT count(*)::bigint FROM seller_memberships WHERE status = 'revoked'),
			(SELECT count(*)::bigint FROM users WHERE status = 'active'),
			(SELECT count(*)::bigint FROM users WHERE status = 'pending_verification'),
			(SELECT count(*)::bigint FROM users WHERE status = 'disabled'),
			(SELECT count(*)::bigint FROM vendor_price_history),
			(SELECT count(*)::bigint
			 FROM vendor_products AS vp
			 WHERE NOT EXISTS (
			 	SELECT 1 FROM vendor_price_history AS vph
			 	WHERE vph.vendor_product_id = vp.id
			 )),
			(SELECT max(changed_at) FROM vendor_price_history),
			(SELECT count(*)::bigint FROM audit_events WHERE created_at >= now() - INTERVAL '24 hours'),
			(SELECT max(created_at) FROM audit_events)
	`).Scan(
		&overview.GeneratedAt,
		&overview.Catalog.ProductCount,
		&overview.Catalog.CategoryCount,
		&overview.Catalog.VariantCount,
		&overview.Catalog.ActiveVariantCount,
		&overview.Catalog.MissingDescriptionCount,
		&overview.Catalog.UncategorizedProductCount,
		&overview.Catalog.ZeroPriceProductCount,
		&lastImportedAt,
		&overview.Vendors.ActiveVendorCount,
		&overview.Vendors.PendingVendorCount,
		&overview.Vendors.SuspendedVendorCount,
		&overview.Vendors.ClosedVendorCount,
		&overview.Vendors.ActiveOfferCount,
		&overview.Vendors.DraftOfferCount,
		&overview.Vendors.PausedOfferCount,
		&overview.Vendors.ArchivedOfferCount,
		&overview.SellerMemberships.PendingCount,
		&overview.SellerMemberships.ActiveCount,
		&overview.SellerMemberships.SuspendedCount,
		&overview.SellerMemberships.RevokedCount,
		&overview.Identity.ActiveUserCount,
		&overview.Identity.PendingVerificationCount,
		&overview.Identity.DisabledUserCount,
		&overview.Pricing.PriceHistoryEntryCount,
		&overview.Pricing.OffersWithoutPriceHistory,
		&lastPriceChangeAt,
		&overview.Activity.AuditEventsLast24Hours,
		&lastAuditEventAt,
	)
	if err != nil {
		return Overview{}, fmt.Errorf("read admin overview aggregates: %w", err)
	}

	overview.Catalog.LastImportedAt = nullableTime(lastImportedAt)
	overview.Pricing.LastPriceChangeAt = nullableTime(lastPriceChangeAt)
	overview.Activity.LastAuditEventAt = nullableTime(lastAuditEventAt)
	overview.Translation = TranslationDiagnostics{
		Status: "paused",
		Detail: "Translation is intentionally paused; no translation worker or provider is configured.",
	}

	rows, err := pool.Query(queryCtx, `
		SELECT source, count(*)::bigint AS product_count
		FROM products
		GROUP BY source
		ORDER BY product_count DESC, source
		LIMIT $1
	`, sourceSummaryLimit)
	if err != nil {
		return Overview{}, fmt.Errorf("read admin catalog source summary: %w", err)
	}
	defer rows.Close()

	overview.Catalog.TopSources = make([]SourceCount, 0, sourceSummaryLimit)
	for rows.Next() {
		var source SourceCount
		if err := rows.Scan(&source.Source, &source.ProductCount); err != nil {
			return Overview{}, fmt.Errorf("scan admin catalog source summary: %w", err)
		}
		overview.Catalog.TopSources = append(overview.Catalog.TopSources, source)
	}
	if err := rows.Err(); err != nil {
		return Overview{}, fmt.Errorf("iterate admin catalog source summary: %w", err)
	}
	return overview, nil
}

func nullableTime(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time.UTC()
	return &result
}
