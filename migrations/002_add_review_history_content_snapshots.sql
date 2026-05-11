-- Add per-review content snapshots so history detail reflects the exact data at review time.

ALTER TABLE review_history
  ADD COLUMN IF NOT EXISTS article_content_snapshot TEXT,
  ADD COLUMN IF NOT EXISTS summary_snapshot TEXT,
  ADD COLUMN IF NOT EXISTS detailed_information_snapshot TEXT;
