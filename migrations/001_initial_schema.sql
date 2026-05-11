-- SEO Premium Backend - Initial Schema
-- PostgreSQL Migration: Articles, SEO Reviews, and Review History

-- =====================================================
-- ARTICLES & CONTENT
-- =====================================================

CREATE TABLE articles (
  id SERIAL PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  slug VARCHAR(255) UNIQUE NOT NULL,
  permanent_link VARCHAR(500) NOT NULL UNIQUE,
  content TEXT NOT NULL,
  summary TEXT,
  detailed_information TEXT,
  status VARCHAR(50) DEFAULT 'draft', -- draft, published, archived
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  published_at TIMESTAMP,
  deleted_at TIMESTAMP
);

CREATE INDEX idx_articles_status ON articles(status);
CREATE INDEX idx_articles_created_at ON articles(created_at DESC);
CREATE INDEX idx_articles_slug ON articles(slug);

-- =====================================================
-- SEO METADATA
-- =====================================================

CREATE TABLE article_seo_metadata (
  id SERIAL PRIMARY KEY,
  article_id INTEGER NOT NULL UNIQUE REFERENCES articles(id) ON DELETE CASCADE,
  seo_title VARCHAR(255),
  meta_description VARCHAR(500),
  slug VARCHAR(255),
  primary_keyword VARCHAR(255),
  secondary_keywords TEXT, -- comma-separated
  synonyms TEXT, -- comma-separated
  canonical_url VARCHAR(500),
  og_title VARCHAR(255),
  og_description VARCHAR(500),
  og_image_url TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_article_seo_metadata_article_id ON article_seo_metadata(article_id);
CREATE INDEX idx_article_seo_metadata_primary_keyword ON article_seo_metadata(primary_keyword);

-- =====================================================
-- ARTICLE IMAGES
-- =====================================================

CREATE TABLE article_images (
  id SERIAL PRIMARY KEY,
  article_id INTEGER NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
  image_name VARCHAR(255),
  mime_type VARCHAR(50),
  data_url TEXT,
  alt_text VARCHAR(255),
  title VARCHAR(255),
  caption TEXT,
  description TEXT,
  sort_order INTEGER DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_article_images_article_id ON article_images(article_id);

-- =====================================================
-- CONTENT METRICS & ANALYTICS
-- =====================================================

CREATE TABLE content_metrics (
  id SERIAL PRIMARY KEY,
  article_id INTEGER NOT NULL UNIQUE REFERENCES articles(id) ON DELETE CASCADE,
  word_count INTEGER,
  sentence_count INTEGER,
  heading_count INTEGER,
  internal_link_count INTEGER,
  outbound_link_count INTEGER,
  image_count INTEGER,
  average_sentence_length DECIMAL(5,2),
  passive_voice_ratio DECIMAL(5,3),
  repeated_starts_ratio DECIMAL(5,3),
  transition_word_count INTEGER,
  keyword_density DECIMAL(5,3),
  readability_level VARCHAR(50), -- elementary, intermediate, advanced
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_content_metrics_article_id ON content_metrics(article_id);

-- =====================================================
-- SEO REVIEWS & SCORING
-- =====================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE seo_reviews (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  article_id INTEGER NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
  overall_score INTEGER CHECK (overall_score >= 0 AND overall_score <= 100),
  seo_score INTEGER CHECK (seo_score >= 0 AND seo_score <= 100),
  readability_score INTEGER CHECK (readability_score >= 0 AND readability_score <= 100),
  advanced_score INTEGER CHECK (advanced_score >= 0 AND advanced_score <= 100),
  status VARCHAR(50) DEFAULT 'good', -- good, needs_improvement, poor
  notes TEXT,
  is_final BOOLEAN DEFAULT false,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_seo_reviews_article_id ON seo_reviews(article_id);
CREATE INDEX idx_seo_reviews_overall_score ON seo_reviews(overall_score);
CREATE INDEX idx_seo_reviews_created_at ON seo_reviews(created_at DESC);
CREATE INDEX idx_seo_reviews_status ON seo_reviews(status);

-- =====================================================
-- CHECKLIST ITEMS & RESULTS
-- =====================================================

CREATE TABLE review_checklist_items (
  id SERIAL PRIMARY KEY,
  check_code VARCHAR(100) UNIQUE NOT NULL,
  check_name VARCHAR(255) NOT NULL,
  check_group VARCHAR(50) NOT NULL, -- SEO, Readability, Advanced
  default_reason TEXT,
  default_improvement TEXT,
  sort_order INTEGER DEFAULT 0,
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_review_checklist_items_group ON review_checklist_items(check_group);

CREATE TABLE review_checklist_results (
  id SERIAL PRIMARY KEY,
  review_id UUID NOT NULL REFERENCES seo_reviews(id) ON DELETE CASCADE,
  checklist_item_id INTEGER NOT NULL REFERENCES review_checklist_items(id) ON DELETE CASCADE,
  result VARCHAR(100), -- passed, failed, warning
  status VARCHAR(50), -- success, needs_improvement, failed
  reason TEXT,
  improvement TEXT,
  affected_fields TEXT, -- JSON array of field names
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(review_id, checklist_item_id)
);

CREATE INDEX idx_review_checklist_results_review_id ON review_checklist_results(review_id);
CREATE INDEX idx_review_checklist_results_status ON review_checklist_results(status);

-- =====================================================
-- FIELD FEEDBACK
-- =====================================================

CREATE TABLE review_field_feedback (
  id SERIAL PRIMARY KEY,
  review_id UUID NOT NULL REFERENCES seo_reviews(id) ON DELETE CASCADE,
  field_name VARCHAR(100) NOT NULL,
  field_label VARCHAR(255),
  messages TEXT, -- JSON array of feedback messages
  severity VARCHAR(50), -- info, warning, error
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(review_id, field_name)
);

CREATE INDEX idx_review_field_feedback_review_id ON review_field_feedback(review_id);
CREATE INDEX idx_review_field_feedback_field_name ON review_field_feedback(field_name);

-- =====================================================
-- IMPROVEMENT RECOMMENDATIONS
-- =====================================================

CREATE TABLE improvement_recommendations (
  id SERIAL PRIMARY KEY,
  review_id UUID NOT NULL REFERENCES seo_reviews(id) ON DELETE CASCADE,
  recommendation TEXT NOT NULL,
  priority VARCHAR(50) DEFAULT 'medium', -- low, medium, high, critical
  estimated_impact VARCHAR(50), -- low, medium, high
  is_completed BOOLEAN DEFAULT false,
  completed_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_improvement_recommendations_review_id ON improvement_recommendations(review_id);
CREATE INDEX idx_improvement_recommendations_priority ON improvement_recommendations(priority);

-- =====================================================
-- REVIEW HISTORY & AUDIT TRAIL
-- =====================================================

CREATE TABLE review_history (
  id SERIAL PRIMARY KEY,
  review_id UUID NOT NULL REFERENCES seo_reviews(id) ON DELETE CASCADE,
  article_id INTEGER NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
  action VARCHAR(100), -- created, updated, scored, finalized, archived, improved
  notes TEXT,
  -- SEO Scores at time of review
  seo_score_snapshot INTEGER,
  readability_score_snapshot INTEGER,
  advanced_score_snapshot INTEGER,
  overall_score_snapshot INTEGER,
  status_snapshot VARCHAR(50),
  -- Key metrics snapshot
  primary_keyword_snapshot VARCHAR(255),
  seo_title_snapshot VARCHAR(255),
  meta_description_snapshot VARCHAR(500),
  slug_snapshot VARCHAR(255),
  secondary_keywords_snapshot TEXT,
  synonyms_snapshot TEXT,
  image_metadata_snapshot JSONB,
  keyword_density_snapshot DECIMAL(5,3),
  word_count_snapshot INTEGER,
  internal_links_snapshot INTEGER,
  outbound_links_snapshot INTEGER,
  -- Checklist changes (JSON for detailed tracking)
  checklist_changes JSONB, -- Stores which checks passed/failed
  recommendations JSONB, -- Stores recommendations at time of review
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_review_history_review_id ON review_history(review_id);
CREATE INDEX idx_review_history_article_id ON review_history(article_id);
CREATE INDEX idx_review_history_action ON review_history(action);
CREATE INDEX idx_review_history_created_at ON review_history(created_at DESC);

-- =====================================================
-- ARTICLE REVIEW SUMMARY (GROUP BY ARTICLE)
-- =====================================================

CREATE TABLE article_review_summary (
  id SERIAL PRIMARY KEY,
  article_id INTEGER NOT NULL UNIQUE REFERENCES articles(id) ON DELETE CASCADE,
  total_reviews INTEGER DEFAULT 0,
  latest_review_id UUID REFERENCES seo_reviews(id) ON DELETE SET NULL,
  latest_overall_score INTEGER,
  latest_seo_score INTEGER,
  latest_readability_score INTEGER,
  latest_advanced_score INTEGER,
  latest_status VARCHAR(50),
  best_overall_score INTEGER,
  worst_overall_score INTEGER,
  avg_overall_score DECIMAL(5,2),
  score_trend VARCHAR(50), -- improving, declining, stable
  last_reviewed_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_article_review_summary_article_id ON article_review_summary(article_id);
CREATE INDEX idx_article_review_summary_latest_overall_score ON article_review_summary(latest_overall_score);

-- =====================================================
-- SETTINGS & ORGANIZATION-WIDE CONFIG
-- =====================================================

CREATE TABLE review_settings (
  id SERIAL PRIMARY KEY,
  setting_key VARCHAR(100) UNIQUE NOT NULL,
  setting_value TEXT,
  description TEXT,
  data_type VARCHAR(50), -- string, integer, decimal, boolean, json
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Default review settings
INSERT INTO review_settings (setting_key, setting_value, description, data_type) VALUES
('min_article_length', '180', 'Minimum word count for articles', 'integer'),
('max_keyword_density', '2.5', 'Maximum allowed keyword density percentage', 'decimal'),
('min_keyword_density', '0.5', 'Minimum keyword density percentage', 'decimal'),
('meta_description_min_length', '120', 'Minimum meta description length', 'integer'),
('meta_description_max_length', '160', 'Maximum meta description length', 'integer'),
('seo_title_min_length', '45', 'Minimum SEO title length', 'integer'),
('seo_title_max_length', '65', 'Maximum SEO title length', 'integer'),
('avg_sentence_length_max', '22', 'Maximum average sentence length', 'decimal'),
('passive_voice_max_ratio', '0.2', 'Maximum allowed passive voice ratio', 'decimal'),
('required_internal_links', '1', 'Minimum internal links required', 'integer'),
('required_outbound_links', '1', 'Minimum outbound links required', 'integer');

-- =====================================================
-- VIEWS FOR COMMON QUERIES
-- =====================================================

-- View: Articles with latest review scores
CREATE VIEW articles_with_latest_reviews AS
SELECT 
  a.id,
  a.title,
  a.slug,
  a.status,
  a.created_at as article_created_at,
  sr.id as latest_review_id,
  sr.overall_score,
  sr.seo_score,
  sr.readability_score,
  sr.advanced_score,
  sr.status as review_status,
  sr.created_at as review_created_at,
  asm.total_reviews,
  asm.avg_overall_score,
  asm.score_trend
FROM articles a
LEFT JOIN seo_reviews sr ON a.id = sr.article_id
LEFT JOIN article_review_summary asm ON a.id = asm.article_id
WHERE sr.id = (
  SELECT id FROM seo_reviews sr2 
  WHERE sr2.article_id = a.id 
  ORDER BY sr2.created_at DESC 
  LIMIT 1
)
OR sr.id IS NULL;

-- View: Review history with article details
CREATE VIEW review_history_details AS
SELECT 
  rh.id,
  rh.review_id,
  rh.article_id,
  a.title as article_title,
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

-- View: Articles grouped by score trend
CREATE VIEW articles_by_score_trend AS
SELECT 
  a.id,
  a.title,
  a.slug,
  a.status,
  asm.latest_overall_score,
  asm.best_overall_score,
  asm.worst_overall_score,
  asm.avg_overall_score,
  asm.score_trend,
  asm.total_reviews,
  asm.last_reviewed_at
FROM articles a
JOIN article_review_summary asm ON a.id = asm.article_id
ORDER BY asm.score_trend DESC, asm.latest_overall_score DESC;
