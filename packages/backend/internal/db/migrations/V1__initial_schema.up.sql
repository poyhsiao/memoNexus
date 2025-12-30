-- V1__initial_schema.up.sql
-- MemoNexus Core Platform - Initial Database Schema
-- Constitution Requirements:
-- - All tables use UUID v4 as Primary Keys
-- - Soft deletion for sync-participating records (is_deleted flag)
-- - Version field for conflict detection
-- - FTS5 external content table for search

-- =====================================================
-- Schema Version Tracking
-- =====================================================

CREATE TABLE schema_migrations (
    version INTEGER PRIMARY KEY CHECK(version > 0),
    applied_at INTEGER NOT NULL CHECK(applied_at > 0),
    description TEXT NOT NULL CHECK(length(description) > 0),
    checksum TEXT NOT NULL CHECK(length(checksum) = 64)  -- SHA-256 of migration SQL
);

-- =====================================================
-- Core Tables
-- =====================================================

-- content_items: Primary table for all captured content
CREATE TABLE content_items (
    -- Primary Key: UUID v4 (constitution requirement)
    id TEXT PRIMARY KEY NOT NULL CHECK(length(id) = 36),

    -- Content Metadata
    title TEXT NOT NULL CHECK(length(title) > 0 AND length(title) <= 500),
    content_text TEXT NOT NULL DEFAULT '',
    source_url TEXT CHECK(source_url IS NULL OR length(source_url) <= 2048),
    media_type TEXT NOT NULL CHECK(media_type IN ('web', 'image', 'video', 'pdf', 'markdown')),

    -- Analysis Results
    tags TEXT DEFAULT '',  -- Comma-separated tag names (denormalized for FTS)
    summary TEXT,  -- AI-generated summary (NULL if not generated)

    -- Soft Deletion (constitution requirement)
    is_deleted INTEGER NOT NULL DEFAULT 0 CHECK(is_deleted IN (0, 1)),

    -- Timestamps
    created_at INTEGER NOT NULL CHECK(created_at > 0),
    updated_at INTEGER NOT NULL CHECK(updated_at > 0 AND updated_at >= created_at),

    -- Conflict Resolution
    version INTEGER NOT NULL DEFAULT 1 CHECK(version > 0),

    -- Content Addressing (for deduplication during sync)
    content_hash TEXT  -- SHA-256 of content_text
);

-- Indexes for content_items
CREATE INDEX idx_content_items_created_at ON content_items(created_at DESC);
CREATE INDEX idx_content_items_updated_at ON content_items(updated_at DESC);
CREATE INDEX idx_content_items_is_deleted ON content_items(is_deleted);
CREATE INDEX idx_content_items_media_type ON content_items(media_type);
CREATE INDEX idx_content_items_content_hash ON content_items(content_hash);

-- =====================================================
-- Tags & Content-Tag Relationships
-- =====================================================

-- tags: User-defined labels for organizing content
CREATE TABLE tags (
    id TEXT PRIMARY KEY NOT NULL CHECK(length(id) = 36),
    name TEXT NOT NULL UNIQUE CHECK(length(name) > 0 AND length(name) <= 50),
    color TEXT DEFAULT '#3B82F6' CHECK(length(color) = 7 AND color LIKE '#%'),
    created_at INTEGER NOT NULL CHECK(created_at > 0),
    updated_at INTEGER NOT NULL CHECK(updated_at >= created_at),
    is_deleted INTEGER NOT NULL DEFAULT 0 CHECK(is_deleted IN (0, 1))
);

CREATE INDEX idx_tags_name ON tags(name);
CREATE INDEX idx_tags_is_deleted ON tags(is_deleted);

-- content_tags: Many-to-many relationship
CREATE TABLE content_tags (
    content_id TEXT NOT NULL CHECK(length(content_id) = 36),
    tag_id TEXT NOT NULL CHECK(length(tag_id) = 36),
    assigned_at INTEGER NOT NULL CHECK(assigned_at > 0),
    FOREIGN KEY (content_id) REFERENCES content_items(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (content_id, tag_id)
);

CREATE INDEX idx_content_tags_content_id ON content_tags(content_id);
CREATE INDEX idx_content_tags_tag_id ON content_tags(tag_id);

-- =====================================================
-- Sync & Conflict Tables
-- =====================================================

-- change_log: Tracks all mutations for incremental sync
CREATE TABLE change_log (
    id TEXT PRIMARY KEY NOT NULL CHECK(length(id) = 36),
    item_id TEXT NOT NULL CHECK(length(item_id) = 36),
    operation TEXT NOT NULL CHECK(operation IN ('create', 'update', 'delete')),
    version INTEGER NOT NULL CHECK(version > 0),
    timestamp INTEGER NOT NULL CHECK(timestamp > 0),
    FOREIGN KEY (item_id) REFERENCES content_items(id) ON DELETE CASCADE
);

CREATE INDEX idx_change_log_item_id ON change_log(item_id);
CREATE INDEX idx_change_log_timestamp ON change_log(timestamp DESC);
CREATE INDEX idx_change_log_operation ON change_log(operation);

-- conflict_log: Records resolved concurrent edits
CREATE TABLE conflict_log (
    id TEXT PRIMARY KEY NOT NULL CHECK(length(id) = 36),
    item_id TEXT NOT NULL CHECK(length(item_id) = 36),
    local_timestamp INTEGER NOT NULL CHECK(local_timestamp > 0),
    remote_timestamp INTEGER NOT NULL CHECK(remote_timestamp > 0),
    resolution TEXT NOT NULL DEFAULT 'last_write_wins' CHECK(resolution IN ('last_write_wins', 'manual')),
    detected_at INTEGER NOT NULL CHECK(detected_at > 0),
    FOREIGN KEY (item_id) REFERENCES content_items(id) ON DELETE CASCADE
);

CREATE INDEX idx_conflict_log_item_id ON conflict_log(item_id);
CREATE INDEX idx_conflict_log_detected_at ON conflict_log(detected_at DESC);

-- =====================================================
-- Configuration Tables
-- =====================================================

-- sync_credentials: Encrypted S3-compatible storage credentials
CREATE TABLE sync_credentials (
    id TEXT PRIMARY KEY NOT NULL CHECK(length(id) = 36),
    endpoint TEXT NOT NULL CHECK(length(endpoint) > 0),
    bucket_name TEXT NOT NULL CHECK(length(bucket_name) > 0),
    region TEXT,
    access_key_encrypted TEXT NOT NULL CHECK(length(access_key_encrypted) > 0),
    secret_key_encrypted TEXT NOT NULL CHECK(length(secret_key_encrypted) > 0),
    is_enabled INTEGER NOT NULL DEFAULT 0 CHECK(is_enabled IN (0, 1)),
    created_at INTEGER NOT NULL CHECK(created_at > 0),
    updated_at INTEGER NOT NULL CHECK(updated_at >= created_at)
);

CREATE INDEX idx_sync_credentials_is_enabled ON sync_credentials(is_enabled);

-- ai_config: Encrypted AI service configuration
CREATE TABLE ai_config (
    id TEXT PRIMARY KEY NOT NULL CHECK(length(id) = 36),
    provider TEXT NOT NULL CHECK(provider IN ('openai', 'claude', 'ollama')),
    api_endpoint TEXT NOT NULL CHECK(length(api_endpoint) > 0),
    api_key_encrypted TEXT NOT NULL CHECK(length(api_key_encrypted) > 0),
    model_name TEXT NOT NULL CHECK(length(model_name) > 0),
    max_tokens INTEGER DEFAULT 1000 CHECK(max_tokens > 0),
    is_enabled INTEGER NOT NULL DEFAULT 0 CHECK(is_enabled IN (0, 1)),
    created_at INTEGER NOT NULL CHECK(created_at > 0),
    updated_at INTEGER NOT NULL CHECK(updated_at >= created_at)
);

CREATE INDEX idx_ai_config_is_enabled ON ai_config(is_enabled);

-- =====================================================
-- Sync Queue Table
-- =====================================================

-- sync_queue: Pending sync operations queued during network unavailability
CREATE TABLE sync_queue (
    id TEXT PRIMARY KEY NOT NULL CHECK(length(id) = 36),
    operation TEXT NOT NULL CHECK(operation IN ('upload', 'download', 'delete')),
    payload TEXT NOT NULL CHECK(length(payload) > 0),
    retry_count INTEGER NOT NULL DEFAULT 0 CHECK(retry_count >= 0),
    max_retries INTEGER NOT NULL DEFAULT 3 CHECK(max_retries > 0),
    next_retry_at INTEGER NOT NULL CHECK(next_retry_at > 0),
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'in_progress', 'failed', 'completed')),
    created_at INTEGER NOT NULL CHECK(created_at > 0),
    updated_at INTEGER NOT NULL CHECK(updated_at >= created_at)
);

CREATE INDEX idx_sync_queue_status ON sync_queue(status);
CREATE INDEX idx_sync_queue_next_retry_at ON sync_queue(next_retry_at);

-- =====================================================
-- Export Tables
-- =====================================================

-- export_archives: Metadata for exported encrypted archives
CREATE TABLE export_archives (
    id TEXT PRIMARY KEY NOT NULL CHECK(length(id) = 36),
    file_path TEXT NOT NULL CHECK(length(file_path) > 0),
    checksum TEXT NOT NULL CHECK(length(checksum) = 64),
    size_bytes INTEGER NOT NULL CHECK(size_bytes > 0),
    item_count INTEGER NOT NULL CHECK(item_count >= 0),
    is_encrypted INTEGER NOT NULL DEFAULT 1 CHECK(is_encrypted IN (0, 1)),
    created_at INTEGER NOT NULL CHECK(created_at > 0)
);

CREATE INDEX idx_export_archives_created_at ON export_archives(created_at DESC);

-- =====================================================
-- Full-Text Search Table
-- =====================================================

-- content_fts: FTS5 virtual table for offline search
CREATE VIRTUAL TABLE content_fts USING fts5(
    title,
    content_text,
    tags,
    content=content_items,
    content_rowid=rowid,
    tokenize='unicode61 remove_diacritics 1'
);

-- Triggers to keep FTS synchronized
CREATE TRIGGER content_items_ai AFTER INSERT ON content_items BEGIN
    INSERT INTO content_fts(rowid, title, content_text, tags)
    VALUES (new.rowid, new.title, new.content_text, new.tags);
END;

CREATE TRIGGER content_items_ad AFTER DELETE ON content_items BEGIN
    INSERT INTO content_fts(content_fts, rowid, title, content_text, tags)
    VALUES ('delete', old.rowid, old.title, old.content_text, old.tags);
END;

CREATE TRIGGER content_items_au AFTER UPDATE ON content_items BEGIN
    INSERT INTO content_fts(content_fts, rowid, title, content_text, tags)
    VALUES ('delete', old.rowid, old.title, old.content_text, old.tags);
    INSERT INTO content_fts(rowid, title, content_text, tags)
    VALUES (new.rowid, new.title, new.content_text, new.tags);
END;

-- =====================================================
-- Seed Initial Migration Record
-- =====================================================

INSERT INTO schema_migrations (version, applied_at, description, checksum)
VALUES (1, strftime('%s', 'now'), 'Initial schema with FTS5 and sync tables',
        -- SHA-256 checksum placeholder (computed from migration SQL content)
        '0000000000000000000000000000000000000000000000000000000000000000');
