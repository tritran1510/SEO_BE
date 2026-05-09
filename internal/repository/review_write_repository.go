package repository

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/seo/backend/internal/model"
	"github.com/seo/backend/internal/service"
	"gorm.io/datatypes"
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

		if err := saveChecklistResults(tx, review.ID, result.ChecklistResults); err != nil {
			return err
		}

		if err := saveFieldFeedback(tx, review.ID, result.FieldFeedback); err != nil {
			return err
		}

		if err := createReviewHistory(tx, review.ID, article.ID, req, result); err != nil {
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

func saveChecklistResults(tx *gorm.DB, reviewID string, checklist []service.ChecklistResult) error {
	for i, item := range checklist {
		checkCode := strings.TrimSpace(item.Code)
		checkName := strings.TrimSpace(item.CheckName)
		checkGroup := strings.TrimSpace(item.Group)
		defaultReason := stringPtrOrNil(item.Reason)
		defaultImprovement := stringPtrOrNil(item.Improvement)

		checklistItem := model.ReviewChecklistItem{
			CheckCode:           checkCode,
			CheckName:           checkName,
			CheckGroup:          checkGroup,
			DefaultReason:       defaultReason,
			DefaultImprovement:  defaultImprovement,
			SortOrder:           i,
			IsActive:            true,
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "check_code"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"check_name",
				"check_group",
				"default_reason",
				"default_improvement",
				"sort_order",
				"is_active",
			}),
		}).Create(&checklistItem).Error; err != nil {
			return err
		}
		if err := tx.Where("check_code = ?", checkCode).First(&checklistItem).Error; err != nil {
			return err
		}

		affectedFieldsJSON, err := toJSON(item.AffectedFields)
		if err != nil {
			return err
		}

		resultValue := normalizeChecklistResult(item.Result)
		statusValue := normalizeChecklistStatus(item.Status)
		row := model.ReviewChecklistResult{
			ReviewID:        reviewID,
			ChecklistItemID: checklistItem.ID,
			Result:          &resultValue,
			Status:          &statusValue,
			Reason:          stringPtrOrNil(item.Reason),
			Improvement:     stringPtrOrNil(item.Improvement),
			AffectedFields:  affectedFieldsJSON,
		}

		if err := tx.Create(&row).Error; err != nil {
			return err
		}
	}
	return nil
}

func saveFieldFeedback(tx *gorm.DB, reviewID string, feedback []service.FieldFeedback) error {
	for _, item := range feedback {
		messagesJSON, err := toJSON(item.Messages)
		if err != nil {
			return err
		}

		fieldFeedback := model.ReviewFieldFeedback{
			ReviewID:   reviewID,
			FieldName:  strings.TrimSpace(item.Field),
			FieldLabel: stringPtrOrNil(item.Label),
			Messages:   messagesJSON,
			Severity:   stringPtr("warning"),
		}
		if err := tx.Create(&fieldFeedback).Error; err != nil {
			return err
		}
	}
	return nil
}

func createReviewHistory(tx *gorm.DB, reviewID string, articleID int, req service.ReviewRequest, result service.ReviewResult) error {
	overall := result.OverallScore
	seo := result.SEOScore
	readability := result.ReadabilityScore
	advanced := result.AdvancedScore
	status := strings.TrimSpace(result.Status)

	wordCount := len(strings.Fields(req.ArticleContent))
	internalLinks, outboundLinks := countLinksForHistory(req.ArticleContent, req.PermanentLink)
	keywordDensity := computeKeywordDensity(req.ArticleContent, req.KeywordSet.PrimaryKeyword)

	checklistJSON, err := toJSON(result.ChecklistResults)
	if err != nil {
		return err
	}
	recommendationsJSON, err := toJSON(result.ImprovementRecommendations)
	if err != nil {
		return err
	}

	history := model.ReviewHistory{
		ReviewID:                 reviewID,
		ArticleID:                articleID,
		Action:                   "scored",
		SEOScoreSnapshot:         &seo,
		ReadabilityScoreSnapshot: &readability,
		AdvancedScoreSnapshot:    &advanced,
		OverallScoreSnapshot:     &overall,
		StatusSnapshot:           &status,
		PrimaryKeywordSnapshot:   stringPtrOrNil(req.KeywordSet.PrimaryKeyword),
		KeywordDensitySnapshot:   &keywordDensity,
		WordCountSnapshot:        &wordCount,
		InternalLinksSnapshot:    &internalLinks,
		OutboundLinksSnapshot:    &outboundLinks,
		ChecklistChanges:         checklistJSON,
		Recommendations:          recommendationsJSON,
	}

	return tx.Create(&history).Error
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

func normalizeChecklistResult(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "pass", "passed":
		return "passed"
	case "warning":
		return "warning"
	default:
		return "failed"
	}
}

func normalizeChecklistStatus(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "good", "success":
		return "success"
	default:
		return "needs_improvement"
	}
}

func countLinksForHistory(content string, permanentLink string) (int, int) {
	baseHost := extractHostForHistory(permanentLink)
	internalLinks := 0
	outboundLinks := 0

	linkPattern := regexp.MustCompile(`https?://[^\s)]+`)
	for _, link := range linkPattern.FindAllString(content, -1) {
		linkHost := extractHostForHistory(link)
		switch {
		case linkHost == "":
			continue
		case baseHost != "" && linkHost == baseHost:
			internalLinks++
		default:
			outboundLinks++
		}
	}
	return internalLinks, outboundLinks
}

func extractHostForHistory(rawURL string) string {
	value := strings.ToLower(strings.TrimSpace(rawURL))
	value = strings.TrimPrefix(value, "https://")
	value = strings.TrimPrefix(value, "http://")
	if slash := strings.Index(value, "/"); slash >= 0 {
		value = value[:slash]
	}
	if colon := strings.Index(value, ":"); colon >= 0 {
		value = value[:colon]
	}
	return strings.TrimSpace(value)
}

func computeKeywordDensity(content string, primaryKeyword string) float32 {
	words := len(strings.Fields(content))
	if words == 0 {
		return 0
	}
	keyword := strings.TrimSpace(strings.ToLower(primaryKeyword))
	if keyword == "" {
		return 0
	}
	mentions := strings.Count(strings.ToLower(content), keyword)
	return float32(mentions) / float32(words) * 100
}

func toJSON(v interface{}) (datatypes.JSON, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(raw), nil
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
