package model

import (
	"time"

	"gorm.io/datatypes"
)

// SEOReview represents a single review of an article
type SEOReview struct {
	ID                  int              `gorm:"primaryKey" json:"id"`
	ArticleID           int              `gorm:"index;foreignKey:ArticleID" json:"article_id"`
	OverallScore        *int             `json:"overall_score"`
	SEOScore            *int             `json:"seo_score"`
	ReadabilityScore    *int             `json:"readability_score"`
	AdvancedScore       *int             `json:"advanced_score"`
	Status              string           `gorm:"default:'good'" json:"status"` // good, needs_improvement, poor
	Notes               *string          `json:"notes"`
	IsFinal             bool             `gorm:"default:false" json:"is_final"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`

	// Relationships
	Article               *Article                        `gorm:"foreignKey:ArticleID;constraint:OnDelete:CASCADE" json:"-"`
	ChecklistResults      []ReviewChecklistResult         `gorm:"foreignKey:ReviewID;constraint:OnDelete:CASCADE" json:"checklist_results,omitempty"`
	FieldFeedback         []ReviewFieldFeedback           `gorm:"foreignKey:ReviewID;constraint:OnDelete:CASCADE" json:"field_feedback,omitempty"`
	Recommendations       []ImprovementRecommendation     `gorm:"foreignKey:ReviewID;constraint:OnDelete:CASCADE" json:"recommendations,omitempty"`
}

// ReviewChecklistItem defines a single checklist item
type ReviewChecklistItem struct {
	ID               int       `gorm:"primaryKey" json:"id"`
	CheckCode        string    `gorm:"uniqueIndex" json:"check_code"`
	CheckName        string    `json:"check_name"`
	CheckGroup       string    `gorm:"index" json:"check_group"` // SEO, Readability, Advanced
	DefaultReason    *string   `json:"default_reason"`
	DefaultImprovement *string `json:"default_improvement"`
	SortOrder        int       `gorm:"default:0" json:"sort_order"`
	IsActive         bool      `gorm:"default:true" json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`

	// Relationships
	Results []ReviewChecklistResult `gorm:"foreignKey:ChecklistItemID;constraint:OnDelete:CASCADE" json:"results,omitempty"`
}

// ReviewChecklistResult stores the result of a single checklist item for a review
type ReviewChecklistResult struct {
	ID              int            `gorm:"primaryKey" json:"id"`
	ReviewID        int            `gorm:"index;foreignKey:ReviewID" json:"review_id"`
	ChecklistItemID int            `gorm:"foreignKey:ChecklistItemID" json:"checklist_item_id"`
	Result          *string        `json:"result"` // passed, failed, warning
	Status          *string        `json:"status"` // success, needs_improvement, failed
	Reason          *string        `json:"reason"`
	Improvement     *string        `json:"improvement"`
	AffectedFields  datatypes.JSON `gorm:"type:jsonb" json:"affected_fields"`
	CreatedAt       time.Time      `json:"created_at"`

	Review          *SEOReview         `gorm:"foreignKey:ReviewID;constraint:OnDelete:CASCADE" json:"-"`
	ChecklistItem   *ReviewChecklistItem `gorm:"foreignKey:ChecklistItemID;constraint:OnDelete:CASCADE" json:"checklist_item,omitempty"`
}

// ReviewFieldFeedback stores feedback for specific article fields
type ReviewFieldFeedback struct {
	ID        int            `gorm:"primaryKey" json:"id"`
	ReviewID  int            `gorm:"index;foreignKey:ReviewID;uniqueIndex:idx_review_field,composite:review_id,field_name" json:"review_id"`
	FieldName string         `gorm:"uniqueIndex:idx_review_field,composite:review_id,field_name" json:"field_name"`
	FieldLabel *string       `json:"field_label"`
	Messages  datatypes.JSON `gorm:"type:jsonb" json:"messages"` // JSON array of feedback strings
	Severity  *string        `json:"severity"`  // info, warning, error
	CreatedAt time.Time      `json:"created_at"`

	Review *SEOReview `gorm:"foreignKey:ReviewID;constraint:OnDelete:CASCADE" json:"-"`
}

// ImprovementRecommendation stores actionable recommendations from reviews
type ImprovementRecommendation struct {
	ID              int        `gorm:"primaryKey" json:"id"`
	ReviewID        int        `gorm:"index;foreignKey:ReviewID" json:"review_id"`
	Recommendation  string     `json:"recommendation"`
	Priority        string     `gorm:"default:'medium'" json:"priority"` // low, medium, high, critical
	EstimatedImpact *string    `json:"estimated_impact"`  // low, medium, high
	IsCompleted     bool       `gorm:"default:false" json:"is_completed"`
	CompletedAt     *time.Time `json:"completed_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	Review *SEOReview `gorm:"foreignKey:ReviewID;constraint:OnDelete:CASCADE" json:"-"`
}

// ReviewHistory stores complete revision history
type ReviewHistory struct {
	ID                       int            `gorm:"primaryKey" json:"id"`
	ReviewID                 int            `gorm:"index;foreignKey:ReviewID" json:"review_id"`
	ArticleID                int            `gorm:"index;foreignKey:ArticleID" json:"article_id"`
	Action                   string         `gorm:"index" json:"action"` // created, updated, scored, finalized, archived, improved
	Notes                    *string        `json:"notes"`
	SEOScoreSnapshot         *int           `json:"seo_score_snapshot"`
	ReadabilityScoreSnapshot *int           `json:"readability_score_snapshot"`
	AdvancedScoreSnapshot    *int           `json:"advanced_score_snapshot"`
	OverallScoreSnapshot     *int           `json:"overall_score_snapshot"`
	StatusSnapshot           *string        `json:"status_snapshot"`
	PrimaryKeywordSnapshot   *string        `json:"primary_keyword_snapshot"`
	KeywordDensitySnapshot   *float32       `json:"keyword_density_snapshot"`
	WordCountSnapshot        *int           `json:"word_count_snapshot"`
	InternalLinksSnapshot    *int           `json:"internal_links_snapshot"`
	OutboundLinksSnapshot    *int           `json:"outbound_links_snapshot"`
	ChecklistChanges         datatypes.JSON `gorm:"type:jsonb" json:"checklist_changes"`
	Recommendations          datatypes.JSON `gorm:"type:jsonb" json:"recommendations"`
	CreatedAt                time.Time      `json:"created_at"`

	Review   *SEOReview `gorm:"foreignKey:ReviewID;constraint:OnDelete:CASCADE" json:"-"`
	Article  *Article   `gorm:"foreignKey:ArticleID;constraint:OnDelete:CASCADE" json:"-"`
}

// ArticleReviewSummary groups review data by article
type ArticleReviewSummary struct {
	ID                     int       `gorm:"primaryKey" json:"id"`
	ArticleID              int       `gorm:"uniqueIndex;foreignKey:ArticleID" json:"article_id"`
	TotalReviews           int       `gorm:"default:0" json:"total_reviews"`
	LatestReviewID         *int      `gorm:"foreignKey:LatestReviewID" json:"latest_review_id"`
	LatestOverallScore     *int      `json:"latest_overall_score"`
	LatestSEOScore         *int      `json:"latest_seo_score"`
	LatestReadabilityScore *int      `json:"latest_readability_score"`
	LatestAdvancedScore    *int      `json:"latest_advanced_score"`
	LatestStatus           *string   `json:"latest_status"`
	BestOverallScore          *int      `json:"best_overall_score"`
	WorstOverallScore         *int      `json:"worst_overall_score"`
	AvgOverallScore           *float32  `json:"avg_overall_score"`
	ScoreTrend                *string   `json:"score_trend"` // improving, declining, stable
	LastReviewedAt            *time.Time `json:"last_reviewed_at"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`

	Article      *Article   `gorm:"foreignKey:ArticleID;constraint:OnDelete:CASCADE" json:"-"`
	LatestReview *SEOReview `gorm:"foreignKey:LatestReviewID;constraint:OnDelete:SET NULL" json:"latest_review,omitempty"`
}

// ReviewSetting stores global review configuration
type ReviewSetting struct {
	ID          int       `gorm:"primaryKey" json:"id"`
	SettingKey  string    `gorm:"uniqueIndex" json:"setting_key"`
	SettingValue string   `json:"setting_value"`
	Description *string   `json:"description"`
	DataType    *string   `json:"data_type"` // string, integer, decimal, boolean, json
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies table names
func (SEOReview) TableName() string {
	return "seo_reviews"
}

func (ReviewChecklistItem) TableName() string {
	return "review_checklist_items"
}

func (ReviewChecklistResult) TableName() string {
	return "review_checklist_results"
}

func (ReviewFieldFeedback) TableName() string {
	return "review_field_feedback"
}

func (ImprovementRecommendation) TableName() string {
	return "improvement_recommendations"
}

func (ReviewHistory) TableName() string {
	return "review_history"
}

func (ArticleReviewSummary) TableName() string {
	return "article_review_summary"
}

func (ReviewSetting) TableName() string {
	return "review_settings"
}
