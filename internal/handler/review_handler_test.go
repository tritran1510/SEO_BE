package handler

import (
	"errors"
	"net/http"
	"testing"

	"github.com/seo/backend/internal/repository"
)

func TestMapReviewHistoryError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		wantStatusCode int
		wantMessage    string
	}{
		{
			name:           "article not found maps to 404",
			err:            repository.ErrArticleNotFound,
			wantStatusCode: http.StatusNotFound,
			wantMessage:    "Article not found or has no reviews",
		},
		{
			name:           "no reviews maps to 404",
			err:            repository.ErrNoReviewsForArticle,
			wantStatusCode: http.StatusNotFound,
			wantMessage:    "Article not found or has no reviews",
		},
		{
			name:           "wrapped not found maps to 404",
			err:            errors.Join(errors.New("wrapped"), repository.ErrArticleNotFound),
			wantStatusCode: http.StatusNotFound,
			wantMessage:    "Article not found or has no reviews",
		},
		{
			name:           "unexpected error maps to 500",
			err:            errors.New("db unavailable"),
			wantStatusCode: http.StatusInternalServerError,
			wantMessage:    "Failed to retrieve review history",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			statusCode, message := mapReviewHistoryError(testCase.err)
			if statusCode != testCase.wantStatusCode {
				t.Fatalf("expected status %d, got %d", testCase.wantStatusCode, statusCode)
			}
			if message != testCase.wantMessage {
				t.Fatalf("expected message %q, got %q", testCase.wantMessage, message)
			}
		})
	}
}

