package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pooya16pg-spec/globalcart/backend/internal/catalog"
	"github.com/pooya16pg-spec/globalcart/backend/internal/dbgen"
)

type Handler struct {
	db      *pgxpool.Pool
	origins map[string]bool
	queries *dbgen.Queries
}

func New(db *pgxpool.Pool, origins map[string]bool) http.Handler {
	h := &Handler{db: db, origins: origins, queries: dbgen.New(db)}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("GET /readyz", h.ready)
	mux.HandleFunc("GET /api/v1/products", h.listProducts)
	mux.HandleFunc("GET /api/v1/categories", h.listCategories)
	mux.HandleFunc("GET /api/v1/products/{id}", h.getProduct)
	mux.HandleFunc("GET /api/v1/products/{id}/variants", h.listProductVariants)
	return h.cors(mux)
}

func (h *Handler) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && h.origins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "globalcart-api"})
}

func (h *Handler) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.db.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready", "database": "connected"})
}

type categoryResponse struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	ParentID     *int64 `json:"parentId,omitempty"`
	ProductCount int64  `json:"productCount"`
}

func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	rows, err := h.queries.ListCategories(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "category_query_failed")
		return
	}
	items := make([]categoryResponse, 0, len(rows))
	for _, row := range rows {
		item := categoryResponse{ID: row.ID, Name: row.Name, ProductCount: row.ProductCount}
		if row.ParentID != 0 {
			parentID := row.ParentID
			item.ParentID = &parentID
		}
		items = append(items, item)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "count": len(items)})
}

func (h *Handler) listProducts(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	category := strings.TrimSpace(r.URL.Query().Get("category"))
	brand := strings.TrimSpace(r.URL.Query().Get("brand"))
	minPrice, minPriceValue, validMinPrice := parseOptionalDecimal(r.URL.Query().Get("minPrice"), 9_999_999_999.99)
	maxPrice, maxPriceValue, validMaxPrice := parseOptionalDecimal(r.URL.Query().Get("maxPrice"), 9_999_999_999.99)
	minRating, minRatingValue, validMinRating := parseOptionalDecimal(r.URL.Query().Get("minRating"), 5)
	maxRating, maxRatingValue, validMaxRating := parseOptionalDecimal(r.URL.Query().Get("maxRating"), 5)
	limit, validLimit := parseBoundedInt(r.URL.Query().Get("limit"), 24, 1, 100)
	if !validLimit {
		writeError(w, http.StatusBadRequest, "invalid_limit")
		return
	}
	offset, validOffset := parseBoundedInt(r.URL.Query().Get("offset"), 0, 0, 100000)
	if !validOffset {
		writeError(w, http.StatusBadRequest, "invalid_offset")
		return
	}
	if len(q) > 200 {
		writeError(w, http.StatusBadRequest, "query_too_long")
		return
	}
	if len(category) > 120 {
		writeError(w, http.StatusBadRequest, "category_too_long")
		return
	}
	if len(brand) > 120 {
		writeError(w, http.StatusBadRequest, "brand_too_long")
		return
	}
	if !validMinPrice {
		writeError(w, http.StatusBadRequest, "invalid_min_price")
		return
	}
	if !validMaxPrice {
		writeError(w, http.StatusBadRequest, "invalid_max_price")
		return
	}
	if minPrice != "" && maxPrice != "" && minPriceValue > maxPriceValue {
		writeError(w, http.StatusBadRequest, "invalid_price_range")
		return
	}
	if !validMinRating {
		writeError(w, http.StatusBadRequest, "invalid_min_rating")
		return
	}
	if !validMaxRating {
		writeError(w, http.StatusBadRequest, "invalid_max_rating")
		return
	}
	if minRating != "" && maxRating != "" && minRatingValue > maxRatingValue {
		writeError(w, http.StatusBadRequest, "invalid_rating_range")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()
	params := dbgen.ListProductsParams{
		Category: category, Query: q, Brand: brand, MinPrice: minPrice, MaxPrice: maxPrice,
		MinRating: minRating, MaxRating: maxRating, PageLimit: int32(limit), PageOffset: int32(offset),
	}
	total, err := h.queries.CountProducts(ctx, dbgen.CountProductsParams{
		Category: category, Query: q, Brand: brand, MinPrice: minPrice, MaxPrice: maxPrice,
		MinRating: minRating, MaxRating: maxRating,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "catalog_query_failed")
		return
	}
	rows, err := h.queries.ListProducts(ctx, params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "catalog_query_failed")
		return
	}
	products := make([]catalog.Product, 0, len(rows))
	for _, row := range rows {
		product, err := mapProduct(row.ID, row.Title, row.Description, row.Category,
			row.Price, row.DiscountPercentage, row.Rating, row.Stock, row.Brand,
			row.Thumbnail, row.Images, row.Source, row.SourceID, row.SourceUrl,
			row.ImportedAt.Time)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "catalog_read_failed")
			return
		}
		products = append(products, product)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": products, "limit": limit, "offset": offset, "count": len(products), "total": total})
}

func (h *Handler) getProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_product_id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	row, err := h.queries.GetProduct(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "product_not_found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "catalog_read_failed")
		return
	}
	product, err := mapProduct(row.ID, row.Title, row.Description, row.Category,
		row.Price, row.DiscountPercentage, row.Rating, row.Stock, row.Brand,
		row.Thumbnail, row.Images, row.Source, row.SourceID, row.SourceUrl,
		row.ImportedAt.Time)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "catalog_read_failed")
		return
	}
	writeJSON(w, http.StatusOK, product)
}

func (h *Handler) listProductVariants(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_product_id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	if _, err := h.queries.GetProduct(ctx, id); errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "product_not_found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "catalog_read_failed")
		return
	}
	rows, err := h.queries.ListProductVariants(ctx, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "catalog_read_failed")
		return
	}
	items := make([]productVariantResponse, 0, len(rows))
	for _, row := range rows {
		item, err := mapProductVariant(row.ID, row.ProductID, row.Sku, row.Title,
			row.Price, row.Stock, row.Attributes)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "catalog_read_failed")
			return
		}
		items = append(items, item)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "count": len(items)})
}

type productVariantResponse struct {
	ID         int64          `json:"id"`
	ProductID  int64          `json:"productId"`
	SKU        *string        `json:"sku,omitempty"`
	Title      string         `json:"title"`
	Price      float64        `json:"price"`
	Stock      int32          `json:"stock"`
	Attributes map[string]any `json:"attributes"`
}

func mapProductVariant(id, productID int64, sku *string, title string, price float64,
	stock int32, rawAttributes []byte) (productVariantResponse, error) {
	attributes := make(map[string]any)
	if len(rawAttributes) > 0 {
		if err := json.Unmarshal(rawAttributes, &attributes); err != nil {
			return productVariantResponse{}, err
		}
	}
	return productVariantResponse{
		ID: id, ProductID: productID, SKU: sku, Title: title,
		Price: price, Stock: stock, Attributes: attributes,
	}, nil
}

func mapProduct(id int64, title, description, category string, price, discount, rating float64,
	stock int32, brand, thumbnail string, rawImages []byte, source, sourceID, sourceURL string,
	importedAt time.Time) (catalog.Product, error) {
	images := make([]string, 0)
	if len(rawImages) > 0 {
		if err := json.Unmarshal(rawImages, &images); err != nil {
			return catalog.Product{}, err
		}
	}
	return catalog.Product{
		ID: id, Title: title, Description: description, Category: category,
		Price: price, DiscountPercentage: discount, Rating: rating, Stock: int(stock),
		Brand: brand, Thumbnail: thumbnail, Images: images, Source: source,
		SourceID: sourceID, SourceURL: sourceURL, ImportedAt: importedAt,
	}, nil
}

func parseBoundedInt(raw string, fallback, min, max int) (int, bool) {
	if raw == "" {
		return fallback, true
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < min || n > max {
		return 0, false
	}
	return n, true
}

// parseOptionalDecimal accepts non-negative decimal values with at most two
// fractional digits. It keeps the original decimal text for PostgreSQL NUMERIC
// comparisons so filtering does not depend on binary floating-point rounding.
func parseOptionalDecimal(raw string, maximum float64) (string, float64, bool) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", 0, true
	}
	parts := strings.Split(value, ".")
	if len(parts) > 2 || len(parts[0]) == 0 || len(parts[0]) > 10 {
		return "", 0, false
	}
	for _, part := range parts {
		if part == "" {
			return "", 0, false
		}
		for _, char := range part {
			if char < '0' || char > '9' {
				return "", 0, false
			}
		}
	}
	if len(parts) == 2 && len(parts[1]) > 2 {
		return "", 0, false
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) || parsed < 0 || parsed > maximum {
		return "", 0, false
	}
	return value, parsed, true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]string{"error": code})
}
