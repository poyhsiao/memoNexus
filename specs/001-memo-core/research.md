# Technical Research: MemoNexus Core Platform

**Feature Branch**: `001-memo-core` | **Date**: 2024-12-30 | **Phase**: 0 - Research

This document consolidates technical research findings that inform the implementation plan. Each technology choice includes the decision, rationale, and alternatives considered.

---

## 1. Framework & Language Versions

### 1.1 Flutter

**Decision**: Use **Flutter 3.27.0** (Stable channel)

**Rationale**:
- Latest stable release as of December 2024
- Enhanced performance improvements for desktop platforms (Windows, macOS, Linux)
- Improved Impeller rendering engine support for desktop
- Better integration with platform-native accessibility APIs (crucial for WCAG 2.1 AA compliance)
- Stable Dart 3.8.0 SDK with enhanced pattern matching and records

**Alternatives Considered**:
| Alternative | Why Rejected |
|-------------|--------------|
| Flutter 3.24 (LTS) | Missing critical desktop accessibility improvements needed for FR-064-FR-073 |
| Flutter 3.27 Beta/Dev channels | Not production-ready; stability is paramount for local-first data storage |

**Migration Path**: Pin to `3.27.0` in `pubspec.yaml` with `>=3.27.0 <4.0.0` for minor version updates.

---

### 1.2 Go

**Decision**: Use **Go 1.25.5** (Latest stable)

**Rationale**:
- Latest stable release as of December 2025
- Enhanced `go test` coverage instrumentation (supports 80% coverage requirement from constitution)
- Improved FFI c-library compilation for mobile platforms (Dart FFI integration)
- Better SQLite driver support with `modernc.org/sqlite` (pure Go, no CGO for mobile)
- Enhanced standard library for HTTP/2 and WebSocket (PocketBase communication)

**Alternatives Considered**| Alternative | Why Rejected |
|-------------|--------------|
| Go 1.21 (LTS) | Missing improved FFI tooling needed for mobile Dart FFI bridge |
| Go 1.24 | Less mature SQLite pure-go driver; version 1.25.5 includes critical bug fixes |

**Migration Path**: Use `.go-version` file to pin version. CI/CD should enforce `go version` check.

---

### 1.3 PocketBase

**Decision**: Use **PocketBase v0.35.0**

**Rationale**:
- Latest stable release as of December 2025
- Enhanced SQLite FTS5 virtual table support (built-in to underlying SQLite)
- Improved Go hooks system for custom business logic injection (`internal/` directory)
- Better embedded mode support for desktop packaging
- Enhanced WebSocket API for real-time sync state updates

**Alternatives Considered**:
| Alternative | Why Rejected |
|-------------|--------------|
| Direct SQLite (no PocketBase) | Would require building REST/WebSocket layer from scratch; PocketBase provides mature auth, logging, and CRUD |
| Raw Go with custom framework | Reinventing wheel; PocketBase aligns with constitution's "Platform Bridge" requirement |

**Integration Strategy**:
- Desktop: Embed PocketBase as subprocess; communicate via REST/WebSocket
- Mobile: Extract Go Core logic as shared library; bypass PocketBase runtime

---

## 2. Database & Search

### 2.1 SQLite FTS5 Implementation

**Decision**: Use **external content table pattern** with **BM25 ranking** and **unicode61 tokenizer**

**Key Configuration**:
```sql
-- Main content table (with UUID v4 PK)
CREATE TABLE content_items (
    id TEXT PRIMARY KEY,  -- UUID v4
    title TEXT NOT NULL,
    content_text TEXT,
    tags TEXT,  -- Comma-separated for FTS
    source_url TEXT,
    media_type TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    is_deleted INTEGER DEFAULT 0  -- Soft deletion flag
);

-- FTS5 virtual table (external content)
CREATE VIRTUAL TABLE content_fts USING fts5(
    title,
    content_text,
    tags,
    content=content_items,
    content_rowid=rowid,
    tokenize='unicode61 remove_diacritics 1'
);

-- Triggers to keep FTS synchronized
CREATE TRIGGER content_ai AFTER INSERT ON content_items BEGIN
    INSERT INTO content_fts(rowid, title, content_text, tags)
    VALUES (new.rowid, new.title, new.content_text, new.tags);
END;

CREATE TRIGGER content_ad AFTER DELETE ON content_items BEGIN
    INSERT INTO content_fts(content_fts, rowid, title, content_text, tags)
    VALUES ('delete', old.rowid, old.title, old.content_text, old.tags);
END;

CREATE TRIGGER content_au AFTER UPDATE ON content_items BEGIN
    INSERT INTO content_fts(content_fts, rowid, title, content_text, tags)
    VALUES ('delete', old.rowid, old.title, old.content_text, old.tags);
    INSERT INTO content_fts(rowid, title, content_text, tags)
    VALUES (new.rowid, new.title, new.content_text, new.tags);
END;
```

**Performance Optimizations**:
- **WAL Mode**: `PRAGMA journal_mode=WAL;` for concurrent reads/writes
- **Incremental FTS**: Use `INSERT INTO content_fts(content_fts, rank)` for large datasets
- **Query Optimization**: Limit results with `MATCH ... ORDER BY bm25(content_fts) LIMIT 20`
- **Index Strategy**: Create index on `created_at DESC` for reverse-chronological list views

**Unicode Handling**:
- `tokenize='unicode61 remove_diacritics 1'` ensures language-agnostic search
- Supports Chinese, Japanese, Korean (CJK) text segmentation
- Diacritic removal ensures "café" matches "cafe"

**Alternatives Considered**:
| Alternative | Why Rejected |
|-------------|--------------|
| Internal content table | Doubles storage; external content is more efficient |
| Simple tokenizer | Doesn't handle Unicode or diacritics; breaks search for international users |
| Custom Go-based search | Violates constitution's "Offline Search via SQLite FTS5" requirement |

---

### 2.2 Database Schema Versioning

**Decision**: Implement **version tracking table** with **V/R (Version/Rollback) migration convention**

**Schema Version Table**:
```sql
CREATE TABLE schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at INTEGER NOT NULL,
    description TEXT,
    checksum TEXT NOT NULL  -- SHA-256 of migration SQL
);

-- Seed initial version
INSERT INTO schema_migrations (version, applied_at, description, checksum)
VALUES (1, strftime('%s', 'now'), 'Initial schema with FTS5', '...SHA256...');
```

**Migration File Convention**:
```
internal/db/migrations/
├── V1__initial_schema.up.sql      -- Forward migration
├── V1__initial_schema.down.sql    -- Rollback migration
├── V2__add_ai_config.up.sql
├── V2__add_ai_config.down.sql
└── ...
```

**Migration Execution Strategy**:
```go
type Migration struct {
    Version    int
    AppliedAt  time.Time
    Description string
    Checksum    string
}

func ApplyMigration(db *sql.DB, m Migration) error {
    tx, _ := db.Begin()
    defer tx.Rollback()

    // Apply migration
    if _, err := tx.Exec(m.SQL); err != nil {
        return err
    }

    // Record version
    query := `INSERT INTO schema_migrations (version, applied_at, description, checksum)
              VALUES (?, ?, ?, ?)`
    if _, err := tx.Exec(query, m.Version, time.Now().Unix(), m.Description, m.Checksum); err != nil {
        return err
    }

    return tx.Commit()
}

func RollbackMigration(db *sql.DB, version int) error {
    tx, _ := db.Begin()
    defer tx.Rollback()

    // Execute down migration
    downSQL := loadMigrationSQL(version, "down")
    if _, err := tx.Exec(downSQL); err != nil {
        return err
    }

    // Remove version record
    if _, err := tx.Exec(`DELETE FROM schema_migrations WHERE version = ?`, version); err != nil {
        return err
    }

    return tx.Commit()
}
```

**Pre-set Data Versioning**:
```sql
-- Track preset data separately
CREATE TABLE preset_data (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    version INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

-- Example: Default AI configuration presets
INSERT INTO preset_data (key, value, version, updated_at)
VALUES
    ('ai_providers', '["openai","claude","ollama"]', 1, strftime('%s', 'now')),
    ('sync_intervals', '["immediate","5m","15m","1h","manual"]', 1, strftime('%s', 'now'));
```

**Alternatives Considered**:
| Alternative | Why Rejected |
|-------------|--------------|
| No version tracking | Cannot rollback; risk of data corruption during schema changes |
| Ad-hoc version column per table | No atomic rollback capability; inconsistent state across tables |
| External migration tools (golang-migrate) | Adds dependency; built-in solution is simpler for local-first app |

---

## 3. Package Management & Monorepo

### 3.1 pnpm Workspace Configuration

**Decision**: Use **pnpm workspace** for monorepo coordination with **separate lockfiles per language**

**Root `package.json`**:
```json
{
  "name": "memonexus",
  "private": true,
  "scripts": {
    "build": "pnpm --filter ./apps/desktop build",
    "test": "pnpm --filter ./packages/backend test && pnpm --filter ./apps/frontend test",
    "lint": "pnpm --filter './packages/*' lint",
    "clean": "pnpm --filter './apps/*' clean && pnpm --filter './packages/*' clean"
  },
  "devDependencies": {
    "@types/node": "^20.11.0",
    "typescript": "^5.3.3"
  }
}
```

**`pnpm-workspace.yaml`**:
```yaml
packages:
  - 'apps/*'
  - 'packages/*'

# Note: Go and Flutter have their own package management
# pnpm workspace ONLY manages:
# - Desktop build scripts (Electron/Tauri if applicable)
# - Shared tooling (ESLint, Prettier)
# - CI/CD orchestration scripts
```

**Directory Structure**:
```
memonexus/
├── apps/
│   ├── frontend/          # Flutter app (uses pubspec.yaml)
│   └── desktop/           # Desktop wrapper (uses pnpm)
├── packages/
│   ├── backend/           # Go Core (uses go.mod)
│   └── shared/            # Shared types (TypeScript)
├── pnpm-workspace.yaml
├── pnpm-lock.yaml         # Only for desktop/shared
├── go.work                # Go workspace
└── pubspec.yaml           # Flutter root (optional)
```

**Unified Build Scripts** (`scripts/build.sh`):
```bash
#!/usr/bin/env bash
set -euo pipefail

echo "Building MemoNexus..."

# Build Go Core (shared library for mobile, binary for desktop)
cd packages/backend
go build -o ../../build/bin/core ./cmd/core

# Build Flutter app
cd ../../apps/frontend
flutter build desktop --release

# Package desktop wrapper
cd ../desktop
pnpm build

echo "Build complete: build/dist/"
```

**Alternatives Considered**:
| Alternative | Why Rejected |
|-------------|--------------|
| npm workspaces | Slower, less disk-efficient; pnpm is constitution-mandated |
| Single package.json | Doesn't separate language-specific dependencies cleanly |
| Turborepo | Overkill for small team; adds complexity |

---

### 3.2 Cross-Platform Go Strategy

**Decision**: Use **Go workspace** (`go.work`) with **conditional compilation tags**

**`go.work`**:
```go
go 1.25.5

use (
    ./packages/backend
    ./packages/backend/internal
)
```

**Conditional Compilation**:
```go
// +build !mobile

package desktop

import "github.com/pocketbase/pocketbase"

// StartPocketBase launches embedded PocketBase (desktop only)
func StartPocketBase(dataDir string) error {
    app := pocketbase.NewWithConfig(pocketbase.Config{
        DataDir: dataDir,
    })
    return app.Start()
}
```

```go
// +build mobile

package mobile

import "database/sql"

// OpenSQLite opens raw SQLite connection (mobile only)
func OpenSQLite(dbPath string) (*sql.DB, error) {
    return sql.Open("sqlite", dbPath)
}
```

**Alternatives Considered**:
| Alternative | Why Rejected |
|-------------|--------------|
| Separate Go modules per platform | Code duplication; violates "Modular Core" principle |
| Single binary for all platforms | Cannot embed PocketBase in mobile (no subprocess support) |

---

## 4. Technology Stack Summary

| Component | Technology | Version | Rationale |
|-----------|-----------|---------|-----------|
| **UI Framework** | Flutter | 3.27.0 | Latest stable with desktop accessibility improvements |
| **Backend Logic** | Go | 1.25.5 | Latest stable with FFI tooling and pure-go SQLite |
| **Embedded Backend** | PocketBase | v0.35.0 | Enhanced FTS5 and WebSocket support |
| **Database** | SQLite | 3.x | FTS5 extension for offline search |
| **Search** | FTS5 | Built-in | BM25 ranking, <100ms queries |
| **Package Manager** | pnpm | Latest | Constitution-mandated, workspace support |
| **State Management** | Riverpod | Latest | Predictable data flow, testable |
| **Cryptographic** | AES-256 | Standard | Export encryption (FR-033) |

---

## 5. Open Questions Resolved

| Question | Answer | Impact |
|----------|--------|--------|
| How to handle database migrations? | Version tracking table with V/R files | Supports rollback (constitution requirement) |
| How to achieve <100ms search? | FTS5 with BM25, WAL mode, incremental indexing | Meets FR-014 performance requirement |
| How to structure monorepo? | pnpm workspace + Go workspace, separate lockfiles | Clean separation of language ecosystems |
| How to ensure schema rollback? | `.down.sql` migration files with transactions | Atomic rollback capability |
| How to handle Unicode search? | `unicode61` tokenizer with diacritic removal | International language support |

---

## 6. Technical Debt & Risks

### Known Risks

| Risk | Mitigation |
|------|------------|
| **SQLite on mobile** | Use `modernc.org/sqlite` (pure Go, no CGO) |
| **FTS5 scalability** | Implement incremental FTS for large datasets (>10K items) |
| **PocketBase mobile incompatibility** | Extract Go Core logic, bypass PocketBase runtime on mobile |
| **Dart FFI complexity** | Use `dart:ffi` with generated bindings via `ffigen` |
| **pnpm + Go + Flutter triad** | Separate build phases; use unified shell script orchestration |

### Future Considerations

1. **Zero-knowledge sync**: Client-side encryption before S3 upload (constitution mentions this as future work)
2. **Web PWA**: Currently out of scope; would require WASM compilation or separate architecture
3. **Real-time sync**: PocketBase WebSockets support desktop; mobile would need custom Go WebSocket server

---

**Next Phase**: Proceed to Phase 1 (Data Model & Contracts) using these technology decisions.
