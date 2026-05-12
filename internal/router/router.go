package router

import (
	"net/http"

	"github.com/seo/backend/internal/handler"
)

func New() *http.ServeMux {
	mux := http.NewServeMux()

	// Health check routes
	mux.HandleFunc("GET /", handler.HealthCheck)
	mux.HandleFunc("GET /api/health", handler.HealthCheck)

	// Review routes
	// Keep both /api/* and bare /* aliases to tolerate different reverse-proxy setups in production.
	mux.HandleFunc("POST /api/review", handler.ReviewContent)
	mux.HandleFunc("POST /review", handler.ReviewContent)
	mux.HandleFunc("GET /api/reviews", handler.GetReviewedArticles)
	mux.HandleFunc("GET /reviews", handler.GetReviewedArticles)
	mux.HandleFunc("GET /api/reviews/{article_id}", handler.GetArticleReviewHistory)
	mux.HandleFunc("GET /reviews/{article_id}", handler.GetArticleReviewHistory)

	return mux
}
