package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pooya16pg-spec/globalcart/backend/internal/catalog"
)

func UpsertProducts(ctx context.Context, pool *pgxpool.Pool, products []catalog.Product) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var platformVendorID int64
	if err := tx.QueryRow(ctx, "SELECT id FROM vendors WHERE slug='globalcart-direct' AND vendor_type='platform' AND status='active'").Scan(&platformVendorID); err != nil {
		return fmt.Errorf("resolve active platform vendor: %w", err)
	}
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
			if err := upsertPlatformOffer(ctx, tx, platformVendorID, p); err != nil {
				return fmt.Errorf("sync platform offer for product %d: %w", p.ID, err)
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
		if err := upsertPlatformOffer(ctx, tx, platformVendorID, p); err != nil {
			return fmt.Errorf("sync platform offer for product %d: %w", p.ID, err)
		}
		if err := syncPlatformVariantOffers(ctx, tx, platformVendorID, p.ID); err != nil {
			return fmt.Errorf("sync platform variant offers for product %d: %w", p.ID, err)
		}
	}
	return tx.Commit(ctx)
}

func upsertPlatformOffer(ctx context.Context, tx pgx.Tx, vendorID int64, p catalog.Product) error {
	var offerID int64
	var oldPrice string
	var oldStock int32
	err := tx.QueryRow(ctx, "SELECT id,price::text,stock FROM vendor_products WHERE vendor_id=$1 AND product_id=$2 AND variant_id IS NULL FOR UPDATE", vendorID, p.ID).Scan(&offerID, &oldPrice, &oldStock)
	if errors.Is(err, pgx.ErrNoRows) {
		err = tx.QueryRow(ctx, "INSERT INTO vendor_products (vendor_id,product_id,price,currency,stock,status) VALUES($1,$2,$3,'USD',$4,'active') RETURNING id,price::text,stock", vendorID, p.ID, p.Price, p.Stock).Scan(&offerID, &oldPrice, &oldStock)
		if err != nil {
			return fmt.Errorf("insert platform offer: %w", err)
		}
		if _, err := tx.Exec(ctx, "INSERT INTO vendor_price_history (vendor_product_id,old_price,new_price,currency,change_source) VALUES($1,NULL,$2::numeric,'USD','initial_import')", offerID, oldPrice); err != nil {
			return fmt.Errorf("record initial platform price: %w", err)
		}
		if _, err := tx.Exec(ctx, "INSERT INTO vendor_inventory_events (vendor_product_id,quantity_delta,resulting_stock,reason,reference) VALUES($1,$2,$2,'opening_balance','catalog_import')", offerID, oldStock); err != nil {
			return fmt.Errorf("record initial platform stock: %w", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("read platform offer: %w", err)
	}
	var newPrice string
	var newStock int32
	if err := tx.QueryRow(ctx, "UPDATE vendor_products SET price=$2,currency='USD',stock=$3,status='active',updated_at=now() WHERE id=$1 RETURNING price::text,stock", offerID, p.Price, p.Stock).Scan(&newPrice, &newStock); err != nil {
		return fmt.Errorf("update platform offer: %w", err)
	}
	if oldPrice != newPrice {
		if _, err := tx.Exec(ctx, "INSERT INTO vendor_price_history (vendor_product_id,old_price,new_price,currency,change_source) VALUES($1,$2::numeric,$3::numeric,'USD','system_sync')", offerID, oldPrice, newPrice); err != nil {
			return fmt.Errorf("record platform price change: %w", err)
		}
	}
	if oldStock != newStock {
		if _, err := tx.Exec(ctx, "INSERT INTO vendor_inventory_events (vendor_product_id,quantity_delta,resulting_stock,reason,reference) VALUES($1,$2,$3,'correction','catalog_import')", offerID, int64(newStock)-int64(oldStock), newStock); err != nil {
			return fmt.Errorf("record platform stock change: %w", err)
		}
	}
	return nil
}

func syncPlatformVariantOffers(ctx context.Context, tx pgx.Tx, vendorID, productID int64) error {
	rows, err := tx.Query(ctx, `SELECT id, sku, price::text, stock, is_active
		FROM product_variants WHERE product_id=$1 ORDER BY id`, productID)
	if err != nil {
		return fmt.Errorf("list product variants: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var variantID int64
		var sku, sourcePrice string
		var sourceStock int32
		var active bool
		if err := rows.Scan(&variantID, &sku, &sourcePrice, &sourceStock, &active); err != nil {
			return fmt.Errorf("scan product variant: %w", err)
		}
		var offerID int64
		var oldPrice string
		var oldStock int32
		err := tx.QueryRow(ctx, `SELECT id,price::text,stock FROM vendor_products
			WHERE vendor_id=$1 AND variant_id=$2 FOR UPDATE`, vendorID, variantID).
			Scan(&offerID, &oldPrice, &oldStock)
		status := "paused"
		if active {
			status = "active"
		}
		if errors.Is(err, pgx.ErrNoRows) {
			err = tx.QueryRow(ctx, `INSERT INTO vendor_products
				(vendor_id,product_id,variant_id,seller_sku,price,currency,stock,status)
				VALUES($1,$2,$3,NULLIF($4,''),$5,'USD',$6,$7)
				RETURNING id,price::text,stock`, vendorID, productID, variantID, sku, sourcePrice, sourceStock, status).
				Scan(&offerID, &oldPrice, &oldStock)
			if err != nil {
				return fmt.Errorf("insert variant offer %d: %w", variantID, err)
			}
			if _, err := tx.Exec(ctx, `INSERT INTO vendor_price_history
				(vendor_product_id,old_price,new_price,currency,change_source)
				VALUES($1,NULL,$2::numeric,'USD','initial_import')`, offerID, oldPrice); err != nil {
				return fmt.Errorf("record initial variant price: %w", err)
			}
			if _, err := tx.Exec(ctx, `INSERT INTO vendor_inventory_events
				(vendor_product_id,quantity_delta,resulting_stock,reason,reference)
				VALUES($1,$2,$2,'opening_balance','catalog_import')`, offerID, oldStock); err != nil {
				return fmt.Errorf("record initial variant stock: %w", err)
			}
			continue
		}
		if err != nil {
			return fmt.Errorf("read variant offer %d: %w", variantID, err)
		}
		var newPrice string
		var newStock int32
		if err := tx.QueryRow(ctx, `UPDATE vendor_products
			SET seller_sku=NULLIF($2,''),price=$3,currency='USD',stock=$4,status=$5,updated_at=now()
			WHERE id=$1 RETURNING price::text,stock`, offerID, sku, sourcePrice, sourceStock, status).
			Scan(&newPrice, &newStock); err != nil {
			return fmt.Errorf("update variant offer %d: %w", variantID, err)
		}
		if oldPrice != newPrice {
			if _, err := tx.Exec(ctx, `INSERT INTO vendor_price_history
				(vendor_product_id,old_price,new_price,currency,change_source)
				VALUES($1,$2::numeric,$3::numeric,'USD','system_sync')`, offerID, oldPrice, newPrice); err != nil {
				return fmt.Errorf("record variant price change: %w", err)
			}
		}
		if oldStock != newStock {
			if _, err := tx.Exec(ctx, `INSERT INTO vendor_inventory_events
				(vendor_product_id,quantity_delta,resulting_stock,reason,reference)
				VALUES($1,$2,$3,'correction','catalog_import')`,
				offerID, int64(newStock)-int64(oldStock), newStock); err != nil {
				return fmt.Errorf("record variant inventory change: %w", err)
			}
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate product variants: %w", err)
	}
	return nil
}
