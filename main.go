package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/seo/backend/internal/config"
	"github.com/seo/backend/internal/middleware"
	"github.com/seo/backend/internal/repository"
	"github.com/seo/backend/internal/router"
)

func main() {
	// Load .env.local file if it exists
	_ = godotenv.Load(".env.local")

	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	if err := repository.InitDB(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer repository.CloseDB()

	// The backend is stateless for now: it receives SEO input and returns a review result.
	mux := router.New()

	// CORS lets the Vite frontend call this API from a different local port during development.
	handler := middleware.CORSMiddleware(middleware.LoggingMiddleware(mux))

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	fmt.Printf("Server running on http://%s\n", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
