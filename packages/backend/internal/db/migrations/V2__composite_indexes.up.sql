-- V2__composite_indexes.up.sql
-- T222: Add composite indexes for common query patterns
-- These indexes optimize queries that filter on multiple columns together

-- =====================================================
-- Composite Indexes for content_items
-- =====================================================

-- Covering index for ListContentItems query: WHERE is_deleted = 0 ORDER BY created_at DESC
-- This index covers both the filter and sort, avoiding separate index lookups
CREATE INDEX IF NOT EXISTS idx_content_items_is_deleted_created_at
ON content_items(is_deleted, created_at DESC);

-- Covering index for filtered list queries: WHERE is_deleted = 0 AND media_type = ? ORDER BY created_at DESC
-- Supports content list filtered by media type
CREATE INDEX IF NOT EXISTS idx_content_items_is_deleted_media_type_created_at
ON content_items(is_deleted, media_type, created_at DESC);

-- =====================================================
-- Composite Indexes for tags
-- =====================================================

-- Covering index for ListTags query: WHERE is_deleted = 0 ORDER BY name
-- Supports tag listing with soft delete filter
CREATE INDEX IF NOT EXISTS idx_tags_is_deleted_name
ON tags(is_deleted, name);

-- =====================================================
-- Composite Indexes for sync queue
-- =====================================================

-- Covering index for pending queue items: WHERE status = 'pending' AND next_retry_at <= ?
-- Supports sync queue processing for retry logic
CREATE INDEX IF NOT EXISTS idx_sync_queue_status_next_retry_at
ON sync_queue(status, next_retry_at);

-- =====================================================
-- Migration Tracking Note
-- =====================================================
-- The migration system (migrate.go) automatically records this migration
-- in the schema_migrations table with version, timestamp, and checksum.
