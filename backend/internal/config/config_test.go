package config

import (
	"strings"
	"testing"
)

func setBaseEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("APP_ENV", "")
	t.Setenv("DATABASE_URL", "postgres://globalcart:local-only@localhost:5432/globalcart?sslmode=disable")
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("CORS_ORIGINS", "http://localhost:3000")
}

func TestLoadDefaultsToDevelopmentAndLoopbackFriendlyAddress(t *testing.T) {
	setBaseEnvironment(t)

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Environment != Development {
		t.Fatalf("environment = %q, want %q", got.Environment, Development)
	}
	if got.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q, want :8080", got.HTTPAddr)
	}
	if !got.AllowedOrigin["http://localhost:3000"] {
		t.Fatalf("localhost origin is not allowed: %#v", got.AllowedOrigin)
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	setBaseEnvironment(t)
	t.Setenv("DATABASE_URL", "  ")

	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "DATABASE_URL is required") {
		t.Fatalf("Load() error = %v, want missing DATABASE_URL error", err)
	}
}

func TestLoadRejectsMalformedDatabaseURLs(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{name: "wrong scheme", url: "https://db.example/globalcart"},
		{name: "missing host", url: "postgres:///globalcart"},
		{name: "missing database", url: "postgres://localhost:5432/"},
		{name: "fragment", url: "postgres://localhost:5432/globalcart#admin"},
		{name: "unsupported ssl mode", url: "postgres://localhost:5432/globalcart?sslmode=unsafe"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setBaseEnvironment(t)
			t.Setenv("DATABASE_URL", tt.url)
			if _, err := Load(); err == nil {
				t.Fatal("Load() unexpectedly accepted malformed DATABASE_URL")
			}
		})
	}
}

func TestLoadRejectsInvalidHTTPAddresses(t *testing.T) {
	for _, value := range []string{"8080", ":0", ":65536", ":abc", "localhost:8080/path", "localhost: 8080"} {
		t.Run(value, func(t *testing.T) {
			setBaseEnvironment(t)
			t.Setenv("HTTP_ADDR", value)
			if _, err := Load(); err == nil {
				t.Fatalf("Load() unexpectedly accepted HTTP_ADDR %q", value)
			}
		})
	}
}

func TestLoadNormalizesExactCORSOrigins(t *testing.T) {
	setBaseEnvironment(t)
	t.Setenv("CORS_ORIGINS", " https://Shop.Example:443/, http://127.0.0.1:3000, https://shop.example ")

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(got.AllowedOrigin) != 2 {
		t.Fatalf("origin count = %d, want 2: %#v", len(got.AllowedOrigin), got.AllowedOrigin)
	}
	if !got.AllowedOrigin["https://shop.example"] {
		t.Fatalf("normalized HTTPS origin missing: %#v", got.AllowedOrigin)
	}
	if !got.AllowedOrigin["http://127.0.0.1:3000"] {
		t.Fatalf("loopback origin missing: %#v", got.AllowedOrigin)
	}
}

func TestLoadRejectsUnsafeCORSOrigins(t *testing.T) {
	tests := []struct {
		name   string
		origin string
	}{
		{name: "wildcard", origin: "*"},
		{name: "null", origin: "null"},
		{name: "wildcard subdomain", origin: "https://*.example.com"},
		{name: "remote HTTP", origin: "http://shop.example.com"},
		{name: "path", origin: "https://shop.example.com/path"},
		{name: "query", origin: "https://shop.example.com?mode=1"},
		{name: "credentials", origin: "https://user:pass@shop.example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setBaseEnvironment(t)
			t.Setenv("CORS_ORIGINS", tt.origin)
			if _, err := Load(); err == nil {
				t.Fatalf("Load() unexpectedly accepted CORS origin %q", tt.origin)
			}
		})
	}
}

func TestProductionRequiresPostgresTLSAndHTTPSOrigins(t *testing.T) {
	setBaseEnvironment(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("CORS_ORIGINS", "https://shop.example.com")
	t.Setenv("DATABASE_URL", "postgres://globalcart:secret@db.example:5432/globalcart?sslmode=verify-full")

	if _, err := Load(); err != nil {
		t.Fatalf("Load() rejected valid production settings: %v", err)
	}

	t.Setenv("DATABASE_URL", "postgres://globalcart:secret@db.example:5432/globalcart?sslmode=disable")
	if _, err := Load(); err == nil {
		t.Fatal("production accepted sslmode=disable")
	}

	t.Setenv("DATABASE_URL", "postgres://globalcart:secret@db.example:5432/globalcart?sslmode=verify-full")
	t.Setenv("CORS_ORIGINS", "http://localhost:3000")
	if _, err := Load(); err == nil {
		t.Fatal("production accepted an insecure HTTP origin")
	}
}

func TestLoadRejectsUnknownEnvironment(t *testing.T) {
	setBaseEnvironment(t)
	t.Setenv("APP_ENV", "production-ish")
	if _, err := Load(); err == nil {
		t.Fatal("Load() unexpectedly accepted unknown APP_ENV")
	}
}
