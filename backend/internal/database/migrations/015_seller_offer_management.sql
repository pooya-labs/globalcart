-- M06.5 seller offer lifecycle and optimistic concurrency.
ALTER TABLE vendor_products ADD COLUMN revision BIGINT NOT NULL DEFAULT 0 CHECK (revision >= 0);

CREATE TABLE vendor_product_revisions (
    id BIGSERIAL PRIMARY KEY,
    vendor_product_id BIGINT NOT NULL REFERENCES vendor_products(id) ON DELETE RESTRICT,
    revision BIGINT NOT NULL CHECK (revision > 0),
    action TEXT NOT NULL CHECK (action IN ('created', 'updated', 'archived')),
    before_state JSONB NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(before_state) = 'object'),
    after_state JSONB NOT NULL CHECK (jsonb_typeof(after_state) = 'object'),
    actor_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(vendor_product_id, revision)
);
CREATE INDEX vendor_product_revisions_offer_time_idx
    ON vendor_product_revisions(vendor_product_id, revision DESC, created_at DESC);

CREATE OR REPLACE FUNCTION advance_vendor_product_revision()
RETURNS trigger LANGUAGE plpgsql AS $vendor_product_revision$
BEGIN
    NEW.revision := OLD.revision + 1;
    RETURN NEW;
END;
$vendor_product_revision$;
CREATE TRIGGER vendor_products_revision_increment
    BEFORE UPDATE ON vendor_products FOR EACH ROW
    EXECUTE FUNCTION advance_vendor_product_revision();

CREATE OR REPLACE FUNCTION reject_vendor_product_revision_mutation()
RETURNS trigger LANGUAGE plpgsql AS $vendor_product_revision_guard$
BEGIN
    RAISE EXCEPTION 'vendor_product_revisions is append-only' USING ERRCODE = '55000';
END;
$vendor_product_revision_guard$;
CREATE TRIGGER vendor_product_revisions_no_update_delete
    BEFORE UPDATE OR DELETE ON vendor_product_revisions FOR EACH ROW
    EXECUTE FUNCTION reject_vendor_product_revision_mutation();
CREATE TRIGGER vendor_product_revisions_no_truncate
    BEFORE TRUNCATE ON vendor_product_revisions FOR EACH STATEMENT
    EXECUTE FUNCTION reject_vendor_product_revision_mutation();
