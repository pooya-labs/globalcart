package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pooya16pg-spec/globalcart/backend/internal/dbgen"
	"github.com/pooya16pg-spec/globalcart/backend/internal/catalog"
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
	mux.HandleFunc("GET /api/v1/products/{id}", h.getProduct)
	return h.cors(mux)
}

func (h *Handler) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && h.origins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
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

func (h *Handler) listProducts(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	category := strings.TrimSpace(r.URL.Query().Get("category"))
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
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()
	rows, err := h.queries.ListProducts(ctx, dbgen.ListProductsParams{
		Category: category, Query: q, PageLimit: int32(limit), PageOffset: int32(offset),
	})
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
	writeJSON(w, http.StatusOK, map[string]any{"items": products, "limit": limit, "offset": offset, "count": len(products)})
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

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]string{"error": code})
}
