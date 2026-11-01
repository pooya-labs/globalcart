package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
