import 'package:uuid/uuid.dart';

// AI configuration model
enum AIProvider {
  openai,
  claude,
  ollama,
}

// UUID helper
final _uuid = Uuid();

class AIConfig {
  final String id;
  final AIProvider provider;
  final String apiEndpoint;
  final String modelName;
  final int maxTokens;
  final bool isEnabled;
  final DateTime createdAt;
  final DateTime updatedAt;

  AIConfig({
    String? id,
    required this.provider,
    required this.apiEndpoint,
    required this.modelName,
    this.maxTokens = 1000,
    this.isEnabled = false,
    required this.createdAt,
    required this.updatedAt,
  }) : id = id ?? _uuid.v4();

  factory AIConfig.fromJson(Map<String, dynamic> json) {
    return AIConfig(
      id: json['id'] as String,
      provider: AIProvider.values.firstWhere(
        (e) => e.name == json['provider'],
        orElse: () => AIProvider.openai,
      ),
      apiEndpoint: json['api_endpoint'] as String,
      modelName: json['model_name'] as String,
      maxTokens: json['max_tokens'] as int? ?? 1000,
      isEnabled: json['is_enabled'] == 1,
      createdAt: DateTime.fromMillisecondsSinceEpoch(json['created_at'] * 1000),
      updatedAt: DateTime.fromMillisecondsSinceEpoch(json['updated_at'] * 1000),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'provider': provider.name,
      'api_endpoint': apiEndpoint,
      'model_name': modelName,
      'max_tokens': maxTokens,
      'is_enabled': isEnabled ? 1 : 0,
      'created_at': createdAt.millisecondsSinceEpoch ~/ 1000,
      'updated_at': updatedAt.millisecondsSinceEpoch ~/ 1000,
    };
  }
}

// SyncCredential model
class SyncCredential {
  final String id;
  final String endpoint;
  final String bucketName;
  final String? region;
  final bool isEnabled;
  final DateTime createdAt;
  final DateTime updatedAt;

  SyncCredential({
    String? id,
    required this.endpoint,
    required this.bucketName,
    this.region,
    this.isEnabled = false,
    required this.createdAt,
    required this.updatedAt,
  }) : id = id ?? _uuid.v4();

  factory SyncCredential.fromJson(Map<String, dynamic> json) {
    return SyncCredential(
      id: json['id'] as String,
      endpoint: json['endpoint'] as String,
      bucketName: json['bucket_name'] as String,
      region: json['region'] as String?,
      isEnabled: json['is_enabled'] == 1,
      createdAt: DateTime.fromMillisecondsSinceEpoch(json['created_at'] * 1000),
      updatedAt: DateTime.fromMillisecondsSinceEpoch(json['updated_at'] * 1000),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'endpoint': endpoint,
      'bucket_name': bucketName,
      'region': region,
      'is_enabled': isEnabled ? 1 : 0,
      'created_at': createdAt.millisecondsSinceEpoch ~/ 1000,
      'updated_at': updatedAt.millisecondsSinceEpoch ~/ 1000,
    };
  }
}

// ExportArchive model
class ExportArchive {
  final String id;
  final String filePath;
  final String checksum;
  final int sizeBytes;
  final int itemCount;
  final bool isEncrypted;
  final DateTime createdAt;

  ExportArchive({
    String? id,
    required this.filePath,
    required this.checksum,
    required this.sizeBytes,
    required this.itemCount,
    this.isEncrypted = true,
    required this.createdAt,
  }) : id = id ?? _uuid.v4();

  factory ExportArchive.fromJson(Map<String, dynamic> json) {
    return ExportArchive(
      id: json['id'] as String,
      filePath: json['file_path'] as String,
      checksum: json['checksum'] as String,
      sizeBytes: json['size_bytes'] as int,
      itemCount: json['item_count'] as int,
      isEncrypted: json['is_encrypted'] == 1,
      createdAt: DateTime.fromMillisecondsSinceEpoch(json['created_at'] * 1000),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'file_path': filePath,
      'checksum': checksum,
      'size_bytes': sizeBytes,
      'item_count': itemCount,
      'is_encrypted': isEncrypted ? 1 : 0,
      'created_at': createdAt.millisecondsSinceEpoch ~/ 1000,
    };
  }
}
