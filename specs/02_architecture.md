## 1. System Overview

MemoNexus is a **Distributed Local-First** personal knowledge management system. The architecture ensures that the "Brain" (logic) and "Memory" (data) reside on the local device, using the cloud solely as a relay for synchronization and backup.

---

## 2. Layered Architecture

### 2.1 Presentation Layer (Flutter)

- **Framework**: Flutter (Multi-platform)
- **Responsibility**: UI/UX, user input handling, and local state management.
- **Communication Bridge**:
- **Desktop (Win/Mac/Linux)**: Communicates with an embedded **PocketBase** instance via REST and WebSockets.
- **Mobile (Android/iOS)**: Communicates with the **Go Core** via **Dart FFI** calling a compiled static library (`.a` or `.framework`).

### 2.2 Logic Core (Go Core)

The Go Core is a platform-agnostic library that handles all business logic:

- **Parser Engine**: Scrapes web content (HTML), extracts metadata (OpenGraph), and parses documents (PDF, Docx).
- **Analysis Engine**:
- **Standard (Offline)**: Uses TF-IDF and TextRank for keyword extraction and tag generation.
- **AI (Optional)**: Provides a gateway to OpenAI/Ollama for summaries and deep insights.

- **Sync Engine**: Manages incremental data synchronization with S3-compatible storage.

### 2.3 Storage Layer (SQLite + FTS5)

- **Relational DB**: Uses SQLite (managed by PocketBase on Desktop, raw SQLite on Mobile).
- **Full-Text Search (FTS5)**: A dedicated virtual table mirrors all text content to enable sub-second offline searches.
- **Blob Storage**: Local file system storage for images/videos, indexed by **SHA-256 Content Addressing**.

---

## 3. Platform Implementation Matrix

| Component        | Desktop (Win/Mac/Linux)  | Mobile (Android/iOS)        |
| ---------------- | ------------------------ | --------------------------- |
| **Backend**      | Embedded PocketBase (Go) | Shared Go Library (via FFI) |
| **Database**     | SQLite (via PocketBase)  | SQLite (via sqflite/ffi)    |
| **Search**       | SQLite FTS5              | SQLite FTS5                 |
| **Connectivity** | Localhost HTTP/WS        | Binary C-Bridge             |

---

## 4. Key Data Flows

### 4.1 Content Ingestion Flow

1. **Input**: User shares a URL or uploads a file.
2. **Analysis**:

- The system checks if **AI Mode** is enabled.
- If **OFF**: Go Core runs TF-IDF to extract tags locally.
- If **ON**: Go Core calls the configured LLM API.

3. **Persistence**: The record is assigned a **UUID v4** and saved to the local SQLite.
4. **Indexing**: The content is immediately pushed to the **FTS5** index for search availability.

### 4.2 Synchronization Flow (S3 Relay)

1. **Mutation**: Any change creates a `change_log` entry with a version timestamp.
2. **Push**: When online, the Sync Engine uploads the `change_log` and new files (named by SHA-256) to S3.
3. **Pull**: Other devices poll S3, download the `change_log`, and reconcile the local database.

---

## 5. Technical Abstractions (Interfaces)

To support the "Optional AI" and "Optional S3" requirements, the Go Core must implement the following interfaces:

```go
// Intelligence Interface
type ContentAnalyzer interface {
    ExtractKeywords(text string) ([]string, error)
    Summarize(text string) (string, error)
}

// Storage Interface
type ObjectStore interface {
    Upload(key string, data []byte) error
    Download(key string) ([]byte, error)
}

```

---

## 6. Conflict & Versioning

- **Conflict Resolution**: **Last Write Wins (LWW)** based on the `updated_at` field.
- **Soft Deletion**: Uses an `is_deleted` boolean flag. Records are only purged after a successful sync across all registered devices.

---

## 7. Security Constraints

- **Zero-Knowledge Sync (Future)**: The Sync Engine should support client-side encryption (AES-256) before uploading data to S3.
- **Local Privacy**: No data is sent to external APIs (OpenAI/S3) unless explicitly configured by the user.

---

### Next Action for Claude Code

> **Would you like me to generate the `internal/db/schema.go` file to define the SQLite tables and FTS5 virtual table as specified in this architecture?**
