package model

import (
	"time"
)

// Article represents a blog post
type Article struct {
	ID                    int       `gorm:"primaryKey" json:"id"`
	Title                 string    `json:"title"`
	Slug                  string    `gorm:"uniqueIndex" json:"slug"`
	PermanentLink         string    `gorm:"unique" json:"permanent_link"`
	Content               string    `json:"content"`
	Summary               *string   `json:"summary"`
	DetailedInformation   *string   `json:"detailed_information"`
	Status                string    `gorm:"default:'draft'" json:"status"` // draft, published, archived
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
	PublishedAt           *time.Time `json:"published_at"`
	DeletedAt             *time.Time `json:"deleted_at"`

	// Relationships
	SEOMetadata  *ArticleSEOMetadata       `gorm:"foreignKey:ArticleID;constraint:OnDelete:CASCADE" json:"seo_metadata,omitempty"`
	Images       []ArticleImage            `gorm:"foreignKey:ArticleID;constraint:OnDelete:CASCADE" json:"images,omitempty"`
	Metrics      *ContentMetrics           `gorm:"foreignKey:ArticleID;constraint:OnDelete:CASCADE" json:"metrics,omitempty"`
	Reviews      []SEOReview               `gorm:"foreignKey:ArticleID;constraint:OnDelete:CASCADE" json:"reviews,omitempty"`
	ReviewSummary *ArticleReviewSummary    `gorm:"foreignKey:ArticleID;constraint:OnDelete:CASCADE" json:"review_summary,omitempty"`
}

// ArticleSEOMetadata stores SEO-specific data
type ArticleSEOMetadata struct {
	ID                 int       `gorm:"primaryKey" json:"id"`
	ArticleID          int       `gorm:"uniqueIndex;foreignKey:ArticleID" json:"article_id"`
	SEOTitle           *string   `json:"seo_title"`
	MetaDescription    *string   `json:"meta_description"`
	Slug               *string   `json:"slug"`
	PrimaryKeyword     *string   `json:"primary_keyword"`
	SecondaryKeywords  *string   `json:"secondary_keywords"` // comma-separated
	Synonyms           *string   `json:"synonyms"`            // comma-separated
	CanonicalURL       *string   `json:"canonical_url"`
	OGTitle            *string   `json:"og_title"`
	OGDescription      *string   `json:"og_description"`
	OGImageURL         *string   `json:"og_image_url"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`

	Article *Article `gorm:"foreignKey:ArticleID;constraint:OnDelete:CASCADE" json:"-"`
}

// ArticleImage stores images associated with articles
type ArticleImage struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	ArticleID int       `json:"article_id"`
	ImageName *string   `json:"image_name"`
	MimeType  *string   `json:"mime_type"`
	DataURL   *string   `json:"data_url"`
	AltText   *string   `json:"alt_text"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`

	Article *Article `gorm:"foreignKey:ArticleID;constraint:OnDelete:CASCADE" json:"-"`
}

// ContentMetrics stores pre-calculated content analytics
type ContentMetrics struct {
	ID                    int       `gorm:"primaryKey" json:"id"`
	ArticleID             int       `gorm:"uniqueIndex;foreignKey:ArticleID" json:"article_id"`
	WordCount             *int      `json:"word_count"`
	SentenceCount         *int      `json:"sentence_count"`
	HeadingCount          *int      `json:"heading_count"`
	InternalLinkCount     *int      `json:"internal_link_count"`
	OutboundLinkCount     *int      `json:"outbound_link_count"`
	ImageCount            *int      `json:"image_count"`
	AverageSentenceLength *float32  `json:"average_sentence_length"`
	PassiveVoiceRatio     *float32  `json:"passive_voice_ratio"`
	RepeatedStartsRatio   *float32  `json:"repeated_starts_ratio"`
	TransitionWordCount   *int      `json:"transition_word_count"`
	KeywordDensity        *float32  `json:"keyword_density"`
	ReadabilityLevel      *string   `json:"readability_level"` // elementary, intermediate, advanced
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`

	Article *Article `gorm:"foreignKey:ArticleID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName specifies table name for ContentMetrics
func (ContentMetrics) TableName() string {
	return "content_metrics"
}

// TableName specifies table name for ArticleImage
func (ArticleImage) TableName() string {
	return "article_images"
}

// TableName specifies table name for ArticleSEOMetadata
func (ArticleSEOMetadata) TableName() string {
	return "article_seo_metadata"
}
