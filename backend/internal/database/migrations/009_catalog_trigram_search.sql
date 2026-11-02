-- PostgreSQL trigram search is a fallback to full-text search for typos.
-- The pg_trgm extension provides similarity, word_similarity and index operators.
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX products_title_trgm_idx
    ON products USING GIN (lower(title) gin_trgm_ops);

CREATE INDEX products_brand_trgm_idx
    ON products USING GIN (lower(brand) gin_trgm_ops);
