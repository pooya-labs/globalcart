-- Demo commerce persistence primitives. No public write API is enabled by this migration.
CREATE UNIQUE INDEX product_variants_id_product_uidx ON product_variants(id, product_id);

CREATE TABLE carts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_token_hash BYTEA NOT NULL UNIQUE CHECK (octet_length(owner_token_hash) = 32),
    status TEXT NOT NULL DEFAULT 'open'
        CHECK (status IN ('open', 'checked_out', 'abandoned', 'expired')),
    currency CHAR(3) NOT NULL DEFAULT 'USD' CHECK (currency ~ '^[A-Z]{3}$'),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX carts_expiry_idx ON carts(expires_at) WHERE status = 'open';

CREATE TABLE cart_items (
    id BIGSERIAL PRIMARY KEY,
    cart_id UUID NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    variant_id BIGINT,
    quantity INTEGER NOT NULL CHECK (quantity BETWEEN 1 AND 99),
    unit_price_snapshot NUMERIC(12,2) NOT NULL CHECK (unit_price_snapshot >= 0),
    currency CHAR(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    product_title_snapshot TEXT NOT NULL,
    variant_title_snapshot TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT cart_items_variant_product_fk
        FOREIGN KEY (variant_id, product_id)
        REFERENCES product_variants(id, product_id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX cart_items_product_uidx
    ON cart_items(cart_id, product_id) WHERE variant_id IS NULL;
CREATE UNIQUE INDEX cart_items_variant_uidx
    ON cart_items(cart_id, variant_id) WHERE variant_id IS NOT NULL;
CREATE INDEX cart_items_cart_idx ON cart_items(cart_id);

CREATE TABLE demo_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id UUID NOT NULL REFERENCES carts(id) ON DELETE RESTRICT,
    idempotency_key_hash BYTEA NOT NULL CHECK (octet_length(idempotency_key_hash) = 32),
    status TEXT NOT NULL DEFAULT 'awaiting_demo_payment'
        CHECK (status IN ('awaiting_demo_payment', 'paid_simulated', 'cancelled', 'expired')),
    currency CHAR(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    subtotal NUMERIC(12,2) NOT NULL CHECK (subtotal >= 0),
    total NUMERIC(12,2) NOT NULL CHECK (total >= 0 AND total = subtotal),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (cart_id, idempotency_key_hash)
);
CREATE INDEX demo_orders_status_created_idx ON demo_orders(status, created_at DESC);

CREATE TABLE demo_order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES demo_orders(id) ON DELETE RESTRICT,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    variant_id BIGINT,
    quantity INTEGER NOT NULL CHECK (quantity BETWEEN 1 AND 99),
    unit_price_snapshot NUMERIC(12,2) NOT NULL CHECK (unit_price_snapshot >= 0),
    currency CHAR(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    product_title_snapshot TEXT NOT NULL,
    variant_title_snapshot TEXT NOT NULL DEFAULT '',
    CONSTRAINT demo_order_items_variant_product_fk
        FOREIGN KEY (variant_id, product_id)
        REFERENCES product_variants(id, product_id) ON DELETE RESTRICT
);
CREATE INDEX demo_order_items_order_idx ON demo_order_items(order_id);

CREATE TABLE inventory_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES demo_orders(id) ON DELETE RESTRICT,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    variant_id BIGINT,
    quantity INTEGER NOT NULL CHECK (quantity BETWEEN 1 AND 99),
    status TEXT NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'released', 'consumed', 'expired')),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT inventory_reservations_variant_product_fk
        FOREIGN KEY (variant_id, product_id)
        REFERENCES product_variants(id, product_id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX inventory_reservations_product_active_uidx
    ON inventory_reservations(order_id, product_id)
    WHERE status = 'active' AND variant_id IS NULL;
CREATE UNIQUE INDEX inventory_reservations_variant_active_uidx
    ON inventory_reservations(order_id, variant_id)
    WHERE status = 'active' AND variant_id IS NOT NULL;
CREATE INDEX inventory_reservations_expiry_idx
    ON inventory_reservations(expires_at) WHERE status = 'active';

CREATE TABLE demo_order_events (
    id BIGSERIAL PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES demo_orders(id) ON DELETE RESTRICT,
    event_type TEXT NOT NULL CHECK (length(btrim(event_type)) > 0),
    payload JSONB NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(payload) = 'object'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX demo_order_events_order_idx ON demo_order_events(order_id, created_at, id);
