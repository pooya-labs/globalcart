package commerce

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrInvalidCoupon   = errors.New("invalid coupon configuration")
	ErrCouponUnavailable = errors.New("coupon is unavailable")
	ErrCouponMinimum   = errors.New("coupon minimum subtotal not met")
)

var couponCodePattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9-]{3,63}$`)

// CouponSpec uses decimal strings to avoid floating-point money arithmetic.
// UsageLimit nil means unlimited; PerCustomerLimit is scoped to the guest
// cart capability until a verified customer identity system exists.
type CouponSpec struct {
	DiscountType      string
	DiscountValue     string
	MinimumSubtotal   string
	MaximumDiscount   *string
	StartsAt          time.Time
	EndsAt            time.Time
	UsageLimit        *int64
	PerCustomerLimit  int32
}

// CreateCoupon stores only a digest of the normalized coupon code. This is an
// internal administrative primitive and is not exposed over HTTP.
func (s *Service) CreateCoupon(ctx context.Context, code string, spec CouponSpec) error {
	normalized := strings.ToUpper(strings.TrimSpace(code))
	if !couponCodePattern.MatchString(normalized) ||
		(spec.DiscountType != "percentage" && spec.DiscountType != "fixed") ||
		strings.TrimSpace(spec.DiscountValue) == "" ||
		strings.TrimSpace(spec.MinimumSubtotal) == "" ||
		!spec.StartsAt.Before(spec.EndsAt) ||
		spec.PerCustomerLimit < 1 ||
		(spec.UsageLimit != nil && *spec.UsageLimit < 1) {
		return ErrInvalidCoupon
	}
	if spec.MaximumDiscount != nil && strings.TrimSpace(*spec.MaximumDiscount) == "" {
		return ErrInvalidCoupon
	}
	digest := sha256.Sum256([]byte(normalized))
	_, err := s.pool.Exec(ctx, `INSERT INTO coupons
		(code_hash,discount_type,discount_value,minimum_subtotal,maximum_discount,
		 starts_at,ends_at,usage_limit,per_customer_limit)
		VALUES($1,$2,$3::numeric,$4::numeric,$5::numeric,$6,$7,$8,$9)`,
		digest[:], spec.DiscountType, spec.DiscountValue, spec.MinimumSubtotal,
		spec.MaximumDiscount, spec.StartsAt, spec.EndsAt, spec.UsageLimit, spec.PerCustomerLimit)
	if err != nil {
		return fmt.Errorf("create coupon: %w", err)
	}
	return nil
}

// ApplyCoupon attaches an eligible coupon to an open cart. The coupon row is
// locked to serialize redemption-limit checks with concurrent checkouts.
func (s *Service) ApplyCoupon(ctx context.Context, ownerTokenHash []byte, code string) error {
	if len(ownerTokenHash) != 32 {
		return errors.New("cart capability digest must be 32 bytes")
	}
	normalized := strings.ToUpper(strings.TrimSpace(code))
	if !couponCodePattern.MatchString(normalized) {
		return ErrCouponUnavailable
	}
	digest := sha256.Sum256([]byte(normalized))
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("begin coupon application: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	var cartID, status string
	var expired bool
	err = tx.QueryRow(ctx, `SELECT id::text,status,expires_at<=now()
		FROM carts WHERE owner_token_hash=$1 FOR UPDATE`, ownerTokenHash).Scan(&cartID, &status, &expired)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrCartNotFound
	}
	if err != nil {
		return fmt.Errorf("lock cart for coupon: %w", err)
	}
	if status != "open" {
		return ErrCartClosed
	}
	if expired {
		return ErrCartExpired
	}

	var couponID int64
	var active bool
	var startsAt, endsAt time.Time
	var usageLimit pgtype.Int8
	var dbNow time.Time
	var perCustomerLimit int32
	err = tx.QueryRow(ctx, `SELECT id,is_active,starts_at,ends_at,usage_limit,per_customer_limit,now()
		FROM coupons WHERE code_hash=$1 FOR UPDATE`, digest[:]).
		Scan(&couponID, &active, &startsAt, &endsAt, &usageLimit, &perCustomerLimit, &dbNow)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && (!active || dbNow.Before(startsAt) || !dbNow.Before(endsAt))) {
		return ErrCouponUnavailable
	}
	if err != nil {
		return fmt.Errorf("lock coupon: %w", err)
	}
	var globalUses, ownerUses int64
	if err := tx.QueryRow(ctx, "SELECT count(*) FROM coupon_redemptions WHERE coupon_id=$1", couponID).Scan(&globalUses); err != nil {
		return fmt.Errorf("count coupon redemptions: %w", err)
	}
	if err := tx.QueryRow(ctx, "SELECT count(*) FROM coupon_redemptions WHERE coupon_id=$1 AND owner_token_hash=$2", couponID, ownerTokenHash).Scan(&ownerUses); err != nil {
		return fmt.Errorf("count owner coupon redemptions: %w", err)
	}
	if (usageLimit.Valid && globalUses >= usageLimit.Int64) || ownerUses >= int64(perCustomerLimit) {
		return ErrCouponUnavailable
	}
	if _, err := tx.Exec(ctx, "UPDATE carts SET coupon_id=$2,updated_at=now() WHERE id=$1", cartID, couponID); err != nil {
		return fmt.Errorf("attach coupon to cart: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit coupon application: %w", err)
	}
	return nil
}

// RemoveCoupon detaches a coupon from an open cart.
func (s *Service) RemoveCoupon(ctx context.Context, ownerTokenHash []byte) error {
	if len(ownerTokenHash) != 32 {
		return errors.New("cart capability digest must be 32 bytes")
	}
	tag, err := s.pool.Exec(ctx, `UPDATE carts SET coupon_id=NULL,updated_at=now()
		WHERE owner_token_hash=$1 AND status='open' AND expires_at>now()`, ownerTokenHash)
	if err != nil {
		return fmt.Errorf("remove cart coupon: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrCartNotFound
	}
	return nil
}
