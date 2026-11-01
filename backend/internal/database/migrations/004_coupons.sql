-- Internal coupon primitives. No public coupon administration or checkout routes are enabled.
CREATE TABLE coupons (
    id BIGSERIAL PRIMARY KEY,
    code_hash BYTEA NOT NULL UNIQUE CHECK (octet_length(code_hash) = 32),
    discount_type TEXT NOT NULL CHECK (discount_type IN ('percentage', 'fixed')),
    discount_value NUMERIC(12,2) NOT NULL CHECK (discount_value > 0),
    minimum_subtotal NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (minimum_subtotal >= 0),
    maximum_discount NUMERIC(12,2),
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    usage_limit BIGINT CHECK (usage_limit IS NULL OR usage_limit > 0),
    per_customer_limit INTEGER NOT NULL DEFAULT 1 CHECK (per_customer_limit > 0),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (ends_at > starts_at),
    CHECK (maximum_discount IS NULL OR maximum_discount > 0),
    CHECK (discount_type <> 'percentage' OR discount_value <= 100)
);
ALTER TABLE carts ADD COLUMN coupon_id BIGINT REFERENCES coupons(id) ON DELETE RESTRICT;
ALTER TABLE demo_orders ADD COLUMN coupon_id BIGINT REFERENCES coupons(id) ON DELETE RESTRICT;
ALTER TABLE demo_orders ADD COLUMN discount_amount NUMERIC(12,2) NOT NULL DEFAULT 0
    CHECK (discount_amount >= 0 AND discount_amount <= subtotal);
CREATE TABLE coupon_redemptions (
    id BIGSERIAL PRIMARY KEY,
    coupon_id BIGINT NOT NULL REFERENCES coupons(id) ON DELETE RESTRICT,
    order_id UUID NOT NULL UNIQUE REFERENCES demo_orders(id) ON DELETE RESTRICT,
    owner_token_hash BYTEA NOT NULL CHECK (octet_length(owner_token_hash) = 32),
    discount_amount NUMERIC(12,2) NOT NULL CHECK (discount_amount >= 0),
    redeemed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX coupon_redemptions_coupon_idx ON coupon_redemptions(coupon_id, redeemed_at);
CREATE INDEX coupon_redemptions_owner_idx ON coupon_redemptions(coupon_id, owner_token_hash);
CREATE INDEX carts_coupon_idx ON carts(coupon_id) WHERE coupon_id IS NOT NULL;
