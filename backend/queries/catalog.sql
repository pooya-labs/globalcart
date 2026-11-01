-- name: ListProducts :many
SELECT id, title, description, category, price::float8 AS price,
       discount_percentage::float8 AS discount_percentage, rating::float8 AS rating,
       stock, brand, thumbnail, images, source, source_id, source_url, imported_at
FROM products
WHERE (sqlc.arg(category)::text = '' OR category = sqlc.arg(category)::text)
  AND (sqlc.arg(query)::text = '' OR
       to_tsvector('simple', title || ' ' || description) @@
       plainto_tsquery('simple', sqlc.arg(query)::text))
ORDER BY id
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
