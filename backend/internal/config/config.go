package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Environment string

const (
	Development Environment = "development"
	Test        Environment = "test"
	Production  Environment = "production"
)

type Config struct {
	Environment   Environment
	HTTPAddr      string
	DatabaseURL   string
	AllowedOrigin map[string]bool
}

func Load() (Config, error) {
	env, err := parseEnvironment(os.Getenv("APP_ENV"))
	if err != nil {
		return Config{}, err
	}

	databaseURL, err := validateDatabaseURL(os.Getenv("DATABASE_URL"), env)
	if err != nil {
		return Config{}, err
	}

	httpAddr, err := validateHTTPAddr(os.Getenv("HTTP_ADDR"))
	if err != nil {
		return Config{}, err
	}

	origins, err := parseAllowedOrigins(os.Getenv("CORS_ORIGINS"), env)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Environment:   env,
		HTTPAddr:      httpAddr,
		DatabaseURL:   databaseURL,
		AllowedOrigin: origins,
	}, nil
}

func parseEnvironment(raw string) (Environment, error) {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return Development, nil
	}
	switch Environment(value) {
	case Development, Test, Production:
		return Environment(value), nil
	default:
		return "", fmt.Errorf("APP_ENV must be development, test, or production")
	}
}

func validateDatabaseURL(raw string, env Environment) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", fmt.Errorf("DATABASE_URL is required")
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("DATABASE_URL is invalid")
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "postgres" && scheme != "postgresql" {
		return "", fmt.Errorf("DATABASE_URL must use postgres or postgresql")
	}
	if parsed.Hostname() == "" {
		return "", fmt.Errorf("DATABASE_URL must include a database host")
	}
	if strings.Trim(parsed.Path, "/") == "" {
		return "", fmt.Errorf("DATABASE_URL must include a database name")
	}
	if parsed.Fragment != "" || parsed.RawFragment != "" {
		return "", fmt.Errorf("DATABASE_URL must not include a fragment")
	}

	query := parsed.Query()
	sslMode := strings.ToLower(strings.TrimSpace(query.Get("sslmode")))
	switch sslMode {
	case "", "disable", "allow", "prefer", "require", "verify-ca", "verify-full":
	default:
		return "", fmt.Errorf("DATABASE_URL contains an unsupported sslmode")
	}
	if env == Production {
		switch sslMode {
		case "require", "verify-ca", "verify-full":
		default:
			return "", fmt.Errorf("production DATABASE_URL must enable PostgreSQL TLS with sslmode=require, verify-ca, or verify-full")
		}
	}

	return value, nil
}

func validateHTTPAddr(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		value = ":8080"
	}
	if strings.ContainsAny(value, "/\\\r\n\t ") {
		return "", fmt.Errorf("HTTP_ADDR must be a host:port address")
	}
	host, port, err := net.SplitHostPort(value)
	if err != nil {
		return "", fmt.Errorf("HTTP_ADDR must be a valid host:port address")
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return "", fmt.Errorf("HTTP_ADDR port must be between 1 and 65535")
	}
	if strings.ContainsAny(host, "/\\") {
		return "", fmt.Errorf("HTTP_ADDR host is invalid")
	}
	return value, nil
}

func parseAllowedOrigins(raw string, env Environment) (map[string]bool, error) {
	origins := make(map[string]bool)
	for _, item := range strings.Split(raw, ",") {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		origin, err := normalizeOrigin(trimmed, env)
		if err != nil {
			return nil, err
		}
		origins[origin] = true
	}
	return origins, nil
}

func normalizeOrigin(raw string, env Environment) (string, error) {
	if raw == "*" || strings.EqualFold(raw, "null") || strings.Contains(raw, "*") {
		return "", fmt.Errorf("CORS_ORIGINS must contain exact origins; wildcard and null origins are not allowed")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("CORS_ORIGINS contains an invalid origin")
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", fmt.Errorf("CORS_ORIGINS entries must use http or https")
	}
	if parsed.Opaque != "" || parsed.User != nil || parsed.Hostname() == "" ||
		(parsed.Path != "" && parsed.Path != "/") || parsed.RawQuery != "" ||
		parsed.ForceQuery || parsed.Fragment != "" || parsed.RawFragment != "" ||
		strings.Contains(raw, "#") {
		return "", fmt.Errorf("CORS_ORIGINS entries must be bare origins without credentials, paths, queries, or fragments")
	}

	hostname := strings.ToLower(parsed.Hostname())
	if scheme == "http" {
		if env == Production {
			return "", fmt.Errorf("production CORS_ORIGINS entries must use HTTPS")
		}
		if !isLoopbackHost(hostname) {
			return "", fmt.Errorf("HTTP CORS origins are allowed only for localhost or loopback IP addresses")
		}
	}

	port := parsed.Port()
	if port != "" {
		portNumber, parseErr := strconv.Atoi(port)
		if parseErr != nil || portNumber < 1 || portNumber > 65535 {
			return "", fmt.Errorf("CORS_ORIGINS contains an invalid port")
		}
	}
	if (scheme == "https" && port == "443") || (scheme == "http" && port == "80") {
		port = ""
	}

	host := hostname
	if strings.Contains(hostname, ":") {
		host = "[" + hostname + "]"
	}
	if port != "" {
		host += ":" + port
	}
	return scheme + "://" + host, nil
}

func isLoopbackHost(hostname string) bool {
	host := strings.TrimSuffix(strings.ToLower(hostname), ".")
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
