-- Add filter indexes for the read-only catalog API facets.
-- The normalized brand expression matches the case/whitespace-insensitive
-- predicate used by the generated sqlc queries.
CREATE INDEX products_brand_normalized_idx
    ON products (lower(btrim(brand)))
    WHERE length(btrim(brand)) > 0;

CREATE INDEX products_price_idx ON products(price);
CREATE INDEX products_rating_idx ON products(rating);
