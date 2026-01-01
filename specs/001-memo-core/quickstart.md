# Developer Quickstart: MemoNexus Core Platform

**Feature Branch**: `001-memo-core` | **Date**: 2024-12-31

This guide helps you set up your development environment and understand the codebase architecture.

---

## 1. Prerequisites

### 1.1 Required Tools

| Tool | Version | Purpose |
|------|---------|---------|
| **Go** | 1.25.5+ | Backend logic (Go Core) |
| **Flutter** | 3.27.0+ | Multi-platform UI |
| **Dart** | 3.8.0+ | Flutter language SDK |
| **pnpm** | 8.0.0+ | Monorepo package management |
| **SQLite** | 3.x (built-in) | Database (via PocketBase or direct) |
| **Git** | Latest | Version control |

### 1.2 Installation

#### Go (1.25.5)
```bash
# macOS
brew install go

# Linux
wget https://go.dev/dl/go1.25.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.25.5.linux-amd64.tar.gz

# Windows
# Download from https://go.dev/dl/

# Verify
go version
```

#### Flutter (3.27.0)
```bash
# macOS
brew install --cask flutter

# Linux
git clone https://github.com/flutter/flutter.git -b stable
export PATH="$PATH:`pwd`/flutter/bin"

# Windows
# Download from https://flutter.dev/docs/get-started/install

# Verify
flutter --version
flutter doctor  # Fix any reported issues
```

#### pnpm
```bash
npm install -g pnpm

# Verify
pnpm --version
```

---

## 2. Repository Setup

### 2.1 Clone Repository

```bash
git clone https://github.com/kimhsiao/memoNexus.git
cd memoNexus
```

### 2.2 Install Dependencies

```bash
# Install Dart dependencies
cd apps/frontend
flutter pub get
cd ../..

# Install Go dependencies
cd packages/backend
go mod download
go mod verify
cd ../..

# Install desktop wrapper dependencies (if applicable)
cd apps/desktop
pnpm install
cd ../..
```

### 2.3 Verify Installation

```bash
# Run all tests
pnpm test  # Runs Go tests and Flutter tests

# Run linters
pnpm lint
```

---

## 3. Project Structure

```
memonexus/
├── apps/
│   ├── frontend/           # Flutter app (multi-platform)
│   │   ├── lib/
│   │   │   ├── models/     # Dart data models
│   │   │   ├── screens/    # UI screens
│   │   │   ├── widgets/    # Reusable widgets
│   │   │   ├── services/   # API clients, WebSocket
│   │   │   └── main.dart
│   │   ├── test/
│   │   └── pubspec.yaml
│   └── desktop/            # Desktop wrapper (Electron/Tauri)
│       ├── src/
│       ├── package.json
│       └── pnpm-lock.yaml
│
├── packages/
│   ├── backend/            # Go Core (shared library)
│   │   ├── cmd/
│   │   │   ├── core/       # Core library entry point
│   │   │   ├── desktop/    # Desktop embedded server (PocketBase)
│   │   │   ├── mobile/     # Mobile FFI exports
│   │   │   └── migrate/    # Migration tool
│   │   ├── internal/
│   │   │   ├── db/         # Database schema, migrations
│   │   │   ├── models/     # Data models with encryption
│   │   │   ├── parser/     # Web scraping, content extraction
│   │   │   ├── analysis/   # TF-IDF, AI integration
│   │   │   ├── sync/       # S3 sync logic
│   │   │   ├── export/     # Export/import logic
│   │   │   ├── crypto/     # Platform secure storage, encryption
│   │   │   ├── telemetry/  # No-op telemetry (opt-in only)
│   │   │   ├── logging/    # Structured logging
│   │   │   └── uuid/       # UUID utilities
│   │   ├── go.mod
│   │   └── go.sum
│   │
│   └── shared/             # Shared TypeScript types
│       ├── src/
│       │   ├── types.ts    # API contracts
│       │   └── api.ts      # API client (unused on mobile)
│       └── package.json
│
├── scripts/
│   ├── build.sh            # Unified build script
│   ├── build-mobile-lib.sh # Mobile FFI library build
│   ├── dev.sh              # Development environment launcher
│   └── test.sh             # Unified test script
│
├── pnpm-workspace.yaml     # Monorepo configuration
├── go.work                 # Go workspace
├── .clauderules            # AI agent development rules
├── CLAUDE.md               # Project-specific instructions
└── README.md
```

---

## 4. Architecture Overview

### 4.1 Layered Architecture

```
┌─────────────────────────────────────────────────────────┐
│              Presentation Layer (Flutter)               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │   Desktop    │  │   Mobile     │  │  Future:     │ │
│  │ (REST/WS)    │  │  (Dart FFI)  │  │  Web PWA     │ │
│  └──────────────┘  └──────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│               Logic Core (Go Library)                   │
│  ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌────────┐ │
│  │  Parser   │ │ Analysis  │ │   Sync    │ │ Export │ │
│  │  Engine   │ │  Engine   │ │  Engine   │ │        │ │
│  └───────────┘ └───────────┘ └───────────┘ └────────┘ │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│              Storage Layer (SQLite + FTS5)              │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Relational Tables  │  FTS5 Virtual Table       │  │
│  │  (content_items,    │  (content_fts)            │  │
│  │   tags, etc.)       │                           │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### 4.2 Platform Communication

| Platform | Bridge to Go Core | Data Access |
|----------|-------------------|-------------|
| **Desktop (Win/Mac/Linux)** | REST/WebSocket → Embedded PocketBase | SQLite via PocketBase |
| **Mobile (Android/iOS)** | Dart FFI → Go Core Shared Library | SQLite via `sqflite` (mobile) |

### 4.3 Key Interfaces

```go
// Content analysis interface (supports standard + AI modes)
type ContentAnalyzer interface {
    ExtractKeywords(text string) ([]string, error)
    Summarize(text string) (string, error)
    SetConfig(config AIConfig) error
}

// S3-compatible storage interface
type ObjectStore interface {
    Upload(key string, data []byte) error
    Download(key string) ([]byte, error)
    Delete(key string) error
    List(prefix string) ([]string, error)
}
```

---

## 5. Development Workflow

### 5.1 Start Development Environment

**Desktop Development**:
```bash
# Terminal 1: Start embedded PocketBase
cd packages/backend
go run cmd/desktop/main.go

# Terminal 2: Start Flutter app
cd apps/frontend
flutter run -d macos  # or windows, linux
```

**Mobile Development**:
```bash
# Terminal 1: Build Go Core shared library
cd packages/backend
./scripts/build-mobile-lib.sh

# Terminal 2: Start Flutter app
cd apps/frontend
flutter run -d emulator  # or connected device
```

### 5.2 Running Tests

**Go Tests**:
```bash
cd packages/backend

# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/analysis

# Run with race detection
go test -race ./...
```

**Flutter Tests**:
```bash
cd apps/frontend

# Run all tests
flutter test

# Run with coverage
flutter test --coverage

# Run specific test file
flutter test test/models/content_item_test.dart
```

**All Tests (Monorepo)**:
```bash
# From repository root
pnpm test
```

### 5.3 Code Style

**Go**:
- Follow [uber-go/guide](https://github.com/uber-go/guide)
- Use `gofmt` for formatting
- Run `go vet` before committing

```bash
gofmt -w .
go vet ./...
```

**Dart/Flutter**:
- Follow [Effective Dart](https://dart.dev/guides/language/effective-dart)
- Use `flutter analyze` before committing

```bash
flutter analyze
dart format .
```

---

## 6. Common Tasks

### 6.1 Adding a New Database Migration

```bash
cd packages/backend/internal/db/migrations

# Create migration files
touch V2__add_new_feature.up.sql
touch V2__add_new_feature.down.sql
```

**V2__add_new_feature.up.sql**:
```sql
-- Add new column
ALTER TABLE content_items ADD COLUMN new_field TEXT;

-- Update version
UPDATE schema_migrations SET version = 2;
```

**V2__add_new_feature.down.sql**:
```sql
-- Rollback
CREATE TABLE content_items_backup AS SELECT id, title, ... FROM content_items;
DROP TABLE content_items;
ALTER TABLE content_items_backup RENAME TO content_items;

-- Downgrade version
UPDATE schema_migrations SET version = 1;
```

### 6.2 Adding a New Go Service

```bash
cd packages/backend/internal

# Create new package directory
mkdir myservice

# Create service file
cat > myservice/service.go << 'EOF'
package myservice

import "context"

type Service struct {
    db *sql.DB
}

func NewService(db *sql.DB) *Service {
    return &Service{db: db}
}

func (s *Service) DoSomething(ctx context.Context) error {
    // Implementation
    return nil
}
EOF
```

### 6.3 Adding a New Flutter Screen

```bash
cd apps/frontend/lib/screens

# Create new screen file
cat > my_screen.dart << 'EOF'
import 'package:flutter/material.dart';

class MyScreen extends StatelessWidget {
  const MyScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('My Screen')),
      body: Container(),
    );
  }
}
EOF
```

Register in routing:
```dart
// lib/router.dart
import 'screens/my_screen.dart';

// Add route
final routes = {
  '/': (context) => const HomeScreen(),
  '/my-screen': (context) => const MyScreen(),
  // ...
};
```

---

## 7. Environment Configuration

### 7.1 Local Development

**No `.env` file needed** for local development (all defaults to `localhost`).

**Optional Configuration** (`apps/desktop/.env`):
```bash
# PocketBase (embedded)
POCKETBASE_HOST=localhost
POCKETBASE_PORT=8090

# Database
DB_PATH=./data/memonexus.db

# Logging
LOG_LEVEL=debug
LOG_FILE=./logs/memonexus.log
```

### 7.2 AI Configuration (Optional)

AI mode is **disabled by default**. To enable:

1. Via UI:
   - Open Settings → AI Configuration
   - Enter provider (OpenAI/Claude/Ollama), API endpoint, API key

2. Via API (desktop only):
```bash
curl -X POST http://localhost:8090/api/ai/config \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "openai",
    "api_endpoint": "https://api.openai.com/v1",
    "api_key": "sk-...",
    "model_name": "gpt-4"
  }'
```

### 7.3 Sync Configuration (Optional)

Cloud sync is **disabled by default**. To enable:

1. Via UI:
   - Open Settings → Sync Configuration
   - Enter S3-compatible credentials

2. Via API (desktop only):
```bash
curl -X POST http://localhost:8090/api/sync/credentials \
  -H "Content-Type: application/json" \
  -d '{
    "endpoint": "https://s3.amazonaws.com",
    "bucket_name": "my-memonexus",
    "access_key": "AKIA...",
    "secret_key": "..."
  }'
```

---

## 8. Debugging

### 8.1 Go Core Debugging

```bash
cd packages/backend

# Run with Delve debugger
dlv debug cmd/core/main.go

# Run with race detection
go run -race cmd/core/main.go

# Enable verbose logging
LOG_LEVEL=debug go run cmd/core/main.go
```

### 8.2 Flutter Debugging

```bash
cd apps/frontend

# Run with observatory
flutter run --profile

# Debug on device
flutter run -d emulator-5554

# Hot reload (while app is running)
# Press 'r' in terminal to hot reload
# Press 'R' to hot restart
```

### 8.3 Database Inspection

```bash
# Using sqlite3 CLI
sqlite3 data/memonexus.db

# Query content items
SELECT id, title, media_type, created_at FROM content_items WHERE is_deleted = 0;

# Query FTS5 search
SELECT * FROM content_fts WHERE content_fts MATCH 'machine learning' LIMIT 10;

# Check schema version
SELECT * FROM schema_migrations;
```

---

## 9. Building for Production

### 9.1 Desktop Build

```bash
# Build Go Core binary
cd packages/backend
go build -o ../../build/bin/core ./cmd/desktop

# Build Flutter desktop app
cd ../../apps/frontend
flutter build desktop --release

# Output: build/macos/Build/Products/Release/memonexus.app
```

### 9.2 Mobile Build

```bash
# Build Go Core shared library
cd packages/backend
./scripts/build-mobile-lib.sh --arch arm64 --output ../mobile/lib/

# Build Flutter app
cd ../../apps/frontend

# Android
flutter build apk --release

# iOS
flutter build ios --release
```

### 9.3 Unified Build

```bash
# From repository root
./scripts/build.sh --release

# Output in build/dist/
```

---

## 10. Troubleshooting

### Problem: Go modules not found

```bash
# Clean and re-download
cd packages/backend
go clean -modcache
go mod download
```

### Problem: Flutter dependencies conflict

```bash
# Clean and reinstall
cd apps/frontend
flutter clean
flutter pub get
```

### Problem: Database locked

```bash
# Stop all running instances
pkill -f memonexus
pkill -f pocketbase

# Remove lock file
rm -f data/memonexus.db-wal
rm -f data/memonexus.db-shm
```

### Problem: WebSocket connection failed

**Desktop Only**: Ensure PocketBase is running on `localhost:8090`.

```bash
curl http://localhost:8090/api/health
```

**Mobile**: WebSocket is NOT used on mobile. Check FFI bridge instead.

---

## 11. Next Steps

1. **Read the Constitution** (`.specify/memory/constitution.md`) to understand core principles
2. **Review the Data Model** (`data-model.md`) to understand the database schema
3. **Study the API Contracts** (`contracts/openapi.yaml`) to understand the API surface
4. **Explore the Codebase**:
   - Start with `packages/backend/internal/parser` (content ingestion)
   - Then `packages/backend/internal/analysis` (TF-IDF, AI)
   - Finally `apps/frontend/lib/screens` (UI implementation)

---

## 12. Getting Help

- **Architecture Questions**: See `specs/02_architecture.md`
- **Product Questions**: See `specs/01_prd.md`
- **Development Guidelines**: See `specs/03_dev_guidelines.md`
- **Constitution**: See `.specify/memory/constitution.md`
- **Claude Rules**: See `.clauderules`

---

**Happy Coding! 🚀**

---

## Implementation Status (as of 2024-12-31)

| Phase | Description | Status |
|-------|-------------|--------|
| Phase 1-2 | Setup & Foundational Infrastructure | ✅ Complete |
| Phase 3 | User Story 1 - Content Capture & Organization | ✅ Complete |
| Phase 4 | User Story 2 - Instant Offline Search | ✅ Complete |
| Phase 5 | User Story 3 - Intelligent Content Analysis | ✅ Complete |
| Phase 6 | User Story 4 - Multi-Device Synchronization | ✅ Complete |
| Phase 7 | User Story 5 - Data Export & Portability | ✅ Complete |
| Phase 8 | Polish & Cross-Cutting Concerns | 🚧 In Progress |

**Phase 8 Sub-tasks**:
- ✅ Error logging (T210-T214)
- ✅ Graceful degradation (T215-T219)
- ✅ Performance optimization (T220-T223)
- ✅ Security hardening (T225-T228)
- ✅ Application launch time tracking (T224)
- ⏳ Accessibility improvements (T203-T204, T206)
- ✅ Documentation updates (T229-T231)
- ⏳ Final testing (T232-T237)
