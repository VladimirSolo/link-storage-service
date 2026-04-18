package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"link-storage-service/internal/config"
	"link-storage-service/internal/handler"
	"link-storage-service/internal/repository/cache"
	"link-storage-service/internal/repository/postgres"
	"link-storage-service/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	if err := migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	repo := postgres.NewLinkRepository(db)
	memCache := cache.NewMemoryCache(cfg.CacheTTL, cfg.CacheSweep)
	svc := service.NewLinkService(repo, memCache)
	h := handler.NewLinkHandler(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	log.Printf("listening on %s", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, mux); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS links (
			id           BIGSERIAL PRIMARY KEY,
			short_code   VARCHAR(16) UNIQUE NOT NULL,
			original_url TEXT        NOT NULL,
			created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			visits       BIGINT      NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS idx_links_short_code ON links(short_code);
	`)
	return err
}
