import 'package:flutter_test/flutter_test.dart';
import 'package:memonexus_frontend/models/ai_config.dart';

void main() {
  group('AIConfig', () {
    test('should create AIConfig with required fields', () {
      final now = DateTime.now();
      final config = AIConfig(
        provider: AIProvider.openai,
        apiEndpoint: 'https://api.openai.com/v1',
        modelName: 'gpt-4',
        createdAt: now,
        updatedAt: now,
      );

      expect(config.provider, AIProvider.openai);
      expect(config.apiEndpoint, 'https://api.openai.com/v1');
      expect(config.modelName, 'gpt-4');
      expect(config.maxTokens, 1000); // default
      expect(config.isEnabled, false); // default
    });

    test('should generate UUID if not provided', () {
      final now = DateTime.now();
      final config1 = AIConfig(
        provider: AIProvider.openai,
        apiEndpoint: 'https://api.example.com',
        modelName: 'model',
        createdAt: now,
        updatedAt: now,
      );

      final config2 = AIConfig(
        provider: AIProvider.openai,
        apiEndpoint: 'https://api.example.com',
        modelName: 'model',
        createdAt: now,
        updatedAt: now,
      );

      expect(config1.id, isNotEmpty);
      expect(config2.id, isNotEmpty);
      expect(config1.id, isNot(config2.id));
    });

    test('should serialize to JSON correctly', () {
      final now = DateTime.utc(2024, 1, 1);
      final config = AIConfig(
        id: 'config-id',
        provider: AIProvider.claude,
        apiEndpoint: 'https://api.anthropic.com/v1',
        modelName: 'claude-3-opus',
        maxTokens: 2000,
        isEnabled: true,
        createdAt: now,
        updatedAt: now,
      );

      final json = config.toJson();

      expect(json['id'], 'config-id');
      expect(json['provider'], 'claude');
      expect(json['api_endpoint'], 'https://api.anthropic.com/v1');
      expect(json['model_name'], 'claude-3-opus');
      expect(json['max_tokens'], 2000);
      expect(json['is_enabled'], 1);
    });

    test('should deserialize from JSON correctly', () {
      final json = {
        'id': 'config-id',
        'provider': 'ollama',
        'api_endpoint': 'http://localhost:11434',
        'model_name': 'llama2',
        'max_tokens': 500,
        'is_enabled': 1,
        'created_at': 1704110400,
        'updated_at': 1704110400,
      };

      final config = AIConfig.fromJson(json);

      expect(config.id, 'config-id');
      expect(config.provider, AIProvider.ollama);
      expect(config.apiEndpoint, 'http://localhost:11434');
      expect(config.modelName, 'llama2');
      expect(config.maxTokens, 500);
      expect(config.isEnabled, true);
    });

    test('should use default values when not provided in JSON', () {
      final json = {
        'id': 'config-id',
        'provider': 'openai',
        'api_endpoint': 'https://api.openai.com/v1',
        'model_name': 'gpt-4',
        'created_at': 1704110400,
        'updated_at': 1704110400,
      };

      final config = AIConfig.fromJson(json);

      expect(config.maxTokens, 1000);
      expect(config.isEnabled, false);
    });

    test('should support all AIProvider values', () {
      final providers = AIProvider.values;
      expect(providers.length, greaterThan(0));
      expect(providers.contains(AIProvider.openai), true);
      expect(providers.contains(AIProvider.claude), true);
      expect(providers.contains(AIProvider.ollama), true);
    });
  });

  group('SyncCredential', () {
    test('should create SyncCredential with required fields', () {
      final now = DateTime.now();
      final credential = SyncCredential(
        endpoint: 'https://s3.amazonaws.com',
        bucketName: 'my-bucket',
        createdAt: now,
        updatedAt: now,
      );

      expect(credential.endpoint, 'https://s3.amazonaws.com');
      expect(credential.bucketName, 'my-bucket');
      expect(credential.region, null);
      expect(credential.isEnabled, false);
    });

    test('should generate UUID if not provided', () {
      final now = DateTime.now();
      final cred1 = SyncCredential(
        endpoint: 'https://s3.amazonaws.com',
        bucketName: 'bucket',
        createdAt: now,
        updatedAt: now,
      );

      final cred2 = SyncCredential(
        endpoint: 'https://s3.amazonaws.com',
        bucketName: 'bucket',
        createdAt: now,
        updatedAt: now,
      );

      expect(cred1.id, isNotEmpty);
      expect(cred2.id, isNotEmpty);
      expect(cred1.id, isNot(cred2.id));
    });

    test('should serialize to JSON correctly', () {
      final now = DateTime.utc(2024, 1, 1);
      final credential = SyncCredential(
        id: 'cred-id',
        endpoint: 'https://s3.amazonaws.com',
        bucketName: 'my-bucket',
        region: 'us-west-2',
        isEnabled: true,
        createdAt: now,
        updatedAt: now,
      );

      final json = credential.toJson();

      expect(json['id'], 'cred-id');
      expect(json['endpoint'], 'https://s3.amazonaws.com');
      expect(json['bucket_name'], 'my-bucket');
      expect(json['region'], 'us-west-2');
      expect(json['is_enabled'], 1);
    });

    test('should deserialize from JSON correctly', () {
      final json = {
        'id': 'cred-id',
        'endpoint': 'https://minio.example.com',
        'bucket_name': 'test-bucket',
        'region': 'eu-central-1',
        'is_enabled': 1,
        'created_at': 1704110400,
        'updated_at': 1704110400,
      };

      final credential = SyncCredential.fromJson(json);

      expect(credential.id, 'cred-id');
      expect(credential.endpoint, 'https://minio.example.com');
      expect(credential.bucketName, 'test-bucket');
      expect(credential.region, 'eu-central-1');
      expect(credential.isEnabled, true);
    });

    test('should handle null region in JSON', () {
      final json = {
        'id': 'cred-id',
        'endpoint': 'https://s3.amazonaws.com',
        'bucket_name': 'bucket',
        'is_enabled': 0,
        'created_at': 1704110400,
        'updated_at': 1704110400,
      };

      final credential = SyncCredential.fromJson(json);

      expect(credential.region, null);
    });
  });

  group('ExportArchive', () {
    test('should create ExportArchive with required fields', () {
      final now = DateTime.now();
      final archive = ExportArchive(
        filePath: '/exports/backup.tar.gz',
        checksum: 'abc123def456',
        sizeBytes: 1024000,
        itemCount: 42,
        createdAt: now,
      );

      expect(archive.filePath, '/exports/backup.tar.gz');
      expect(archive.checksum, 'abc123def456');
      expect(archive.sizeBytes, 1024000);
      expect(archive.itemCount, 42);
      expect(archive.isEncrypted, true); // default
    });

    test('should generate UUID if not provided', () {
      final now = DateTime.now();
      final archive1 = ExportArchive(
        filePath: '/exports/1.tar.gz',
        checksum: 'checksum1',
        sizeBytes: 1000,
        itemCount: 1,
        createdAt: now,
      );

      final archive2 = ExportArchive(
        filePath: '/exports/2.tar.gz',
        checksum: 'checksum2',
        sizeBytes: 2000,
        itemCount: 2,
        createdAt: now,
      );

      expect(archive1.id, isNotEmpty);
      expect(archive2.id, isNotEmpty);
      expect(archive1.id, isNot(archive2.id));
    });

    test('should serialize to JSON correctly', () {
      final now = DateTime.utc(2024, 1, 1);
      final archive = ExportArchive(
        id: 'archive-id',
        filePath: '/exports/backup.tar.gz',
        checksum: 'abc123',
        sizeBytes: 2048000,
        itemCount: 100,
        isEncrypted: true,
        createdAt: now,
      );

      final json = archive.toJson();

      expect(json['id'], 'archive-id');
      expect(json['file_path'], '/exports/backup.tar.gz');
      expect(json['checksum'], 'abc123');
      expect(json['size_bytes'], 2048000);
      expect(json['item_count'], 100);
      expect(json['is_encrypted'], 1);
    });

    test('should deserialize from JSON correctly', () {
      final json = {
        'id': 'archive-id',
        'file_path': '/exports/manual.tar.gz',
        'checksum': 'xyz789',
        'size_bytes': 512000,
        'item_count': 25,
        'is_encrypted': 0,
        'created_at': 1704110400,
      };

      final archive = ExportArchive.fromJson(json);

      expect(archive.id, 'archive-id');
      expect(archive.filePath, '/exports/manual.tar.gz');
      expect(archive.checksum, 'xyz789');
      expect(archive.sizeBytes, 512000);
      expect(archive.itemCount, 25);
      expect(archive.isEncrypted, false);
    });

    test('should handle unencrypted archives', () {
      final now = DateTime.now();
      final archive = ExportArchive(
        filePath: '/exports/plain.tar.gz',
        checksum: 'checksum',
        sizeBytes: 5000,
        itemCount: 10,
        isEncrypted: false,
        createdAt: now,
      );

      expect(archive.isEncrypted, false);

      final json = archive.toJson();
      expect(json['is_encrypted'], 0);
    });
  });
}
