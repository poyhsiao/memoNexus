# Product Requirement Document: MemoNexus

**Project Code Name:** MemoNexus
**Status:** Initial Draft
**Core Concept:** A local-first, privacy-centric personal knowledge base and media vault that works offline, searches instantly, and syncs optionally to the cloud.

---

## 1. Executive Summary

MemoNexus is a cross-platform application designed for users to capture, organize, and retrieve meaningful digital content (websites, documents, photos, and videos). Unlike traditional cloud-only services, MemoNexus prioritizes local execution and storage, using AI only as an optional enhancement.

## 2. Core Features

### 2.1 Content Ingestion & Parsing

- **Multi-Source Support**: Accept URLs, Images, Videos, PDFs, and Markdown files.
- **Web Scraping**: Automatically extract clean text, titles, and metadata from URLs, bypassing ads and clutter.
- **File Processing**: Generate thumbnails for media and perform text extraction (OCR/PDF-to-text) for documents.

### 2.2 Intelligent Analysis (Hybrid Approach)

- **Standard Mode (Default/Free)**:
  - Perform offline keyword extraction using $TF-IDF$ (Term Frequency-Inverse Document Frequency) algorithms.
  - Automatic metadata parsing (EXIF for photos, OpenGraph for web).
- **AI Mode (Optional)**:
  - Integration with LLMs (OpenAI, Claude, or local Ollama) for summary generation and "Deep Insights."
  - Suggested categorization based on content semantics.

### 2.3 Powerful Search

- **Full-Text Search (FTS)**: Native SQLite FTS5 support for lightning-fast searches across all indexed text.
- **Offline Retrieval**: Searching must be functional without any internet connection on all devices.
- **Advanced Filtering**: Filter by tags, date ranges, media types, and source domains.

### 2.4 Synchronization & Storage

- **Local-First Storage**: Primary data resides in a local SQLite database.
- **S3 Sync (Optional)**: Support for S3-compatible storage (AWS S3, R2, MinIO) for encrypted backups and multi-device syncing.
- **On-Demand Export**: Option to bundle the database and assets into a compressed archive for manual migration.

## 3. Platform Support

- **Desktop**: Windows, macOS, Linux (via Flutter + Embedded PocketBase/Go).
- **Mobile**: Android, iOS (via Flutter + Go-C-Archive).
- **Web**: Progressive Web App (PWA) connecting to a self-hosted PocketBase instance.

## 4. Non-Functional Requirements

- **Privacy**: No data is sent to third-party servers unless AI or S3 sync is explicitly enabled.
- **Performance**: Search results must return in < 100ms for up to 10,000 entries.
- **Consistency**: Unified UI/UX across mobile and desktop.
