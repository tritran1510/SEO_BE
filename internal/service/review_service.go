package service

import (
	"errors"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type KeywordSet struct {
	SEOTitle          string `json:"seoTitle"`
	Slug              string `json:"slug"`
	MetaDescription   string `json:"metaDescription"`
	PrimaryKeyword    string `json:"primaryKeyword"`
	SecondaryKeywords string `json:"secondaryKeywords"`
	Synonyms          string `json:"synonyms"`
}

type ImportedImage struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	MimeType string `json:"mimeType"`
	DataURL  string `json:"dataUrl"`
}

type ReviewRequest struct {
	ArticleTitle        string          `json:"articleTitle"`
	PermanentLink       string          `json:"permanentLink"`
	ArticleContent      string          `json:"articleContent"`
	ContentImages       []ImportedImage `json:"contentImages"`
	DetailedInformation string          `json:"detailedInformation"`
	Summary             string          `json:"summary"`
	KeywordSet          KeywordSet      `json:"keywordSet"`
}

type ChecklistResult struct {
	Code           string   `json:"code"`
	Group          string   `json:"group"`
	CheckName      string   `json:"checkName"`
	Result         string   `json:"result"`
	Status         string   `json:"status"`
	Reason         string   `json:"reason"`
	Improvement    string   `json:"improvement"`
	AffectedFields []string `json:"affectedFields"`
}

type FieldFeedback struct {
	Field    string   `json:"field"`
	Label    string   `json:"label"`
	Messages []string `json:"messages"`
}

type ReviewResult struct {
	OverallScore               int               `json:"overallScore"`
	SEOScore                   int               `json:"seoScore"`
	ReadabilityScore           int               `json:"readabilityScore"`
	AdvancedScore              int               `json:"advancedScore"`
	Status                     string            `json:"status"`
	ChecklistResults           []ChecklistResult `json:"checklistResults"`
	ImprovementRecommendations []string          `json:"improvementRecommendations"`
	FieldsNeedingImprovement   []string          `json:"fieldsNeedingImprovement"`
	FieldFeedback              []FieldFeedback   `json:"fieldFeedback"`
}

type contentStats struct {
	Sentences             []string
	Words                 []string
	AverageSentenceLength float64
}

var (
	headingPattern = regexp.MustCompile(`(?m)^#+\s.*$`)
	linkPattern    = regexp.MustCompile(`https?://[^\s)]+`)
	passivePattern = regexp.MustCompile(`\b(is|are|was|were|be|been|being)\s+\w+ed\b`)
	fieldLabels    = map[string]string{
		"articleTitle":        "Article Title",
		"permanentLink":       "Permanent Link",
		"articleContent":      "Article Content",
		"contentImages":       "Content Images",
		"detailedInformation": "Detailed Information",
		"summary":             "Summary",
		"seoTitle":            "SEO Title",
		"slug":                "Slug",
		"metaDescription":     "Meta Description",
		"primaryKeyword":      "Primary Keyword",
		"secondaryKeywords":   "Secondary Keywords",
		"synonyms":            "Synonyms",
	}
	fieldOrder = []string{
		"articleTitle", "permanentLink", "articleContent", "contentImages",
		"detailedInformation", "summary", "seoTitle", "slug",
		"metaDescription", "primaryKeyword", "secondaryKeywords", "synonyms",
	}
	transitionWords = []string{
		"however", "therefore", "moreover", "meanwhile", "in addition",
		"for example", "as a result", "in contrast", "first", "second", "finally",
	}
)

func ValidateReviewRequest(input ReviewRequest) error {
	// Check required fields
	switch {
	case strings.TrimSpace(input.ArticleTitle) == "":
		return errors.New("Article title is required.")
	case strings.TrimSpace(input.PermanentLink) == "":
		return errors.New("Permanent link is required.")
	case strings.TrimSpace(input.ArticleContent) == "":
		return errors.New("Article content is required.")
	case strings.TrimSpace(input.Summary) == "":
		return errors.New("Summary is required.")
	case strings.TrimSpace(input.KeywordSet.PrimaryKeyword) == "":
		return errors.New("Primary keyword is required.")
	default:
		break
	}

	// Check field size limits to prevent DoS
	const (
		maxArticleContentSize = 5 * 1024 * 1024  // 5MB
		maxDataURLSize        = 10 * 1024 * 1024 // 10MB per image
		maxTotalImagesSize    = 20 * 1024 * 1024 // 20MB total
	)

	// Validate article content size
	if len(input.ArticleContent) > maxArticleContentSize {
		return errors.New("Article content exceeds maximum allowed size of 5MB.")
	}

	// Validate content images
	totalImagesSize := 0
	for i, img := range input.ContentImages {
		dataURLSize := len(img.DataURL)
		if dataURLSize > maxDataURLSize {
			return errors.New("Image #" + strconv.Itoa(i+1) + " exceeds maximum allowed size of 10MB.")
		}
		totalImagesSize += dataURLSize
	}

	if totalImagesSize > maxTotalImagesSize {
		return errors.New("Total size of all images exceeds maximum allowed size of 20MB.")
	}

	return nil
}

func GenerateReview(input ReviewRequest) ReviewResult {
	primaryKeyword := strings.TrimSpace(input.KeywordSet.PrimaryKeyword)
	secondaryKeywords := splitKeywords(input.KeywordSet.SecondaryKeywords)
	synonyms := splitKeywords(input.KeywordSet.Synonyms)
	stats := buildContentStats(input.ArticleContent)
	intro := firstIntroSegment(input.ArticleContent)
	headings := headingPattern.FindAllString(input.ArticleContent, -1)
	internalLinks, outboundLinks := countLinks(input.ArticleContent, input.PermanentLink)

	primaryInTitle := containsPhrase(input.ArticleTitle, primaryKeyword)
	primaryInSEOTitle := containsPhrase(input.KeywordSet.SEOTitle, primaryKeyword)
	primaryInSlug := containsPhrase(input.KeywordSet.Slug, strings.ReplaceAll(primaryKeyword, " ", "-"))
	primaryInPermanentLink := containsPhrase(input.PermanentLink, input.KeywordSet.Slug)
	primaryInDescription := containsPhrase(input.KeywordSet.MetaDescription, primaryKeyword)
	primaryInSummary := containsPhrase(input.Summary, primaryKeyword)
	primaryInIntro := containsPhrase(intro, primaryKeyword)
	primaryInHeadings := containsPhrase(strings.Join(headings, " "), primaryKeyword)
	primaryMentions := countMatches(input.ArticleContent, primaryKeyword)
	keywordDensity := 0.0
	if len(stats.Words) > 0 {
		keywordDensity = float64(primaryMentions) / float64(len(stats.Words)) * 100
	}

	// These groups mirror the product requirement document so the frontend can render each section directly.
	seoChecks := []ChecklistResult{
		makeItem("seoTitleContainsPrimaryKeyword", "SEO", "SEO title contains primary keyword", primaryInSEOTitle, "The SEO title should reinforce the main search phrase.", "Add the primary keyword near the beginning of the SEO title.", []string{"seoTitle", "primaryKeyword"}),
		makeItem("seoTitleLengthStrong", "SEO", "SEO title length is within a strong range", len(strings.TrimSpace(input.KeywordSet.SEOTitle)) >= 45 && len(strings.TrimSpace(input.KeywordSet.SEOTitle)) <= 65, "The SEO title should be concise enough for search previews.", "Keep the SEO title close to 45 to 65 characters.", []string{"seoTitle"}),
		makeItem("titleContainsPrimaryKeyword", "SEO", "Title contains primary keyword", primaryInTitle, "The article title should clearly align with the primary keyword.", "Work the primary keyword naturally into the article title.", []string{"articleTitle", "primaryKeyword"}),
		makeItem("slugShortAndClear", "SEO", "Slug is short and clear", len(input.KeywordSet.Slug) >= 12 && len(input.KeywordSet.Slug) <= 45 && primaryInSlug, "The slug should be readable and include the main phrase.", "Keep the slug concise and include the primary keyword in hyphenated form.", []string{"slug", "primaryKeyword"}),
		makeItem("permanentLinkUsesSlug", "SEO", "Permanent link uses the chosen slug", primaryInPermanentLink, "The permanent link should stay aligned with the SEO slug.", "Update the permanent link so it uses the same clean slug as the SEO field.", []string{"permanentLink", "slug"}),
		makeItem("metaDescriptionLengthValid", "SEO", "Meta description length is valid", len(strings.TrimSpace(input.KeywordSet.MetaDescription)) >= 120 && len(strings.TrimSpace(input.KeywordSet.MetaDescription)) <= 160 && primaryInDescription, "The meta description should fit search snippets and include the target phrase.", "Aim for 120 to 160 characters and include the primary keyword once.", []string{"metaDescription", "primaryKeyword"}),
		makeItem("primaryKeywordInIntroduction", "SEO", "Primary keyword appears in introduction", primaryInIntro, "The introduction should establish the topic quickly.", "Mention the primary keyword early in the opening section.", []string{"articleContent", "primaryKeyword"}),
		makeItem("primaryKeywordInHeadings", "SEO", "Primary keyword appears in headings", primaryInHeadings, "Subheadings help both readers and search engines scan the topic structure.", "Add the primary keyword or a close variation into at least one subheading.", []string{"articleContent", "primaryKeyword"}),
		makeItem("primaryKeywordDensityBalanced", "SEO", "Primary keyword density is balanced", keywordDensity >= 0.5 && keywordDensity <= 2.5, "Keyword density should stay visible without becoming repetitive.", "Keep the primary keyword visible without repeating it unnaturally.", []string{"articleContent", "primaryKeyword"}),
		makeItem("contentLengthSufficient", "SEO", "Content length is sufficient", len(stats.Words) >= 180, "The article needs enough depth to support the topic well.", "Expand the article with more useful detail, examples, or structure.", []string{"articleContent"}),
		makeItem("summarySupportsTargetTopic", "SEO", "Summary supports the target topic", primaryInSummary, "The summary should reinforce the same search intent as the main article.", "Rephrase the summary so it supports the main keyword and intent.", []string{"summary", "primaryKeyword"}),
		makeItem("internalLinksPresent", "SEO", "Internal links are present", internalLinks > 0, "Internal links help connect the article to the rest of the site.", "Add at least one relevant internal link to related content on the same site.", []string{"articleContent", "permanentLink"}),
		makeItem("outboundLinksPresent", "SEO", "Outbound links are present", outboundLinks > 0, "Outbound links can strengthen credibility when they cite helpful external sources.", "Add at least one useful outbound link to a credible external source when appropriate.", []string{"articleContent"}),
	}

	advancedChecks := []ChecklistResult{
		makeItem("detailedInformationProvidesContext", "Advanced", "Detailed information provides useful context", len(strings.TrimSpace(input.DetailedInformation)) >= 40, "Detailed information helps the reviewer understand the article context.", "Add more supporting detail so the review can understand the article context more clearly.", []string{"detailedInformation"}),
		makeItem("secondaryKeywordsDistributed", "Advanced", "Secondary keywords are distributed naturally", len(secondaryKeywords) > 0 && countAnyPhraseMatches(input.ArticleContent, secondaryKeywords) >= 1, "Supporting keyphrases should appear naturally across the content.", "Use one or two secondary keywords in headings or supporting paragraphs where they fit naturally.", []string{"secondaryKeywords", "articleContent"}),
		makeItem("synonymSupportPresent", "Advanced", "Synonym support is present", len(synonyms) > 0 && countAnyPhraseMatches(input.ArticleContent, synonyms) >= 1, "Synonyms help broaden topical relevance and avoid repetition.", "Introduce one or two synonym phrases in the body content when they match the meaning.", []string{"synonyms", "articleContent"}),
		makeItem("topicConsistencyMaintained", "Advanced", "Topic consistency is maintained", containsPhrase(input.DetailedInformation, primaryKeyword) || containsPhrase(input.Summary, primaryKeyword) || primaryInTitle, "The supporting context should stay closely tied to the target topic.", "Use the detailed information and summary fields to strengthen topic framing and search intent alignment.", []string{"detailedInformation", "summary", "primaryKeyword"}),
	}

	passiveRatio := estimatePassiveVoiceRatio(stats.Sentences)
	repeatedStartsRatio := estimateRepeatedSentenceStarts(stats.Sentences)
	transitionCount := countAnyPhraseMatches(input.ArticleContent, transitionWords)

	readabilityChecks := []ChecklistResult{
		makeItem("sentenceLengthManageable", "Readability", "Sentence length is manageable", stats.AverageSentenceLength > 0 && stats.AverageSentenceLength <= 22, "Average sentence length should stay readable for editors and readers.", "Break longer sentences into shorter ideas to improve scanability.", []string{"articleContent"}),
		makeItem("paragraphFlowScannable", "Readability", "Paragraph flow is easy to scan", paragraphsAreScannable(input.ArticleContent), "Dense text blocks make the article harder to scan.", "Split long paragraphs into shorter sections with clearer pacing.", []string{"articleContent"}),
		makeItem("headingDistributionExists", "Readability", "Heading distribution exists", len(headings) >= 2, "Subheadings make longer content easier to navigate.", "Add subheadings to improve navigation and keyword distribution.", []string{"articleContent"}),
		makeItem("transitionWordsSupportFlow", "Readability", "Transition words support flow", transitionCount >= 2, "Transition words help readers follow the logic between ideas.", "Add a few more transition phrases so the article flows more naturally.", []string{"articleContent"}),
		makeItem("passiveVoiceLimited", "Readability", "Passive voice usage stays limited", passiveRatio <= 0.2, "Too much passive voice can make content feel indirect or harder to follow.", "Rewrite some passive constructions in a more direct active voice.", []string{"articleContent"}),
		makeItem("sentenceStartsVaried", "Readability", "Sentence starts are varied", repeatedStartsRatio <= 0.35, "Repeated sentence starts can make the writing feel mechanical.", "Vary sentence openings so the article sounds more natural to read.", []string{"articleContent"}),
	}

	allChecks := append(append(seoChecks, advancedChecks...), readabilityChecks...)
	seoScore := scoreChecks(seoChecks)
	readabilityScore := scoreChecks(readabilityChecks)
	advancedScore := scoreChecks(advancedChecks)
	overallScore := clampInt(int(float64(seoScore)*0.45+float64(readabilityScore)*0.35+float64(advancedScore)*0.2), 0, 100)

	status := "good"
	switch {
	case overallScore < 55:
		status = "poor"
	case overallScore < 80:
		status = "needs improvement"
	}

	improvements := make([]string, 0, 5)
	for _, item := range allChecks {
		if item.Status == "attention" {
			improvements = append(improvements, item.CheckName+": "+item.Improvement)
		}
		if len(improvements) == 5 {
			break
		}
	}

	fieldFeedback := buildFieldFeedback(allChecks)
	fieldsNeedingImprovement := make([]string, 0, len(fieldFeedback))
	for _, item := range fieldFeedback {
		fieldsNeedingImprovement = append(fieldsNeedingImprovement, item.Field)
	}

	return ReviewResult{
		OverallScore:               overallScore,
		SEOScore:                   seoScore,
		ReadabilityScore:           readabilityScore,
		AdvancedScore:              advancedScore,
		Status:                     status,
		ChecklistResults:           allChecks,
		ImprovementRecommendations: improvements,
		FieldsNeedingImprovement:   fieldsNeedingImprovement,
		FieldFeedback:              fieldFeedback,
	}
}

func makeItem(code string, group string, checkName string, passed bool, reason string, improvement string, affectedFields []string) ChecklistResult {
	result := "Needs work"
	status := "attention"
	if passed {
		result = "Pass"
		status = "good"
	}

	return ChecklistResult{Code: code, Group: group, CheckName: checkName, Result: result, Status: status, Reason: reason, Improvement: improvement, AffectedFields: affectedFields}
}

func buildFieldFeedback(checks []ChecklistResult) []FieldFeedback {
	messagesByField := map[string][]string{}

	for _, check := range checks {
		if check.Status == "good" {
			continue
		}
		for _, field := range check.AffectedFields {
			if !slices.Contains(messagesByField[field], check.Improvement) {
				messagesByField[field] = append(messagesByField[field], check.Improvement)
			}
		}
	}

	feedback := make([]FieldFeedback, 0, len(messagesByField))
	for _, field := range fieldOrder {
		messages, exists := messagesByField[field]
		if !exists {
			continue
		}
		feedback = append(feedback, FieldFeedback{Field: field, Label: fieldLabels[field], Messages: messages})
	}
	return feedback
}

func splitKeywords(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func tokenize(value string) []string {
	replacer := strings.NewReplacer(".", " ", ",", " ", "!", " ", "?", " ", "(", " ", ")", " ", "\"", " ", "'", " ", ":", " ", ";", " ")
	return strings.Fields(strings.ToLower(replacer.Replace(value)))
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func containsPhrase(text string, phrase string) bool {
	normalizedPhrase := normalize(phrase)
	return normalizedPhrase != "" && strings.Contains(normalize(text), normalizedPhrase)
}

func countMatches(text string, phrase string) int {
	normalizedPhrase := normalize(phrase)
	if normalizedPhrase == "" {
		return 0
	}
	return strings.Count(normalize(text), normalizedPhrase)
}

func countAnyPhraseMatches(text string, phrases []string) int {
	count := 0
	for _, phrase := range phrases {
		if containsPhrase(text, phrase) {
			count++
		}
	}
	return count
}

func buildContentStats(content string) contentStats {
	rawSentences := regexp.MustCompile(`[.!?]+`).Split(content, -1)
	sentences := make([]string, 0, len(rawSentences))
	for _, sentence := range rawSentences {
		trimmed := strings.TrimSpace(sentence)
		if trimmed != "" {
			sentences = append(sentences, trimmed)
		}
	}

	words := tokenize(content)
	averageSentenceLength := 0.0
	if len(sentences) > 0 {
		averageSentenceLength = float64(len(words)) / float64(len(sentences))
	}

	return contentStats{Sentences: sentences, Words: words, AverageSentenceLength: averageSentenceLength}
}

func firstIntroSegment(content string) string {
	for _, segment := range regexp.MustCompile(`[\n\.]+`).Split(content, -1) {
		trimmed := strings.TrimSpace(segment)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func paragraphsAreScannable(content string) bool {
	paragraphs := regexp.MustCompile(`\n\s*\n`).Split(content, -1)
	for _, paragraph := range paragraphs {
		if len(tokenize(paragraph)) > 120 {
			return false
		}
	}
	return true
}

func countLinks(content string, permanentLink string) (int, int) {
	internalLinks := 0
	outboundLinks := 0
	baseHost := extractHost(permanentLink)

	for _, link := range linkPattern.FindAllString(content, -1) {
		linkHost := extractHost(link)
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

func extractHost(rawURL string) string {
	parsedURL, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}
	return strings.ToLower(parsedURL.Hostname())
}

func estimatePassiveVoiceRatio(sentences []string) float64 {
	if len(sentences) == 0 {
		return 0
	}

	passiveCount := 0
	for _, sentence := range sentences {
		if passivePattern.MatchString(strings.ToLower(sentence)) {
			passiveCount++
		}
	}

	return float64(passiveCount) / float64(len(sentences))
}

func estimateRepeatedSentenceStarts(sentences []string) float64 {
	if len(sentences) == 0 {
		return 0
	}

	startCounts := map[string]int{}
	for _, sentence := range sentences {
		words := tokenize(sentence)
		if len(words) == 0 {
			continue
		}
		startCounts[words[0]]++
	}

	repeatedCount := 0
	for _, count := range startCounts {
		if count > 1 {
			repeatedCount += count
		}
	}

	return float64(repeatedCount) / float64(len(sentences))
}

func scoreChecks(checks []ChecklistResult) int {
	if len(checks) == 0 {
		return 0
	}

	passed := 0
	for _, check := range checks {
		if check.Status == "good" {
			passed++
		}
	}

	return clampInt(int(float64(passed)/float64(len(checks))*100), 0, 100)
}

func clampInt(value int, min int, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
