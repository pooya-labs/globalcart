package database

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pooya16pg-spec/globalcart/backend/internal/catalog"
)

func UpsertProducts(ctx context.Context, pool *pgxpool.Pool, products []catalog.Product) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, p := range products {
		images, err := json.Marshal(p.Images)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `INSERT INTO products
			(id,title,description,category,price,discount_percentage,rating,stock,brand,thumbnail,images,source,source_id,source_url,imported_at)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
			ON CONFLICT(id) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description,
			category=EXCLUDED.category, price=EXCLUDED.price, discount_percentage=EXCLUDED.discount_percentage,
			rating=EXCLUDED.rating, stock=EXCLUDED.stock, brand=EXCLUDED.brand, thumbnail=EXCLUDED.thumbnail,
			images=EXCLUDED.images, source=EXCLUDED.source, source_id=EXCLUDED.source_id,
			source_url=EXCLUDED.source_url, imported_at=EXCLUDED.imported_at, updated_at=now()`,
			p.ID, p.Title, p.Description, p.Category, p.Price, p.DiscountPercentage, p.Rating, p.Stock, p.Brand, p.Thumbnail, images, p.Source, p.SourceID, p.SourceURL, p.ImportedAt)
		if err != nil {
			return fmt.Errorf("upsert product %d: %w", p.ID, err)
		}
		categoryName := strings.TrimSpace(p.Category)
		if categoryName == "" {
			if _, err := tx.Exec(ctx, "UPDATE products SET category_id=NULL WHERE id=$1", p.ID); err != nil {
				return fmt.Errorf("clear product %d category relation: %w", p.ID, err)
			}
			continue
		}
		var categoryID int64
		err = tx.QueryRow(ctx, `INSERT INTO categories(name) VALUES($1)
			ON CONFLICT (lower(btrim(name))) DO UPDATE SET name=EXCLUDED.name
			RETURNING id`, categoryName).Scan(&categoryID)
		if err != nil {
			return fmt.Errorf("upsert category for product %d: %w", p.ID, err)
		}
		if _, err := tx.Exec(ctx, "UPDATE products SET category_id=$1 WHERE id=$2", categoryID, p.ID); err != nil {
			return fmt.Errorf("link category for product %d: %w", p.ID, err)
		}
	}
	return tx.Commit(ctx)
}
