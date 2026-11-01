package httpapi

import (
	"net/http"
	"net/http/httptest"
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
