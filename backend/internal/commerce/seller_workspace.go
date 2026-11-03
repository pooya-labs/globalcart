package commerce

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/pooya16pg-spec/globalcart/backend/internal/auth"
)

const sellerWorkspaceTimeout = 5 * time.Second

var (
	ErrInvalidSellerProfileInput     = errors.New("invalid seller profile input")
	ErrSellerProfileRevisionConflict = errors.New("seller profile has changed")
	ErrSellerProfileNoChanges        = errors.New("seller profile contains no changes")
)

type SellerVendorProfile struct {
	VendorID    int64     `json:"vendorId"`
	Slug        string    `json:"slug"`
	DisplayName string    `json:"displayName"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type SellerOfferAnalytics struct {
	Total       int64 `json:"total"`
	Active      int64 `json:"active"`
	Draft       int64 `json:"draft"`
	Paused      int64 `json:"paused"`
	Archived    int64 `json:"archived"`
	ActiveStock int64 `json:"activeStock"`
}

type SellerPaidOrderTotal struct {
	Currency        string `json:"currency"`
	Orders          int64  `json:"orders"`
	SimulatedAmount string `json:"simulatedAmount"`
}

type SellerOrderAnalytics struct {
	Total               int64                  `json:"total"`
	AwaitingDemoPayment int64                  `json:"awaitingDemoPayment"`
	PaidSimulated       int64                  `json:"paidSimulated"`
	Cancelled           int64                  `json:"cancelled"`
	Expired             int64                  `json:"expired"`
	RefundedSimulated   int64                  `json:"refundedSimulated"`
	Last30Days          int64                  `json:"last30Days"`
	PaidTotals          []SellerPaidOrderTotal `json:"paidTotals"`
}

type SellerWorkspace struct {
	Profile SellerVendorProfile  `json:"profile"`
	Offers  SellerOfferAnalytics `json:"offers"`
	Orders  SellerOrderAnalytics `json:"orders"`
}

type SellerProfileUpdateInput struct {
	ExpectedUpdatedAt time.Time `json:"expectedUpdatedAt"`
	DisplayName       string    `json:"displayName"`
}

func normalizeSellerDisplayName(value string) (string, error) {
	value = strings.TrimSpace(value)
	if !utf8.ValidString(value) || strings.ContainsRune(value, 0) {
		return "", ErrInvalidSellerProfileInput
	}
	length := utf8.RuneCountInString(value)
	if length < 1 || length > 200 {
		return "", ErrInvalidSellerProfileInput
	}
	return value, nil
}

func readSellerVendorProfileTx(ctx context.Context, tx pgx.Tx, vendorID int64, lock bool) (SellerVendorProfile, error) {
	var p SellerVendorProfile
	query := `SELECT id,slug,display_name,vendor_type,status,created_at,updated_at FROM vendors WHERE id=$1`
	if lock {
		query += " FOR UPDATE"
	}
	err := tx.QueryRow(ctx, query, vendorID).Scan(&p.VendorID, &p.Slug, &p.DisplayName, &p.Type, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return SellerVendorProfile{}, ErrVendorNotFound
	}
	if err != nil {
		return SellerVendorProfile{}, fmt.Errorf("read seller vendor profile: %w", err)
	}
	return p, nil
}

// GetSellerWorkspaceForSeller returns only the selected, currently authorized
// vendor's profile and aggregate demo metrics. It never returns customer data.
func (s *Service) GetSellerWorkspaceForSeller(ctx context.Context, principal auth.SessionPrincipal, vendorID int64) (SellerWorkspace, error) {
	if vendorID <= 0 {
		return SellerWorkspace{}, ErrInvalidSellerProfileInput
	}
	if s == nil || s.pool == nil {
		return SellerWorkspace{}, errors.New("seller workspace database is unavailable")
	}
	qctx, cancel := context.WithTimeout(ctx, sellerWorkspaceTimeout)
	defer cancel()
	tx, err := s.pool.BeginTx(qctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return SellerWorkspace{}, fmt.Errorf("begin seller workspace read: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if err := auth.AuthorizeSellerVendorTx(qctx, tx, principal, auth.AuthorizationSellerReadVendor, vendorID); err != nil {
		return SellerWorkspace{}, err
	}
	profile, err := readSellerVendorProfileTx(qctx, tx, vendorID, false)
	if err != nil {
		return SellerWorkspace{}, err
	}
	out := SellerWorkspace{Profile: profile, Orders: SellerOrderAnalytics{PaidTotals: make([]SellerPaidOrderTotal, 0)}}
	err = tx.QueryRow(qctx, `
		SELECT count(*)::bigint,
			count(*) FILTER(WHERE status='active')::bigint,
			count(*) FILTER(WHERE status='draft')::bigint,
			count(*) FILTER(WHERE status='paused')::bigint,
			count(*) FILTER(WHERE status='archived')::bigint,
			COALESCE(sum(stock) FILTER(WHERE status='active'),0)::bigint
		FROM vendor_products WHERE vendor_id=$1
	`, vendorID).Scan(&out.Offers.Total, &out.Offers.Active, &out.Offers.Draft, &out.Offers.Paused, &out.Offers.Archived, &out.Offers.ActiveStock)
	if err != nil {
		return SellerWorkspace{}, fmt.Errorf("aggregate seller offers: %w", err)
	}
	err = tx.QueryRow(qctx, `
		SELECT count(*)::bigint,
			count(*) FILTER(WHERE status='awaiting_demo_payment')::bigint,
			count(*) FILTER(WHERE status='paid_simulated')::bigint,
			count(*) FILTER(WHERE status='cancelled')::bigint,
			count(*) FILTER(WHERE status='expired')::bigint,
			count(*) FILTER(WHERE status='refunded_simulated')::bigint,
			count(*) FILTER(WHERE created_at >= now()-interval '30 days')::bigint
		FROM demo_vendor_orders WHERE vendor_id=$1
	`, vendorID).Scan(&out.Orders.Total, &out.Orders.AwaitingDemoPayment, &out.Orders.PaidSimulated,
		&out.Orders.Cancelled, &out.Orders.Expired, &out.Orders.RefundedSimulated, &out.Orders.Last30Days)
	if err != nil {
		return SellerWorkspace{}, fmt.Errorf("aggregate seller demo orders: %w", err)
	}
	rows, err := tx.Query(qctx, `
		SELECT currency,count(*) FILTER(WHERE status='paid_simulated')::bigint,
			COALESCE(sum(total) FILTER(WHERE status='paid_simulated'),0)::numeric(12,2)::text
		FROM demo_vendor_orders WHERE vendor_id=$1 AND status='paid_simulated' GROUP BY currency ORDER BY currency
	`, vendorID)
	if err != nil {
		return SellerWorkspace{}, fmt.Errorf("aggregate simulated paid totals: %w", err)
	}
	for rows.Next() {
		var item SellerPaidOrderTotal
		if err := rows.Scan(&item.Currency, &item.Orders, &item.SimulatedAmount); err != nil {
			rows.Close()
			return SellerWorkspace{}, fmt.Errorf("scan seller simulated paid totals: %w", err)
		}
		out.Orders.PaidTotals = append(out.Orders.PaidTotals, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return SellerWorkspace{}, fmt.Errorf("iterate seller simulated paid totals: %w", err)
	}
	rows.Close()
	if err := tx.Commit(qctx); err != nil {
		return SellerWorkspace{}, fmt.Errorf("commit seller workspace read: %w", err)
	}
	return out, nil
}

// UpdateSellerVendorProfileForSeller changes only the marketplace display name.
// The vendor's updated_at is the optimistic revision token; stale edits are rejected.
func (s *Service) UpdateSellerVendorProfileForSeller(ctx context.Context, principal auth.SessionPrincipal,
	vendorID int64, input SellerProfileUpdateInput) (SellerVendorProfile, error) {
	if vendorID <= 0 || input.ExpectedUpdatedAt.IsZero() {
		return SellerVendorProfile{}, ErrInvalidSellerProfileInput
	}
	name, err := normalizeSellerDisplayName(input.DisplayName)
	if err != nil {
		return SellerVendorProfile{}, err
	}
	if s == nil || s.pool == nil {
		return SellerVendorProfile{}, errors.New("seller workspace database is unavailable")
	}
	qctx, cancel := context.WithTimeout(ctx, sellerWorkspaceTimeout)
	defer cancel()
	tx, err := s.pool.BeginTx(qctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return SellerVendorProfile{}, fmt.Errorf("begin seller profile update: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if err := auth.AuthorizeSellerVendorTx(qctx, tx, principal, auth.AuthorizationSellerWriteVendor, vendorID); err != nil {
		return SellerVendorProfile{}, err
	}
	before, err := readSellerVendorProfileTx(qctx, tx, vendorID, true)
	if err != nil {
		return SellerVendorProfile{}, err
	}
	if !before.UpdatedAt.Equal(input.ExpectedUpdatedAt) {
		return SellerVendorProfile{}, ErrSellerProfileRevisionConflict
	}
	if before.DisplayName == name {
		return SellerVendorProfile{}, ErrSellerProfileNoChanges
	}
	var after SellerVendorProfile
	err = tx.QueryRow(qctx, `
		UPDATE vendors SET display_name=$3,updated_at=now()
		WHERE id=$1 AND updated_at=$2
		RETURNING id,slug,display_name,vendor_type,status,created_at,updated_at
	`, vendorID, input.ExpectedUpdatedAt, name).Scan(
		&after.VendorID, &after.Slug, &after.DisplayName, &after.Type, &after.Status, &after.CreatedAt, &after.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SellerVendorProfile{}, ErrSellerProfileRevisionConflict
		}
		return SellerVendorProfile{}, fmt.Errorf("update seller vendor profile: %w", err)
	}
	metadata, err := json.Marshal(map[string]string{"previous_display_name": before.DisplayName, "display_name": after.DisplayName})
	if err != nil {
		return SellerVendorProfile{}, fmt.Errorf("encode seller profile audit: %w", err)
	}
	if _, err := tx.Exec(qctx, `
		INSERT INTO audit_events(event_type,actor_user_id,resource_type,resource_id,metadata)
		VALUES('seller.vendor_profile.updated',$2,'vendor',$1,$3::jsonb)
	`, fmt.Sprint(vendorID), principal.UserID, metadata); err != nil {
		return SellerVendorProfile{}, fmt.Errorf("append seller profile audit: %w", err)
	}
	if err := tx.Commit(qctx); err != nil {
		return SellerVendorProfile{}, fmt.Errorf("commit seller profile update: %w", err)
	}
	return after, nil
}
