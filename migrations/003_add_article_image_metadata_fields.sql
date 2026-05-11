-- Add missing image metadata fields to align backend with frontend image dialog.
ALTER TABLE article_images
  ADD COLUMN IF NOT EXISTS title VARCHAR(255),
  ADD COLUMN IF NOT EXISTS caption TEXT,
  ADD COLUMN IF NOT EXISTS description TEXT;

