-- Vendor-aware cart lines and parent/child order structure.
-- Legacy cart/order rows remain vendor_product_id=NULL and keep their old lifecycle.
ALTER TABLE vendor_products
    ADD CONSTRAINT vendor_products_id_product_uidx UNIQUE (id, product_id);

ALTER TABLE cart_items
    ADD COLUMN vendor_product_id BIGINT,
    ADD CONSTRAINT cart_items_vendor_product_fk
        FOREIGN KEY (vendor_product_id, product_id)
        REFERENCES vendor_products(id, product_id) ON DELETE RESTRICT;

DROP INDEX cart_items_product_uidx;
DROP INDEX cart_items_variant_uidx;
CREATE UNIQUE INDEX cart_items_legacy_product_uidx
    ON cart_items(cart_id, product_id)
    WHERE vendor_product_id IS NULL AND variant_id IS NULL;
CREATE UNIQUE INDEX cart_items_legacy_variant_uidx
    ON cart_items(cart_id, variant_id)
    WHERE vendor_product_id IS NULL AND variant_id IS NOT NULL;
CREATE UNIQUE INDEX cart_items_vendor_offer_uidx
    ON cart_items(cart_id, vendor_product_id)
    WHERE vendor_product_id IS NOT NULL;
CREATE INDEX cart_items_vendor_offer_idx
    ON cart_items(vendor_product_id, cart_id) WHERE vendor_product_id IS NOT NULL;

CREATE TABLE demo_vendor_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES demo_orders(id) ON DELETE RESTRICT,
    vendor_id BIGINT NOT NULL REFERENCES vendors(id) ON DELETE RESTRICT,
    status TEXT NOT NULL DEFAULT 'awaiting_demo_payment'
        CHECK (status IN ('awaiting_demo_payment', 'paid_simulated', 'cancelled', 'expired')),
    currency CHAR(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    subtotal NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (subtotal >= 0),
    discount_amount NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (discount_amount >= 0 AND discount_amount <= subtotal),
    total NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (total = subtotal - discount_amount),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(order_id, vendor_id),
    UNIQUE(id, order_id)
);
CREATE INDEX demo_vendor_orders_vendor_status_idx
    ON demo_vendor_orders(vendor_id, status, created_at DESC);

ALTER TABLE demo_order_items
    ADD COLUMN vendor_product_id BIGINT,
    ADD COLUMN vendor_order_id UUID,
    ADD CONSTRAINT demo_order_items_vendor_product_fk
        FOREIGN KEY (vendor_product_id, product_id)
        REFERENCES vendor_products(id, product_id) ON DELETE RESTRICT,
    ADD CONSTRAINT demo_order_items_vendor_order_fk
        FOREIGN KEY (vendor_order_id, order_id)
        REFERENCES demo_vendor_orders(id, order_id) ON DELETE RESTRICT,
    ADD CONSTRAINT demo_order_items_vendor_pair_check
        CHECK ((vendor_product_id IS NULL) = (vendor_order_id IS NULL));

ALTER TABLE inventory_reservations
    ADD COLUMN vendor_product_id BIGINT,
    ADD COLUMN vendor_order_id UUID,
    ADD CONSTRAINT inventory_reservations_vendor_product_fk
        FOREIGN KEY (vendor_product_id, product_id)
        REFERENCES vendor_products(id, product_id) ON DELETE RESTRICT,
    ADD CONSTRAINT inventory_reservations_vendor_order_fk
        FOREIGN KEY (vendor_order_id, order_id)
        REFERENCES demo_vendor_orders(id, order_id) ON DELETE RESTRICT,
    ADD CONSTRAINT inventory_reservations_vendor_pair_check
        CHECK ((vendor_product_id IS NULL) = (vendor_order_id IS NULL));

DROP INDEX inventory_reservations_product_active_uidx;
DROP INDEX inventory_reservations_variant_active_uidx;
CREATE UNIQUE INDEX inventory_reservations_legacy_product_active_uidx
    ON inventory_reservations(order_id, product_id)
    WHERE status='active' AND vendor_product_id IS NULL AND variant_id IS NULL;
CREATE UNIQUE INDEX inventory_reservations_legacy_variant_active_uidx
    ON inventory_reservations(order_id, variant_id)
    WHERE status='active' AND vendor_product_id IS NULL AND variant_id IS NOT NULL;
CREATE UNIQUE INDEX inventory_reservations_vendor_offer_active_uidx
    ON inventory_reservations(order_id, vendor_product_id)
    WHERE status='active' AND vendor_product_id IS NOT NULL;
CREATE INDEX inventory_reservations_vendor_expiry_idx
    ON inventory_reservations(vendor_product_id, expires_at)
    WHERE status='active' AND vendor_product_id IS NOT NULL;
