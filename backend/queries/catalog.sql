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
SELECT p.id, COALESCE(apo.title_override, p.title) AS title,
       COALESCE(apo.description_override, p.description) AS description,
       p.category, p.price::float8 AS price,
       p.discount_percentage::float8 AS discount_percentage, p.rating::float8 AS rating,
       p.stock, COALESCE(apo.brand_override, p.brand) AS brand,
       p.thumbnail, p.images, p.source, p.source_id, p.source_url, p.imported_at
FROM products AS p
LEFT JOIN admin_product_overrides AS apo ON apo.product_id = p.id
CROSS JOIN filter_values AS f
WHERE COALESCE(apo.visibility_status, 'active') <> 'archived'
  AND (f.category IS NULL OR EXISTS (
    SELECT 1 FROM categories AS c
    WHERE c.id = p.category_id
      AND lower(btrim(c.name)) = lower(btrim(f.category))
  ))
  AND (
       f.query IS NULL
       OR to_tsvector('simple', COALESCE(apo.title_override, p.title) || ' ' || COALESCE(apo.description_override, p.description)) @@ plainto_tsquery('simple', f.query)
       OR (
           length(f.query) >= 3
           AND (
               lower(COALESCE(apo.title_override, p.title)) % lower(f.query)
               OR lower(f.query) <% lower(COALESCE(apo.title_override, p.title))
               OR lower(COALESCE(apo.brand_override, p.brand)) % lower(f.query)
           )
       )
  )
  AND (f.brand IS NULL OR lower(btrim(COALESCE(apo.brand_override, p.brand))) = lower(btrim(f.brand)))
  AND (f.min_price IS NULL OR p.price >= f.min_price)
  AND (f.max_price IS NULL OR p.price <= f.max_price)
  AND (f.min_rating IS NULL OR p.rating >= f.min_rating)
  AND (f.max_rating IS NULL OR p.rating <= f.max_rating)
ORDER BY
    CASE
        WHEN f.query IS NULL THEN 0::real
        WHEN to_tsvector('simple', p.title || ' ' || p.description) @@ plainto_tsquery('simple', f.query)
            THEN 2::real + ts_rank(to_tsvector('simple', p.title || ' ' || p.description), plainto_tsquery('simple', f.query))
        ELSE GREATEST(
            word_similarity(lower(f.query), lower(COALESCE(apo.title_override, p.title))),
            similarity(lower(COALESCE(apo.title_override, p.title)), lower(f.query)),
            similarity(lower(COALESCE(apo.brand_override, p.brand)), lower(f.query))
        )
    END DESC,
    p.id
LIMIT sqlc.arg(page_limit)::integer
OFFSET sqlc.arg(page_offset)::integer;

-- name: GetProduct :one
SELECT p.id, COALESCE(apo.title_override, p.title) AS title,
       COALESCE(apo.description_override, p.description) AS description,
       p.category, p.price::float8 AS price,
       p.discount_percentage::float8 AS discount_percentage, p.rating::float8 AS rating,
       p.stock, COALESCE(apo.brand_override, p.brand) AS brand,
       p.thumbnail, p.images, p.source, p.source_id, p.source_url, p.imported_at
FROM products AS p
LEFT JOIN admin_product_overrides AS apo ON apo.product_id = p.id
WHERE p.id = sqlc.arg(id)::bigint
  AND COALESCE(apo.visibility_status, 'active') <> 'archived';

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
    AND NOT EXISTS (SELECT 1 FROM admin_product_overrides AS apo WHERE apo.product_id = p.id AND apo.visibility_status = 'archived')
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
LEFT JOIN admin_product_overrides AS apo ON apo.product_id = p.id
CROSS JOIN filter_values AS f
WHERE COALESCE(apo.visibility_status, 'active') <> 'archived'
  AND (f.category IS NULL OR EXISTS (
    SELECT 1 FROM categories AS c
    WHERE c.id = p.category_id
      AND lower(btrim(c.name)) = lower(btrim(f.category))
  ))
  AND (
       f.query IS NULL
       OR to_tsvector('simple', p.title || ' ' || p.description) @@ plainto_tsquery('simple', f.query)
       OR (
           length(f.query) >= 3
           AND (
               lower(COALESCE(apo.title_override, p.title)) % lower(f.query)
               OR lower(f.query) <% lower(COALESCE(apo.title_override, p.title))
               OR lower(COALESCE(apo.brand_override, p.brand)) % lower(f.query)
           )
       )
  )
  AND (f.brand IS NULL OR lower(btrim(COALESCE(apo.brand_override, p.brand))) = lower(btrim(f.brand)))
  AND (f.min_price IS NULL OR p.price >= f.min_price)
  AND (f.max_price IS NULL OR p.price <= f.max_price)
  AND (f.min_rating IS NULL OR p.rating >= f.min_rating)
  AND (f.max_rating IS NULL OR p.rating <= f.max_rating);
