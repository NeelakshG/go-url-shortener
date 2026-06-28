package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/NeelakshG/snippy/internal/db"
	"github.com/NeelakshG/snippy/internal/handler"
)

func main() {
	// Load config from environment
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Connect to Postgres
	pool, err := db.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Wire handlers
	links := &handler.Links{Store: db.NewStore(pool)}

	// Register routes — Go 1.22 method+path routing
	mux := http.NewServeMux()
	mux.HandleFunc("POST /links", links.CreateLink)
	mux.HandleFunc("GET /{code}", links.Resolve)
	mux.HandleFunc("GET /stats/{code}", links.Stats)

	log.Printf("snippy listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
