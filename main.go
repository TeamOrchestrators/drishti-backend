package main

import (
	"context"
	"log"
	"net/http"
	"os"

	db "github.com/TeamOrchestrators/drishti-backend/db/db/generated"
	"github.com/TeamOrchestrators/drishti-backend/internal/emergency"
	"github.com/TeamOrchestrators/drishti-backend/internal/expedition"
	"github.com/TeamOrchestrators/drishti-backend/internal/inventory"
	"github.com/TeamOrchestrators/drishti-backend/internal/logistics"
	"github.com/TeamOrchestrators/drishti-backend/internal/personnel"
	"github.com/TeamOrchestrators/drishti-backend/internal/station"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// APIConfig holds application-wide dependencies shared by handlers and services.
type APIConfig struct {
	database *db.Queries
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal("Unable to create database pool:", err)
	}
	defer pool.Close()

	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}

	queries := db.New(pool)
	cfg := &APIConfig{database: queries}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	station.RegisterRoutes(mux, cfg.database)
	personnel.RegisterRoutes(mux, cfg.database)
	expedition.RegisterRoutes(mux, cfg.database)
	inventory.RegisterRoutes(mux, cfg.database)
	logistics.RegisterRoutes(mux, cfg.database)
	emergency.RegisterRoutes(mux, cfg.database)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Starting server on port %s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
