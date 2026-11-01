-- Extend the imported catalog without changing the existing read API contract.
CREATE TABLE categories (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    parent_id BIGINT REFERENCES categories(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT categories_name_nonempty CHECK (length(btrim(name)) > 0),
    CONSTRAINT categories_not_own_parent CHECK (parent_id IS NULL OR parent_id <> id)
);
CREATE UNIQUE INDEX categories_normalized_name_uidx ON categories (lower(btrim(name)));
INSERT INTO categories (name)
SELECT DISTINCT ON (lower(btrim(category))) btrim(category)
FROM products
WHERE length(btrim(category)) > 0
ORDER BY lower(btrim(category)), btrim(category);
ALTER TABLE products ADD COLUMN category_id BIGINT REFERENCES categories(id) ON DELETE SET NULL;
UPDATE products AS p
SET category_id = c.id
FROM categories AS c
WHERE lower(btrim(p.category)) = lower(btrim(c.name));
CREATE INDEX products_category_id_idx ON products(category_id);
CREATE TABLE product_variants (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sku TEXT,
    title TEXT NOT NULL DEFAULT '',
    price NUMERIC(12,2) NOT NULL CHECK (price >= 0),
    stock INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
    attributes JSONB NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(attributes) = 'object'),
    source TEXT NOT NULL DEFAULT '',
    source_id TEXT NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT product_variants_source_identity_unique UNIQUE (source, source_id)
);
CREATE UNIQUE INDEX product_variants_sku_uidx ON product_variants(sku) WHERE sku IS NOT NULL AND sku <> '';
CREATE INDEX product_variants_product_id_idx ON product_variants(product_id);
