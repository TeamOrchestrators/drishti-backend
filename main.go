package main

import (
	"context"
	"log"
	"net/http"
	"os"

	db "github.com/TeamOrchestrators/drishti-backend/db/db/generated"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

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

	_ = &APIConfig{database: queries}

	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Route Handler

	log.Println("Starting server on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
