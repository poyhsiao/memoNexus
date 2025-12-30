// WebSocket client for real-time events (desktop only)
// Mobile platforms use polling via FFI bridge instead

import 'dart:async';
import 'dart:convert';
import 'package:web_socket_channel/web_socket_channel.dart';

class MemoNexusWebSocket {
  final String url;
  final String token;
  WebSocketChannel? _channel;
  final StreamController<Map<String, dynamic>> _eventController =
      StreamController.broadcast();
  Timer? _pingTimer;
  bool _isConnected = false;

  MemoNexusWebSocket({
    this.url = 'ws://localhost:8090/api/realtime',
    this.token = 'X-Local-Token',
  });

  Stream<Map<String, dynamic>> get events => _eventController.stream;
  bool get isConnected => _isConnected;

  Future<void> connect() async {
    if (_isConnected) return;

    final uri = Uri.parse('$url?token=$token');
    _channel = WebSocketChannel.connect(uri);

    // Listen for messages
    _channel!.stream.listen(
      (message) {
        final data = jsonDecode(message as String) as Map<String, dynamic>;
        _handleMessage(data);
      },
      onError: (error) {
        _isConnected = false;
        _eventController.addError(error);
      },
      onDone: () {
        _isConnected = false;
        _pingTimer?.cancel();
      },
    );

    _isConnected = true;

    // Start keep-alive ping
    _pingTimer = Timer.periodic(const Duration(seconds: 30), (_) {
      sendPing();
    });
  }

  void subscribe(List<String> events) {
    if (!_isConnected) {
      throw StateError('WebSocket is not connected');
    }

    _channel!.sink.add(jsonEncode({
      'action': 'subscribe',
      'events': events,
    }));
  }

  void unsubscribe(List<String> events) {
    if (!_isConnected) {
      throw StateError('WebSocket is not connected');
    }

    _channel!.sink.add(jsonEncode({
      'action': 'unsubscribe',
      'events': events,
    }));
  }

  void sendPing() {
    if (!_isConnected) return;

    _channel!.sink.add(jsonEncode({
      'action': 'ping',
      'timestamp': DateTime.now().millisecondsSinceEpoch ~/ 1000,
    }));
  }

  void disconnect() {
    _pingTimer?.cancel();
    _channel?.sink.close();
    _isConnected = false;
  }

  void _handleMessage(Map<String, dynamic> data) {
    // Handle action responses
    final action = data['action'] as String?;
    if (action != null) {
      if (action == 'pong') {
        // Pong received, connection is alive
        return;
      }
      if (action == 'subscribe_ack') {
        // Subscription acknowledged
        return;
      }
    }

    // Emit event to stream
    _eventController.add(data);
  }
}

// =====================================================
// Event Models
// =====================================================

// Sync Events
class SyncStartedEvent {
  final String syncId;
  final String direction;

  SyncStartedEvent({
    required this.syncId,
    required this.direction,
  });

  factory SyncStartedEvent.fromJson(Map<String, dynamic> json) {
    final data = json['data'] as Map<String, dynamic>;
    return SyncStartedEvent(
      syncId: data['sync_id'] as String,
      direction: data['direction'] as String,
    );
  }
}

class SyncProgressEvent {
  final String syncId;
  final int percent;
  final int completed;
  final int total;
  final String currentItem;

  SyncProgressEvent({
    required this.syncId,
    required this.percent,
    required this.completed,
    required this.total,
    required this.currentItem,
  });

  factory SyncProgressEvent.fromJson(Map<String, dynamic> json) {
    final data = json['data'] as Map<String, dynamic>;
    return SyncProgressEvent(
      syncId: data['sync_id'] as String,
      percent: data['percent'] as int,
      completed: data['completed'] as int,
      total: data['total'] as int,
      currentItem: data['current_item'] as String,
    );
  }
}

class SyncCompletedEvent {
  final String syncId;
  final int uploaded;
  final int downloaded;
  final int durationSeconds;

  SyncCompletedEvent({
    required this.syncId,
    required this.uploaded,
    required this.downloaded,
    required this.durationSeconds,
  });

  factory SyncCompletedEvent.fromJson(Map<String, dynamic> json) {
    final data = json['data'] as Map<String, dynamic>;
    return SyncCompletedEvent(
      syncId: data['sync_id'] as String,
      uploaded: data['uploaded'] as int,
      downloaded: data['downloaded'] as int,
      durationSeconds: data['duration_seconds'] as int,
    );
  }
}

class SyncFailedEvent {
  final String syncId;
  final String errorCode;
  final String errorMessage;
  final bool retryable;

  SyncFailedEvent({
    required this.syncId,
    required this.errorCode,
    required this.errorMessage,
    required this.retryable,
  });

  factory SyncFailedEvent.fromJson(Map<String, dynamic> json) {
    final data = json['data'] as Map<String, dynamic>;
    return SyncFailedEvent(
      syncId: data['sync_id'] as String,
      errorCode: data['error_code'] as String,
      errorMessage: data['error_message'] as String,
      retryable: data['retryable'] as bool? ?? false,
    );
  }
}

// Analysis Events
class AnalysisStartedEvent {
  final String itemId;
  final String mode;

  AnalysisStartedEvent({
    required this.itemId,
    required this.mode,
  });

  factory AnalysisStartedEvent.fromJson(Map<String, dynamic> json) {
    final data = json['data'] as Map<String, dynamic>;
    return AnalysisStartedEvent(
      itemId: data['item_id'] as String,
      mode: data['mode'] as String,
    );
  }
}

class AnalysisCompletedEvent {
  final String itemId;
  final String mode;
  final Map<String, dynamic> result;

  AnalysisCompletedEvent({
    required this.itemId,
    required this.mode,
    required this.result,
  });

  factory AnalysisCompletedEvent.fromJson(Map<String, dynamic> json) {
    final data = json['data'] as Map<String, dynamic>;
    return AnalysisCompletedEvent(
      itemId: data['item_id'] as String,
      mode: data['mode'] as String,
      result: data['result'] as Map<String, dynamic>,
    );
  }
}

class AnalysisFailedEvent {
  final String itemId;
  final String attemptedMode;
  final String fallbackMode;
  final String errorCode;
  final String errorMessage;
  final Map<String, dynamic>? fallbackResult;

  AnalysisFailedEvent({
    required this.itemId,
    required this.attemptedMode,
    required this.fallbackMode,
    required this.errorCode,
    required this.errorMessage,
    this.fallbackResult,
  });

  factory AnalysisFailedEvent.fromJson(Map<String, dynamic> json) {
    final data = json['data'] as Map<String, dynamic>;
    return AnalysisFailedEvent(
      itemId: data['item_id'] as String,
      attemptedMode: data['attempted_mode'] as String,
      fallbackMode: data['fallback_mode'] as String,
      errorCode: data['error_code'] as String,
      errorMessage: data['error_message'] as String,
      fallbackResult: data['fallback_result'] as Map<String, dynamic>?,
    );
  }
}

// Content Events
class ContentCreatedEvent {
  final Map<String, dynamic> item;

  ContentCreatedEvent({required this.item});

  factory ContentCreatedEvent.fromJson(Map<String, dynamic> json) {
    return ContentCreatedEvent(
      item: json['data']['item'] as Map<String, dynamic>,
    );
  }
}

class ContentUpdatedEvent {
  final Map<String, dynamic> item;
  final int yourVersion;
  final bool conflictDetected;

  ContentUpdatedEvent({
    required this.item,
    required this.yourVersion,
    required this.conflictDetected,
  });

  factory ContentUpdatedEvent.fromJson(Map<String, dynamic> json) {
    final data = json['data'] as Map<String, dynamic>;
    return ContentUpdatedEvent(
      item: data['item'] as Map<String, dynamic>,
      yourVersion: data['your_version'] as int,
      conflictDetected: data['conflict_detected'] as bool? ?? false,
    );
  }
}

class ContentDeletedEvent {
  final String itemId;
  final int deletedAt;

  ContentDeletedEvent({
    required this.itemId,
    required this.deletedAt,
  });

  factory ContentDeletedEvent.fromJson(Map<String, dynamic> json) {
    final data = json['data'] as Map<String, dynamic>;
    return ContentDeletedEvent(
      itemId: data['item_id'] as String,
      deletedAt: data['deleted_at'] as int,
    );
  }
}

// Export Events
class ExportStartedEvent {
  final String exportId;
  final bool includeMedia;

  ExportStartedEvent({
    required this.exportId,
    required this.includeMedia,
  });

  factory ExportStartedEvent.fromJson(Map<String, dynamic> json) {
    final data = json['data'] as Map<String, dynamic>;
    return ExportStartedEvent(
      exportId: data['export_id'] as String,
      includeMedia: data['include_media'] as bool? ?? true,
    );
  }
}

class ExportProgressEvent {
  final String exportId;
  final String stage;
  final int percent;
  final String currentFile;

  ExportProgressEvent({
    required this.exportId,
    required this.stage,
    required this.percent,
    required this.currentFile,
  });

  factory ExportProgressEvent.fromJson(Map<String, dynamic> json) {
    final data = json['data'] as Map<String, dynamic>;
    return ExportProgressEvent(
      exportId: data['export_id'] as String,
      stage: data['stage'] as String,
      percent: data['percent'] as int,
      currentFile: data['current_file'] as String,
    );
  }
}

class ExportCompletedEvent {
  final String exportId;
  final String filePath;
  final int sizeBytes;
  final int itemCount;
  final String checksum;

  ExportCompletedEvent({
    required this.exportId,
    required this.filePath,
    required this.sizeBytes,
    required this.itemCount,
    required this.checksum,
  });

  factory ExportCompletedEvent.fromJson(Map<String, dynamic> json) {
    final data = json['data'] as Map<String, dynamic>;
    return ExportCompletedEvent(
      exportId: data['export_id'] as String,
      filePath: data['file_path'] as String,
      sizeBytes: data['size_bytes'] as int,
      itemCount: data['item_count'] as int,
      checksum: data['checksum'] as String,
    );
  }
}

class ExportFailedEvent {
  final String exportId;
  final String errorCode;
  final String errorMessage;

  ExportFailedEvent({
    required this.exportId,
    required this.errorCode,
    required this.errorMessage,
  });

  factory ExportFailedEvent.fromJson(Map<String, dynamic> json) {
    final data = json['data'] as Map<String, dynamic>;
    return ExportFailedEvent(
      exportId: data['export_id'] as String,
      errorCode: data['error_code'] as String,
      errorMessage: data['error_message'] as String,
    );
  }
}

// Import Events
class ImportStartedEvent {
  final String importId;
  final String archivePath;

  ImportStartedEvent({
    required this.importId,
    required this.archivePath,
  });

  factory ImportStartedEvent.fromJson(Map<String, dynamic> json) {
    final data = json['data'] as Map<String, dynamic>;
    return ImportStartedEvent(
      importId: data['import_id'] as String,
      archivePath: data['archive_path'] as String,
    );
  }
}

class ImportCompletedEvent {
  final String importId;
  final int importedCount;
  final int skippedCount;
  final int durationSeconds;

  ImportCompletedEvent({
    required this.importId,
    required this.importedCount,
    required this.skippedCount,
    required this.durationSeconds,
  });

  factory ImportCompletedEvent.fromJson(Map<String, dynamic> json) {
    final data = json['data'] as Map<String, dynamic>;
    return ImportCompletedEvent(
      importId: data['import_id'] as String,
      importedCount: data['imported_count'] as int,
      skippedCount: data['skipped_count'] as int,
      durationSeconds: data['duration_seconds'] as int,
    );
  }
}

class ImportFailedEvent {
  final String importId;
  final String errorCode;
  final String errorMessage;

  ImportFailedEvent({
    required this.importId,
    required this.errorCode,
    required this.errorMessage,
  });

  factory ImportFailedEvent.fromJson(Map<String, dynamic> json) {
    final data = json['data'] as Map<String, dynamic>;
    return ImportFailedEvent(
      importId: data['import_id'] as String,
      errorCode: data['error_code'] as String,
      errorMessage: data['error_message'] as String,
    );
  }
}
