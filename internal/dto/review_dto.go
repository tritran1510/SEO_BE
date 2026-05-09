package dto

import (
	"time"
)

// ReviewListItemDTO represents a single article in the reviewed articles list
type ReviewListItemDTO struct {
	ArticleID              int        `json:"article_id"`
	Title                  string     `json:"title"`
	Slug                   string     `json:"slug"`
	PermanentLink          string     `json:"permanent_link"`
	PrimaryKeyword         *string    `json:"primary_keyword"`
	LatestReviewID         *string    `json:"latest_review_id"`
	LatestOverallScore     *int       `json:"latest_overall_score"`
	LatestSEOScore         *int       `json:"latest_seo_score"`
	LatestReadabilityScore *int       `json:"latest_readability_score"`
	LatestAdvancedScore    *int       `json:"latest_advanced_score"`
	LatestStatus           *string    `json:"latest_status"`
	TotalReviews           int        `json:"total_reviews"`
	AvgOverallScore        *float32   `json:"avg_overall_score"`
	BestOverallScore       *int       `json:"best_overall_score"`
	WorstOverallScore      *int       `json:"worst_overall_score"`
	ScoreTrend             *string    `json:"score_trend"` // improving, declining, stable
	LastReviewedAt         *time.Time `json:"last_reviewed_at"`
}

// ReviewHistoryItemDTO represents a single review in the review history
type ReviewHistoryItemDTO struct {
	ReviewID                   string                   `json:"review_id"`
	CreatedAt                  time.Time                `json:"created_at"`
	OverallScore               *int                     `json:"overall_score"`
	SEOScore                   *int                     `json:"seo_score"`
	ReadabilityScore           *int                     `json:"readability_score"`
	AdvancedScore              *int                     `json:"advanced_score"`
	Status                     *string                  `json:"status"`
	Notes                      *string                  `json:"notes"`
	ArticleContent             *string                  `json:"article_content,omitempty"`
	Summary                    *string                  `json:"summary,omitempty"`
	DetailedInformation        *string                  `json:"detailed_information,omitempty"`
	SEOTitle                   *string                  `json:"seo_title,omitempty"`
	MetaDescription            *string                  `json:"meta_description,omitempty"`
	PrimaryKeyword             *string                  `json:"primary_keyword,omitempty"`
	SecondaryKeywords          *string                  `json:"secondary_keywords,omitempty"`
	Synonyms                   *string                  `json:"synonyms,omitempty"`
	ImprovementRecommendations []string                 `json:"improvement_recommendations,omitempty"`
	ChecklistResults           []map[string]interface{} `json:"checklist_results,omitempty"`
}

// ArticleDetailDTO represents the article details
type ArticleDetailDTO struct {
	ArticleID     int        `json:"article_id"`
	Title         string     `json:"title"`
	Slug          string     `json:"slug"`
	PermanentLink string     `json:"permanent_link"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	PublishedAt   *time.Time `json:"published_at"`
}

// ReviewListResponseDTO is the response for GET /api/reviews
type ReviewListResponseDTO struct {
	Items      []ReviewListItemDTO `json:"items"`
	Pagination PaginationDTO       `json:"pagination"`
}

// ReviewHistoryResponseDTO is the response for GET /api/reviews/:article_id
type ReviewHistoryResponseDTO struct {
	Article    ArticleDetailDTO       `json:"article"`
	Reviews    []ReviewHistoryItemDTO `json:"reviews"`
	Pagination PaginationDTO          `json:"pagination"`
	Summary    ReviewSummaryDTO       `json:"summary"`
}

// ReviewSummaryDTO provides aggregated data for a review set
type ReviewSummaryDTO struct {
	TotalReviews int      `json:"total_reviews"`
	BestScore    *int     `json:"best_score"`
	WorstScore   *int     `json:"worst_score"`
	AvgScore     *float32 `json:"avg_score"`
	Trend        *string  `json:"trend"`
}

// PaginationDTO contains pagination information
type PaginationDTO struct {
	TotalCount  int `json:"totalCount"`
	CurrentPage int `json:"currentPage"`
	PageSize    int `json:"pageSize"`
	TotalPages  int `json:"totalPages"`
}
