package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"

	"github.com/seo/backend/internal/dto"
	"github.com/seo/backend/internal/model"
	"gorm.io/gorm"
)

var (
	ErrArticleNotFound   = errors.New("article not found")
	ErrNoReviewsForArticle = errors.New("no reviews for article")
)

// GetAllReviewsGrouped retrieves all articles grouped by reviews with pagination
func GetAllReviewsGrouped(page, pageSize int) (*dto.ReviewListResponseDTO, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	// Count total reviewed articles (with at least one review)
	var totalCount int64
	if err := DB.Model(&model.ArticleReviewSummary{}).
		Where("total_reviews > 0").
		Count(&totalCount).Error; err != nil {
		return nil, err
	}

	// Get reviewed articles with pagination
	var summaries []model.ArticleReviewSummary
	if err := DB.Where("total_reviews > 0").
		Order("last_reviewed_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&summaries).Error; err != nil {
		return nil, err
	}

	// Get article IDs from summaries
	var articleIDs []int
	for _, s := range summaries {
		articleIDs = append(articleIDs, s.ArticleID)
	}

	// Get articles
	var articles []model.Article
	if len(articleIDs) > 0 {
		if err := DB.Where("id IN ?", articleIDs).Find(&articles).Error; err != nil {
			return nil, err
		}
	}

	// Get SEO metadata for articles
	var seoMetadataMap = make(map[int]*model.ArticleSEOMetadata)
	if len(articleIDs) > 0 {
		var metadata []model.ArticleSEOMetadata
		if err := DB.Where("article_id IN ?", articleIDs).Find(&metadata).Error; err != nil {
			return nil, err
		}
		for i := range metadata {
			seoMetadataMap[metadata[i].ArticleID] = &metadata[i]
		}
	}

	// Build DTOs
	items := make([]dto.ReviewListItemDTO, 0, len(articles))
	summaryMap := make(map[int]*model.ArticleReviewSummary)
	for i := range summaries {
		summaryMap[summaries[i].ArticleID] = &summaries[i]
	}

	for _, article := range articles {
		summary := summaryMap[article.ID]
		var primaryKeyword *string
		if seoMeta, ok := seoMetadataMap[article.ID]; ok {
			primaryKeyword = seoMeta.PrimaryKeyword
		}

		item := dto.ReviewListItemDTO{
			ArticleID:              article.ID,
			Title:                  article.Title,
			Slug:                   article.Slug,
			PermanentLink:          article.PermanentLink,
			PrimaryKeyword:         primaryKeyword,
			TotalReviews:           0,
			AvgOverallScore:        nil,
			BestOverallScore:       nil,
			WorstOverallScore:      nil,
			ScoreTrend:             nil,
			LastReviewedAt:         nil,
			LatestReviewID:         nil,
			LatestOverallScore:     nil,
			LatestSEOScore:         nil,
			LatestReadabilityScore: nil,
			LatestAdvancedScore:    nil,
			LatestStatus:           nil,
		}

		if summary != nil {
			item.TotalReviews = summary.TotalReviews
			item.LatestReviewID = summary.LatestReviewID
			item.LatestOverallScore = summary.LatestOverallScore
			item.LatestSEOScore = summary.LatestSEOScore
			item.LatestReadabilityScore = summary.LatestReadabilityScore
			item.LatestAdvancedScore = summary.LatestAdvancedScore
			item.LatestStatus = summary.LatestStatus
			item.AvgOverallScore = summary.AvgOverallScore
			item.BestOverallScore = summary.BestOverallScore
			item.WorstOverallScore = summary.WorstOverallScore
			item.ScoreTrend = summary.ScoreTrend
			item.LastReviewedAt = summary.LastReviewedAt
		}

		items = append(items, item)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	return &dto.ReviewListResponseDTO{
		Items: items,
		Pagination: dto.PaginationDTO{
			TotalCount:  int(totalCount),
			CurrentPage: page,
			PageSize:    pageSize,
			TotalPages:  totalPages,
		},
	}, nil
}

// GetReviewHistoryByArticleID retrieves all reviews for a specific article
func GetReviewHistoryByArticleID(articleID, page, pageSize int) (*dto.ReviewHistoryResponseDTO, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// Get article
	var article model.Article
	if err := DB.First(&article, articleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %d", ErrArticleNotFound, articleID)
		}
		return nil, err
	}

	// Count total reviews
	var totalCount int64
	if err := DB.Model(&model.SEOReview{}).Where("article_id = ?", articleID).Count(&totalCount).Error; err != nil {
		return nil, err
	}
	if totalCount == 0 {
		return nil, fmt.Errorf("%w: %d", ErrNoReviewsForArticle, articleID)
	}

	// Get reviews with pagination, ordered by created_at DESC
	var reviews []model.SEOReview
	if err := DB.Where("article_id = ?", articleID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&reviews).Error; err != nil {
		return nil, err
	}

	// Get article review summary
	var summary model.ArticleReviewSummary
	DB.Where("article_id = ?", articleID).First(&summary)

	// Backward compatibility for old history rows created before SEO snapshots existed.
	var seoMetadata model.ArticleSEOMetadata
	_ = DB.Where("article_id = ?", articleID).First(&seoMetadata).Error

	// Build DTOs
	historyItems := make([]dto.ReviewHistoryItemDTO, 0, len(reviews))
	for _, review := range reviews {
		var history model.ReviewHistory
		historyErr := DB.Where("review_id = ?", review.ID).
			Order("created_at DESC").
			First(&history).Error
		if historyErr != nil && !errors.Is(historyErr, gorm.ErrRecordNotFound) {
			return nil, historyErr
		}
		isLegacyHistory := history.ID == 0 || history.ArticleContentSnapshot == nil

		recommendations := make([]string, 0)
		parseSnapshotJSON(history.Recommendations, &recommendations, review.ID, articleID, "recommendations")
		imageMetadata := make([]dto.ReviewImageMetadataDTO, 0)
		parseSnapshotJSON(history.ImageMetadataSnapshot, &imageMetadata, review.ID, articleID, "image_metadata_snapshot")
		checklistResults := make([]map[string]interface{}, 0)
		parseSnapshotJSON(history.ChecklistChanges, &checklistResults, review.ID, articleID, "checklist_changes")

		articleContent := &article.Content
		if history.ArticleContentSnapshot != nil {
			articleContent = history.ArticleContentSnapshot
		}
		summaryText := article.Summary
		if history.SummarySnapshot != nil {
			summaryText = history.SummarySnapshot
		}
		detailedInfo := article.DetailedInformation
		if history.DetailedInfoSnapshot != nil {
			detailedInfo = history.DetailedInfoSnapshot
		}

		item := dto.ReviewHistoryItemDTO{
			ReviewID:                   review.ID,
			CreatedAt:                  review.CreatedAt,
			OverallScore:               review.OverallScore,
			SEOScore:                   review.SEOScore,
			ReadabilityScore:           review.ReadabilityScore,
			AdvancedScore:              review.AdvancedScore,
			Status:                     &review.Status,
			Notes:                      review.Notes,
			ArticleContent:             articleContent,
			Summary:                    summaryText,
			DetailedInformation:        detailedInfo,
			SEOTitle:                   coalesceHistorySnapshot(history.SEOTitleSnapshot, seoMetadata.SEOTitle, isLegacyHistory),
			MetaDescription:            coalesceHistorySnapshot(history.MetaDescriptionSnapshot, seoMetadata.MetaDescription, isLegacyHistory),
			PrimaryKeyword:             coalesceHistorySnapshot(history.PrimaryKeywordSnapshot, seoMetadata.PrimaryKeyword, isLegacyHistory),
			Slug:                       coalesceHistorySnapshot(history.SlugSnapshot, seoMetadata.Slug, isLegacyHistory),
			SecondaryKeywords:          coalesceHistorySnapshot(history.SecondaryKeywordsSnapshot, seoMetadata.SecondaryKeywords, isLegacyHistory),
			Synonyms:                   coalesceHistorySnapshot(history.SynonymsSnapshot, seoMetadata.Synonyms, isLegacyHistory),
			ImageMetadata:              imageMetadata,
			ImprovementRecommendations: recommendations,
			ChecklistResults:           checklistResults,
		}
		historyItems = append(historyItems, item)
	}

	// Build summary DTO
	summaryDTO := dto.ReviewSummaryDTO{
		TotalReviews: summary.TotalReviews,
		BestScore:    summary.BestOverallScore,
		WorstScore:   summary.WorstOverallScore,
		AvgScore:     summary.AvgOverallScore,
		Trend:        summary.ScoreTrend,
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	return &dto.ReviewHistoryResponseDTO{
		Article: dto.ArticleDetailDTO{
			ArticleID:     article.ID,
			Title:         article.Title,
			Slug:          article.Slug,
			PermanentLink: article.PermanentLink,
			Status:        article.Status,
			CreatedAt:     article.CreatedAt,
			UpdatedAt:     article.UpdatedAt,
			PublishedAt:   article.PublishedAt,
		},
		Reviews: historyItems,
		Pagination: dto.PaginationDTO{
			TotalCount:  int(totalCount),
			CurrentPage: page,
			PageSize:    pageSize,
			TotalPages:  totalPages,
		},
		Summary: summaryDTO,
	}, nil
}

func parseSnapshotJSON(raw []byte, target interface{}, reviewID string, articleID int, fieldName string) {
	if len(raw) == 0 {
		return
	}
	if err := json.Unmarshal(raw, target); err != nil {
		log.Printf(
			"warning: failed to parse review history snapshot field=%s article_id=%d review_id=%s err=%v",
			fieldName,
			articleID,
			reviewID,
			err,
		)
	}
}

func coalesceHistorySnapshot(primary *string, fallback *string, allowLegacyFallback bool) *string {
	if primary != nil {
		return primary
	}
	if !allowLegacyFallback {
		return nil
	}
	return fallback
}

// GetReviewByID retrieves a single review by ID
func GetReviewByID(reviewID string) (*model.SEOReview, error) {
	var review model.SEOReview
	if err := DB.Where("id = ?", reviewID).First(&review).Error; err != nil {
		return nil, err
	}
	return &review, nil
}

// GetArticleByID retrieves an article by ID
func GetArticleByID(articleID int) (*model.Article, error) {
	var article model.Article
	if err := DB.Preload("SEOMetadata").
		Preload("Images").
		Preload("Metrics").
		Preload("ReviewSummary").
		First(&article, articleID).Error; err != nil {
		return nil, err
	}
	return &article, nil
}
