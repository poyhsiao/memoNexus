import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:memonexus_frontend/providers/sync_provider.dart';
import 'package:memonexus_frontend/services/api_client.dart';

void main() {
  group('SyncConfigState', () {
    test('should have correct default values', () {
      const state = SyncConfigState();

      expect(state.isConfigured, false);
      expect(state.endpoint, isNull);
      expect(state.bucketName, isNull);
      expect(state.region, isNull);
      expect(state.isLoading, false);
      expect(state.error, isNull);
    });

    test('copyWith should create new state with updated values', () {
      const original = SyncConfigState(
        isConfigured: true,
        endpoint: 's3.amazonaws.com',
      );
      final updated = original.copyWith(
        isLoading: true,
        bucketName: 'my-bucket',
      );

      expect(original.isConfigured, true);
      expect(original.endpoint, 's3.amazonaws.com');
      expect(original.bucketName, isNull);
      expect(original.isLoading, false);
      expect(updated.isConfigured, true);
      expect(updated.bucketName, 'my-bucket');
      expect(updated.isLoading, true);
    });

    test('copyWith with null error should clear error', () {
      const original = SyncConfigState(error: 'Connection error');
      final updated = original.copyWith(isLoading: true);

      // Note: error field is directly assigned (no ??), so null clears it
      expect(updated.error, isNull);
    });
  });

  group('SyncStatusState', () {
    test('should have correct default values', () {
      const state = SyncStatusState();

      expect(state.status, 'idle');
      expect(state.lastSync, isNull);
      expect(state.pendingChanges, 0);
      expect(state.isLoading, false);
      expect(state.error, isNull);
    });

    test('copyWith should create new state with updated values', () {
      const original = SyncStatusState(
        status: 'syncing',
        pendingChanges: 5,
      );
      final updated = original.copyWith(
        isLoading: true,
        status: 'idle',
      );

      expect(original.status, 'syncing');
      expect(original.pendingChanges, 5);
      expect(original.isLoading, false);
      expect(updated.status, 'idle');
      expect(updated.pendingChanges, 5);
      expect(updated.isLoading, true);
    });

    test('isSyncing should return true when status is syncing', () {
      const state = SyncStatusState(status: 'syncing');
      expect(state.isSyncing, isTrue);
      expect(state.isIdle, isFalse);
      expect(state.hasFailed, isFalse);
    });

    test('isIdle should return true when status is idle', () {
      const state = SyncStatusState(status: 'idle');
      expect(state.isIdle, isTrue);
      expect(state.isSyncing, isFalse);
      expect(state.hasFailed, isFalse);
    });

    test('hasFailed should return true when status is failed', () {
      const state = SyncStatusState(status: 'failed');
      expect(state.hasFailed, isTrue);
      expect(state.isSyncing, isFalse);
      expect(state.isIdle, isFalse);
    });

    test('copyWith with null error should clear error', () {
      const original = SyncStatusState(error: 'Sync failed');
      final updated = original.copyWith(isLoading: true);

      // Note: error field is directly assigned (no ??), so null clears it
      expect(updated.error, isNull);
    });
  });

  group('SyncConfigNotifier', () {
    test('loadConfig should update state with API data', () async {
      final api = _FakeMemoNexusAPIClient();
      api.mockSyncCredentials = {
        'configured': true,
        'endpoint': 's3.amazonaws.com',
        'bucket_name': 'test-bucket',
        'region': 'us-east-1',
      };
      final notifier = SyncConfigNotifier(api);

      await notifier.loadConfig();

      expect(notifier.state.isConfigured, true);
      expect(notifier.state.endpoint, 's3.amazonaws.com');
      expect(notifier.state.bucketName, 'test-bucket');
      expect(notifier.state.region, 'us-east-1');
      expect(notifier.state.isLoading, false);
      expect(notifier.state.error, isNull);
    });

    test('loadConfig should handle missing optional fields', () async {
      final api = _FakeMemoNexusAPIClient();
      api.mockSyncCredentials = {'configured': false};
      final notifier = SyncConfigNotifier(api);

      await notifier.loadConfig();

      expect(notifier.state.isConfigured, false);
      expect(notifier.state.endpoint, isNull);
      expect(notifier.state.bucketName, isNull);
      expect(notifier.state.isLoading, false);
    });

    test('loadConfig should handle API errors', () async {
      final api = _FakeMemoNexusAPIClient();
      api.shouldFail = true;
      api.errorMessage = 'Network error';
      final notifier = SyncConfigNotifier(api);

      await notifier.loadConfig();

      expect(notifier.state.isConfigured, false);
      expect(notifier.state.isLoading, false);
      expect(notifier.state.error, isNotNull);
      expect(notifier.state.error, contains('Network error'));
    });

    test('configureSync should return true on success', () async {
      final api = _FakeMemoNexusAPIClient();
      api.mockSyncCredentials = {
        'configured': true,
        'endpoint': 's3.amazonaws.com',
        'bucket_name': 'test-bucket',
      };
      final notifier = SyncConfigNotifier(api);

      final result = await notifier.configureSync(
        endpoint: 's3.amazonaws.com',
        bucketName: 'test-bucket',
        accessKey: 'key',
        secretKey: 'secret',
      );

      expect(result, isTrue);
      expect(notifier.state.isConfigured, true);
      expect(notifier.state.error, isNull);
    });

    test('configureSync should handle API errors', () async {
      final api = _FakeMemoNexusAPIClient();
      api.shouldFail = true;
      api.errorMessage = 'Invalid credentials';
      final notifier = SyncConfigNotifier(api);

      final result = await notifier.configureSync(
        endpoint: 's3.amazonaws.com',
        bucketName: 'test-bucket',
        accessKey: 'key',
        secretKey: 'secret',
      );

      expect(result, isFalse);
      expect(notifier.state.error, isNotNull);
      expect(notifier.state.error, contains('Invalid credentials'));
    });

    test('disableSync should reset state to default', () async {
      final api = _FakeMemoNexusAPIClient();
      api.mockSyncCredentials = {
        'configured': true,
        'endpoint': 'minio:9000',
      };
      final notifier = SyncConfigNotifier(api);

      // First load a config
      await notifier.loadConfig();
      expect(notifier.state.isConfigured, true);

      // Then disable
      await notifier.disableSync();

      expect(notifier.state.isConfigured, false);
      expect(notifier.state.endpoint, isNull);
      expect(notifier.state.isLoading, false);
    });

    test('disableSync should handle API errors', () async {
      final api = _FakeMemoNexusAPIClient();
      api.shouldFail = true;
      api.errorMessage = 'Failed to disable';
      final notifier = SyncConfigNotifier(api);

      notifier.state = const SyncConfigState(isConfigured: true);

      await notifier.disableSync();

      expect(notifier.state.error, isNotNull);
      expect(notifier.state.error, contains('Failed to disable'));
    });

    test('clearError should clear error state', () async {
      final api = _FakeMemoNexusAPIClient();
      api.shouldFail = true;
      final notifier = SyncConfigNotifier(api);

      await notifier.loadConfig();
      expect(notifier.state.error, isNotNull);

      notifier.clearError();
      expect(notifier.state.error, isNull);
    });
  });

  group('SyncStatusNotifier', () {
    test('loadStatus should update state with API data', () async {
      final api = _FakeMemoNexusAPIClient();
      final now = DateTime.now().millisecondsSinceEpoch;
      api.mockSyncStatus = {
        'status': 'idle',
        'last_sync': now,
        'pending_changes': 3,
      };
      final notifier = SyncStatusNotifier(api);

      await notifier.loadStatus();

      expect(notifier.state.status, 'idle');
      expect(notifier.state.lastSync, now);
      expect(notifier.state.pendingChanges, 3);
      expect(notifier.state.isLoading, false);
      expect(notifier.state.error, isNull);
    });

    test('loadStatus should handle missing optional fields', () async {
      final api = _FakeMemoNexusAPIClient();
      api.mockSyncStatus = {'status': 'idle'};
      final notifier = SyncStatusNotifier(api);

      await notifier.loadStatus();

      expect(notifier.state.status, 'idle');
      expect(notifier.state.lastSync, isNull);
      expect(notifier.state.pendingChanges, 0);
    });

    test('loadStatus should handle API errors', () async {
      final api = _FakeMemoNexusAPIClient();
      api.shouldFail = true;
      api.errorMessage = 'Failed to load status';
      final notifier = SyncStatusNotifier(api);

      await notifier.loadStatus();

      expect(notifier.state.isLoading, false);
      expect(notifier.state.error, isNotNull);
      expect(notifier.state.error, contains('Failed to load status'));
    });

    test('triggerSync should update state on success', () async {
      final api = _FakeMemoNexusAPIClient();
      api.mockSyncStatus = {
        'pending_changes': 0,
      };
      final notifier = SyncStatusNotifier(api);

      final result = await notifier.triggerSync();

      expect(result, isTrue);
      expect(notifier.state.status, 'idle');
      expect(notifier.state.pendingChanges, 0);
      expect(notifier.state.lastSync, isNotNull);
      expect(notifier.state.error, isNull);
    });

    test('triggerSync should handle API errors', () async {
      final api = _FakeMemoNexusAPIClient();
      api.shouldFail = true;
      api.errorMessage = 'Sync failed';
      final notifier = SyncStatusNotifier(api);

      final result = await notifier.triggerSync();

      expect(result, isFalse);
      expect(notifier.state.status, 'failed');
      expect(notifier.state.error, isNotNull);
      expect(notifier.state.error, contains('Sync failed'));
    });

    test('clearError should clear error state', () async {
      final api = _FakeMemoNexusAPIClient();
      api.shouldFail = true;
      final notifier = SyncStatusNotifier(api);

      await notifier.loadStatus();
      expect(notifier.state.error, isNotNull);

      notifier.clearError();
      expect(notifier.state.error, isNull);
    });
  });

  group('Provider Integration', () {
    test('syncConfigProvider should provide SyncConfigNotifier', () async {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      // First read the provider to trigger auto-loading, then wait for it
      container.read(syncConfigProvider);
      await Future.delayed(const Duration(milliseconds: 10));

      // Now safely access the notifier
      final notifier = container.read(syncConfigProvider.notifier);
      expect(notifier, isA<SyncConfigNotifier>());

      // Wait for any pending async operations
      await Future.delayed(const Duration(milliseconds: 10));
    });

    test('syncStatusProvider should provide SyncStatusNotifier', () async {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      // First read the provider to trigger auto-loading, then wait for it
      container.read(syncStatusProvider);
      await Future.delayed(const Duration(milliseconds: 10));

      // Now safely access the notifier
      final notifier = container.read(syncStatusProvider.notifier);
      expect(notifier, isA<SyncStatusNotifier>());

      // Wait for any pending async operations
      await Future.delayed(const Duration(milliseconds: 10));
    });

    // Note: Skipping auto-load tests due to async initialization issues
  });

  group('StorageProvider Enum', () {
    test('should have correct values for all providers', () {
      expect(StorageProvider.aws.displayName, 'AWS S3');
      expect(StorageProvider.aws.defaultEndpoint, 's3.amazonaws.com');
      expect(StorageProvider.aws.defaultRegion, 'us-east-1');

      expect(StorageProvider.cloudflare.displayName, 'Cloudflare R2');
      expect(StorageProvider.cloudflare.defaultRegion, 'auto');

      expect(StorageProvider.minio.displayName, 'MinIO');
      expect(StorageProvider.minio.defaultEndpoint, 'localhost:9000');

      expect(StorageProvider.custom.displayName, 'Custom S3-Compatible');
      expect(StorageProvider.custom.defaultEndpoint, isEmpty);
    });
  });

  group('Storage Provider Names', () {
    test('should contain all storage providers', () {
      expect(storageProviderNames['aws'], 'AWS S3');
      expect(storageProviderNames['cloudflare'], 'Cloudflare R2');
      expect(storageProviderNames['minio'], 'MinIO (Self-hosted)');
      expect(storageProviderNames['custom'], 'Custom S3-Compatible');
    });
  });
}

// =====================================================
// Fake API Client
// =====================================================

class _FakeMemoNexusAPIClient extends MemoNexusAPIClient {
  bool shouldFail = false;
  String? errorMessage;
  Map<String, dynamic>? mockSyncCredentials;
  Map<String, dynamic>? mockSyncStatus;

  @override
  Future<Map<String, dynamic>> getSyncCredentials() async {
    if (shouldFail) {
      throw Exception(errorMessage ?? 'Failed to load sync config');
    }
    return mockSyncCredentials ?? {};
  }

  @override
  Future<Map<String, dynamic>> configureSync({
    required String endpoint,
    required String bucketName,
    required String accessKey,
    required String secretKey,
    String? region,
  }) async {
    if (shouldFail) {
      throw Exception(errorMessage ?? 'Failed to configure sync');
    }
    mockSyncCredentials = {
      'configured': true,
      'endpoint': endpoint,
      'bucket_name': bucketName,
      if (region != null) 'region': region,
    };
    return {'success': true};
  }

  @override
  Future<void> disableSync() async {
    if (shouldFail) {
      throw Exception(errorMessage ?? 'Failed to disable sync');
    }
    mockSyncCredentials = {'configured': false};
  }

  @override
  Future<Map<String, dynamic>> getSyncStatus() async {
    if (shouldFail) {
      throw Exception(errorMessage ?? 'Failed to load sync status');
    }
    return mockSyncStatus ?? {};
  }

  @override
  Future<Map<String, dynamic>> triggerSync() async {
    if (shouldFail) {
      throw Exception(errorMessage ?? 'Failed to trigger sync');
    }
    return mockSyncStatus ?? {'pending_changes': 0};
  }
}
