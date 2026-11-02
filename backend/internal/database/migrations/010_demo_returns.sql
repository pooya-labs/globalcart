-- Internal-only full-order return and simulated refund lifecycle.
-- No payment provider, card data, payout or inventory mutation is introduced.
ALTER TABLE demo_orders
    DROP CONSTRAINT demo_orders_status_check,
    ADD CONSTRAINT demo_orders_status_check
        CHECK (status IN ('awaiting_demo_payment', 'paid_simulated', 'cancelled', 'expired', 'refunded_simulated'));

ALTER TABLE demo_vendor_orders
    DROP CONSTRAINT demo_vendor_orders_status_check,
    ADD CONSTRAINT demo_vendor_orders_status_check
        CHECK (status IN ('awaiting_demo_payment', 'paid_simulated', 'cancelled', 'expired', 'refunded_simulated'));

CREATE TABLE demo_returns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL UNIQUE REFERENCES demo_orders(id) ON DELETE RESTRICT,
    reason TEXT NOT NULL
        CHECK (reason IN ('changed_mind', 'damaged_demo', 'not_as_described', 'wrong_item')),
    status TEXT NOT NULL DEFAULT 'requested'
        CHECK (status IN ('requested', 'approved', 'rejected', 'refunded_simulated')),
    currency CHAR(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    refund_amount NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (refund_amount >= 0),
    reviewed_at TIMESTAMPTZ,
    refunded_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT demo_returns_lifecycle_check CHECK (
        (status = 'requested' AND reviewed_at IS NULL AND refunded_at IS NULL AND refund_amount = 0)
        OR (status = 'approved' AND reviewed_at IS NOT NULL AND refunded_at IS NULL)
        OR (status = 'rejected' AND reviewed_at IS NOT NULL AND refunded_at IS NULL AND refund_amount = 0)
        OR (status = 'refunded_simulated' AND reviewed_at IS NOT NULL AND refunded_at IS NOT NULL)
    )
);
CREATE INDEX demo_returns_status_created_idx ON demo_returns(status, created_at DESC);
