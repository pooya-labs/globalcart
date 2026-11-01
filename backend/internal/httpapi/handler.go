package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pooya16pg-spec/globalcart/backend/internal/catalog"
)

type Handler struct {
	db      *pgxpool.Pool
	origins map[string]bool
}

func New(db *pgxpool.Pool, origins map[string]bool) http.Handler {
	h := &Handler{db: db, origins: origins}
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
	rows, err := h.db.Query(ctx, `SELECT id,title,description,category,price::float8,discount_percentage::float8,
		rating::float8,stock,brand,thumbnail,images,source,source_id,source_url,imported_at
		FROM products WHERE ($1='' OR category=$1) AND ($2='' OR to_tsvector('simple', title || ' ' || description) @@ plainto_tsquery('simple',$2))
		ORDER BY id LIMIT $3 OFFSET $4`, category, q, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "catalog_query_failed")
		return
	}
	defer rows.Close()
	products := make([]catalog.Product, 0)
	for rows.Next() {
		var p catalog.Product
		if err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.Category, &p.Price, &p.DiscountPercentage, &p.Rating, &p.Stock, &p.Brand, &p.Thumbnail, &p.Images, &p.Source, &p.SourceID, &p.SourceURL, &p.ImportedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "catalog_read_failed")
			return
		}
		products = append(products, p)
	}
	if rows.Err() != nil {
		writeError(w, http.StatusInternalServerError, "catalog_read_failed")
		return
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
	var p catalog.Product
	err = h.db.QueryRow(ctx, `SELECT id,title,description,category,price::float8,discount_percentage::float8,
		rating::float8,stock,brand,thumbnail,images,source,source_id,source_url,imported_at
		FROM products WHERE id=$1`, id).Scan(&p.ID, &p.Title, &p.Description, &p.Category, &p.Price, &p.DiscountPercentage, &p.Rating, &p.Stock, &p.Brand, &p.Thumbnail, &p.Images, &p.Source, &p.SourceID, &p.SourceURL, &p.ImportedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "product_not_found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "catalog_read_failed")
		return
	}
	writeJSON(w, http.StatusOK, p)
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
