# MemoNexus

> **Local-First Personal Knowledge Base** with AI-Powered Analysis and Multi-Device Sync

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.25.5+-00ADD8?logo=go)](https://go.dev/)
[![Flutter Version](https://img.shields.io/badge/Flutter-3.27.0+-02569B?logo=flutter)](https://flutter.dev/)

## Overview

MemoNexus is a privacy-focused, local-first personal knowledge base that helps you capture, organize, and search through your digital content. It features:

- **🌐 Offline-First**: All data stored locally with full-text search (<100ms for 10K items)
- **🤖 AI-Optional**: TF-IDF keyword extraction by default, optional AI summaries (OpenAI/Claude/Ollama)
- **☁️ Optional Sync**: S3-compatible cloud sync with conflict resolution
- **🔒 Privacy First**: AES-256-GCM encryption, platform-native secure storage, zero telemetry
- **📱 Cross-Platform**: Desktop (Win/Mac/Linux), Mobile (Android/iOS)

## Architecture

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
└─────────────────────────────────────────────────────────┘
```

## Quick Start

### Prerequisites

- **Go** 1.25.5+
- **Flutter** 3.27.0+
- **pnpm** 8.0.0+

### Installation

```bash
# Clone repository
git clone https://github.com/kimhsiao/memoNexus.git
cd memoNexus

# Install dependencies
cd apps/frontend && flutter pub get && cd ../..
cd packages/backend && go mod download && cd ../..

# Run tests
pnpm test
```

### Development

**Desktop Development**:
```bash
# Terminal 1: Start embedded PocketBase server
cd packages/backend
go run cmd/desktop/main.go

# Terminal 2: Start Flutter app
cd apps/frontend
flutter run -d macos  # or windows, linux
```

**Mobile Development**:
```bash
# Build Go Core shared library
cd packages/backend
./scripts/build-mobile-lib.sh

# Run Flutter app
cd ../../apps/frontend
flutter run -d emulator  # or connected device
```

For detailed setup instructions, see [Developer Quickstart](specs/001-memo-core/quickstart.md).

## Features

### Content Capture

- **URL Import**: Fetch and extract content from any URL
- **File Import**: Support for PDF, Markdown, images, videos
- **Automatic Metadata**: Title extraction, OpenGraph parsing
- **Tag Organization**: Custom tags with color coding
- **Thumbnail Generation**: Automatic thumbnails for media files

### Instant Search

- **Full-Text Search**: SQLite FTS5 with BM25 ranking
- **Sub-100ms Response**: Search 10K+ items instantly
- **Unicode Support**: Multi-language content handling
- **Advanced Filters**: Media type, date range, tag filtering
- **Term Highlighting**: See matched terms in results

### Intelligent Analysis

- **TF-IDF Keywords**: Automatic keyword extraction (offline)
- **AI Summaries**: Optional summaries via OpenAI/Claude/Ollama
- **TextRank Algorithm**: Graph-based phrase ranking
- **Graceful Degradation**: Falls back to TF-IDF if AI unavailable

### Data Portability

- **Encrypted Export**: AES-256-GCM encrypted archives
- **Import/Restore**: Full data restoration with password
- **Scheduled Exports**: Automatic backup with retention policy
- **Cross-Platform**: Standard tar.gz format

## Project Status

| Phase | Description | Status |
|-------|-------------|--------|
| Phase 1-2 | Setup & Foundational Infrastructure | ✅ Complete |
| Phase 3 | Content Capture & Organization | ✅ Complete |
| Phase 4 | Instant Offline Search | ✅ Complete |
| Phase 5 | Intelligent Content Analysis | ✅ Complete |
| Phase 6 | Multi-Device Synchronization | ✅ Complete |
| Phase 7 | Data Export & Portability | ✅ Complete |
| Phase 8 | Polish & Cross-Cutting Concerns | 🚧 In Progress |

See [Implementation Tasks](specs/001-memo-core/tasks.md) for detailed progress.

## Documentation

- [Quickstart Guide](specs/001-memo-core/quickstart.md) - Developer setup
- [Tasks](specs/001-memo-core/tasks.md) - Implementation task list
- [CLAUDE.md](CLAUDE.md) - Project-specific instructions
- [OpenSpec](openspec/) - Change proposal workflow

## Development

### Project Structure

```
memonexus/
├── apps/
│   ├── frontend/      # Flutter app (multi-platform)
│   └── desktop/       # Desktop wrapper
├── packages/
│   ├── backend/       # Go Core (shared library)
│   └── shared/        # Shared TypeScript types
├── scripts/           # Build and test scripts
├── specs/             # Feature specifications
└── tests/             # Integration tests
```

### Testing

```bash
# Run all tests
pnpm test

# Go tests with coverage
cd packages/backend
go test -cover ./...

# Flutter tests
cd apps/frontend
flutter test --coverage
```

### Building

```bash
# Desktop build
./scripts/build.sh --release

# Mobile build
./scripts/build-mobile-lib.sh --arch arm64
flutter build apk --release
```

## Security & Privacy

- **Zero Telemetry**: No data transmission without explicit opt-in
- **Encryption at Rest**: AES-256-GCM for API keys, export archives
- **Platform Secure Storage**: Keychain (macOS), Credential Manager (Windows)
- **No Phone Home**: All features work offline by default
- **Open Source**: Fully auditable codebase

## Contributing

1. Read the [Constitution](.specify/memory/constitution.md)
2. Check [Open Issues](https://github.com/kimhsiao/memoNexus/issues)
3. Create an [OpenSpec Proposal](openspec/AGENTS.md)
4. Submit Pull Request

## License

MIT License - see [LICENSE](LICENSE) for details.

---

**Built with ❤️ for privacy-focused knowledge management**
