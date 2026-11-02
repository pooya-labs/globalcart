-- name: ListProducts :many
WITH filter_values AS (
    SELECT
        NULLIF(sqlc.arg(category)::text, '') AS category,
        NULLIF(sqlc.arg(query)::text, '') AS query,
        NULLIF(sqlc.arg(brand)::text, '') AS brand,
        NULLIF(sqlc.arg(min_price)::text, '')::numeric AS min_price,
        NULLIF(sqlc.arg(max_price)::text, '')::numeric AS max_price,
        NULLIF(sqlc.arg(min_rating)::text, '')::numeric AS min_rating,
        NULLIF(sqlc.arg(max_rating)::text, '')::numeric AS max_rating
)
SELECT p.id, p.title, p.description, p.category, p.price::float8 AS price,
       p.discount_percentage::float8 AS discount_percentage, p.rating::float8 AS rating,
       p.stock, p.brand, p.thumbnail, p.images, p.source, p.source_id, p.source_url, p.imported_at
FROM products AS p
CROSS JOIN filter_values AS f
WHERE (f.category IS NULL OR EXISTS (
    SELECT 1 FROM categories AS c
    WHERE c.id = p.category_id
      AND lower(btrim(c.name)) = lower(btrim(f.category))
  ))
  AND (f.query IS NULL OR
       to_tsvector('simple', p.title || ' ' || p.description) @@
       plainto_tsquery('simple', f.query))
  AND (f.brand IS NULL OR lower(btrim(p.brand)) = lower(btrim(f.brand)))
  AND (f.min_price IS NULL OR p.price >= f.min_price)
  AND (f.max_price IS NULL OR p.price <= f.max_price)
  AND (f.min_rating IS NULL OR p.rating >= f.min_rating)
  AND (f.max_rating IS NULL OR p.rating <= f.max_rating)
ORDER BY p.id
LIMIT sqlc.arg(page_limit)::integer
OFFSET sqlc.arg(page_offset)::integer;

-- name: GetProduct :one
SELECT id, title, description, category, price::float8 AS price,
       discount_percentage::float8 AS discount_percentage, rating::float8 AS rating,
       stock, brand, thumbnail, images, source, source_id, source_url, imported_at
FROM products
WHERE id = sqlc.arg(id)::bigint;

-- name: ListProductVariants :many
SELECT id, product_id, sku, title, price::float8 AS price, stock, attributes, source, source_id,
       is_active, created_at, updated_at
FROM product_variants
WHERE product_id = sqlc.arg(product_id)::bigint
  AND is_active = TRUE
ORDER BY id;

-- name: ListCategories :many
SELECT c.id, c.name, COALESCE(c.parent_id, 0)::bigint AS parent_id,
       count(p.id)::bigint AS product_count
FROM categories AS c
LEFT JOIN products AS p ON p.category_id = c.id
GROUP BY c.id
ORDER BY lower(c.name), c.id;

-- name: CountProducts :one
WITH filter_values AS (
    SELECT
        NULLIF(sqlc.arg(category)::text, '') AS category,
        NULLIF(sqlc.arg(query)::text, '') AS query,
        NULLIF(sqlc.arg(brand)::text, '') AS brand,
        NULLIF(sqlc.arg(min_price)::text, '')::numeric AS min_price,
        NULLIF(sqlc.arg(max_price)::text, '')::numeric AS max_price,
        NULLIF(sqlc.arg(min_rating)::text, '')::numeric AS min_rating,
        NULLIF(sqlc.arg(max_rating)::text, '')::numeric AS max_rating
)
SELECT count(*)::bigint
FROM products AS p
CROSS JOIN filter_values AS f
WHERE (f.category IS NULL OR EXISTS (
    SELECT 1 FROM categories AS c
    WHERE c.id = p.category_id
      AND lower(btrim(c.name)) = lower(btrim(f.category))
  ))
  AND (f.query IS NULL OR
       to_tsvector('simple', p.title || ' ' || p.description) @@
       plainto_tsquery('simple', f.query))
  AND (f.brand IS NULL OR lower(btrim(p.brand)) = lower(btrim(f.brand)))
  AND (f.min_price IS NULL OR p.price >= f.min_price)
  AND (f.max_price IS NULL OR p.price <= f.max_price)
  AND (f.min_rating IS NULL OR p.rating >= f.min_rating)
  AND (f.max_rating IS NULL OR p.rating <= f.max_rating);
