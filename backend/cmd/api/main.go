package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pooya16pg-spec/globalcart/backend/internal/config"
	"github.com/pooya16pg-spec/globalcart/backend/internal/database"
	"github.com/pooya16pg-spec/globalcart/backend/internal/httpapi"
)

func main() {
	cfg, err := config.Load()
	if err != nil { log.Fatal(err) }
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil { log.Fatalf("database config: %v", err) }
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil { log.Fatalf("database unavailable: %v", err) }
	if err := database.Migrate(ctx, pool); err != nil { log.Fatalf("database migration: %v", err) }

	server := &http.Server{
		Addr: cfg.HTTPAddr,
		Handler: httpapi.New(pool, cfg.AllowedOrigin),
		ReadHeaderTimeout: 5*time.Second,
		ReadTimeout: 10*time.Second,
		WriteTimeout: 15*time.Second,
		IdleTimeout: 60*time.Second,
	}
	go func() {
		log.Printf("GlobalCart API listening on %s", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) { log.Fatalf("http server: %v", err) }
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	shutdown, done := context.WithTimeout(context.Background(), 10*time.Second)
	defer done()
	if err := server.Shutdown(shutdown); err != nil { log.Printf("graceful shutdown: %v", err) }
}
