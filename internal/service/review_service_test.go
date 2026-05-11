package service

import "testing"

func TestContainsKeyphraseUsesWordBoundaries(t *testing.T) {
	if containsKeyphrase("artisan approach", "art") {
		t.Fatalf("expected partial token not to match")
	}
	if !containsKeyphrase("seo premium content review", "premium content") {
		t.Fatalf("expected phrase token sequence to match")
	}
}

func TestLongSentenceRatio(t *testing.T) {
	sentences := []string{
		"This sentence is intentionally very long and includes more than twenty words to trigger readability threshold logic in the review engine.",
		"Short sentence here.",
	}

	ratio := longSentenceRatio(sentences)
	if ratio <= 0.4 || ratio >= 0.6 {
		t.Fatalf("expected ratio near 0.5, got %f", ratio)
	}
}

func TestDetectContentLanguage(t *testing.T) {
	if detectContentLanguage("Đây là nội dung tiếng Việt") != "vi" {
		t.Fatalf("expected vietnamese detection")
	}
	if detectContentLanguage("This is plain english content") != "en" {
		t.Fatalf("expected english detection")
	}
}

func TestComputeImageAltCoverage(t *testing.T) {
	coverage := computeImageAltCoverage([]ImportedImage{
		{AltText: "product image alt"},
		{AltText: ""},
		{AltText: "another alt"},
	})

	if coverage <= 0.65 || coverage >= 0.68 {
		t.Fatalf("expected coverage about 0.666, got %f", coverage)
	}
}

func TestGenerateReviewAddsImageAltChecklist(t *testing.T) {
	input := ReviewRequest{
		ArticleTitle:   "Premium SEO Product Review Guide",
		PermanentLink:  "https://example.com/premium-seo-review-guide",
		ArticleContent: "Premium SEO product review guide. Premium SEO product review guide helps teams. Premium SEO product review guide is practical.",
		ContentImages: []ImportedImage{
			{AltText: "first image alt"},
			{AltText: ""},
		},
		DetailedInformation: "Premium SEO product review guide details for teams.",
		Summary:             "Premium SEO product review guide summary.",
		KeywordSet: KeywordSet{
			SEOTitle:          "Premium SEO Product Review Guide for Teams",
			Slug:              "premium-seo-product-review-guide",
			MetaDescription:   "Premium SEO product review guide for teams. Learn practical checks and better structure for optimized pages that convert and rank.",
			PrimaryKeyword:    "premium seo product review guide",
			SecondaryKeywords: "seo review, content quality",
			Synonyms:          "seo audit",
		},
	}

	review := GenerateReview(input)
	found := false
	for _, item := range review.ChecklistResults {
		if item.Code == "imageAltCoverageStrong" {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected imageAltCoverageStrong checklist item")
	}
}
