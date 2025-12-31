// Sync Provider for Riverpod state management
// Manages cloud synchronization configuration and status with S3-compatible storage
// T173: SyncConfig Riverpod provider

import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../services/api_client.dart';
import 'content_provider.dart' show apiClientProvider;

// =====================================================
// Sync Config State
// =====================================================

class SyncConfigState {
  final bool isConfigured;
  final String? endpoint;
  final String? bucketName;
  final String? region;
  final bool isLoading;
  final String? error;

  const SyncConfigState({
    this.isConfigured = false,
    this.endpoint,
    this.bucketName,
    this.region,
    this.isLoading = false,
    this.error,
  });

  SyncConfigState copyWith({
    bool? isConfigured,
    String? endpoint,
    String? bucketName,
    String? region,
    bool? isLoading,
    String? error,
  }) {
    return SyncConfigState(
      isConfigured: isConfigured ?? this.isConfigured,
      endpoint: endpoint ?? this.endpoint,
      bucketName: bucketName ?? this.bucketName,
      region: region ?? this.region,
      isLoading: isLoading ?? this.isLoading,
      error: error,
    );
  }
}

// =====================================================
// Sync Status State
// =====================================================

class SyncStatusState {
  final String status; // idle, syncing, failed
  final int? lastSync; // Unix timestamp
  final int pendingChanges;
  final bool isLoading;
  final String? error;

  const SyncStatusState({
    this.status = 'idle',
    this.lastSync,
    this.pendingChanges = 0,
    this.isLoading = false,
    this.error,
  });

  SyncStatusState copyWith({
    String? status,
    int? lastSync,
    int? pendingChanges,
    bool? isLoading,
    String? error,
  }) {
    return SyncStatusState(
      status: status ?? this.status,
      lastSync: lastSync ?? this.lastSync,
      pendingChanges: pendingChanges ?? this.pendingChanges,
      isLoading: isLoading ?? this.isLoading,
      error: error,
    );
  }

  bool get isSyncing => status == 'syncing';
  bool get isIdle => status == 'idle';
  bool get hasFailed => status == 'failed';
}

// =====================================================
// Sync Config Notifier
// =====================================================

class SyncConfigNotifier extends StateNotifier<SyncConfigState> {
  final MemoNexusAPIClient _api;

  SyncConfigNotifier(this._api) : super(const SyncConfigState());

  /// Load current sync configuration
  Future<void> loadConfig() async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final data = await _api.getSyncCredentials();

      state = SyncConfigState(
        isConfigured: data['configured'] as bool? ?? false,
        endpoint: data['endpoint'] as String?,
        bucketName: data['bucket_name'] as String?,
        region: data['region'] as String?,
        isLoading: false,
      );
    } catch (e) {
      state = state.copyWith(
        isConfigured: false,
        isLoading: false,
        error: e.toString(),
      );
    }
  }

  /// Configure sync with S3-compatible credentials
  Future<bool> configureSync({
    required String endpoint,
    required String bucketName,
    required String accessKey,
    required String secretKey,
    String? region,
  }) async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      await _api.configureSync(
        endpoint: endpoint,
        bucketName: bucketName,
        accessKey: accessKey,
        secretKey: secretKey,
        region: region,
      );

      // Reload config to get the saved state
      await loadConfig();
      return true;
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: e.toString(),
      );
      return false;
    }
  }

  /// Disable sync and remove credentials
  Future<void> disableSync() async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      await _api.disableSync();
      state = const SyncConfigState();
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: e.toString(),
      );
    }
  }

  /// Clear error state
  void clearError() {
    state = state.copyWith(error: null);
  }
}

// =====================================================
// Sync Status Notifier
// =====================================================

class SyncStatusNotifier extends StateNotifier<SyncStatusState> {
  final MemoNexusAPIClient _api;

  SyncStatusNotifier(this._api) : super(const SyncStatusState());

  /// Load current sync status
  Future<void> loadStatus() async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final data = await _api.getSyncStatus();

      state = SyncStatusState(
        status: data['status'] as String? ?? 'idle',
        lastSync: data['last_sync'] as int?,
        pendingChanges: data['pending_changes'] as int? ?? 0,
        isLoading: false,
      );
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: e.toString(),
      );
    }
  }

  /// Trigger immediate sync
  Future<bool> triggerSync() async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final data = await _api.triggerSync();

      state = SyncStatusState(
        status: 'idle', // Sync completed immediately
        lastSync: DateTime.now().millisecondsSinceEpoch,
        pendingChanges: data['pending_changes'] as int? ?? 0,
        isLoading: false,
      );

      return true;
    } catch (e) {
      state = state.copyWith(
        status: 'failed',
        isLoading: false,
        error: e.toString(),
      );
      return false;
    }
  }

  /// Clear error state
  void clearError() {
    state = state.copyWith(error: null);
  }
}

// =====================================================
// Providers
// =====================================================

/// Sync config provider - sync configuration state and operations
final syncConfigProvider =
    StateNotifierProvider<SyncConfigNotifier, SyncConfigState>((ref) {
  final api = ref.watch(apiClientProvider);
  final notifier = SyncConfigNotifier(api);

  // Load config on initialization
  notifier.loadConfig();

  return notifier;
});

/// Sync status provider - sync status and operations
final syncStatusProvider =
    StateNotifierProvider<SyncStatusNotifier, SyncStatusState>((ref) {
  final api = ref.watch(apiClientProvider);
  final notifier = SyncStatusNotifier(api);

  // Load status on initialization
  notifier.loadStatus();

  return notifier;
});

// =====================================================
// Storage Type Options
// =====================================================

/// Storage provider options for sync configuration
enum StorageProvider {
  aws('AWS S3', 's3.amazonaws.com', 'us-east-1'),
  cloudflare('Cloudflare R2', '<account>.r2.cloudflarestorage.com', 'auto'),
  minio('MinIO', 'localhost:9000', 'us-east-1'),
  custom('Custom S3-Compatible', '', '');

  const StorageProvider(this.displayName, this.defaultEndpoint, this.defaultRegion);

  final String displayName;
  final String defaultEndpoint;
  final String defaultRegion;
}

/// Storage provider names for dropdown display
final storageProviderNames = <String, String>{
  'aws': 'AWS S3',
  'cloudflare': 'Cloudflare R2',
  'minio': 'MinIO (Self-hosted)',
  'custom': 'Custom S3-Compatible',
};
