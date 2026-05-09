package repository

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/seo/backend/internal/model"
	"github.com/seo/backend/internal/service"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var sentenceSplitPattern = regexp.MustCompile(`[.!?]+`)

// SaveReviewSubmission persists incoming review input and generated result.
func SaveReviewSubmission(req service.ReviewRequest, result service.ReviewResult) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		article, err := upsertArticle(tx, req)
		if err != nil {
			return err
		}

		if err := upsertSEOMetadata(tx, article.ID, req); err != nil {
			return err
		}

		if err := replaceArticleImages(tx, article.ID, req.ContentImages); err != nil {
			return err
		}

		if err := upsertContentMetrics(tx, article.ID, req); err != nil {
			return err
		}

		review, err := createReview(tx, article.ID, result)
		if err != nil {
			return err
		}

		if err := replaceRecommendations(tx, review.ID, result.ImprovementRecommendations); err != nil {
			return err
		}

		return upsertReviewSummary(tx, article.ID, review.ID, result)
	})
}

func upsertArticle(tx *gorm.DB, req service.ReviewRequest) (*model.Article, error) {
	var article model.Article
	err := tx.Where("permanent_link = ?", strings.TrimSpace(req.PermanentLink)).First(&article).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	slug := strings.TrimSpace(req.KeywordSet.Slug)
	if slug == "" {
		slug = strings.TrimSpace(req.PermanentLink)
	}

	article.Title = strings.TrimSpace(req.ArticleTitle)
	article.Slug = slug
	article.PermanentLink = strings.TrimSpace(req.PermanentLink)
	article.Content = req.ArticleContent
	article.Summary = stringPtrOrNil(req.Summary)
	article.DetailedInformation = stringPtrOrNil(req.DetailedInformation)
	if article.Status == "" {
		article.Status = "draft"
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := tx.Create(&article).Error; err != nil {
			return nil, err
		}
		return &article, nil
	}

	if err := tx.Save(&article).Error; err != nil {
		return nil, err
	}
	return &article, nil
}

func upsertSEOMetadata(tx *gorm.DB, articleID int, req service.ReviewRequest) error {
	meta := model.ArticleSEOMetadata{
		ArticleID:         articleID,
		SEOTitle:          stringPtrOrNil(req.KeywordSet.SEOTitle),
		MetaDescription:   stringPtrOrNil(req.KeywordSet.MetaDescription),
		Slug:              stringPtrOrNil(req.KeywordSet.Slug),
		PrimaryKeyword:    stringPtrOrNil(req.KeywordSet.PrimaryKeyword),
		SecondaryKeywords: stringPtrOrNil(req.KeywordSet.SecondaryKeywords),
		Synonyms:          stringPtrOrNil(req.KeywordSet.Synonyms),
	}

	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "article_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"seo_title",
			"meta_description",
			"slug",
			"primary_keyword",
			"secondary_keywords",
			"synonyms",
			"updated_at",
		}),
	}).Create(&meta).Error
}

func replaceArticleImages(tx *gorm.DB, articleID int, images []service.ImportedImage) error {
	if err := tx.Where("article_id = ?", articleID).Delete(&model.ArticleImage{}).Error; err != nil {
		return err
	}
	for i, img := range images {
		image := model.ArticleImage{
			ArticleID: articleID,
			ImageName: stringPtrOrNil(img.Name),
			MimeType:  stringPtrOrNil(img.MimeType),
			DataURL:   stringPtrOrNil(img.DataURL),
			SortOrder: i,
		}
		if err := tx.Create(&image).Error; err != nil {
			return err
		}
	}
	return nil
}

func upsertContentMetrics(tx *gorm.DB, articleID int, req service.ReviewRequest) error {
	wordCount := len(strings.Fields(req.ArticleContent))
	sentenceCount := 0
	for _, sentence := range sentenceSplitPattern.Split(req.ArticleContent, -1) {
		if strings.TrimSpace(sentence) != "" {
			sentenceCount++
		}
	}
	imageCount := len(req.ContentImages)
	keywordDensity := float32(0)
	keyword := strings.TrimSpace(strings.ToLower(req.KeywordSet.PrimaryKeyword))
	if wordCount > 0 && keyword != "" {
		lowered := strings.ToLower(req.ArticleContent)
		mentions := strings.Count(lowered, keyword)
		keywordDensity = float32(mentions) / float32(wordCount) * 100
	}

	metrics := model.ContentMetrics{
		ArticleID:      articleID,
		WordCount:      intPtr(wordCount),
		SentenceCount:  intPtr(sentenceCount),
		ImageCount:     intPtr(imageCount),
		KeywordDensity: float32Ptr(keywordDensity),
	}

	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "article_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"word_count",
			"sentence_count",
			"image_count",
			"keyword_density",
			"updated_at",
		}),
	}).Create(&metrics).Error
}

func createReview(tx *gorm.DB, articleID int, result service.ReviewResult) (*model.SEOReview, error) {
	overall := result.OverallScore
	seo := result.SEOScore
	readability := result.ReadabilityScore
	advanced := result.AdvancedScore

	review := model.SEOReview{
		ArticleID:        articleID,
		OverallScore:     &overall,
		SEOScore:         &seo,
		ReadabilityScore: &readability,
		AdvancedScore:    &advanced,
		Status:           strings.TrimSpace(result.Status),
	}

	if err := tx.Create(&review).Error; err != nil {
		return nil, err
	}
	return &review, nil
}

func replaceRecommendations(tx *gorm.DB, reviewID string, recommendations []string) error {
	for _, recommendation := range recommendations {
		rec := model.ImprovementRecommendation{
			ReviewID:       reviewID,
			Recommendation: recommendation,
			Priority:       "medium",
		}
		if err := tx.Create(&rec).Error; err != nil {
			return err
		}
	}
	return nil
}

func upsertReviewSummary(tx *gorm.DB, articleID int, reviewID string, result service.ReviewResult) error {
	var summary model.ArticleReviewSummary
	err := tx.Where("article_id = ?", articleID).First(&summary).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	now := time.Now()
	overall := result.OverallScore
	status := strings.TrimSpace(result.Status)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		avg := float32(overall)
		summary = model.ArticleReviewSummary{
			ArticleID:              articleID,
			TotalReviews:           1,
			LatestReviewID:         &reviewID,
			LatestOverallScore:     &overall,
			LatestSEOScore:         intPtr(result.SEOScore),
			LatestReadabilityScore: intPtr(result.ReadabilityScore),
			LatestAdvancedScore:    intPtr(result.AdvancedScore),
			LatestStatus:           &status,
			BestOverallScore:       &overall,
			WorstOverallScore:      &overall,
			AvgOverallScore:        &avg,
			ScoreTrend:             stringPtr("stable"),
			LastReviewedAt:         &now,
		}
		return tx.Create(&summary).Error
	}

	previousLatest := intValue(summary.LatestOverallScore)
	total := summary.TotalReviews + 1
	avg := float32((float64(float32Value(summary.AvgOverallScore))*float64(summary.TotalReviews) + float64(overall)) / float64(total))
	best := overall
	if summary.BestOverallScore != nil && *summary.BestOverallScore > best {
		best = *summary.BestOverallScore
	}
	worst := overall
	if summary.WorstOverallScore != nil && *summary.WorstOverallScore < worst {
		worst = *summary.WorstOverallScore
	}
	trend := "stable"
	if overall > previousLatest {
		trend = "improving"
	} else if overall < previousLatest {
		trend = "declining"
	}

	summary.TotalReviews = total
	summary.LatestReviewID = &reviewID
	summary.LatestOverallScore = &overall
	summary.LatestSEOScore = intPtr(result.SEOScore)
	summary.LatestReadabilityScore = intPtr(result.ReadabilityScore)
	summary.LatestAdvancedScore = intPtr(result.AdvancedScore)
	summary.LatestStatus = &status
	summary.BestOverallScore = &best
	summary.WorstOverallScore = &worst
	summary.AvgOverallScore = &avg
	summary.ScoreTrend = &trend
	summary.LastReviewedAt = &now

	return tx.Save(&summary).Error
}

func stringPtrOrNil(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func intPtr(v int) *int {
	return &v
}

func float32Ptr(v float32) *float32 {
	return &v
}

func stringPtr(v string) *string {
	return &v
}

func intValue(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

func float32Value(v *float32) float32 {
	if v == nil {
		return 0
	}
	return *v
}
