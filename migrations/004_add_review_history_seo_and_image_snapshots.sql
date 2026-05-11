-- Add per-review SEO and image snapshots so history responses stay immutable over time.
ALTER TABLE review_history
  ADD COLUMN IF NOT EXISTS seo_title_snapshot VARCHAR(255),
  ADD COLUMN IF NOT EXISTS meta_description_snapshot VARCHAR(500),
  ADD COLUMN IF NOT EXISTS slug_snapshot VARCHAR(255),
  ADD COLUMN IF NOT EXISTS secondary_keywords_snapshot TEXT,
  ADD COLUMN IF NOT EXISTS synonyms_snapshot TEXT,
  ADD COLUMN IF NOT EXISTS image_metadata_snapshot JSONB;

DROP VIEW IF EXISTS review_history_details;

CREATE VIEW review_history_details AS
SELECT
  rh.id,
  rh.review_id,
  rh.article_id,
  a.title AS article_title,
  a.slug,
  a.permanent_link,
  rh.action,
  rh.notes,
  rh.seo_score_snapshot,
  rh.readability_score_snapshot,
  rh.advanced_score_snapshot,
  rh.overall_score_snapshot,
  rh.status_snapshot,
  rh.primary_keyword_snapshot,
  rh.seo_title_snapshot,
  rh.meta_description_snapshot,
  rh.slug_snapshot,
  rh.secondary_keywords_snapshot,
  rh.synonyms_snapshot,
  rh.article_content_snapshot,
  rh.summary_snapshot,
  rh.detailed_information_snapshot,
  rh.image_metadata_snapshot,
  rh.keyword_density_snapshot,
  rh.word_count_snapshot,
  rh.created_at
FROM review_history rh
JOIN articles a ON rh.article_id = a.id;

