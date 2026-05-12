-- Align production schema with the current SEO_BE runtime expectations.
-- Safe to run on an existing database that may already contain some of these changes.

BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ---------------------------------------------------------------------------
-- 1) Ensure article_images has the metadata columns used by the backend/UI.
-- ---------------------------------------------------------------------------
ALTER TABLE article_images
  ADD COLUMN IF NOT EXISTS title VARCHAR(255),
  ADD COLUMN IF NOT EXISTS caption TEXT,
  ADD COLUMN IF NOT EXISTS description TEXT;

-- ---------------------------------------------------------------------------
-- 2) Ensure review_history stores the full immutable snapshots expected by code.
-- ---------------------------------------------------------------------------
ALTER TABLE review_history
  ADD COLUMN IF NOT EXISTS seo_title_snapshot VARCHAR(255),
  ADD COLUMN IF NOT EXISTS meta_description_snapshot VARCHAR(500),
  ADD COLUMN IF NOT EXISTS slug_snapshot VARCHAR(255),
  ADD COLUMN IF NOT EXISTS secondary_keywords_snapshot TEXT,
  ADD COLUMN IF NOT EXISTS synonyms_snapshot TEXT,
  ADD COLUMN IF NOT EXISTS article_content_snapshot TEXT,
  ADD COLUMN IF NOT EXISTS summary_snapshot TEXT,
  ADD COLUMN IF NOT EXISTS detailed_information_snapshot TEXT,
  ADD COLUMN IF NOT EXISTS image_metadata_snapshot JSONB;

-- ---------------------------------------------------------------------------
-- 3) Remove incorrect per-column UNIQUE constraints that block multi-row inserts.
--    The backend expects composite uniqueness only.
-- ---------------------------------------------------------------------------
ALTER TABLE review_checklist_results
  DROP CONSTRAINT IF EXISTS review_checklist_results_review_id_key,
  DROP CONSTRAINT IF EXISTS review_checklist_results_checklist_item_id_key;

ALTER TABLE review_field_feedback
  DROP CONSTRAINT IF EXISTS review_field_feedback_review_id_key,
  DROP CONSTRAINT IF EXISTS review_field_feedback_field_name_key;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'review_checklist_results_review_id_checklist_item_id_key'
  ) THEN
    ALTER TABLE review_checklist_results
      ADD CONSTRAINT review_checklist_results_review_id_checklist_item_id_key
      UNIQUE (review_id, checklist_item_id);
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'review_field_feedback_review_id_field_name_key'
  ) THEN
    ALTER TABLE review_field_feedback
      ADD CONSTRAINT review_field_feedback_review_id_field_name_key
      UNIQUE (review_id, field_name);
  END IF;
END $$;

-- ---------------------------------------------------------------------------
-- 4) Recreate the history view so it exposes the current snapshot shape.
-- ---------------------------------------------------------------------------
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
  rh.internal_links_snapshot,
  rh.outbound_links_snapshot,
  rh.checklist_changes,
  rh.recommendations,
  rh.created_at
FROM review_history rh
JOIN articles a ON rh.article_id = a.id;

COMMIT;
