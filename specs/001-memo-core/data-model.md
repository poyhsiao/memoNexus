# Data Model: MemoNexus Core Platform

**Feature Branch**: `001-memo-core` | **Date**: 2024-12-30 | **Phase**: 1 - Design

This document defines the complete data model for the MemoNexus platform, including entities, fields, relationships, validation rules, and database schema.

---

## 1. Entity Relationship Diagram

```mermaid
erDiagram
    CONTENT_ITEMS ||--o{ TAGS : has
    CONTENT_ITEMS ||--|| CHANGE_LOG : tracks
    CONTENT_ITEMS ||--o{ SEARCH_INDEX : mirrors
    CONTENT_ITEMS ||--o{ EXPORT_ARCHIVES : exports
    CONTENT_ITEMS ||--o{ CONFLICT_LOG : conflicts
    CONTENT_ITEMS ||--o{ SYNC_QUEUE : queues

    AI_CONFIG ||--|| CONTENT_ITEMS : analyzes
    SYNC_CREDENTIAL ||--o{ SYNC_QUEUE : authenticates

    CONTENT_ITEMS {
        UUID id PK
        TEXT title
        TEXT content_text
        TEXT tags
        TEXT source_url
        TEXT media_type
        INTEGER created_at
        INTEGER updated_at
        BOOLEAN is_deleted
        TEXT summary
    }

    TAGS {
        UUID id PK
        TEXT name UK
        TEXT color
    }

    CHANGE_LOG {
        UUID id PK
        UUID item_id FK
        TEXT operation
        INTEGER version
        INTEGER timestamp
    }

    SEARCH_INDEX {
        INTEGER rowid PK
        TEXT title
        TEXT content_text
        TEXT tags
    }

    AI_CONFIG {
        UUID id PK
        TEXT provider
        TEXT api_endpoint
        TEXT api_key_encrypted
        BOOLEAN enabled
    }

    SYNC_CREDENTIAL {
        UUID id PK
        TEXT endpoint
        TEXT access_key_encrypted
        TEXT secret_key_encrypted
        TEXT bucket_name
    }

    EXPORT_ARCHIVES {
        UUID id PK
        TEXT file_path
        TEXT checksum
        INTEGER size_bytes
        INTEGER created_at
    }

    CONFLICT_LOG {
        UUID id PK
        UUID item_id FK
        INTEGER local_timestamp
        INTEGER remote_timestamp
        TEXT resolution
    }

    SYNC_QUEUE {
        UUID id PK
        TEXT operation
        TEXT payload
        INTEGER retry_count
        INTEGER next_retry_at
        TEXT status
    }
}
```

---

## 2. Database Schema (SQLite)

### 2.1 Core Tables

#### `content_items`

Primary table for all captured content (web pages, files, etc.).

```sql
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
    content_hash TEXT,  -- SHA-256 of content_text

    -- Indexes
    UNIQUE(id)
);

CREATE INDEX idx_content_items_created_at ON content_items(created_at DESC);
CREATE INDEX idx_content_items_updated_at ON content_items(updated_at DESC);
CREATE INDEX idx_content_items_is_deleted ON content_items(is_deleted);
CREATE INDEX idx_content_items_media_type ON content_items(media_type);
CREATE INDEX idx_content_items_content_hash ON content_items(content_hash);
```

**Validation Rules**:
- `id`: Must be valid UUID v4 (regex: `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
- `title`: Required, max 500 characters
- `media_type`: Enum (web, image, video, pdf, markdown)
- `created_at`, `updated_at`: Unix timestamps
- `version`: Incremented on each update (for conflict detection)

---

#### `tags`

User-defined labels for organizing content.

```sql
CREATE TABLE tags (
    -- Primary Key: UUID v4
    id TEXT PRIMARY KEY NOT NULL CHECK(length(id) = 36),

    -- Tag Metadata
    name TEXT NOT NULL UNIQUE CHECK(length(name) > 0 AND length(name) <= 50),
    color TEXT DEFAULT '#3B82F6' CHECK(length(color) = 7 AND color LIKE '#%'),

    -- Timestamps
    created_at INTEGER NOT NULL CHECK(created_at > 0),
    updated_at INTEGER NOT NULL CHECK(updated_at >= created_at),

    -- Soft Deletion
    is_deleted INTEGER NOT NULL DEFAULT 0 CHECK(is_deleted IN (0, 1))
);

CREATE INDEX idx_tags_name ON tags(name);
CREATE INDEX idx_tags_is_deleted ON tags(is_deleted);
```

**Validation Rules**:
- `name`: Unique, max 50 characters
- `color`: Hex color code (e.g., #3B82F6)

---

#### `content_tags`

Many-to-many relationship between content items and tags.

```sql
CREATE TABLE content_tags (
    -- Composite Primary Key
    content_id TEXT NOT NULL CHECK(length(content_id) = 36),
    tag_id TEXT NOT NULL CHECK(length(tag_id) = 36),

    -- Timestamp
    assigned_at INTEGER NOT NULL CHECK(assigned_at > 0),

    -- Foreign Keys
    FOREIGN KEY (content_id) REFERENCES content_items(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE,

    -- Primary Key & Unique Constraint
    PRIMARY KEY (content_id, tag_id)
);

CREATE INDEX idx_content_tags_content_id ON content_tags(content_id);
CREATE INDEX idx_content_tags_tag_id ON content_tags(tag_id);
```

---

### 2.2 Sync & Conflict Tables

#### `change_log`

Tracks all mutations for incremental sync and concurrent edit detection.

```sql
CREATE TABLE change_log (
    -- Primary Key: UUID v4
    id TEXT PRIMARY KEY NOT NULL CHECK(length(id) = 36),

    -- Reference
    item_id TEXT NOT NULL CHECK(length(item_id) = 36),

    -- Change Metadata
    operation TEXT NOT NULL CHECK(operation IN ('create', 'update', 'delete')),

    -- Version Control
    version INTEGER NOT NULL CHECK(version > 0),

    -- Timestamp
    timestamp INTEGER NOT NULL CHECK(timestamp > 0),

    -- Foreign Key
    FOREIGN KEY (item_id) REFERENCES content_items(id) ON DELETE CASCADE
);

CREATE INDEX idx_change_log_item_id ON change_log(item_id);
CREATE INDEX idx_change_log_timestamp ON change_log(timestamp DESC);
CREATE INDEX idx_change_log_operation ON change_log(operation);
```

---

#### `conflict_log`

Records resolved concurrent edits for user awareness.

```sql
CREATE TABLE conflict_log (
    -- Primary Key: UUID v4
    id TEXT PRIMARY KEY NOT NULL CHECK(length(id) = 36),

    -- Reference
    item_id TEXT NOT NULL CHECK(length(item_id) = 36),

    -- Conflicting Timestamps
    local_timestamp INTEGER NOT NULL CHECK(local_timestamp > 0),
    remote_timestamp INTEGER NOT NULL CHECK(remote_timestamp > 0),

    -- Resolution
    resolution TEXT NOT NULL DEFAULT 'last_write_wins' CHECK(resolution IN ('last_write_wins', 'manual')),

    -- Metadata
    detected_at INTEGER NOT NULL CHECK(detected_at > 0),

    -- Foreign Key
    FOREIGN KEY (item_id) REFERENCES content_items(id) ON DELETE CASCADE
);

CREATE INDEX idx_conflict_log_item_id ON conflict_log(item_id);
CREATE INDEX idx_conflict_log_detected_at ON conflict_log(detected_at DESC);
```

---

### 2.3 Configuration Tables

#### `sync_credentials`

Encrypted S3-compatible storage credentials.

```sql
CREATE TABLE sync_credentials (
    -- Primary Key: UUID v4
    id TEXT PRIMARY KEY NOT NULL CHECK(length(id) = 36),

    -- S3 Configuration
    endpoint TEXT NOT NULL CHECK(length(endpoint) > 0),
    bucket_name TEXT NOT NULL CHECK(length(bucket_name) > 0),
    region TEXT,

    -- Encrypted Credentials
    access_key_encrypted TEXT NOT NULL CHECK(length(access_key_encrypted) > 0),
    secret_key_encrypted TEXT NOT NULL CHECK(length(secret_key_encrypted) > 0),

    -- Status
    is_enabled INTEGER NOT NULL DEFAULT 0 CHECK(is_enabled IN (0, 1)),

    -- Timestamps
    created_at INTEGER NOT NULL CHECK(created_at > 0),
    updated_at INTEGER NOT NULL CHECK(updated_at >= created_at)
);

CREATE INDEX idx_sync_credentials_is_enabled ON sync_credentials(is_enabled);
```

**Encryption**: Use platform-secure storage (Keychain on macOS, Credential Manager on Windows, Keystore on Android/iOS) with AES-256-GCM.

---

#### `ai_config`

Encrypted AI service configuration.

```sql
CREATE TABLE ai_config (
    -- Primary Key: UUID v4
    id TEXT PRIMARY KEY NOT NULL CHECK(length(id) = 36),

    -- AI Provider
    provider TEXT NOT NULL CHECK(provider IN ('openai', 'claude', 'ollama')),
    api_endpoint TEXT NOT NULL CHECK(length(api_endpoint) > 0),

    -- Encrypted API Key
    api_key_encrypted TEXT NOT NULL CHECK(length(api_key_encrypted) > 0),

    -- Model Configuration
    model_name TEXT NOT NULL CHECK(length(model_name) > 0),
    max_tokens INTEGER DEFAULT 1000 CHECK(max_tokens > 0),

    -- Status
    is_enabled INTEGER NOT NULL DEFAULT 0 CHECK(is_enabled IN (0, 1)),

    -- Timestamps
    created_at INTEGER NOT NULL CHECK(created_at > 0),
    updated_at INTEGER NOT NULL CHECK(updated_at >= created_at)
);

CREATE INDEX idx_ai_config_is_enabled ON ai_config(is_enabled);
```

---

### 2.4 Sync Queue Table

#### `sync_queue`

Pending sync operations queued during network unavailability.

```sql
CREATE TABLE sync_queue (
    -- Primary Key: UUID v4
    id TEXT PRIMARY KEY NOT NULL CHECK(length(id) = 36),

    -- Operation Metadata
    operation TEXT NOT NULL CHECK(operation IN ('upload', 'download', 'delete')),
    payload TEXT NOT NULL CHECK(length(payload) > 0),  -- JSON payload

    -- Retry Logic
    retry_count INTEGER NOT NULL DEFAULT 0 CHECK(retry_count >= 0),
    max_retries INTEGER NOT NULL DEFAULT 3 CHECK(max_retries > 0),
    next_retry_at INTEGER NOT NULL CHECK(next_retry_at > 0),

    -- Status
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'in_progress', 'failed', 'completed')),

    -- Timestamps
    created_at INTEGER NOT NULL CHECK(created_at > 0),
    updated_at INTEGER NOT NULL CHECK(updated_at >= created_at)
);

CREATE INDEX idx_sync_queue_status ON sync_queue(status);
CREATE INDEX idx_sync_queue_next_retry_at ON sync_queue(next_retry_at);
```

**Exponential Backoff**: Calculate `next_retry_at = created_at + (2^retry_count * 60)` seconds.

---

### 2.5 Export Tables

#### `export_archives`

Metadata for exported encrypted archives.

```sql
CREATE TABLE export_archives (
    -- Primary Key: UUID v4
    id TEXT PRIMARY KEY NOT NULL CHECK(length(id) = 36),

    -- Archive Metadata
    file_path TEXT NOT NULL CHECK(length(file_path) > 0),
    checksum TEXT NOT NULL CHECK(length(checksum) = 64),  -- SHA-256
    size_bytes INTEGER NOT NULL CHECK(size_bytes > 0),
    item_count INTEGER NOT NULL CHECK(item_count >= 0),

    -- Encryption Confirmation
    is_encrypted INTEGER NOT NULL DEFAULT 1 CHECK(is_encrypted IN (0, 1)),

    -- Timestamp
    created_at INTEGER NOT NULL CHECK(created_at > 0)
);

CREATE INDEX idx_export_archives_created_at ON export_archives(created_at DESC);
```

---

### 2.6 Full-Text Search Table

#### `content_fts`

FTS5 virtual table for offline search.

```sql
-- FTS5 Virtual Table (External Content)
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
```

**Query Pattern**:
```sql
-- Search with BM25 ranking
SELECT
    ci.id, ci.title, ci.media_type, ci.created_at,
    bm25(content_fts) as relevance
FROM content_items ci
JOIN content_fts ON ci.rowid = content_fts.rowid
WHERE content_fts MATCH ? AND ci.is_deleted = 0
ORDER BY relevance, ci.created_at DESC
LIMIT 20;
```

---

### 2.7 Schema Migration Tracking

#### `schema_migrations`

Tracks applied schema migrations.

```sql
CREATE TABLE schema_migrations (
    version INTEGER PRIMARY KEY CHECK(version > 0),
    applied_at INTEGER NOT NULL CHECK(applied_at > 0),
    description TEXT NOT NULL CHECK(length(description) > 0),
    checksum TEXT NOT NULL CHECK(length(checksum) = 64)  -- SHA-256 of migration SQL
);

-- Initial migration
INSERT INTO schema_migrations (version, applied_at, description, checksum)
VALUES (1, strftime('%s', 'now'), 'Initial schema with FTS5 and sync tables', '...');
```

---

## 3. Data Model Constraints

### 3.1 Primary Keys (Constitution Requirement)

**ALL tables MUST use UUID v4 as Primary Keys.**

```go
// UUID v4 Generation
import "github.com/google/uuid"

func NewUUID() string {
    return uuid.New().String()  // Format: "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx"
}
```

**Rationale**: Prevents ID collision during multi-device synchronization.

---

### 3.2 Soft Deletion (Constitution Requirement)

**All tables that participate in sync MUST use soft deletion.**

Tables requiring soft deletion:
- `content_items` (`is_deleted`)
- `tags` (`is_deleted`)

**Pattern**:
```sql
-- Delete (soft)
UPDATE content_items SET is_deleted = 1, updated_at = ? WHERE id = ?;

-- Query (exclude deleted)
SELECT * FROM content_items WHERE is_deleted = 0;
```

---

### 3.3 Timestamps (Constitution Requirement)

**All tables MUST use Unix timestamps (`INTEGER`) for time tracking.**

```sql
-- Current timestamp in SQLite
strftime('%s', 'now')
```

**Go Conversion**:
```go
import "time"

func ToUnix(t time.Time) int64 {
    return t.Unix()
}

func FromUnix(ts int64) time.Time {
    return time.Unix(ts, 0)
}
```

---

### 3.4 Conflict Resolution (Constitution Requirement)

**ALL concurrent edits use "Last Write Wins" based on `updated_at` timestamp.**

```go
type ContentItem struct {
    ID        string
    Title     string
    UpdatedAt int64  // Used for conflict resolution
    Version   int    // Incremented on each update
}

// Update with conflict detection
func (item *ContentItem) Update(newTitle string, newUpdatedAt int64) error {
    if newUpdatedAt <= item.UpdatedAt {
        return errors.New("conflict: remote update is older or same")
    }
    item.Title = newTitle
    item.UpdatedAt = newUpdatedAt
    item.Version++
    return nil
}
```

---

## 4. Go Struct Definitions

```go
package models

import (
    "database/sql/driver"
    "encoding/json"
    "time"
)

// UUID wrapper for type safety
type UUID string

func (u UUID) Value() (driver.Value, error) {
    return string(u), nil
}

func (u *UUID) Scan(value interface{}) error {
    if value == nil {
        *u = ""
        return nil
    }
    *u = UUID(value.([]byte))
    return nil
}

// ContentItem represents a captured content
type ContentItem struct {
    ID          UUID    `db:"id" json:"id"`
    Title       string  `db:"title" json:"title"`
    ContentText string  `db:"content_text" json:"content_text"`
    SourceURL   string  `db:"source_url" json:"source_url,omitempty"`
    MediaType   string  `db:"media_type" json:"media_type"`
    Tags        string  `db:"tags" json:"tags"`  // Comma-separated
    Summary     string  `db:"summary" json:"summary,omitempty"`
    IsDeleted   bool    `db:"is_deleted" json:"is_deleted"`
    CreatedAt   int64   `db:"created_at" json:"created_at"`
    UpdatedAt   int64   `db:"updated_at" json:"updated_at"`
    Version     int     `db:"version" json:"version"`
    ContentHash string  `db:"content_hash" json:"content_hash,omitempty"`
}

// Tag represents a user-defined label
type Tag struct {
    ID        UUID    `db:"id" json:"id"`
    Name      string  `db:"name" json:"name"`
    Color     string  `db:"color" json:"color"`
    IsDeleted bool    `db:"is_deleted" json:"is_deleted"`
    CreatedAt int64   `db:"created_at" json:"created_at"`
    UpdatedAt int64   `db:"updated_at" json:"updated_at"`
}

// ChangeLog tracks mutations for sync
type ChangeLog struct {
    ID        UUID    `db:"id" json:"id"`
    ItemID    UUID    `db:"item_id" json:"item_id"`
    Operation string  `db:"operation" json:"operation"`  // create, update, delete
    Version   int     `db:"version" json:"version"`
    Timestamp int64   `db:"timestamp" json:"timestamp"`
}

// SyncCredential holds encrypted S3 configuration
type SyncCredential struct {
    ID                UUID    `db:"id" json:"id"`
    Endpoint          string  `db:"endpoint" json:"endpoint"`
    BucketName        string  `db:"bucket_name" json:"bucket_name"`
    Region            string  `db:"region" json:"region,omitempty"`
    AccessKeyEncrypted string `db:"access_key_encrypted" json:"-"`  // Never expose
    SecretKeyEncrypted string `db:"secret_key_encrypted" json:"-"`  // Never expose
    IsEnabled         bool    `db:"is_enabled" json:"is_enabled"`
    CreatedAt         int64   `db:"created_at" json:"created_at"`
    UpdatedAt         int64   `db:"updated_at" json:"updated_at"`
}

// AIConfig holds encrypted AI service configuration
type AIConfig struct {
    ID              UUID    `db:"id" json:"id"`
    Provider        string  `db:"provider" json:"provider"`  // openai, claude, ollama
    APIEndpoint     string  `db:"api_endpoint" json:"api_endpoint"`
    APIKeyEncrypted string  `db:"api_key_encrypted" json:"-"`  // Never expose
    ModelName       string  `db:"model_name" json:"model_name"`
    MaxTokens       int     `db:"max_tokens" json:"max_tokens"`
    IsEnabled       bool    `db:"is_enabled" json:"is_enabled"`
    CreatedAt       int64   `db:"created_at" json:"created_at"`
    UpdatedAt       int64   `db:"updated_at" json:"updated_at"`
}

// ConflictLog records resolved concurrent edits
type ConflictLog struct {
    ID             UUID    `db:"id" json:"id"`
    ItemID         UUID    `db:"item_id" json:"item_id"`
    LocalTimestamp int64   `db:"local_timestamp" json:"local_timestamp"`
    RemoteTimestamp int64   `db:"remote_timestamp" json:"remote_timestamp"`
    Resolution     string  `db:"resolution" json:"resolution"`  // last_write_wins, manual
    DetectedAt     int64   `db:"detected_at" json:"detected_at"`
}

// SyncQueue represents a pending sync operation
type SyncQueue struct {
    ID          UUID             `db:"id" json:"id"`
    Operation   string           `db:"operation" json:"operation"`  // upload, download, delete
    Payload     json.RawMessage  `db:"payload" json:"payload"`
    RetryCount  int              `db:"retry_count" json:"retry_count"`
    MaxRetries  int              `db:"max_retries" json:"max_retries"`
    NextRetryAt int64            `db:"next_retry_at" json:"next_retry_at"`
    Status      string           `db:"status" json:"status"`  // pending, in_progress, failed, completed
    CreatedAt   int64            `db:"created_at" json:"created_at"`
    UpdatedAt   int64            `db:"updated_at" json:"updated_at"`
}

// ExportArchive holds metadata for exported archives
type ExportArchive struct {
    ID         UUID    `db:"id" json:"id"`
    FilePath   string  `db:"file_path" json:"file_path"`
    Checksum   string  `db:"checksum" json:"checksum"`  // SHA-256
    SizeBytes  int64   `db:"size_bytes" json:"size_bytes"`
    ItemCount  int     `db:"item_count" json:"item_count"`
    IsEncrypted bool   `db:"is_encrypted" json:"is_encrypted"`
    CreatedAt  int64   `db:"created_at" json:"created_at"`
}
```

---

## 5. Flutter Data Models

```dart
import 'package:uuid/uuid.dart';

// UUID helper
final _uuid = Uuid();

// ContentItem model
class ContentItem {
  final String id;
  final String title;
  final String contentText;
  final String? sourceUrl;
  final MediaType mediaType;
  final List<String> tags;
  final String? summary;
  final bool isDeleted;
  final DateTime createdAt;
  final DateTime updatedAt;
  final int version;

  ContentItem({
    String? id,
    required this.title,
    required this.contentText,
    this.sourceUrl,
    required this.mediaType,
    required this.tags,
    this.summary,
    this.isDeleted = false,
    required this.createdAt,
    required this.updatedAt,
    required this.version,
  }) : id = id ?? _uuid.v4();

  factory ContentItem.fromJson(Map<String, dynamic> json) {
    return ContentItem(
      id: json['id'] as String,
      title: json['title'] as String,
      contentText: json['content_text'] as String,
      sourceUrl: json['source_url'] as String?,
      mediaType: MediaType.values.firstWhere(
        (e) => e.name == json['media_type'],
        orElse: () => MediaType.web,
      ),
      tags: (json['tags'] as String).split(',').where((t) => t.isNotEmpty).toList(),
      summary: json['summary'] as String?,
      isDeleted: json['is_deleted'] == 1,
      createdAt: DateTime.fromMillisecondsSinceEpoch(json['created_at'] * 1000),
      updatedAt: DateTime.fromMillisecondsSinceEpoch(json['updated_at'] * 1000),
      version: json['version'] as int,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'title': title,
      'content_text': contentText,
      'source_url': sourceUrl,
      'media_type': mediaType.name,
      'tags': tags.join(','),
      'summary': summary,
      'is_deleted': isDeleted ? 1 : 0,
      'created_at': createdAt.millisecondsSinceEpoch ~/ 1000,
      'updated_at': updatedAt.millisecondsSinceEpoch ~/ 1000,
      'version': version,
    };
  }

  ContentItem copyWith({
    String? title,
    String? contentText,
    List<String>? tags,
    String? summary,
    bool? isDeleted,
    DateTime? updatedAt,
  }) {
    return ContentItem(
      id: id,
      title: title ?? this.title,
      contentText: contentText ?? this.contentText,
      sourceUrl: sourceUrl,
      mediaType: mediaType,
      tags: tags ?? this.tags,
      summary: summary ?? this.summary,
      isDeleted: isDeleted ?? this.isDeleted,
      createdAt: createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
      version: version + 1,
    );
  }
}

enum MediaType { web, image, video, pdf, markdown }

// Tag model
class Tag {
  final String id;
  final String name;
  final String color;
  final bool isDeleted;
  final DateTime createdAt;
  final DateTime updatedAt;

  Tag({
    String? id,
    required this.name,
    this.color = '#3B82F6',
    this.isDeleted = false,
    required this.createdAt,
    required this.updatedAt,
  }) : id = id ?? _uuid.v4();

  factory Tag.fromJson(Map<String, dynamic> json) {
    return Tag(
      id: json['id'] as String,
      name: json['name'] as String,
      color: json['color'] as String? ?? '#3B82F6',
      isDeleted: json['is_deleted'] == 1,
      createdAt: DateTime.fromMillisecondsSinceEpoch(json['created_at'] * 1000),
      updatedAt: DateTime.fromMillisecondsSinceEpoch(json['updated_at'] * 1000),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      'color': color,
      'is_deleted': isDeleted ? 1 : 0,
      'created_at': createdAt.millisecondsSinceEpoch ~/ 1000,
      'updated_at': updatedAt.millisecondsSinceEpoch ~/ 1000,
    };
  }
}

// SearchResult model
class SearchResult {
  final ContentItem item;
  final double relevance;
  final List<String> matchedTerms;

  SearchResult({
    required this.item,
    required this.relevance,
    required this.matchedTerms,
  });
}
```

---

## 6. Validation Rules Summary

| Field | Type | Constraints | Validation |
|-------|------|-------------|------------|
| `content_items.id` | UUID | v4 format | Regex: `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$` |
| `content_items.title` | TEXT | Non-empty, ≤500 chars | `length(title) > 0 AND length(title) <= 500` |
| `content_items.media_type` | ENUM | web, image, video, pdf, markdown | `media_type IN ('web', 'image', 'video', 'pdf', 'markdown')` |
| `content_items.updated_at` | INTEGER | ≥ created_at | `updated_at >= created_at` |
| `tags.name` | TEXT | Unique, ≤50 chars | `UNIQUE`, `length(name) <= 50` |
| `tags.color` | TEXT | Hex color | `length(color) = 7 AND color LIKE '#%'` |
| `sync_credentials.endpoint` | TEXT | Non-empty | `length(endpoint) > 0` |
| `export_archives.checksum` | TEXT | SHA-256 | `length(checksum) = 64` |
| `export_archives.size_bytes` | INTEGER | Positive | `size_bytes > 0` |

---

**Next Phase**: Proceed to API Contracts (`contracts/` directory) and Quickstart Guide (`quickstart.md`).
