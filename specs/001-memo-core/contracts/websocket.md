# WebSocket Protocol: MemoNexus Real-time Events

**Version**: 1.0.0 | **Date**: 2024-12-30

This document defines the WebSocket protocol for real-time communication between the desktop Flutter UI and the embedded PocketBase instance.

---

## 1. Connection

### 1.1 Endpoint

```
ws://localhost:8090/api/realtime
```

**Note**: Mobile platforms use Dart FFI instead of WebSocket. This protocol is desktop-only.

### 1.2 Authentication

WebSocket connection includes authentication token in the initial handshake:

```javascript
const ws = new WebSocket('ws://localhost:8090/api/realtime?token=X-Local-Token');
```

---

## 2. Message Format

All messages follow this JSON structure:

```json
{
  "type": "event_type",
  "data": { /* event-specific payload */ },
  "timestamp": 1735584000
}
```

**Fields**:
- `type` (string): Event type identifier
- `data` (object): Event-specific payload
- `timestamp` (integer): Unix timestamp

---

## 3. Client → Server Messages

### 3.1 Subscribe to Events

Subscribe to specific event types.

```json
{
  "action": "subscribe",
  "events": ["sync", "analysis", "conflict"]
}
```

**Response**:
```json
{
  "action": "subscribe_ack",
  "subscribed": ["sync", "analysis", "conflict"],
  "timestamp": 1735584000
}
```

### 3.2 Unsubscribe from Events

```json
{
  "action": "unsubscribe",
  "events": ["analysis"]
}
```

### 3.3 Ping (Keep-Alive)

Client sends periodic ping to detect connection health.

```json
{
  "action": "ping",
  "timestamp": 1735584000
}
```

**Server Response**:
```json
{
  "action": "pong",
  "timestamp": 1735584000
}
```

**Frequency**: Client should ping every 30 seconds.

---

## 4. Server → Client Events

### 4.1 Sync Events

#### `sync.started`

Sync operation has started.

```json
{
  "type": "sync.started",
  "data": {
    "sync_id": "550e8400-e29b-41d4-a716-446655440000",
    "direction": "upload" // "upload" or "download"
  },
  "timestamp": 1735584000
}
```

#### `sync.progress`

Sync progress update (sent every 5% or 100 items).

```json
{
  "type": "sync.progress",
  "data": {
    "sync_id": "550e8400-e29b-41d4-a716-446655440000",
    "percent": 45, // 0-100
    "completed": 450,
    "total": 1000,
    "current_item": "Uploading: IMG_20241230.jpg"
  },
  "timestamp": 1735584000
}
```

#### `sync.completed`

Sync operation completed successfully.

```json
{
  "type": "sync.completed",
  "data": {
    "sync_id": "550e8400-e29b-41d4-a716-446655440000",
    "uploaded": 500,
    "downloaded": 20,
    "duration_seconds": 120
  },
  "timestamp": 1735584000
}
```

#### `sync.failed`

Sync operation failed.

```json
{
  "type": "sync.failed",
  "data": {
    "sync_id": "550e8400-e29b-41d4-a716-446655440000",
    "error_code": "S3_AUTH_FAILED",
    "error_message": "Authentication failed: Invalid access key",
    "retryable": false,
    "retry_after": null // ISO 8601 duration if retryable (e.g., "PT5M")
  },
  "timestamp": 1735584000
}
```

#### `sync.conflict_detected`

One or more conflicts were detected during sync.

```json
{
  "type": "sync.conflict_detected",
  "data": {
    "sync_id": "550e8400-e29b-41d4-a716-446655440000",
    "conflicts": [
      {
        "item_id": "660e8400-e29b-41d4-a716-446655440000",
        "title": "Interesting Article",
        "local_timestamp": 1735584000,
        "remote_timestamp": 1735584050,
        "resolution": "last_write_wins"
      }
    ],
    "total_conflicts": 1
  },
  "timestamp": 1735584000
}
```

---

### 4.2 Analysis Events

#### `analysis.started`

Content analysis (keyword extraction or AI summary) has started.

```json
{
  "type": "analysis.started",
  "data": {
    "item_id": "550e8400-e29b-41d4-a716-446655440000",
    "mode": "ai" // "standard" (TF-IDF) or "ai"
  },
  "timestamp": 1735584000
}
```

#### `analysis.completed`

Analysis completed successfully.

```json
{
  "type": "analysis.completed",
  "data": {
    "item_id": "550e8400-e29b-41d4-a716-446655440000",
    "mode": "ai",
    "result": {
      "keywords": ["machine learning", "AI", "neural networks"],
      "summary": "This article discusses recent advances in deep learning..."
    }
  },
  "timestamp": 1735584000
}
```

#### `analysis.failed`

Analysis failed (graceful degradation).

```json
{
  "type": "analysis.failed",
  "data": {
    "item_id": "550e8400-e29b-41d4-a716-446655440000",
    "attempted_mode": "ai",
    "fallback_mode": "standard",
    "error_code": "AI_RATE_LIMITED",
    "error_message": "OpenAI API rate limit exceeded. Falling back to TF-IDF.",
    "fallback_result": {
      "keywords": ["machine learning", "AI", "neural networks"]
    }
  },
  "timestamp": 1735584000
}
```

---

### 4.3 Content Events

#### `content.created`

New content item was created (by another window/desktop app instance).

```json
{
  "type": "content.created",
  "data": {
    "item": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "title": "Interesting Article",
      "media_type": "web",
      "created_at": 1735584000
    }
  },
  "timestamp": 1735584000
}
```

#### `content.updated`

Content item was updated (concurrent edit from another window).

```json
{
  "type": "content.updated",
  "data": {
    "item": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "title": "Updated Article Title",
      "version": 3,
      "updated_at": 1735584050
    },
    "your_version": 2,
    "conflict_detected": true
  },
  "timestamp": 1735584000
}
```

#### `content.deleted`

Content item was soft-deleted.

```json
{
  "type": "content.deleted",
  "data": {
    "item_id": "550e8400-e29b-41d4-a716-446655440000",
    "deleted_at": 1735584000
  },
  "timestamp": 1735584000
}
```

---

### 4.4 Export Events

#### `export.started`

Export operation started.

```json
{
  "type": "export.started",
  "data": {
    "export_id": "550e8400-e29b-41d4-a716-446655440000",
    "include_media": true
  },
  "timestamp": 1735584000
}
```

#### `export.progress`

Export progress update.

```json
{
  "type": "export.progress",
  "data": {
    "export_id": "550e8400-e29b-41d4-a716-446655440000",
    "stage": "compressing", // "collecting", "compressing", "encrypting"
    "percent": 60,
    "current_file": "IMG_20241230.jpg"
  },
  "timestamp": 1735584000
}
```

#### `export.completed`

Export completed successfully.

```json
{
  "type": "export.completed",
  "data": {
    "export_id": "550e8400-e29b-41d4-a716-446655440000",
    "file_path": "/exports/memonexus_20241230.tar.gz",
    "size_bytes": 104857600,
    "item_count": 1234,
    "checksum": "a1b2c3d4..." // SHA-256
  },
  "timestamp": 1735584000
}
```

#### `export.failed`

Export failed.

```json
{
  "type": "export.failed",
  "data": {
    "export_id": "550e8400-e29b-41d4-a716-446655440000",
    "error_code": "DISK_FULL",
    "error_message": "Insufficient disk space. Required 500MB, available 100MB."
  },
  "timestamp": 1735584000
}
```

---

### 4.5 Import Events

#### `import.started`

Import operation started.

```json
{
  "type": "import.started",
  "data": {
    "import_id": "550e8400-e29b-41d4-a716-446655440000",
    "archive_path": "/exports/memonexus_20241230.tar.gz"
  },
  "timestamp": 1735584000
}
```

#### `import.completed`

Import completed successfully.

```json
{
  "type": "import.completed",
  "data": {
    "import_id": "550e8400-e29b-41d4-a716-446655440000",
    "imported_count": 1234,
    "skipped_count": 5,
    "duration_seconds": 45
  },
  "timestamp": 1735584000
}
```

#### `import.failed`

Import failed.

```json
{
  "type": "import.failed",
  "data": {
    "import_id": "550e8400-e29b-41d4-a716-446655440000",
    "error_code": "INVALID_PASSWORD",
    "error_message": "Failed to decrypt archive. Password may be incorrect."
  },
  "timestamp": 1735584000
}
```

---

## 5. Error Codes

| Error Code | Category | Retryable | Description |
|------------|----------|-----------|-------------|
| `S3_AUTH_FAILED` | Sync | No | Invalid S3 credentials |
| `S3_TIMEOUT` | Sync | Yes | S3 request timeout |
| `S3_QUOTA_EXCEEDED` | Sync | No | S3 storage quota exceeded |
| `NETWORK_UNAVAILABLE` | Sync | Yes | Network connection failed |
| `AI_RATE_LIMITED` | Analysis | Yes | AI API rate limit exceeded |
| `AI_TIMEOUT` | Analysis | Yes | AI API request timeout |
| `AI_INVALID_CREDENTIALS` | Analysis | No | Invalid AI API key |
| `DISK_FULL` | Export | No | Insufficient disk space |
| `INVALID_PASSWORD` | Import | No | Wrong export password |
| `CORRUPTED_ARCHIVE` | Import | No | Archive checksum failed |

**Retry Strategy**:
- **Retryable errors**: Exponential backoff with max 3 retries
- **Non-retryable errors**: Display error to user with remediation steps

---

## 6. Connection Lifecycle

### 6.1 Connection Establishment

```
Client                  Server
  |                       |
  |----- WebSocket Open -->|
  |                       |
  |<-- Hello (server_id) -|
  |                       |
  |----- Subscribe ------>|
  |    (events: [])       |
  |                       |
  |<-- Subscribe ACK -----|
  |                       |
```

### 6.2 Keep-Alive

```
Client                  Server
  |                       |
  |----- Ping ----------->|
  |                       |
  |<-- Pong ------------- |
  |                       |
  |       [30s later]     |
  |----- Ping ----------->|
  |                       |
```

**Timeout**: If no pong received within 10 seconds, client should reconnect.

### 6.3 Connection Close

```
Client                  Server
  |                       |
  |----- Close (normal) -->|
  |                       |
  |<-- Close (normal) ----|
  |                       |
```

**Close Codes**:
- `1000`: Normal closure
- `1001`: Endpoint going away
- `1006`: Abnormal closure (network error)
- `4000`: Custom: Authentication failed
- `4001`: Custom: Server shutdown

---

## 7. Flutter Integration

### 7.1 WebSocket Client

```dart
import 'dart:async';
import 'dart:convert';
import 'package:web_socket_channel/web_socket_channel.dart';

class MemoNexusWebSocket {
  final String _url = 'ws://localhost:8090/api/realtime';
  final String _token;
  late WebSocketChannel _channel;
  final StreamController<Map<String, dynamic>> _eventController = StreamController.broadcast();
  Timer? _pingTimer;

  MemoNexusWebSocket(this._token);

  Stream<Map<String, dynamic>> get events => _eventController.stream;

  Future<void> connect() async {
    _channel = WebSocketChannel.connect(
      Uri.parse('$_url?token=$_token'),
    );

    // Listen for messages
    _channel.stream.listen(
      (message) {
        final data = jsonDecode(message as String) as Map<String, dynamic>;
        _eventController.add(data);
      },
      onError: (error) {
        print('WebSocket error: $error');
      },
      onDone: () {
        print('WebSocket connection closed');
        _pingTimer?.cancel();
      },
    );

    // Start keep-alive ping
    _pingTimer = Timer.periodic(const Duration(seconds: 30), (_) {
      sendPing();
    });
  }

  void subscribe(List<String> events) {
    _channel.sink.add(jsonEncode({
      'action': 'subscribe',
      'events': events,
    }));
  }

  void sendPing() {
    _channel.sink.add(jsonEncode({
      'action': 'ping',
      'timestamp': DateTime.now().millisecondsSinceEpoch ~/ 1000,
    }));
  }

  void disconnect() {
    _pingTimer?.cancel();
    _channel.sink.close();
  }
}
```

### 7.2 Event Handling

```dart
class SyncStatusWidget extends StatefulWidget {
  @override
  _SyncStatusWidgetState createState() => _SyncStatusWidgetState();
}

class _SyncStatusWidgetState extends State<SyncStatusWidget> {
  final _ws = MemoNexusWebSocket('local-token');
  String _syncStatus = 'idle';
  int _syncProgress = 0;

  @override
  void initState() {
    super.initState();
    _ws.connect();
    _ws.subscribe(['sync']);

    _ws.events.listen((event) {
      switch (event['type']) {
        case 'sync.started':
          setState(() => _syncStatus = 'syncing');
          break;
        case 'sync.progress':
          setState(() => _syncProgress = event['data']['percent'] as int);
          break;
        case 'sync.completed':
          setState(() {
            _syncStatus = 'idle';
            _syncProgress = 0;
          });
          break;
        case 'sync.failed':
          setState(() {
            _syncStatus = 'failed';
            _syncProgress = 0;
          });
          _showError(event['data']['error_message'] as String);
          break;
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    // ... build UI based on _syncStatus and _syncProgress
    return Container();
  }

  @override
  void dispose() {
    _ws.disconnect();
    super.dispose();
  }
}
```

---

## 8. Security Considerations

1. **Local-Only**: WebSocket server binds to `localhost` only (not exposed to network)
2. **Token Validation**: Each connection must include valid `X-Local-Token`
3. **No Sensitive Data**: WebSocket messages never include:
   - API keys (AI or S3)
   - Export passwords
   - Full content (only IDs and metadata)
4. **Rate Limiting**: Client should limit WebSocket message frequency to avoid overwhelming the server

---

**Next**: See [quickstart.md](../quickstart.md) for developer onboarding guide.
