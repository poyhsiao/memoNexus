-- V2__composite_indexes.down.sql
-- T222: Rollback composite indexes

-- Drop composite indexes
DROP INDEX IF EXISTS idx_content_items_is_deleted_created_at;
DROP INDEX IF EXISTS idx_content_items_is_deleted_media_type_created_at;
DROP INDEX IF EXISTS idx_tags_is_deleted_name;
DROP INDEX IF EXISTS idx_sync_queue_status_next_retry_at;

-- Remove migration record
DELETE FROM schema_migrations WHERE version = 2;
