# Tasks: MemoNexus Core Platform

**Input**: Design documents from `/specs/001-memo-core/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: Constitution requires 80% test coverage. Test tasks are included for Go Core logic (Unit tests) and cross-component integration (Integration tests).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4, US5)
- Include exact file paths in descriptions

## Path Conventions

Based on plan.md structure:
- **Go Core**: `packages/backend/` (internal/, cmd/)
- **Flutter UI**: `apps/frontend/lib/`
- **Desktop wrapper**: `apps/desktop/`
- **Shared types**: `packages/shared/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and monorepo structure

- [X] T001 Create monorepo structure: apps/, packages/, scripts/ directories
- [X] T002 Initialize Go workspace with go.work file at repository root
- [X] T003 [P] Initialize pnpm workspace with pnpm-workspace.yaml at repository root
- [X] T004 [P] Create Go module packages/backend with go.mod (Go 1.25.5, PocketBase v0.35.0, pure-Go SQLite)
- [X] T005 [P] Create Flutter project apps/frontend (Flutter 3.27.0, Riverpod, UUID package)
- [X] T006 [P] Create packages/shared TypeScript project for type definitions
- [X] T007 [P] Create unified build script scripts/build.sh for Go + Flutter coordination
- [X] T008 [P] Create unified test script scripts/test.sh with coverage reporting
- [X] T009 [P] Create development launcher scripts/dev.sh for local development
- [X] T010 [P] Create build-mobile-lib.sh script in scripts/build-mobile-lib.sh (Go Core → shared library for mobile FFI: go build -buildmode=c-shared for Android .so, iOS .framework)

**Checkpoint**: ✅ Monorepo structure ready, all package managers initialized, mobile build script ready

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Database Layer (Go Core)

- [X] T011 Create database schema version tracking table in packages/backend/internal/db/migrations/V1__initial_schema.up.sql
- [X] T012 [P] Create content_items table in packages/backend/internal/db/migrations/V1__initial_schema.up.sql (UUID v4 PK, is_deleted, version, FTS5 mirrors)
- [X] T013 [P] Create tags table in packages/backend/internal/db/migrations/V1__initial_schema.up.sql (UUID v4 PK, is_deleted)
- [X] T014 [P] Create content_tags many-to-many table in packages/backend/internal/db/migrations/V1__initial_schema.up.sql
- [X] T015 [P] Create change_log table in packages/backend/internal/db/migrations/V1__initial_schema.up.sql (for sync)
- [X] T016 [P] Create conflict_log table in packages/backend/internal/db/migrations/V1__initial_schema.up.sql (for concurrent edits)
- [X] T017 [P] Create sync_credentials table in packages/backend/internal/db/migrations/V1__initial_schema.up.sql (encrypted S3 config)
- [X] T018 [P] Create ai_config table in packages/backend/internal/db/migrations/V1__initial_schema.up.sql (encrypted AI config)
- [X] T019 [P] Create sync_queue table in packages/backend/internal/db/migrations/V1__initial_schema.up.sql (for offline queuing)
- [X] T020 [P] Create export_archives table in packages/backend/internal/db/migrations/V1__initial_schema.up.sql
- [X] T021 Create content_fts FTS5 virtual table in packages/backend/internal/db/migrations/V1__initial_schema.up.sql (external content, BM25 ranking)
- [X] T022 Create FTS5 synchronization triggers in packages/backend/internal/db/migrations/V1__initial_schema.up.sql
- [X] T023 Create rollback migration V1__initial_schema.down.sql in packages/backend/internal/db/migrations/
- [X] T024 Implement database migration system in packages/backend/internal/db/migrate/migrate.go (version tracking, rollback support)

### Go Core Models (Data Layer)

- [X] T025 [P] Create ContentItem Go struct in packages/backend/internal/models/content.go (UUID, Title, ContentText, Tags, MediaType, IsDeleted, timestamps, version, content_hash)
- [X] T026 [P] Create Tag Go struct in packages/backend/internal/models/tag.go (UUID, Name, Color, IsDeleted, timestamps)
- [X] T027 [P] Create ChangeLog Go struct in packages/backend/internal/models/change_log.go (UUID, ItemID, Operation, Version, Timestamp)
- [X] T028 [P] Create ConflictLog Go struct in packages/backend/internal/models/conflict_log.go (UUID, ItemID, LocalTimestamp, RemoteTimestamp, Resolution)
- [X] T029 [P] Create SyncCredential Go struct in packages/backend/internal/models/sync_credential.go (encrypted fields)
- [X] T030 [P] Create AIConfig Go struct in packages/backend/internal/models/ai_config.go (encrypted fields)
- [X] T031 [P] Create SyncQueue Go struct in packages/backend/internal/models/sync_queue.go
- [X] T032 [P] Create ExportArchive Go struct in packages/backend/internal/models/export_archive.go

### Database Repository (CRUD Operations)

- [X] T033 Implement database connection manager in packages/backend/internal/db/db.go (SQLite with WAL mode, FTS5 enabled)
- [X] T034 Implement ContentItem repository in packages/backend/internal/db/repository.go (CRUD with transactions)
- [X] T035 [P] Implement Tag repository in packages/backend/internal/db/repository.go
- [X] T036 [P] Implement ChangeLog repository in packages/backend/internal/db/repository.go
- [X] T037 [P] Implement ConflictLog repository in packages/backend/internal/db/repository.go
- [X] T038 [P] Implement SyncQueue repository in packages/backend/internal/db/repository.go

### Flutter Data Models (Dart Layer)

- [X] T039 [P] Create ContentItem Dart model in apps/frontend/lib/models/content_item.dart (fromJson, toJson, copyWith, UUID generation)
- [X] T040 [P] Create Tag Dart model in apps/frontend/lib/models/content_item.dart
- [X] T041 [P] Create SearchResult Dart model in apps/frontend/lib/models/content_item.dart
- [X] T042 [P] Create AIConfig Dart model in apps/frontend/lib/models/ai_config.dart
- [X] T043 [P] Create SyncCredential Dart model in apps/frontend/lib/models/ai_config.dart
- [X] T044 [P] Create ExportArchive Dart model in apps/frontend/lib/models/ai_config.dart
- [X] T045 [P] Create MediaType enum in apps/frontend/lib/models/content_item.dart

### Core Services (Go & Flutter)

- [X] T046 Implement UUID generator utility in packages/backend/internal/uuid/uuid.go (UUID v4 format validation)
- [X] T047 [P] Implement error code system in packages/backend/internal/errors/errors.go (error codes bridging Go-Dart boundary)
- [X] T048 [P] Implement structured logging in packages/backend/internal/logging/logger.go (JSON format, log levels)
- [X] T049 [P] Implement API client base in apps/frontend/lib/services/api_client.dart (REST client for desktop)
- [X] T050 [P] Implement WebSocket client in apps/frontend/lib/services/websocket.dart (real-time events)
- [X] T051 [P] Implement local storage service in apps/frontend/lib/services/storage.dart (SQLite for mobile testing)

### Desktop API Server (PocketBase Integration)

- [X] T052 Create embedded PocketBase server in packages/backend/cmd/desktop/main.go (starts on localhost:8090)
- [X] T053 [P] Implement REST API handler for /content in packages/backend/cmd/desktop/handlers/content.go
- [X] T054 [P] Implement REST API handler for /tags in packages/backend/cmd/desktop/handlers/tags.go
- [X] T055 [P] Implement REST API handler for /search in packages/backend/cmd/desktop/handlers/search.go
- [X] T056 [P] Implement WebSocket server for real-time events in packages/backend/cmd/desktop/websocket.go

### FFI Bridge (Mobile - Foundational)

- [X] T057 [P] Implement Dart FFI bridge structure in apps/frontend/lib/services/ffi_bridge.dart (FFI setup, function signatures, error handling)
- [X] T058 [P] Implement Go FFI exports header in packages/backend/cmd/mobile/exports.go (export directives for all Go Core functions)
- [X] T059 [P] Implement Dart FFI function bindings for content operations in apps/frontend/lib/services/ffi_bridge.dart (go_content_create, go_content_list, go_content_get, go_content_update, go_content_delete)
- [X] T060 [P] Implement Go FFI exports for content operations in packages/backend/cmd/mobile/exports.go (ContentCreate, ContentList, ContentGet, ContentUpdate, ContentDelete functions)
- [X] T061 [P] Implement Dart FFI function bindings for search in apps/frontend/lib/services/ffi_bridge.dart (go_search_query, go_search_filters)
- [X] T062 [P] Implement Go FFI exports for search in packages/backend/cmd/mobile/exports.go (SearchQuery function with FTS5)
- [X] T063 [P] Implement Dart FFI function bindings for analysis in apps/frontend/lib/services/ffi_bridge.dart (go_analyze_keywords, go_generate_summary)
- [X] T064 [P] Implement Go FFI exports for analysis in packages/backend/cmd/mobile/exports.go (AnalyzeKeywords, GenerateSummary functions)

### Performance Validation (Constitution Requirement)

- [X] T065 [P] Performance benchmark: search 10,000 items in <100ms in packages/backend/internal/db/search_benchmark_test.go (constitution requirement SC-002, SC-007)
- [X] T066 [P] Performance benchmark: content list render 1,000 items in <500ms in packages/backend/internal/db/list_benchmark_test.go (constitution requirement FR-039, SC-005)
- [X] T067 [P] Performance benchmark: content ingestion 100 items in 10 minutes in packages/backend/internal/parser/ingestion_benchmark_test.go (constitution requirement FR-038, SC-006)
- [X] T068 [P] Memory profiling for Go Core in packages/backend/internal/memory/profile_test.go (constitution requirement: identify leaks)
- [X] T069 [P] Offline verification: test all features without network in tests/integration/offline_test.go (constitution requirement: 100% offline availability)

**Checkpoint**: Foundation ready - database, models, services, API server, FFI bridge, performance benchmarks all implemented. User story implementation can now begin in parallel.

---

## Phase 3: User Story 1 - Content Capture & Organization (Priority: P1) 🎯 MVP

**Goal**: Users can capture content from URLs and files, with automatic metadata extraction and tag-based organization

**Independent Test**:
1. Paste a URL → system fetches and extracts content, displays success confirmation with preview
2. Drag-drop a file → system stores it, generates thumbnail (if media), adds to knowledge base
3. View content list → items displayed with thumbnails, titles, tags, timestamps (reverse chronological)
4. Add custom tags → tags are persisted and displayed in item details

### Unit Tests for User Story 1

- [ ] T083 [P] [US1] Unit test for ContentItem repository CRUD operations in packages/backend/internal/db/repository_test.go
- [ ] T084 [P] [US1] Unit test for Tag repository CRUD operations in packages/backend/internal/db/repository_test.go
- [ ] T085 [P] [US1] Unit test for UUID generation and validation in packages/backend/internal/uuid/uuid_test.go

### Implementation for User Story 1

#### Content Ingestion (Go Core - Parser Engine)

- [X] T086 [P] [US1] Implement web scraper in packages/backend/internal/parser/web/scraper.go (fetch URL, extract clean text, metadata, OpenGraph)
- [ ] T087 [P] [US1] Implement PDF parser in packages/backend/internal/parser/document/pdf.go (extract text from PDF)
- [ ] T088 [P] [US1] Implement Markdown parser in packages/backend/internal/parser/document/markdown.go (parse structured markdown)
- [ ] T089 [P] [US1] Implement image metadata extractor in packages/backend/internal/parser/media/image.go (EXIF data, thumbnail generation)
- [ ] T090 [P] [US1] Implement video metadata extractor in packages/backend/internal/parser/media/video.go (duration, resolution, thumbnail)
- [ ] T091 [P] [US1] Implement file storage manager in packages/backend/internal/parser/storage/storage.go (SHA-256 content addressing, local filesystem)
- [ ] T092 [US1] Create ContentService orchestration layer in packages/backend/internal/services/content_service.go (coordinates parser, storage, database)

#### API Layer (Desktop REST)

- [ ] T093 [US1] Implement POST /content endpoint in packages/backend/cmd/desktop/handlers/content.go (URL and file upload)
- [ ] T094 [US1] Implement GET /content/{id} endpoint in packages/backend/cmd/desktop/handlers/content.go
- [ ] T095 [US1] Implement PUT /content/{id} endpoint in packages/backend/cmd/desktop/handlers/content.go (title, tags update, version increment)
- [ ] T096 [US1] Implement DELETE /content/{id} endpoint in packages/backend/cmd/desktop/handlers/content.go (soft delete)
- [ ] T097 [US1] Implement GET /content endpoint (list) in packages/backend/cmd/desktop/handlers/content.go (pagination, filters, sort)

#### Tag Management (Go Core + API)

- [ ] T098 [US1] Implement POST /tags endpoint in packages/backend/cmd/desktop/handlers/tags.go
- [ ] T099 [US1] Implement PUT /tags/{id} endpoint in packages/backend/cmd/desktop/handlers/tags.go
- [ ] T100 [US1] Implement DELETE /tags/{id} endpoint in packages/backend/cmd/desktop/handlers/tags.go
- [ ] T101 [US1] Implement GET /tags endpoint in packages/backend/cmd/desktop/handlers/tags.go (list all)

#### Flutter UI (Content Capture)

- [ ] T102 [P] [US1] Create content capture screen in apps/frontend/lib/screens/capture_screen.dart (URL input, file upload)
- [ ] T103 [P] [US1] Create content list widget in apps/frontend/lib/widgets/content_list.dart (thumbnails, titles, tags, timestamps)
- [ ] T104 [P] [US1] Create content detail view in apps/frontend/lib/screens/content_detail_screen.dart (full content, metadata, edit tags)
- [ ] T105 [P] [US1] Create tag picker widget in apps/frontend/lib/widgets/tag_picker.dart (multi-select, tag creation)
- [ ] T106 [P] [US1] Create file upload widget in apps/frontend/lib/widgets/file_upload.dart (drag-drop support)
- [ ] T107 [US1] Create ContentList Riverpod provider in apps/frontend/lib/providers/content_provider.dart (state management)
- [ ] T108 [US1] Create CaptureScreen Riverpod provider in apps/frontend/lib/providers/capture_provider.dart

#### Integration & Error Handling

- [ ] T109 [US1] Add error handling for unreachable URLs in packages/backend/internal/parser/web/scraper.go (timeout, 4xx/5xx errors)
- [ ] T110 [US1] Add error handling for corrupted files in packages/backend/internal/parser/storage/storage.go
- [ ] T111 [US1] Add duplicate content detection in packages/backend/internal/services/content_service.go (SHA-256 hash comparison)
- [ ] T112 [US1] Add logging for content ingestion operations in packages/backend/internal/services/content_service.go (FR-048, FR-049)

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently. Users can capture content from URLs/files, view organized list with thumbnails, and manage tags.

---

## Phase 4: User Story 2 - Instant Offline Search (Priority: P1) 🎯 MVP

**Goal**: Users can search across all captured content instantly, even without internet connection, with relevance-ranked results

**Independent Test**:
1. Capture 100+ diverse items → disconnect from internet → enter search query → relevant results appear in <100ms
2. Search for multi-word phrase → items containing all words appear at top with BM25 relevance ranking
3. Apply filters (media type, date range, tags) → result list updates to show only matching items
4. Click search result → full content detail view opens

### Unit Tests for User Story 2

- [ ] T113 [P] [US2] Unit test for FTS5 query execution in packages/backend/internal/db/search_test.go (BM25 ranking, Unicode handling)
- [ ] T114 [P] [US2] Unit test for search filters (media type, date range, tags) in packages/backend/internal/db/search_test.go

### Implementation for User Story 2

#### Search Engine (Go Core - FTS5)

- [ ] T115 [US2] Implement FTS5 search service in packages/backend/internal/db/search.go (BM25 ranking, Unicode support, <100ms response)
- [ ] T116 [US2] Implement search filter builder in packages/backend/internal/db/filters.go (media type, date range, tag filters)
- [ ] T117 [US2] Implement search term highlighting in packages/backend/internal/db/highlighting.go (extract matched terms from FTS5 results)

#### API Layer (Desktop REST)

- [ ] T118 [US2] Implement GET /search endpoint in packages/backend/cmd/desktop/handlers/search.go (FTS5 query with filters, pagination)
- [ ] T119 [US2] Add query validation (max length, special characters) in packages/backend/cmd/desktop/handlers/search.go

#### Flutter UI (Search)

- [ ] T120 [P] [US2] Create search screen in apps/frontend/lib/screens/search_screen.dart (search input, filters, results list)
- [ ] T121 [P] [US2] Create search result item widget in apps/frontend/lib/widgets/search_result_item.dart (relevance score, highlighted terms)
- [ ] T122 [P] [US2] Create search filter widget in apps/frontend/lib/widgets/search_filters.dart (media type, date range, tags)
- [ ] T123 [US2] Create SearchResults Riverpod provider in apps/frontend/lib/providers/search_provider.dart (debounced queries, filter state)
- [ ] T124 [US2] Add search result navigation to content detail screen in apps/frontend/lib/screens/search_screen.dart

#### Performance Optimization

- [ ] T125 [US2] Enable WAL mode for concurrent read/write in packages/backend/internal/db/db.go (PRAGMA journal_mode=WAL)
- [ ] T126 [US2] Add database indexes for list views in packages/backend/internal/db/migrations/V1__initial_schema.up.sql (created_at DESC, media_type)

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently. Users can capture content and instantly search it offline with sub-100ms response times.

---

## Phase 5: User Story 3 - Intelligent Content Analysis (Priority: P2)

**Goal**: Users can optionally enable AI-powered analysis for summaries, while standard mode uses offline TF-IDF for keyword extraction

**Independent Test**:
1. With AI mode disabled → capture long article → system automatically extracts keywords using TF-IDF and suggests tags
2. With AI mode enabled + valid credentials → request summary → system calls LLM API, generates summary, displays in detail view
3. Without AI credentials → attempt summary → helpful message guides to configure AI or use standard mode
4. View any item → standard keyword suggestions always available, AI options clearly marked as optional

### Unit Tests for User Story 3

- [ ] T127 [P] [US3] Unit test for TF-IDF keyword extraction in packages/backend/internal/analysis/tfidf/tfidf_test.go
- [ ] T128 [P] [US3] Unit test for TextRank algorithm in packages/backend/internal/analysis/textrank/textrank_test.go
- [ ] T129 [P] [US3] Unit test for AI client mocking (OpenAI, Claude, Ollama) in packages/backend/internal/analysis/ai/client_test.go

### Implementation for User Story 3

#### Standard Mode (TF-IDF - Offline)

- [X] T130 [P] [US3] Implement TF-IDF calculator in packages/backend/internal/analysis/tfidf/calculator.go (term frequency, inverse document frequency)
- [ ] T131 [P] [US3] Implement TextRank keyword extraction in packages/backend/internal/analysis/textrank/extractor.go (graph-based ranking)
- [ ] T132 [US3] Create AnalysisService orchestration in packages/backend/internal/services/analysis_service.go (TF-IDF mode by default)

#### AI Mode (Optional - Online)

- [X] T133 [P] [US3] Implement OpenAI client in packages/backend/internal/analysis/ai/openai.go (summary generation)
- [X] T134 [P] [US3] Implement Claude client in packages/backend/internal/analysis/ai/claude.go (summary generation)
- [X] T135 [P] [US3] Implement Ollama client in packages/backend/internal/analysis/ai/ollama.go (local LLM support)
- [X] T136 [US3] Implement graceful degradation in packages/backend/internal/services/analysis_service.go (fallback to TF-IDF on AI failure, non-blocking notification per FR-056)

#### API Layer (Desktop REST)

- [ ] T137 [US3] Implement GET /ai/config endpoint in packages/backend/cmd/desktop/handlers/ai.go (get config, API key redacted)
- [ ] T138 [US3] Implement POST /ai/config endpoint in packages/backend/cmd/desktop/handlers/ai.go (set encrypted credentials)
- [ ] T139 [US3] Implement DELETE /ai/config endpoint in packages/backend/cmd/desktop/handlers/ai.go (disable AI mode)
- [ ] T140 [US3] Implement POST /content/{id}/summary endpoint in packages/backend/cmd/desktop/handlers/content.go (generate summary)

#### Flutter UI (Analysis)

- [ ] T141 [P] [US3] Create AI configuration screen in apps/frontend/lib/screens/ai_config_screen.dart (provider selection, API endpoint, key input)
- [ ] T142 [P] [US3] Create summary view widget in apps/frontend/lib/widgets/summary_view.dart (display AI/TF-IDF summaries)
- [ ] T143 [P] [US3] Create keyword suggestions widget in apps/frontend/lib/widgets/keyword_suggestions.dart (auto-generated tags)
- [ ] T144 [US3] Create AIConfig Riverpod provider in apps/frontend/lib/providers/ai_config_provider.dart

#### WebSocket Events (Analysis Progress)

- [ ] T145 [US3] Implement analysis.started WebSocket event in packages/backend/cmd/desktop/websocket.go (notify when analysis starts)
- [ ] T146 [US3] Implement analysis.completed WebSocket event in packages/backend/cmd/desktop/websocket.go (notify with result)
- [ ] T147 [US3] Implement analysis.failed WebSocket event in packages/backend/cmd/desktop/websocket.go (graceful degradation notification)

**Checkpoint**: At this point, User Stories 1, 2, AND 3 should all work independently. Users have automatic keyword extraction (offline) and optional AI summaries.

---

## Phase 6: User Story 4 - Multi-Device Synchronization (Priority: P2)

**Goal**: Users can optionally enable cloud sync to backup knowledge base and synchronize across multiple devices with incremental updates and conflict resolution

**Independent Test**:
1. Capture content on Device A → enable sync with S3 credentials → system uploads database and media files, shows sync progress, displays "last synced" timestamp
2. Enable sync on Device B with same credentials → content from Device A appears after sync completes
3. Edit same item on two devices while offline → both sync → system uses "last write wins" based on updated_at, logs conflict for awareness
4. Disable sync or disconnect internet → application continues fully functional with local data only

### Unit Tests for User Story 4

- [ ] T148 [P] [US4] Unit test for S3 client upload/download in packages/backend/internal/sync/s3/client_test.go
- [ ] T149 [P] [US4] Unit test for conflict resolution (last write wins) in packages/backend/internal/sync/conflict/resolver_test.go
- [ ] T150 [P] [US4] Unit test for sync queue (exponential backoff) in packages/backend/internal/sync/queue/queue_test.go

### Implementation for User Story 4

#### S3-Compatible Storage (Go Core)

- [X] T151 [P] [US4] Implement S3 client interface in packages/backend/internal/sync/s3/client.go (Upload, Download, Delete, List methods)
- [ ] T152 [P] [US4] Implement AWS S3 provider in packages/backend/internal/sync/s3/aws.go
- [ ] T153 [P] [US4] Implement Cloudflare R2 provider in packages/backend/internal/sync/s3/r2.go
- [ ] T154 [P] [US4] Implement MinIO provider in packages/backend/internal/sync/s3/minio.go (self-hosted S3-compatible)

#### Sync Engine (Go Core)

- [X] T155 [US4] Implement incremental sync logic in packages/backend/internal/sync/engine/sync.go (change_log-based incremental uploads)
- [X] T156 [US4] Implement conflict resolver in packages/backend/internal/sync/conflict/resolver.go (last write wins per updated_at, conflict logging)
- [X] T157 [US4] Implement sync queue manager in packages/backend/internal/sync/queue/queue.go (offline queuing, exponential backoff, retry logic)
- [ ] T158 [US4] Implement SHA-256 content addressing for media files in packages/backend/internal/sync/storage/content_addressed.go (deduplication)

#### API Layer (Desktop REST)

- [ ] T159 [US4] Implement GET /sync/credentials endpoint in packages/backend/cmd/desktop/handlers/sync.go (get config, secrets redacted)
- [ ] T160 [US4] Implement POST /sync/credentials endpoint in packages/backend/cmd/desktop/handlers/sync.go (set encrypted credentials, test S3 connection)
- [ ] T161 [US4] Implement DELETE /sync/credentials endpoint in packages/backend/cmd/desktop/handlers/sync.go (disable sync)
- [ ] T162 [US4] Implement GET /sync/status endpoint in packages/backend/cmd/desktop/handlers/sync.go (current status, last sync, pending changes)
- [ ] T163 [US4] Implement POST /sync/now endpoint in packages/backend/cmd/desktop/handlers/sync.go (trigger immediate sync)

#### WebSocket Events (Sync Progress)

- [ ] T164 [US4] Implement sync.started WebSocket event in packages/backend/cmd/desktop/websocket.go
- [ ] T165 [US4] Implement sync.progress WebSocket event in packages/backend/cmd/desktop/websocket.go (percent, completed, total, current item)
- [ ] T166 [US4] Implement sync.completed WebSocket event in packages/backend/cmd/desktop/websocket.go (uploaded, downloaded, duration)
- [ ] T167 [US4] Implement sync.failed WebSocket event in packages/backend/cmd/desktop/websocket.go (error code, retryable, retry_after)
- [ ] T168 [US4] Implement sync.conflict_detected WebSocket event in packages/backend/cmd/desktop/websocket.go (conflicts array, resolution)

#### Flutter UI (Sync)

- [ ] T169 [P] [US4] Create sync configuration screen in apps/frontend/lib/screens/sync_config_screen.dart (S3 endpoint, bucket, credentials)
- [ ] T170 [P] [US4] Create sync status widget in apps/frontend/lib/widgets/sync_status.dart (status indicator, last synced, pending count)
- [ ] T171 [P] [US4] Create sync progress widget in apps/frontend/lib/widgets/sync_progress.dart (progress bar, current item)
- [ ] T172 [P] [US4] Create conflict log viewer in apps/frontend/lib/screens/conflict_log_screen.dart (list of detected conflicts with resolutions)
- [ ] T173 [US4] Create SyncConfig Riverpod provider in apps/frontend/lib/providers/sync_provider.dart

#### Background Sync (Go Core)

- [ ] T174 [US4] Implement background sync scheduler in packages/backend/internal/sync/scheduler/scheduler.go (periodic sync when online, queue processing when offline)
- [ ] T175 [US4] Add graceful degradation for sync failures in packages/backend/internal/sync/engine/sync.go (non-blocking notifications per FR-057)

**Checkpoint**: At this point, User Stories 1-4 should all work independently. Users can sync knowledge base across devices with conflict-aware incremental updates.

---

## Phase 7: User Story 5 - Data Export & Portability (Priority: P3)

**Goal**: Users can export their entire knowledge base as an encrypted portable archive for backup or migration

**Independent Test**:
1. Initiate export → system prompts for password → creates AES-256 encrypted compressed archive containing database and all media files → displays success with file location
2. Import encrypted archive on fresh installation → provide correct password → all content, tags, metadata, media files are decrypted and restored
3. Configure automatic export schedule → system creates encrypted exports at specified intervals, manages storage by retaining configurable number of backups

### Unit Tests for User Story 5

- [ ] T176 [P] [US5] Unit test for archive encryption (AES-256) in packages/backend/internal/export/crypto/crypto_test.go
- [ ] T177 [P] [US5] Unit test for archive compression (tar.gz) in packages/backend/internal/export/compression/compress_test.go
- [ ] T178 [P] [US5] Unit test for checksum validation (SHA-256) in packages/backend/internal/export/checksum/checksum_test.go

### Implementation for User Story 5

#### Export Engine (Go Core)

- [X] T179 [P] [US5] Implement AES-256 encryption in packages/backend/internal/export/crypto/encrypt.go (password-based, AES-256-GCM)
- [X] T180 [P] [US5] Implement archive compression in packages/backend/internal/export/compression/compress.go (tar.gz format, database + media files)
- [X] T181 [P] [US5] Implement checksum generator in packages/backend/internal/export/checksum/checksum.go (SHA-256 of archive)
- [X] T182 [US5] Create ExportService orchestration in packages/backend/internal/services/export_service.go (coordinate encryption, compression, checksum)

#### Import Engine (Go Core)

- [X] T183 [US5] Implement archive decryption in packages/backend/internal/export/crypto/decrypt.go (password validation, AES-256-GCM decryption)
- [X] T184 [US5] Implement archive extraction in packages/backend/internal/export/compression/decompress.go (tar.gz extraction, validation)
- [X] T185 [US5] Implement data restoration in packages/backend/internal/export/restore.go (database import, media file restoration, conflict handling)

#### API Layer (Desktop REST)

- [ ] T186 [US5] Implement POST /export endpoint in packages/backend/cmd/desktop/handlers/export.go (trigger export with password, include_media option)
- [ ] T187 [US5] Implement POST /import endpoint in packages/backend/cmd/desktop/handlers/export.go (import archive with password, validation, restoration)

#### WebSocket Events (Export Progress)

- [ ] T188 [US5] Implement export.started WebSocket event in packages/backend/cmd/desktop/websocket.go
- [ ] T189 [US5] Implement export.progress WebSocket event in packages/backend/cmd/desktop/websocket.go (stage, percent, current file)
- [ ] T190 [US5] Implement export.completed WebSocket event in packages/backend/cmd/desktop/websocket.go (file path, size, item count, checksum)
- [ ] T191 [US5] Implement export.failed WebSocket event in packages/backend/cmd/desktop/websocket.go (error code, message)
- [ ] T192 [US5] Implement import.started WebSocket event in packages/backend/cmd/desktop/websocket.go
- [ ] T193 [US5] Implement import.completed WebSocket event in packages/backend/cmd/desktop/websocket.go (imported count, skipped count)
- [ ] T194 [US5] Implement import.failed WebSocket event in packages/backend/cmd/desktop/websocket.go (error code: INVALID_PASSWORD, CORRUPTED_ARCHIVE)

#### Flutter UI (Export/Import)

- [ ] T195 [P] [US5] Create export screen in apps/frontend/lib/screens/export_screen.dart (password prompt, include media toggle, export button)
- [ ] T196 [P] [US5] Create import screen in apps/frontend/lib/screens/import_screen.dart (archive selection, password input, import button)
- [ ] T197 [P] [US5] Create export progress widget in apps/frontend/lib/widgets/export_progress.dart (stage, percent, current file)
- [ ] T198 [P] [US5] Create export archive list widget in apps/frontend/lib/widgets/export_archive_list.dart (history, file size, date)
- [ ] T199 [US5] Create automatic export scheduler widget in apps/frontend/lib/widgets/auto_export_config.dart (schedule interval, retention count)
- [ ] T200 [US5] Create Export Riverpod provider in apps/frontend/lib/providers/export_provider.dart

#### Scheduled Exports

- [ ] T201 [US5] Implement export scheduler in packages/backend/internal/export/scheduler/scheduler.go (periodic exports, retention management)
- [ ] T202 [US5] Add export archive management in packages/backend/internal/services/export_service.go (list archives, delete old archives per retention policy)

**Checkpoint**: At this point, ALL user stories (1-5) should be independently functional. Users can export/import encrypted archives for backup and migration.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories (accessibility, error handling, observability, performance)

### Accessibility Compliance (FR-064 to FR-073)

- [ ] T203 [P] Add keyboard navigation support to all interactive elements in apps/frontend/lib/widgets/ (WCAG 2.1 AA)
- [ ] T204 [P] Add screen reader labels (semantics) in apps/frontend/lib/screens/ and apps/frontend/lib/widgets/ (NVDA/JAWS, VoiceOver, TalkBack)
- [ ] T205 [P] Verify color contrast ratios (4.5:1 normal, 3:1 large) in apps/frontend/lib/theme/app_theme.dart
- [ ] T206 [P] Add focus management and visible indicators in apps/frontend/lib/widgets/ (logical tab order)
- [ ] T207 [P] Add text scaling support (up to 200%) in apps/frontend/lib/theme/app_theme.dart
- [ ] T208 [P] Configure keyboard shortcuts documentation in apps/frontend/lib/services/keyboard_shortcuts.dart

### Observability (FR-048 to FR-055)

- [ ] T209 [P] Implement local log file writer in packages/backend/internal/logging/file_logger.go (JSON structured logs, 10MB rotation, max 5 files)
- [ ] T210 [P] Add error logging (timestamp, error code, message) in packages/backend/internal/services/ (all services)
- [ ] T211 [P] Add critical operations logging (export, sync) in packages/backend/internal/services/ (start/complete/success/failure)
- [ ] T212 [P] Add concurrent edit conflict logging in packages/backend/internal/sync/conflict/resolver.go (item UUID, both timestamps)
- [ ] T213 [P] Implement "View Logs" functionality in apps/frontend/lib/screens/logs_screen.dart (open log file location)
- [ ] T214 Verify no sensitive data in logs (passwords, API keys, user content) per FR-050

### Graceful Degradation (FR-056 to FR-063)

- [ ] T215 [P] Implement non-blocking notification system in apps/frontend/lib/widgets/notification_banner.dart (dismissable, retry button for temporary failures)
- [ ] T216 [P] Add AI service failure handling in packages/backend/internal/analysis/ai/ (timeout, rate limit, invalid credentials → fall back to standard mode)
- [ ] T217 [P] Add S3 service failure handling in packages/backend/internal/sync/engine/sync.go (auth error, timeout, quota exceeded → continue offline)
- [ ] T218 [P] Implement sync queue for offline operations in packages/backend/internal/sync/queue/queue.go (queue when network unavailable, process when connection resumes)
- [ ] T219 [P] Add automatic retry with exponential backoff in packages/backend/internal/sync/queue/queue.go (max 3 attempts)

### Performance Optimization (FR-038 to FR-041, SC-005 to SC-007)

- [ ] T220 [P] Optimize content list rendering (virtual scrolling) in apps/frontend/lib/widgets/content_list.dart (render within 500ms for 1,000 items)
- [ ] T221 [P] Implement thumbnail generation background queue in packages/backend/internal/parser/media/thumbnail.go (non-blocking UI)
- [ ] T222 [P] Add database query optimization (indexes, prepared statements) in packages/backend/internal/db/repository.go
- [ ] T223 [P] Implement FTS incremental indexing for large datasets in packages/backend/internal/db/search.go (handle 10K+ items)
- [ ] T224 Verify application launch time (<2 seconds with 10K items) in apps/frontend/lib/main.dart

### Security Hardening (FR-042 to FR-047, SC-011 to SC-015)

- [ ] T225 [P] Implement platform-secure storage for credentials (Keychain on macOS, Credential Manager on Windows, Keystore on Android/iOS) in packages/backend/internal/crypto/secure_storage.go
- [ ] T226 [P] Add API key encryption at rest in packages/backend/internal/models/ai_config.go and packages/backend/internal/models/sync_credential.go (AES-256-GCM)
- [ ] T227 [P] Verify export passwords are NOT stored with archive in packages/backend/internal/export/crypto/encrypt.go
- [ ] T228 [P] Add telemetry/metrics verification (zero external transmission without opt-in) in packages/backend/internal/telemetry/telemetry.go (no-op per FR-053)

### Documentation

- [ ] T229 Update CLAUDE.md with implementation status
- [ ] T230 Verify quickstart.md accuracy (run through setup instructions)
- [ ] T231 Update README.md with usage examples and screenshots

### Final Testing

- [ ] T232 Run full test suite (Go tests, Flutter tests) with scripts/test.sh
- [ ] T233 Verify 80% code coverage requirement (constitution)
- [ ] T234 Run quickstart.md validation (follow guide from scratch)
- [ ] T235 Manual accessibility testing (keyboard navigation, screen reader, color contrast, text scaling)
- [ ] T236 Performance testing (search <100ms for 10K items, launch <2s, list render <500ms)
- [ ] T237 Telemetry verification: verify zero external transmission without opt-in in tests/integration/telemetry_test.go (constitution requirement FR-053, SC-011)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-7)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Phase 8)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - Integrates with US1 (searches content_items created by US1)
- **User Story 3 (P2)**: Can start after Foundational (Phase 2) - Integrates with US1 (analyzes content_items created by US1)
- **User Story 4 (P2)**: Can start after Foundational (Phase 2) - Integrates with US1 (syncs content_items created by US1)
- **User Story 5 (P3)**: Can start after Foundational (Phase 2) - Exports all content from US1-US4

### Within Each User Story

- Unit tests MUST be written before implementation (Test-First per constitution)
- Models before services
- Services before API handlers
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- All unit tests for a user story marked [P] can run in parallel
- Models within a story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members
- All Polish phase accessibility tasks marked [P] can run in parallel

---

## Parallel Example: User Story 1

```bash
# Launch all unit tests for User Story 1 together:
Task T057: "Unit test for ContentItem repository CRUD operations"
Task T058: "Unit test for Tag repository CRUD operations"
Task T059: "Unit test for UUID generation and validation"

# Launch all web parsers in parallel:
Task T060: "Implement web scraper"
Task T061: "Implement PDF parser"
Task T062: "Implement Markdown parser"

# Launch all media parsers in parallel:
Task T063: "Implement image metadata extractor"
Task T064: "Implement video metadata extractor"
```

---

## Implementation Strategy

### MVP First (User Stories 1 + 2 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (Content Capture & Organization)
4. Complete Phase 4: User Story 2 (Instant Offline Search)
5. **STOP and VALIDATE**: Test US1 + US2 independently (capture content, search offline, verify <100ms response)
6. Deploy/demo if ready

**MVP delivers**: A fully functional local-first knowledge base with content capture and instant offline search - the core value proposition.

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP core!)
3. Add User Story 2 → Test independently → Deploy/Demo (MVP complete!)
4. Add User Story 3 → Test independently → Deploy/Demo (Enhanced analysis)
5. Add User Story 4 → Test independently → Deploy/Demo (Multi-device sync)
6. Add User Story 5 → Test independently → Deploy/Demo (Export/import)
7. Add Polish phase → Deploy/Demo (Production-ready)

Each story adds value without breaking previous stories.

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (Content Capture)
   - Developer B: User Story 2 (Search)
   - Developer C: User Story 3 (Analysis)
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify unit tests fail before implementing (Test-First per constitution)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Constitution requires 80% test coverage for Go logic
- All tests use `go test` (Go) and `flutter test` (Flutter)
