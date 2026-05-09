package handler

import (
	"encoding/json"
	"net/http"

	"github.com/seo/backend/internal/service"
)

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, Response{
		Status:  "success",
		Message: "Service is healthy",
	})
}

func ReviewContent(w http.ResponseWriter, r *http.Request) {
	// Limit request body size to 50MB to prevent DoS attacks
	const maxBodySize = 50 * 1024 * 1024 // 50MB
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

	var request service.ReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		// Check if it's a max bytes error
		if err.Error() == "http: request body too large" {
			writeJSON(w, http.StatusRequestEntityTooLarge, Response{
				Status:  "error",
				Message: "Request body exceeds maximum allowed size of 50MB.",
			})
			return
		}
		writeJSON(w, http.StatusBadRequest, Response{
			Status:  "error",
			Message: "Invalid review payload.",
		})
		return
	}

	if err := service.ValidateReviewRequest(request); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	// The review engine is pure business logic with no persistence dependency.
	result := service.GenerateReview(request)
	writeJSON(w, http.StatusOK, Response{
		Status:  "success",
		Message: "SEO review generated successfully.",
		Data:    result,
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
