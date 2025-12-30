# Development Guidelines & Technical Standards

## 1. Technical Stack

- **UI Framework**: Flutter (Stable channel)
- **Backend Logic**: Go (Golang 1.21+)
- **Database Engine**: SQLite 3 with FTS5 extension enabled
- **Backend Framework**: PocketBase (extended with custom Go hooks)
- **Inter-op**: Dart FFI for Mobile, REST/WebSockets for Desktop

## 2. System Architecture: Distributed Local-First

The system follows a "Modular Core" design where the business logic is decoupled from the UI.

### 2.1 The Go Core (Shared Logic)

All heavy-lifting tasks (scraping, TF-IDF calculation, sync protocols) must be written in Go.

- **Desktop**: The Go Core runs as an embedded sidecar or part of the PocketBase binary.
- **Mobile**: Compile the Go Core into a C-Archive (`.a` for Android, `.framework` for iOS) using `gomobile`.

### 2.2 Data Schema Standards

- **Primary Keys**: All tables MUST use `UUID v4` to prevent ID collisions during multi-device synchronization.
- **Concurrency**: Use `updated_at` timestamps and a `change_log` table to track incremental mutations for S3 syncing.
- **Soft Deletes**: Use an `is_deleted` flag instead of removing rows to ensure deletion propagates during sync.

## 3. Search Engine Strategy

- **FTS5 Implementation**:
  - Create a virtual table `fts_index` to mirror searchable columns (title, content, tags).
  - Use the $bm25$ ranking function to sort search results by relevance.
- **Offline Indexing**: Indexing must happen immediately upon data ingestion on the local device.

## 4. Synchronization Protocol (S3)

- **Chunking**: Large files (videos/high-res photos) should be uploaded in chunks.
- **Content Addressing**: Files should be named using their SHA-256 hash to prevent duplicate uploads across devices.
- **Conflict Resolution**: Implement a "Last Write Wins" (LWW) policy by default, but log conflicts for manual override if necessary.

## 5. Coding Standards

- **Golang**: Follow `uber-go/guide` for style. Ensure all interfaces are documented.
- **Flutter**: Use `Riverpod` or `Bloc` for state management to ensure a predictable data flow between the UI and the Go Core.
- **Error Handling**: Use a unified error code system that bridges the Go-Dart boundary.

## 6. Directory Structure

```text
/cmd            # Entry points for Desktop/Server
/internal       # Shared Go logic (Parser, Sync, NLP)
/mobile         # Go-mobile wrapper and build scripts
/ui             # Flutter project directory
/pb_data        # PocketBase local storage (ignored by git)
```
