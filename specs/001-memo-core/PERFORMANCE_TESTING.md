# Performance Testing Guide (T236)

**Feature**: 001-memo-core  
**Date**: 2025-01-01  
**Constitution Requirements**:
- Search < 100ms for 10K items (FR-039, SC-005)
- Application launch < 2 seconds (FR-038, SC-005)
- List render < 500ms for 1,000 items (FR-040, SC-006)

## Performance Requirements Summary

| Metric | Target | Implementation Status | Location |
|--------|--------|----------------------|----------|
| Search Response | < 100ms (10K items) | ✅ Implemented | `internal/db/search.go:BM25Search` |
| App Launch Time | < 2 seconds | ✅ Tracking Implemented | `apps/frontend/lib/main.dart` |
| List Render | < 500ms (1K items) | ✅ Implemented | `apps/frontend/lib/widgets/content_list.dart` |

---

## 1. Search Performance (<100ms for 10K items)

### 1.1 Backend Implementation

**File**: `packages/backend/internal/db/search.go`

The search uses SQLite FTS5 with BM25 ranking:

```go
// BM25Search performs full-text search with BM25 ranking
// Performance: < 100ms for 10K content items (per constitution SC-005)
func (r *SQLiteRepository) BM25Search(
    ctx context.Context,
    query string,
    limit int,
    offset int,
) ([]*models.ContentItem, int, error)
```

**Optimizations Implemented**:
- FTS5 virtual table with BM25 ranking
- Incremental indexing for large datasets (T223)
- Prepared statements (T222)
- Database indexes on frequently queried columns

### 1.2 Search Performance Test

**File**: `packages/backend/internal/db/search_test.go`

```bash
cd packages/backend

# Run search performance benchmark
go test -bench=BenchmarkBM25Search -benchmem ./internal/db

# Run with CPU profiling
go test -bench=BenchmarkBM25Search -cpuprofile=cpu.prof ./internal/db
go tool pprof cpu.prof

# Run with 10K items
go test -v -run=TestBM25Search_LargeDataset ./internal/db
```

**Expected Results**:
- 10K items: < 100ms average response time
- 1K items: < 20ms average response time
- 100 items: < 5ms average response time

### 1.3 Verification Steps

1. **Setup test database with 10K items**:
```bash
cd packages/backend
go run cmd/migrate/main.go --seed-size=10000
```

2. **Run search queries**:
```bash
# Direct API test
curl -X POST http://localhost:8090/api/search \
  -H "Content-Type: application/json" \
  -d '{"query": "machine learning", "limit": 20}'

# Measure response time
time curl -X POST http://localhost:8090/api/search \
  -H "Content-Type: application/json" \
  -d '{"query": "test query", "limit": 20}'
```

3. **Verify logs show search duration**:
Check logs for `Search completed in Xms` - should be < 100ms

---

## 2. Application Launch Time (< 2 seconds)

### 2.1 Frontend Implementation

**File**: `apps/frontend/lib/main.dart`

Launch time tracking implemented in T224:

```dart
// Performance tracking for T224: Launch time <2s with 10K items
final _appLaunchStartTime = DateTime.now();

void main() {
  // ... initialization code ...
  
  // Log application launch time for T224 verification
  final launchDuration = DateTime.now().difference(_appLaunchStartTime);
  developer.log(
    'Application launched in ${launchDuration.inMilliseconds}ms '
    '(target: <2000ms per constitution)',
    name: 'performance',
  );
}
```

### 2.2 Launch Time Breakdown

| Component | Expected Time | Notes |
|-----------|---------------|-------|
| Flutter framework init | ~500ms | Platform-dependent |
| Provider initialization | ~100ms | Riverpod setup |
| Database initialization | ~300ms | SQLite + FTS5 setup |
| Initial UI render | ~200ms | First frame render |
| **Total Target** | **< 2000ms** | Constitutional requirement |

### 2.3 Verification Steps

1. **Run app in release mode**:
```bash
cd apps/frontend

# macOS
flutter run -d macos --release

# Windows
flutter run -d windows --release

# Linux
flutter run -d linux --release
```

2. **Check logs for launch time**:
```bash
# View Flutter logs
flutter logs

# Look for performance logs
grep "performance" flutter_logs.txt
```

3. **Expected log output**:
```
[performance] Application launched in 1234ms (target: <2000ms per constitution)
```

4. **If launch time exceeds 2s**:
- Check for heavy initialization in `main()`
- Verify lazy loading of providers
- Profile with `flutter run --profile` + DevTools

---

## 3. List Rendering Performance (< 500ms for 1K items)

### 3.1 Frontend Implementation

**File**: `apps/frontend/lib/widgets/content_list.dart`

Virtual scrolling implementation (T220):

```dart
// VirtualListView for efficient rendering of large lists
// Performance: render within 500ms for 1,000 items (per constitution SC-006)
class VirtualListView extends StatefulWidget {
  final List<ContentItem> items;
  final Widget Function(BuildContext, ContentItem) itemBuilder;
}
```

**Optimizations Implemented**:
- Virtual scrolling (only render visible items)
- Item caching to avoid rebuilds
- Lazy loading of thumbnails
- Background thumbnail generation (T221)

### 3.2 List Performance Test

**File**: `apps/frontend/test/widgets/content_list_test.dart`

```bash
cd apps/frontend

# Run performance test
flutter test test/widgets/content_list_performance_test.dart

# Run with timeline
flutter test --timeline --profile test/widgets/content_list_performance_test.dart
```

### 3.3 Verification Steps

1. **Load test data**:
```bash
# Create 1K test items via API
for i in {1..1000}; do
  curl -X POST http://localhost:8090/api/content \
    -H "Content-Type: application/json" \
    -d "{\"title\": \"Test Item $i\", \"content\": \"Test content\"}"
done
```

2. **Navigate to content list in app**:
- Open Flutter app
- Navigate to content list screen
- Use Flutter DevTools to measure frame time

3. **Expected performance**:
- First frame render: < 500ms for 1K items
- Scroll FPS: > 55 FPS (smooth scrolling)
- Frame build time: < 16ms (60 FPS)

4. **Profile with DevTools**:
```bash
flutter run --profile
# Open DevTools: http://localhost:9100
# Check Performance tab > Frame rendering
```

---

## 4. Backend Performance Benchmarks

### 4.1 Run All Benchmarks

```bash
cd packages/backend

# Run all benchmarks
go test -bench=. -benchmem ./...

# Run specific package benchmarks
go test -bench=. -benchmem ./internal/db
go test -bench=. -benchmem ./internal/analysis
go test -bench=. -benchmem ./internal/parser

# With race detection
go test -race -bench=. -benchmem ./internal/...
```

### 4.2 Key Benchmarks

| Benchmark | Target | Implementation |
|-----------|--------|----------------|
| `BenchmarkBM25Search` | < 100ms | `internal/db/search.go` |
| `BenchmarkExtractKeywords` | < 50ms (1KB text) | `internal/analysis/tfidf.go` |
| `BenchmarkParseWebContent` | < 500ms | `internal/parser/web_extractor.go` |
| `BenchmarkGenerateThumbnail` | Non-blocking | `internal/parser/media/thumbnail.go` |

---

## 5. Memory Performance

### 5.1 Memory Profiling

**Go Backend**:
```bash
cd packages/backend

# Run with memory profiling
go test -bench=. -memprofile=mem.prof ./internal/db

# Analyze memory profile
go tool pprof mem.prof
# Interactive commands:
#   top10    - Show top 10 memory allocations
#   list fn  - Show source code for function
#   pdf      - Generate PDF visualization
```

**Flutter Frontend**:
```bash
cd apps/frontend

# Run with memory tracking
flutter run --profile

# In DevTools, check Memory tab
# Look for memory leaks and usage patterns
```

### 5.2 Memory Targets

| Component | Target | Notes |
|-----------|--------|-------|
| Backend (idle) | < 100MB | PocketBase + SQLite |
| Backend (peak) | < 500MB | During large operations |
| Frontend (idle) | < 100MB | Flutter app memory |
| Frontend (peak) | < 300MB | During large list render |

---

## 6. Performance Testing Scenarios

### Scenario 1: Cold Start Performance
1. Stop all app instances
2. Clear application cache
3. Launch app
4. Measure time to interactive (TTI)
5. **Target**: < 2 seconds

### Scenario 2: Large Dataset Search
1. Seed database with 10K items
2. Perform search query
3. Measure response time
4. **Target**: < 100ms

### Scenario 3: Large List Rendering
1. Load 1K items in list
2. Measure first frame render time
3. Scroll through entire list
4. **Target**: < 500ms initial, > 55 FPS scroll

### Scenario 4: Background Operations
1. Start file upload (100MB)
2. Navigate to other screens
3. Verify UI remains responsive
4. **Target**: No blocking operations

### Scenario 5: Concurrent Operations
1. Perform search while syncing
2. Navigate while exporting
3. **Target**: No UI freezes or crashes

---

## 7. Performance Monitoring in Production

### 7.1 Launch Time Monitoring

Launch time is automatically logged (T224):

```dart
developer.log(
  'Application launched in ${launchDuration.inMilliseconds}ms '
  '(target: <2000ms per constitution)',
  name: 'performance',
);
```

### 7.2 Search Performance Monitoring

Search duration is logged in `search.go`:

```go
logger.Info("search completed",
    "duration_ms", time.Since(start).Milliseconds(),
    "result_count", len(results),
)
```

### 7.3 View Performance Logs

```bash
# Flutter logs
flutter logs | grep performance

# Backend logs
tail -f logs/memonexus.log | grep -E "(duration|performance|ms)"
```

---

## 8. Performance Optimization Checklist

### Backend (Go)
- [x] FTS5 full-text search with BM25 ranking
- [x] Database indexes on frequently queried columns
- [x] Prepared statements for query optimization (T222)
- [x] Incremental FTS indexing for large datasets (T223)
- [x] Background thumbnail generation (T221)
- [x] Connection pooling for database

### Frontend (Flutter)
- [x] ProviderScope with performance observer (T224)
- [x] Virtual scrolling for large lists (T220)
- [x] Lazy loading of thumbnails
- [x] Const widgets for immutable UI elements
- [x] Item caching in lists
- [x] Performance overlay in debug mode

---

## 9. Troubleshooting Performance Issues

### Issue: Search > 100ms

**Possible Causes**:
1. FTS5 index not built
2. Missing database indexes
3. Large result set without limit

**Solutions**:
```sql
-- Verify FTS5 index exists
SELECT * FROM sqlite_master WHERE type='table' AND name='content_fts';

-- Rebuild FTS5 index
INSERT INTO content_fts(content_fts) VALUES('rebuild');

-- Check indexes
PRAGMA index_list('content_items');
```

### Issue: Launch > 2 seconds

**Possible Causes**:
1. Heavy initialization in `main()`
2. Synchronous provider initialization
3. Large database migration on startup

**Solutions**:
- Move heavy initialization to background
- Use async provider loading
- Run migrations in background thread

### Issue: List Render > 500ms

**Possible Causes**:
1. Not using virtual scrolling
2. Building all widgets upfront
3. Heavy image processing on main thread

**Solutions**:
- Enable virtual scrolling (already implemented)
- Use `const` constructors
- Move thumbnail generation to background (already implemented)

---

## 10. Performance Test Results Template

| Test | Target | Actual | Status | Date |
|------|--------|--------|--------|------|
| Search (10K items) | < 100ms | ______ms | ☐ Pass ☐ Fail | |
| App Launch (cold) | < 2000ms | ______ms | ☐ Pass ☐ Fail | |
| List Render (1K items) | < 500ms | ______ms | ☐ Pass ☐ Fail | |
| Memory (idle) | < 100MB | ______MB | ☐ Pass ☐ Fail | |
| Memory (peak) | < 500MB | ______MB | ☐ Pass ☐ Fail | |

---

## 11. Continuous Performance Monitoring

### Automated Benchmarks

Add to CI/CD pipeline:

```yaml
# .github/workflows/performance.yml
name: Performance Tests

on: [push, pull_request]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run Go benchmarks
        run: |
          cd packages/backend
          go test -bench=. -benchmem ./... | tee benchmark.txt
      - name: Run Flutter performance tests
        run: |
          cd apps/frontend
          flutter test --timeline --profile
      - name: Upload results
        uses: actions/upload-artifact@v3
        with:
          name: performance-results
          path: benchmark.txt
```

---

## 12. References

- [Flutter Performance Best Practices](https://docs.flutter.dev/perf)
- [Go Performance Patterns](https://github.com/dgryski/go-perfbook)
- [SQLite FTS5 Documentation](https://www.sqlite.org/fts5.html)
- [WCAG Performance Guidelines](https://www.w3.org/WAI/WCAG21/quickref/)
