# Implementation Plan: MemoNexus Core Platform

**Branch**: `001-memo-core` | **Date**: 2024-12-30 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-memo-core/spec.md`

## Summary

MemoNexus is a **local-first personal knowledge base** and media vault with content ingestion, intelligent analysis (TF-IDF standard, AI optional), offline search via SQLite FTS5, and optional S3 cloud sync. The platform uses Flutter for multi-platform UI, Go for shared business logic (Modular Core), SQLite with FTS5 for storage, and PocketBase for embedded desktop backend. Mobile platforms communicate with Go Core via Dart FFI, while desktop uses REST/WebSocket to an embedded PocketBase instance.

## Technical Context

**Language/Version**: Go 1.25.5, Dart 3.8.0 (Flutter 3.27.0), SQL (SQLite 3.x with FTS5)
**Primary Dependencies**:
- Go: github.com/pocketbase/pocketbase v0.35.0, modernc.org/sqlite (pure Go, no CGO), github.com/google/uuid
- Flutter: flutter 3.27.0, riverpod (state management), uuid, web_socket_channel
- Build: pnpm (workspace coordination), Go workspace

**Storage**: SQLite 3 with FTS5 extension (external content table pattern, BM25 ranking, unicode61 tokenizer), local file system with SHA-256 content addressing for media files

**Testing**: Go testing (`go test`), Flutter testing (`flutter test`), integration tests for cross-component interactions, contract tests for interface boundaries

**Target Platform**:
- Desktop: Windows, macOS, Linux (Flutter desktop with embedded PocketBase)
- Mobile: Android, iOS (Flutter with Go Core via Dart FFI)
- Future: Web PWA (planned for subsequent release)

**Project Type**: Monorepo (hybrid: packages/backend for Go Core, apps/frontend for Flutter UI, apps/desktop for desktop wrapper)

**Performance Goals**:
- Search: <100ms for up to 10,000 items (FTS5 with BM25)
- Content ingestion: <5 seconds for typical web pages and small files
- Content list rendering: <500ms for up to 1,000 items
- Offline availability: 100% (no core feature requires internet)

**Constraints**:
- Must work completely offline for all core features (capture, search, viewing)
- AI/S3 services are optional enhancements with graceful degradation
- Database schema must support rollback (version tracking with V/R migration files)
- All tables must use UUID v4 as Primary Keys (constitution requirement)
- Soft deletion (`is_deleted` flag) for all records participating in sync

**Scale/Scope**:
- Typical users: 1,000-10,000 content items
- System remains performant up to 100,000 items
- Concurrent editing: Single user with multiple windows/devices (last write wins)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Phase 0 Check (Before Research)

| Principle | Requirement | Status | Notes |
|-----------|-------------|--------|-------|
| **I. Local-First** | Core features work offline | ✅ PASS | Research will verify SQLite FTS5 offline capability |
| **II. Modular Core** | Go logic in `internal/` | ✅ PASS | Architecture specifies `packages/backend/internal/` |
| **III. Data Integrity** | UUID v4 Primary Keys | ✅ PASS | Data model defines UUID for all tables |
| **IV. Platform Bridge** | Dart FFI for mobile | ✅ PASS | Architecture confirms FFI for mobile, REST/WebSocket for desktop |
| **V. Offline Search** | SQLite FTS5 required | ✅ PASS | Research confirms FTS5 external content table pattern |
| **VI. Test-First** | Go tests required | ✅ PASS | Testing strategy includes `go test` with 80% coverage |

**Result**: ✅ **GATE PASSED** - Proceed to Phase 0 research

### Post-Phase 1 Check (After Design)

| Principle | Requirement | Status | Evidence |
|-----------|-------------|--------|----------|
| **I. Local-First** | Core features work offline | ✅ PASS | FR-015: Search works completely offline; AI/S3 are optional |
| **II. Modular Core** | Go logic in `internal/` | ✅ PASS | Structure: `packages/backend/internal/{db,parser,analysis,sync,export}` |
| **III. Data Integrity** | UUID v4 Primary Keys | ✅ PASS | All tables use `TEXT PRIMARY KEY` with UUID format (data-model.md) |
| **IV. Platform Bridge** | Dart FFI for mobile | ✅ PASS | WebSocket contract for desktop; FFI planned for mobile (contracts/websocket.md) |
| **V. Offline Search** | SQLite FTS5 required | ✅ PASS | `content_fts` virtual table with BM25 ranking (data-model.md) |
| **VI. Test-First** | Go tests required | ✅ PASS | Testing strategy defined (quickstart.md); 80% coverage requirement |

**Result**: ✅ **GATE PASSED** - Proceed to Phase 2 (tasks)

## Project Structure

### Documentation (this feature)

```text
specs/001-memo-core/
├── spec.md              # Feature specification (73 functional requirements, 24 success criteria)
├── plan.md              # This file (implementation plan)
├── research.md          # Phase 0: Technical research (Flutter 3.27.0, Go 1.25.5, PocketBase v0.35.0)
├── data-model.md        # Phase 1: Database schema (10 entities, FTS5 configuration)
├── quickstart.md        # Phase 1: Developer onboarding guide
├── contracts/           # Phase 1: API contracts
│   ├── openapi.yaml     # REST API specification (OpenAPI 3.0)
│   └── websocket.md     # WebSocket protocol (real-time events)
└── checklists/
    └── requirements.md  # Quality validation checklist
```

### Source Code (repository root)

```text
memonexus/
├── apps/
│   ├── frontend/                    # Flutter multi-platform UI
│   │   ├── lib/
│   │   │   ├── models/              # Dart data models (ContentItem, Tag, etc.)
│   │   │   ├── screens/             # UI screens (home, search, settings)
│   │   │   ├── widgets/             # Reusable widgets
│   │   │   ├── services/            # API clients, WebSocket, FFI bridge
│   │   │   │   ├── api_client.dart  # REST API client (desktop only)
│   │   │   │   ├── websocket.dart   # WebSocket client (desktop only)
│   │   │   │   └── ffi_bridge.dart  # Dart FFI bridge (mobile only)
│   │   │   ├── providers/           # Riverpod state providers
│   │   │   └── main.dart
│   │   ├── test/
│   │   └── pubspec.yaml
│   │
│   └── desktop/                     # Desktop wrapper (Tauri/Electron)
│       ├── src/
│       ├── package.json
│       └── pnpm-lock.yaml
│
├── packages/
│   ├── backend/                     # Go Core (platform-agnostic library)
│   │   ├── cmd/
│   │   │   ├── core/                # Main entry point (shared library)
│   │   │   ├── desktop/             # Desktop server (PocketBase wrapper)
│   │   │   └── migrate/             # Migration tool
│   │   ├── internal/
│   │   │   ├── db/                  # Database layer
│   │   │   │   ├── migrations/      # V/R migration files (V1__*.up.sql, V1__*.down.sql)
│   │   │   │   ├── schema.go        # Schema definitions
│   │   │   │   └── repository.go    # CRUD operations
│   │   │   ├── parser/              # Content ingestion
│   │   │   │   ├── web/             # Web scraping (HTML parsing)
│   │   │   │   ├── document/        # PDF, Docx parsing
│   │   │   │   └── media/           # Image/video metadata extraction
│   │   │   ├── analysis/            # Content analysis
│   │   │   │   ├── tfidf/           # Standard mode (offline)
│   │   │   │   ├── ai/              # AI mode (optional, OpenAI/Claude/Ollama)
│   │   │   │   └── textrank/        # Keyword extraction
│   │   │   ├── sync/                # Cloud synchronization
│   │   │   │   ├── s3/              # S3-compatible client
│   │   │   │   ├── conflict/        # Last write wins resolution
│   │   │   │   └── queue/           # Sync queue management
│   │   │   ├── export/              # Export/import
│   │   │   │   ├── archive/         # AES-256 encrypted archives
│   │   │   │   └── compression/     # tar.gz compression
│   │   │   └── models/              # Go struct definitions
│   │   ├── go.mod
│   │   └── go.sum
│   │
│   └── shared/                      # Shared TypeScript types (desktop)
│       ├── src/
│       │   ├── types.ts             # API contract types
│       │   └── api.ts               # API client
│       └── package.json
│
├── scripts/
│   ├── build.sh                     # Unified build script
│   ├── test.sh                      # Unified test script
│   ├── dev.sh                       # Development environment launcher
│   └── build-mobile-lib.sh          # Go Core → shared library for mobile
│
├── pnpm-workspace.yaml              # Monorepo configuration
├── go.work                          # Go workspace
├── .clauderules                     # AI agent development rules
├── CLAUDE.md                        # Project-specific instructions
├── .specify/                        # SpecKit workflow templates
│   ├── memory/
│   │   └── constitution.md          # Project constitution
│   └── templates/
├── specs/                           # Feature specifications
│   ├── 01_prd.md
│   ├── 02_architecture.md
│   ├── 03_dev_guidelines.md
│   └── 001-memo-core/               # This feature
└── README.md
```

**Structure Decision**: Hybrid monorepo combining Go workspace (`packages/backend`) with pnpm workspace (`apps/`, `packages/shared`). This separation allows:
- Go Core to be compiled as shared library for mobile (Dart FFI)
- Go Core to run as embedded server for desktop (PocketBase)
- Flutter UI to remain platform-agnostic
- Desktop wrapper to handle platform-specific packaging

**Desktop Packaging Strategy Clarification**:
- **Primary approach**: Flutter builds native desktop executables (Windows/macOS/Linux) via Flutter desktop embedding
- **apps/frontend/**: Flutter multi-platform project that builds for desktop (Windows `.exe`, macOS `.app`, Linux binary)
- **packages/backend/cmd/desktop/**: PocketBase server executable that the Flutter desktop app embeds and starts
- **apps/desktop/**: Optional distribution wrapper for advanced packaging needs (installers, code signing, auto-updates, tray icons)
- **Communication**: Flutter desktop → REST/WebSocket → PocketBase (localhost:8090)

The `apps/desktop/` wrapper (Tauri/Electron) is OPTIONAL for production distribution. For initial development, Flutter desktop builds directly provide cross-platform desktop apps with embedded PocketBase server.

### Key Architecture Patterns

1. **Modular Core (Go)**: All business logic in `packages/backend/internal/`
   - `db/`: Database operations with transaction support
   - `parser/`: Content extraction (web, PDF, media)
   - `analysis/`: TF-IDF (standard) + AI (optional)
   - `sync/`: S3-compatible sync with conflict resolution
   - `export/`: Encrypted archive export/import

2. **Platform Bridge**:
   - **Desktop**: REST/WebSocket → `http://localhost:8090` (embedded PocketBase)
   - **Mobile**: Dart FFI → compiled Go shared library (`.a`/`.framework`)

3. **Data Model**:
   - All tables use UUID v4 Primary Keys
   - Soft deletion with `is_deleted` flag
   - Version field for conflict detection
   - FTS5 external content table for search

4. **API Contract**:
   - REST: OpenAPI 3.0 specification (`contracts/openapi.yaml`)
   - WebSocket: Real-time events (`contracts/websocket.md`)
   - Mobile: Direct FFI calls (no HTTP)

## Complexity Tracking

> **No constitution violations detected.** All design decisions align with the 6 core principles defined in `.specify/memory/constitution.md`.
