package repository

import (
	"testing"

	"github.com/seo/backend/internal/service"
)

func TestBuildImageSnapshotsPreservesMetadata(t *testing.T) {
	images := []service.ImportedImage{
		{
			ID:          "img-1",
			Name:        "image-1.png",
			MimeType:    "image/png",
			DataURL:     "data:image/png;base64,AAAA",
			AltText:     "alt one",
			Title:       "title one",
			Caption:     "caption one",
			Description: "description one",
		},
		{
			ID:          "img-2",
			Name:        "image-2.png",
			MimeType:    "image/png",
			DataURL:     "data:image/png;base64,BBBB",
			AltText:     "alt two",
			Title:       "title two",
			Caption:     "caption two",
			Description: "description two",
		},
	}

	snapshots := buildImageSnapshots(images)
	if len(snapshots) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snapshots))
	}
	if snapshots[0].SortOrder != 0 || snapshots[1].SortOrder != 1 {
		t.Fatalf("expected stable sort order")
	}
	if snapshots[1].AltText != "alt two" || snapshots[1].Description != "description two" {
		t.Fatalf("expected metadata fields to be preserved")
	}
}
