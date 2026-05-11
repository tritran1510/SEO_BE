package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/seo/backend/internal/repository"
)

// GetReviewedArticles handles GET /api/reviews - lists all reviewed articles
func GetReviewedArticles(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")

	page := 1
	pageSize := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// Get reviewed articles from repository
	data, err := repository.GetAllReviewsGrouped(page, pageSize)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{
			Status:  "error",
			Message: "Failed to retrieve reviewed articles",
		})
		return
	}

	writeJSON(w, http.StatusOK, Response{
		Status:  "success",
		Message: "Reviewed articles retrieved successfully",
		Data:    data,
	})
}

// GetArticleReviewHistory handles GET /api/reviews/:article_id - shows review history for an article
func GetArticleReviewHistory(w http.ResponseWriter, r *http.Request) {
	// Extract article_id from URL path
	// Go 1.22+ supports path parameters in pattern matching
	articleIDStr := r.PathValue("article_id")
	if articleIDStr == "" {
		writeJSON(w, http.StatusBadRequest, Response{
			Status:  "error",
			Message: "Article ID is required",
		})
		return
	}

	articleID, err := strconv.Atoi(articleIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Response{
			Status:  "error",
			Message: "Invalid article ID",
		})
		return
	}

	// Parse pagination parameters
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")

	page := 1
	pageSize := 20

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// Get review history from repository
	data, err := repository.GetReviewHistoryByArticleID(articleID, page, pageSize)
	if err != nil {
		statusCode, message := mapReviewHistoryError(err)
		writeJSON(w, statusCode, Response{
			Status:  "error",
			Message: message,
		})
		return
	}

	writeJSON(w, http.StatusOK, Response{
		Status:  "success",
		Message: "Review history retrieved successfully",
		Data:    data,
	})
}

func mapReviewHistoryError(err error) (int, string) {
	switch {
	case errors.Is(err, repository.ErrArticleNotFound), errors.Is(err, repository.ErrNoReviewsForArticle):
		return http.StatusNotFound, "Article not found or has no reviews"
	default:
		return http.StatusInternalServerError, "Failed to retrieve review history"
	}
}
