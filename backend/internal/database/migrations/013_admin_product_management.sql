-- M06.2 keeps administrator edits separate from supplier-owned catalog data.
-- The source record remains importable and is never overwritten by editorial changes.
CREATE TABLE admin_product_overrides (
    product_id BIGINT PRIMARY KEY REFERENCES products(id) ON DELETE RESTRICT,
    title_override TEXT,
    description_override TEXT,
    brand_override TEXT,
    visibility_status TEXT NOT NULL DEFAULT 'active'
        CHECK (visibility_status IN ('active', 'archived')),
    revision BIGINT NOT NULL DEFAULT 0 CHECK (revision >= 0),
    created_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    updated_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    archived_by_user_id UUID REFERENCES users(id) ON DELETE RESTRICT,
    archived_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT admin_product_title_override_check CHECK (
        title_override IS NULL OR (
            length(btrim(title_override)) BETWEEN 1 AND 300
        )
    ),
    CONSTRAINT admin_product_description_override_check CHECK (
        description_override IS NULL OR length(description_override) <= 10000
    ),
    CONSTRAINT admin_product_brand_override_check CHECK (
        brand_override IS NULL OR length(brand_override) <= 120
    ),
    CONSTRAINT admin_product_archive_state_check CHECK (
        (visibility_status = 'archived') = (archived_at IS NOT NULL)
        AND (visibility_status = 'archived') = (archived_by_user_id IS NOT NULL)
    )
);
CREATE INDEX admin_product_overrides_visibility_updated_idx
    ON admin_product_overrides(visibility_status, updated_at DESC, product_id);

CREATE TABLE admin_product_revisions (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    revision BIGINT NOT NULL CHECK (revision > 0),
    action TEXT NOT NULL CHECK (action IN ('updated', 'archived', 'restored')),
    before_state JSONB NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(before_state) = 'object'),
    after_state JSONB NOT NULL CHECK (jsonb_typeof(after_state) = 'object'),
    actor_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(product_id, revision)
);
CREATE INDEX admin_product_revisions_product_time_idx
    ON admin_product_revisions(product_id, revision DESC, created_at DESC);

CREATE OR REPLACE FUNCTION reject_admin_product_revision_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $admin_product_revision_guard$
BEGIN
    RAISE EXCEPTION 'admin_product_revisions is append-only' USING ERRCODE = '55000';
END;
$admin_product_revision_guard$;

CREATE TRIGGER admin_product_revisions_no_update_delete
    BEFORE UPDATE OR DELETE ON admin_product_revisions
    FOR EACH ROW EXECUTE FUNCTION reject_admin_product_revision_mutation();
CREATE TRIGGER admin_product_revisions_no_truncate
    BEFORE TRUNCATE ON admin_product_revisions
    FOR EACH STATEMENT EXECUTE FUNCTION reject_admin_product_revision_mutation();
