package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHealthIsPublicAndDoesNotNeedDatabase(t *testing.T) {
	h := New(nil, nil)
	res := httptest.NewRecorder()
	h.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.Code)
	}
	if got := res.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("content type = %q", got)
	}
}

func TestCORSOnlyAllowsConfiguredOrigins(t *testing.T) {
	h := New(nil, map[string]bool{"http://localhost:3000": true})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", "https://evil.example")
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if got := res.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("unexpected allow-origin %q", got)
	}
}

func TestListProductsRejectsInvalidParametersBeforeDatabaseAccess(t *testing.T) {
	h := New(nil, nil)
	for _, target := range []string{
		"/api/v1/products?limit=0",
		"/api/v1/products?limit=abc",
		"/api/v1/products?limit=101",
		"/api/v1/products?offset=-1",
		"/api/v1/products?offset=100001",
		"/api/v1/products?q=" + strings.Repeat("x", 201),
		"/api/v1/products?category=" + strings.Repeat("x", 121),
		"/api/v1/products?brand=" + strings.Repeat("x", 121),
		"/api/v1/products?minPrice=-1",
		"/api/v1/products?maxPrice=10000000000",
		"/api/v1/products?minPrice=1.234",
		"/api/v1/products?minPrice=20&maxPrice=10",
		"/api/v1/products?minRating=5.01",
		"/api/v1/products?minRating=4&maxRating=3",
	} {
		res := httptest.NewRecorder()
		h.ServeHTTP(res, httptest.NewRequest(http.MethodGet, target, nil))
		if res.Code != http.StatusBadRequest {
			t.Errorf("GET %s status = %d, want 400", target, res.Code)
		}
	}
}

func TestParseBoundedInt(t *testing.T) {
	tests := []struct {
		name, raw                string
		fallback, min, max, want int
		valid                    bool
	}{
		{name: "default", fallback: 24, min: 1, max: 100, want: 24, valid: true},
		{name: "valid", raw: "50", min: 1, max: 100, want: 50, valid: true},
		{name: "below minimum", raw: "0", min: 1, max: 100},
		{name: "above maximum", raw: "101", min: 1, max: 100},
		{name: "not a number", raw: "many", min: 1, max: 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, valid := parseBoundedInt(tt.raw, tt.fallback, tt.min, tt.max)
			if got != tt.want || valid != tt.valid {
				t.Fatalf("parseBoundedInt() = (%d, %t), want (%d, %t)", got, valid, tt.want, tt.valid)
			}
		})
	}
}

func TestParseOptionalDecimal(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		maximum float64
		want    string
		valid   bool
	}{
		{name: "omitted", maximum: 100, want: "", valid: true},
		{name: "trimmed exact decimal", raw: " 49.90 ", maximum: 100, want: "49.90", valid: true},
		{name: "integer", raw: "0", maximum: 100, want: "0", valid: true},
		{name: "upper bound", raw: "5.00", maximum: 5, want: "5.00", valid: true},
		{name: "negative", raw: "-1", maximum: 100},
		{name: "too precise", raw: "1.001", maximum: 100},
		{name: "scientific notation", raw: "1e2", maximum: 1000},
		{name: "not a number", raw: "NaN", maximum: 100},
		{name: "infinity", raw: "Inf", maximum: 100},
		{name: "above maximum", raw: "5.01", maximum: 5},
		{name: "empty fraction", raw: "10.", maximum: 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, valid := parseOptionalDecimal(tt.raw, tt.maximum)
			if got != tt.want || valid != tt.valid {
				t.Fatalf("parseOptionalDecimal(%q) = (%q, %t), want (%q, %t)", tt.raw, got, valid, tt.want, tt.valid)
			}
		})
	}
}

func TestMapProductPreservesCatalogJSONShape(t *testing.T) {
	got, err := mapProduct(7, "Phone", "A phone", "Electronics", 99.5, 10, 4.8,
		12, "Brand", "thumb.jpg", []byte(`["a.jpg","b.jpg"]`),
		"feed", "source-7", "https://example.test/p/7", time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
	if err != nil {
		t.Fatalf("mapProduct: %v", err)
	}
	if got.ID != 7 || got.Price != 99.5 || got.Stock != 12 || len(got.Images) != 2 || got.SourceURL != "https://example.test/p/7" {
		t.Fatalf("unexpected mapped product: %+v", got)
	}
	if _, err := mapProduct(7, "Phone", "A phone", "Electronics", 99.5, 10, 4.8,
		12, "Brand", "thumb.jpg", []byte("not-json"), "feed", "source-7", "", time.Time{}); err == nil {
		t.Fatal("expected invalid images JSON to fail")
	}
}

func TestMapProductVariant(t *testing.T) {
	sku := "GC-BLK-M"
	got, err := mapProductVariant(9, 7, &sku, "Black / Medium", 13.99, 2,
		[]byte(`{"color":"black","size":"M"}`))
	if err != nil {
		t.Fatalf("mapProductVariant: %v", err)
	}
	if got.ID != 9 || got.ProductID != 7 || got.SKU == nil || *got.SKU != sku ||
		got.Price != 13.99 || got.Stock != 2 || got.Attributes["color"] != "black" {
		t.Fatalf("unexpected mapped variant: %+v", got)
	}
	if _, err := mapProductVariant(9, 7, nil, "Default", 1, 1, []byte("[]")); err == nil {
		t.Fatal("expected non-object variant attributes to fail")
	}
}

func TestCategoryResponseOmitsRootParent(t *testing.T) {
	item := categoryResponse{ID: 3, Name: "Electronics", ProductCount: 12}
	raw, err := json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "parentId") {
		t.Fatalf("root category should omit parentId: %s", raw)
	}
	parent := int64(3)
	item.ParentID = &parent
	raw, err = json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"parentId":3`) {
		t.Fatalf("child category parentId missing: %s", raw)
	}
}

func TestSellerAndCommerceRoutesAreNotExposed(t *testing.T) {
	h := New(nil, nil)
	for _, tc := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/seller/orders"},
		{http.MethodPost, "/api/v1/seller/orders"},
		{http.MethodGet, "/api/v1/vendors/1/products"},
		{http.MethodPost, "/api/v1/vendors/1/products"},
		{http.MethodGet, "/api/v1/vendor/orders"},
		{http.MethodPost, "/api/v1/cart"},
		{http.MethodPost, "/api/v1/checkout"},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			res := httptest.NewRecorder()
			h.ServeHTTP(res, httptest.NewRequest(tc.method, tc.path, nil))
			if res.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want 404 while route is intentionally unavailable", res.Code)
			}
		})
	}
}

func TestReadOnlyCatalogRoutesRejectMutationMethods(t *testing.T) {
	h := New(nil, nil)
	for _, path := range []string{"/api/v1/products", "/api/v1/categories", "/api/v1/products/1"} {
		res := httptest.NewRecorder()
		h.ServeHTTP(res, httptest.NewRequest(http.MethodPost, path, nil))
		if res.Code != http.StatusMethodNotAllowed {
			t.Errorf("POST %s status = %d, want 405", path, res.Code)
		}
	}
}

func TestAllowedCORSOriginReceivesReadOnlyMethods(t *testing.T) {
	h := New(nil, map[string]bool{"https://shop.example.test": true})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", "https://shop.example.test")
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if got := res.Header().Get("Access-Control-Allow-Origin"); got != "https://shop.example.test" {
		t.Fatalf("allow-origin = %q", got)
	}
	if got := res.Header().Get("Access-Control-Allow-Methods"); got != "GET, OPTIONS" {
		t.Fatalf("allow-methods = %q, want GET, OPTIONS", got)
	}
	if got := res.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type" {
		t.Fatalf("allow-headers = %q, want Content-Type only before auth exists", got)
	}
}
