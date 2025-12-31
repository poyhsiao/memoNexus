// Export Provider - T200: User Story 5
// Simplified version without code generation
//
// ★ Insight ─────────────────────────────────────
// 1. Manual state classes vs freezed: Trade-off between
//    boilerplate code and build-time generation. For
//    initial development, manual classes provide faster
//    iteration and zero build dependencies.
// 2. StateNotifier pattern provides reactive state updates
//    without requiring code generation or build_runner.
// ─────────────────────────────────────────────────

import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uuid/uuid.dart';

// Export status enum
enum ExportStatus {
  idle,
  preparing,
  encrypting,
  compressing,
  uploading,
  completed,
  failed,
}

// Import status enum
enum ImportStatus {
  idle,
  validating,
  decrypting,
  extracting,
  restoring,
  completed,
  failed,
}

// Export state model
class ExportState {
  final ExportStatus status;
  final int progress;
  final String currentStage;
  final String currentFile;
  final int totalItems;
  final int processedItems;
  final String? filePath;
  final String? checksum;
  final int? sizeBytes;
  final bool includeMedia;
  final String? errorMessage;
  final DateTime? startedAt;
  final DateTime? completedAt;

  const ExportState({
    this.status = ExportStatus.idle,
    this.progress = 0,
    this.currentStage = '',
    this.currentFile = '',
    this.totalItems = 0,
    this.processedItems = 0,
    this.filePath,
    this.checksum,
    this.sizeBytes,
    this.includeMedia = false,
    this.errorMessage,
    this.startedAt,
    this.completedAt,
  });

  /// Calculate progress percentage (0-100)
  double get progressPercent {
    if (totalItems == 0) return 0;
    return (processedItems / totalItems * 100).clamp(0, 100);
  }

  /// Check if export is currently running
  bool get isRunning =>
      status == ExportStatus.preparing ||
      status == ExportStatus.encrypting ||
      status == ExportStatus.compressing ||
      status == ExportStatus.uploading;

  /// Check if export completed successfully
  bool get isSuccess => status == ExportStatus.completed;

  /// Check if export failed
  bool get isFailed => status == ExportStatus.failed;

  ExportState copyWith({
    ExportStatus? status,
    int? progress,
    String? currentStage,
    String? currentFile,
    int? totalItems,
    int? processedItems,
    String? filePath,
    String? checksum,
    int? sizeBytes,
    bool? includeMedia,
    String? errorMessage,
    DateTime? startedAt,
    DateTime? completedAt,
  }) {
    return ExportState(
      status: status ?? this.status,
      progress: progress ?? this.progress,
      currentStage: currentStage ?? this.currentStage,
      currentFile: currentFile ?? this.currentFile,
      totalItems: totalItems ?? this.totalItems,
      processedItems: processedItems ?? this.processedItems,
      filePath: filePath ?? this.filePath,
      checksum: checksum ?? this.checksum,
      sizeBytes: sizeBytes ?? this.sizeBytes,
      includeMedia: includeMedia ?? this.includeMedia,
      errorMessage: errorMessage ?? this.errorMessage,
      startedAt: startedAt ?? this.startedAt,
      completedAt: completedAt ?? this.completedAt,
    );
  }
}

// Import state model
class ImportState {
  final ImportStatus status;
  final int progress;
  final String currentStage;
  final int totalItems;
  final int importedItems;
  final int skippedItems;
  final String? filePath;
  final String? errorMessage;
  final DateTime? startedAt;
  final DateTime? completedAt;

  const ImportState({
    this.status = ImportStatus.idle,
    this.progress = 0,
    this.currentStage = '',
    this.totalItems = 0,
    this.importedItems = 0,
    this.skippedItems = 0,
    this.filePath,
    this.errorMessage,
    this.startedAt,
    this.completedAt,
  });

  /// Calculate progress percentage (0-100)
  double get progressPercent {
    if (totalItems == 0) return 0;
    return (importedItems / totalItems * 100).clamp(0, 100);
  }

  /// Check if import is currently running
  bool get isRunning =>
      status == ImportStatus.validating ||
      status == ImportStatus.decrypting ||
      status == ImportStatus.extracting ||
      status == ImportStatus.restoring;

  /// Check if import completed successfully
  bool get isSuccess => status == ImportStatus.completed;

  /// Check if import failed
  bool get isFailed => status == ImportStatus.failed;

  ImportState copyWith({
    ImportStatus? status,
    int? progress,
    String? currentStage,
    int? totalItems,
    int? importedItems,
    int? skippedItems,
    String? filePath,
    String? errorMessage,
    DateTime? startedAt,
    DateTime? completedAt,
  }) {
    return ImportState(
      status: status ?? this.status,
      progress: progress ?? this.progress,
      currentStage: currentStage ?? this.currentStage,
      totalItems: totalItems ?? this.totalItems,
      importedItems: importedItems ?? this.importedItems,
      skippedItems: skippedItems ?? this.skippedItems,
      filePath: filePath ?? this.filePath,
      errorMessage: errorMessage ?? this.errorMessage,
      startedAt: startedAt ?? this.startedAt,
      completedAt: completedAt ?? this.completedAt,
    );
  }
}

// Export archive model
class ExportArchive {
  final String id;
  final String filePath;
  final String checksum;
  final int sizeBytes;
  final int itemCount;
  final DateTime createdAt;
  final bool isEncrypted;

  const ExportArchive({
    required this.id,
    required this.filePath,
    required this.checksum,
    required this.sizeBytes,
    required this.itemCount,
    required this.createdAt,
    this.isEncrypted = false,
  });

  /// Format file size for display
  String get formattedSize {
    const kb = 1024;
    const mb = kb * 1024;
    const gb = mb * 1024;

    if (sizeBytes >= gb) {
      return '${(sizeBytes / gb).toStringAsFixed(2)} GB';
    } else if (sizeBytes >= mb) {
      return '${(sizeBytes / mb).toStringAsFixed(2)} MB';
    } else if (sizeBytes >= kb) {
      return '${(sizeBytes / kb).toStringAsFixed(2)} KB';
    } else {
      return '$sizeBytes bytes';
    }
  }

  /// Format date for display
  String get formattedDate {
    return '${createdAt.year}-${createdAt.month.toString().padLeft(2, '0')}-${createdAt.day.toString().padLeft(2, '0')} '
        '${createdAt.hour.toString().padLeft(2, '0')}:${createdAt.minute.toString().padLeft(2, '0')}';
  }
}

// Auto-export configuration
class AutoExportConfig {
  final bool enabled;
  final AutoExportInterval interval;
  final int retentionCount;
  final bool includeMedia;
  final DateTime? lastExportAt;
  final DateTime? nextExportAt;

  const AutoExportConfig({
    this.enabled = false,
    this.interval = AutoExportInterval.manual,
    this.retentionCount = 5,
    this.includeMedia = false,
    this.lastExportAt,
    this.nextExportAt,
  });

  AutoExportConfig copyWith({
    bool? enabled,
    AutoExportInterval? interval,
    int? retentionCount,
    bool? includeMedia,
    DateTime? lastExportAt,
    DateTime? nextExportAt,
  }) {
    return AutoExportConfig(
      enabled: enabled ?? this.enabled,
      interval: interval ?? this.interval,
      retentionCount: retentionCount ?? this.retentionCount,
      includeMedia: includeMedia ?? this.includeMedia,
      lastExportAt: lastExportAt ?? this.lastExportAt,
      nextExportAt: nextExportAt ?? this.nextExportAt,
    );
  }
}

// Auto-export interval enum
enum AutoExportInterval {
  manual,
  daily,
  weekly,
  monthly,
}

/// Export StateNotifier - manages export operations
class ExportNotifier extends StateNotifier<ExportState> {
  ExportNotifier() : super(const ExportState());

  final _uuid = const Uuid();

  /// Start a new export operation
  Future<void> startExport({
    required String password,
    required bool includeMedia,
  }) async {
    state = state.copyWith(
      status: ExportStatus.preparing,
      startedAt: DateTime.now(),
      includeMedia: includeMedia,
    );

    try {
      // TODO: Implement actual export via API/FFI
      await Future.delayed(const Duration(milliseconds: 100));

      state = state.copyWith(
        status: ExportStatus.encrypting,
        currentStage: 'Encrypting with AES-256',
        progress: 25,
      );

      await Future.delayed(const Duration(milliseconds: 100));

      state = state.copyWith(
        status: ExportStatus.compressing,
        currentStage: 'Compressing archive',
        progress: 50,
      );

      await Future.delayed(const Duration(milliseconds: 100));

      state = state.copyWith(
        status: ExportStatus.completed,
        progress: 100,
        currentStage: 'Completed',
        completedAt: DateTime.now(),
        filePath: '/exports/memonexus_${DateTime.now().millisecondsSinceEpoch}.tar.gz',
        checksum: _uuid.v4().substring(0, 8),
        sizeBytes: 1024000,
      );
    } catch (e) {
      state = state.copyWith(
        status: ExportStatus.failed,
        errorMessage: e.toString(),
        completedAt: DateTime.now(),
      );
    }
  }

  /// Cancel the current export
  void cancelExport() {
    if (state.isRunning) {
      state = state.copyWith(
        status: ExportStatus.idle,
        errorMessage: 'Export cancelled by user',
        completedAt: DateTime.now(),
      );
    }
  }

  /// Reset to idle state
  void reset() {
    state = const ExportState();
  }
}

/// Import StateNotifier - manages import operations
class ImportNotifier extends StateNotifier<ImportState> {
  ImportNotifier() : super(const ImportState());

  /// Start a new import operation
  Future<void> startImport({
    required String filePath,
    required String password,
  }) async {
    state = state.copyWith(
      status: ImportStatus.validating,
      startedAt: DateTime.now(),
      filePath: filePath,
    );

    try {
      // TODO: Implement actual import via API/FFI
      await Future.delayed(const Duration(milliseconds: 100));

      state = state.copyWith(
        status: ImportStatus.decrypting,
        currentStage: 'Decrypting archive',
        progress: 25,
      );

      await Future.delayed(const Duration(milliseconds: 100));

      state = state.copyWith(
        status: ImportStatus.extracting,
        currentStage: 'Extracting files',
        progress: 50,
      );

      await Future.delayed(const Duration(milliseconds: 100));

      state = state.copyWith(
        status: ImportStatus.restoring,
        currentStage: 'Restoring database',
        progress: 80,
        totalItems: 100,
        importedItems: 95,
        skippedItems: 5,
      );

      await Future.delayed(const Duration(milliseconds: 100));

      state = state.copyWith(
        status: ImportStatus.completed,
        progress: 100,
        currentStage: 'Completed',
        completedAt: DateTime.now(),
      );
    } catch (e) {
      state = state.copyWith(
        status: ImportStatus.failed,
        errorMessage: e.toString(),
        completedAt: DateTime.now(),
      );
    }
  }

  /// Cancel the current import
  void cancelImport() {
    if (state.isRunning) {
      state = state.copyWith(
        status: ImportStatus.idle,
        errorMessage: 'Import cancelled by user',
        completedAt: DateTime.now(),
      );
    }
  }

  /// Reset to idle state
  void reset() {
    state = const ImportState();
  }
}

/// Export archives StateNotifier - manages list of export archives
class ExportArchivesNotifier extends StateNotifier<AsyncValue<List<ExportArchive>>> {
  ExportArchivesNotifier() : super(const AsyncValue.loading()) {
    loadArchives();
  }

  /// Load export archives from storage
  Future<void> loadArchives() async {
    state = const AsyncValue.loading();

    try {
      // TODO: Implement actual loading via API/FFI
      await Future.delayed(const Duration(milliseconds: 100));

      // Placeholder data
      final archives = <ExportArchive>[
        ExportArchive(
          id: '1',
          filePath: '/exports/memonexus_20241231.tar.gz',
          checksum: 'abc123def456',
          sizeBytes: 1024000,
          itemCount: 150,
          createdAt: DateTime.now().subtract(const Duration(days: 1)),
          isEncrypted: true,
        ),
        ExportArchive(
          id: '2',
          filePath: '/exports/memonexus_20241230.tar.gz',
          checksum: 'def456ghi789',
          sizeBytes: 980000,
          itemCount: 142,
          createdAt: DateTime.now().subtract(const Duration(days: 2)),
          isEncrypted: true,
        ),
      ];

      state = AsyncValue.data(archives);
    } catch (e, st) {
      state = AsyncValue.error(e, st);
    }
  }

  /// Delete an export archive
  Future<void> deleteArchive(String id) async {
    try {
      // TODO: Implement actual deletion via API/FFI
      state.whenData((archives) {
        final updated = archives.where((a) => a.id != id).toList();
        state = AsyncValue.data(updated);
      });
    } catch (e, st) {
      state = AsyncValue.error(e, st);
    }
  }
}

/// Auto-export config StateNotifier - manages auto-export settings
class AutoExportConfigNotifier extends StateNotifier<AutoExportConfig> {
  AutoExportConfigNotifier() : super(const AutoExportConfig());

  /// Update auto-export configuration
  void updateConfig(AutoExportConfig config) {
    state = config;
    // TODO: Persist to storage
  }

  /// Enable/disable auto-export
  void setEnabled(bool enabled) {
    state = state.copyWith(enabled: enabled);
  }

  /// Set export interval
  void setInterval(AutoExportInterval interval) {
    state = state.copyWith(interval: interval);
  }

  /// Set retention count
  void setRetentionCount(int count) {
    state = state.copyWith(retentionCount: count.clamp(1, 100));
  }

  /// Set whether to include media
  void setIncludeMedia(bool include) {
    state = state.copyWith(includeMedia: include);
  }
}

// Providers
final exportProvider = StateNotifierProvider<ExportNotifier, ExportState>(
  (ref) => ExportNotifier(),
);

final importProvider = StateNotifierProvider<ImportNotifier, ImportState>(
  (ref) => ImportNotifier(),
);

final exportArchivesProvider =
    StateNotifierProvider<ExportArchivesNotifier, AsyncValue<List<ExportArchive>>>(
  (ref) => ExportArchivesNotifier(),
);

final autoExportConfigProvider =
    StateNotifierProvider<AutoExportConfigNotifier, AutoExportConfig>(
  (ref) => AutoExportConfigNotifier(),
);
