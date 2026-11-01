-- Multi-vendor catalog foundation. Existing checkout continues to use legacy product stock
-- until a separately tested vendor-aware order split is introduced.
CREATE TABLE vendors (
    id BIGSERIAL PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    vendor_type TEXT NOT NULL DEFAULT 'merchant'
        CHECK (vendor_type IN ('platform', 'merchant')),
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'active', 'suspended', 'closed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT vendors_slug_format CHECK (slug ~ '^[a-z0-9]+([_-][a-z0-9]+)*$'),
    CONSTRAINT vendors_display_name_nonempty CHECK (length(btrim(display_name)) > 0)
);
CREATE UNIQUE INDEX vendors_single_platform_uidx
    ON vendors (vendor_type) WHERE vendor_type = 'platform';
CREATE INDEX vendors_status_created_idx ON vendors(status, created_at DESC);

CREATE TABLE vendor_products (
    id BIGSERIAL PRIMARY KEY,
    vendor_id BIGINT NOT NULL REFERENCES vendors(id) ON DELETE RESTRICT,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    variant_id BIGINT,
    seller_sku TEXT,
    price NUMERIC(12,2) NOT NULL CHECK (price >= 0),
    currency CHAR(3) NOT NULL DEFAULT 'USD' CHECK (currency ~ '^[A-Z]{3}$'),
    stock INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
    status TEXT NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'active', 'paused', 'archived')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT vendor_products_variant_product_fk
        FOREIGN KEY (variant_id, product_id)
        REFERENCES product_variants(id, product_id) ON DELETE RESTRICT,
    CONSTRAINT vendor_products_sku_nonempty
        CHECK (seller_sku IS NULL OR length(btrim(seller_sku)) > 0)
);
CREATE UNIQUE INDEX vendor_products_base_offer_uidx
    ON vendor_products(vendor_id, product_id) WHERE variant_id IS NULL;
CREATE UNIQUE INDEX vendor_products_variant_offer_uidx
    ON vendor_products(vendor_id, variant_id) WHERE variant_id IS NOT NULL;
CREATE UNIQUE INDEX vendor_products_seller_sku_uidx
    ON vendor_products(vendor_id, lower(btrim(seller_sku)))
    WHERE seller_sku IS NOT NULL AND seller_sku <> '';
CREATE INDEX vendor_products_product_active_idx
    ON vendor_products(product_id, status, price) WHERE status = 'active';
CREATE INDEX vendor_products_vendor_status_idx
    ON vendor_products(vendor_id, status, updated_at DESC);

CREATE TABLE vendor_price_history (
    id BIGSERIAL PRIMARY KEY,
    vendor_product_id BIGINT NOT NULL REFERENCES vendor_products(id) ON DELETE RESTRICT,
    old_price NUMERIC(12,2) CHECK (old_price IS NULL OR old_price >= 0),
    new_price NUMERIC(12,2) NOT NULL CHECK (new_price >= 0),
    currency CHAR(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    change_source TEXT NOT NULL
        CHECK (change_source IN ('initial_import', 'merchant_update', 'admin_adjustment', 'system_sync')),
    changed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX vendor_price_history_offer_time_idx
    ON vendor_price_history(vendor_product_id, changed_at DESC, id DESC);

CREATE TABLE vendor_inventory_events (
    id BIGSERIAL PRIMARY KEY,
    vendor_product_id BIGINT NOT NULL REFERENCES vendor_products(id) ON DELETE RESTRICT,
    quantity_delta INTEGER NOT NULL,
    resulting_stock INTEGER NOT NULL CHECK (resulting_stock >= 0),
    reason TEXT NOT NULL
        CHECK (reason IN ('opening_balance', 'stock_adjustment', 'restock', 'return', 'correction')),
    reference TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT vendor_inventory_event_delta_check
        CHECK (quantity_delta <> 0 OR reason = 'opening_balance')
);
CREATE INDEX vendor_inventory_events_offer_time_idx
    ON vendor_inventory_events(vendor_product_id, created_at DESC, id DESC);

-- Preserve the imported single-seller catalog as an explicit platform offer.
INSERT INTO vendors (slug, display_name, vendor_type, status)
VALUES ('globalcart-direct', 'GlobalCart Direct', 'platform', 'active');

INSERT INTO vendor_products
    (vendor_id, product_id, variant_id, seller_sku, price, currency, stock, status)
SELECT v.id, p.id, NULL, NULL, p.price, 'USD', p.stock, 'active'
FROM vendors AS v
CROSS JOIN products AS p
WHERE v.slug = 'globalcart-direct';

INSERT INTO vendor_products
    (vendor_id, product_id, variant_id, seller_sku, price, currency, stock, status)
SELECT v.id, pv.product_id, pv.id, pv.sku, pv.price, 'USD', pv.stock,
       CASE WHEN pv.is_active THEN 'active' ELSE 'paused' END
FROM vendors AS v
CROSS JOIN product_variants AS pv
WHERE v.slug = 'globalcart-direct';

-- Initial price and stock states are auditable from the moment of migration.
INSERT INTO vendor_price_history
    (vendor_product_id, old_price, new_price, currency, change_source)
SELECT id, NULL, price, currency, 'initial_import'
FROM vendor_products;

INSERT INTO vendor_inventory_events
    (vendor_product_id, quantity_delta, resulting_stock, reason, reference)
SELECT id, stock, stock, 'opening_balance', 'migration:005_vendor_foundation'
FROM vendor_products;
