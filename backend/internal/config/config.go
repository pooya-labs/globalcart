package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	HTTPAddr     string
	DatabaseURL  string
	AllowedOrigin map[string]bool
}

func Load() (Config, error) {
	url := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if url == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	addr := strings.TrimSpace(os.Getenv("HTTP_ADDR"))
	if addr == "" { addr = ":8080" }
	origins := map[string]bool{}
	for _, item := range strings.Split(os.Getenv("CORS_ORIGINS"), ",") {
		if origin := strings.TrimSpace(item); origin != "" { origins[origin] = true }
	}
	return Config{HTTPAddr: addr, DatabaseURL: url, AllowedOrigin: origins}, nil
}
