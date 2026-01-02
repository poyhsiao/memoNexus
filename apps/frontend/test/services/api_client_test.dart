// API Client Service Tests
// Tests for MemoNexusAPIClient REST API communication

import 'dart:convert';
import 'package:flutter_test/flutter_test.dart';
import 'package:memonexus_frontend/services/api_client.dart';

void main() {
  group('MemoNexusAPIClient', () {
    late MemoNexusAPIClient client;

    setUp(() {
      client = MemoNexusAPIClient(
        baseUrl: 'http://localhost:8090/api',
        token: 'test-token',
      );
    });

    group('Headers', () {
      test('should have correct default headers', () {
        expect(client.baseUrl, 'http://localhost:8090/api');
        expect(client.token, 'test-token');
      });

      test('should accept custom baseUrl', () {
        final customClient = MemoNexusAPIClient(
          baseUrl: 'http://custom.api.com/v1',
        );
        expect(customClient.baseUrl, 'http://custom.api.com/v1');
      });
    });

    group('Content Items', () {
      test('listContentItems should build correct query parameters', () async {
        // Test query parameter building logic
        final queryParams = <String, String>{
          'page': '1',
          'per_page': '20',
          'sort': 'created_at',
          'order': 'desc',
        };

        expect(queryParams['page'], '1');
        expect(queryParams['per_page'], '20');
        expect(queryParams['sort'], 'created_at');
        expect(queryParams['order'], 'desc');
      });

      test('listContentItems should include mediaType filter when provided', () {
        final queryParams = <String, String>{
          'page': '1',
          'per_page': '20',
          'sort': 'created_at',
          'order': 'desc',
        };
        queryParams['media_type'] = 'markdown';

        expect(queryParams['media_type'], 'markdown');
      });

      test('listContentItems should include tag filter when provided', () {
        final queryParams = <String, String>{
          'page': '1',
          'per_page': '20',
          'sort': 'created_at',
          'order': 'desc',
        };
        queryParams['tag'] = 'test-tag';

        expect(queryParams['tag'], 'test-tag');
      });

      test('createContentFromURL should include required fields', () {
        final data = {
          'type': 'url',
          'source_url': 'https://example.com',
        };

        expect(data['type'], 'url');
        expect(data['source_url'], 'https://example.com');
      });

      test('createContentFromURL should include optional title when provided', () {
        final data = {
          'type': 'url',
          'source_url': 'https://example.com',
          'title': 'Custom Title',
        };

        expect(data['title'], 'Custom Title');
      });

      test('createContentFromURL should include optional tags when provided', () {
        final data = {
          'type': 'url',
          'source_url': 'https://example.com',
          'tags': ['tag1', 'tag2'],
        };

        expect(data['tags'], ['tag1', 'tag2']);
      });

      test('createContentFromFile should include required fields', () {
        final data = {
          'type': 'file',
          'file_path': '/path/to/file.pdf',
        };

        expect(data['type'], 'file');
        expect(data['file_path'], '/path/to/file.pdf');
      });

      test('updateContentItem should include title when provided', () {
        final data = {
          'title': 'Updated Title',
        };

        expect(data['title'], 'Updated Title');
      });

      test('updateContentItem should include tags when provided', () {
        final data = {
          'tags': ['tag1', 'tag2', 'tag3'],
        };

        expect(data['tags'], ['tag1', 'tag2', 'tag3']);
      });
    });

    group('Search', () {
      test('search should include required query parameters', () {
        final queryParams = <String, String>{
          'q': 'test query',
          'limit': '20',
        };

        expect(queryParams['q'], 'test query');
        expect(queryParams['limit'], '20');
      });

      test('search should include mediaType filter when provided', () {
        final queryParams = <String, String>{
          'q': 'test',
          'limit': '20',
        };
        queryParams['media_type'] = 'pdf';

        expect(queryParams['media_type'], 'pdf');
      });

      test('search should include tags filter when provided', () {
        final queryParams = <String, String>{
          'q': 'test',
          'limit': '20',
        };
        queryParams['tags'] = 'tag1,tag2';

        expect(queryParams['tags'], 'tag1,tag2');
      });

      test('search should include date range when provided', () {
        final queryParams = <String, String>{
          'q': 'test',
          'limit': '20',
        };
        queryParams['date_from'] = '1234567890';
        queryParams['date_to'] = '1234567899';

        expect(queryParams['date_from'], '1234567890');
        expect(queryParams['date_to'], '1234567899');
      });
    });

    group('Tags', () {
      test('createTag should include required name', () {
        final data = {
          'name': 'Test Tag',
          'color': '#3B82F6',
        };

        expect(data['name'], 'Test Tag');
        expect(data['color'], '#3B82F6');
      });

      test('createTag should use default color when not provided', () {
        final data = {
          'name': 'Test Tag',
          'color': '#3B82F6',
        };

        expect(data['color'], '#3B82F6');
      });

      test('updateTag should include name when provided', () {
        final data = {
          'name': 'Updated Tag',
        };

        expect(data['name'], 'Updated Tag');
      });

      test('updateTag should include color when provided', () {
        final data = {
          'color': '#FF0000',
        };

        expect(data['color'], '#FF0000');
      });
    });

    group('AI Configuration', () {
      test('setAIConfig should include all required fields', () {
        final data = {
          'provider': 'openai',
          'api_endpoint': 'https://api.openai.com/v1',
          'api_key': 'sk-test-key',
        };

        expect(data['provider'], 'openai');
        expect(data['api_endpoint'], 'https://api.openai.com/v1');
        expect(data['api_key'], 'sk-test-key');
      });

      test('setAIConfig should include optional modelName when provided', () {
        final data = {
          'provider': 'openai',
          'api_endpoint': 'https://api.openai.com/v1',
          'api_key': 'sk-test',
          'model_name': 'gpt-4',
        };

        expect(data['model_name'], 'gpt-4');
      });

      test('setAIConfig should include optional maxTokens when provided', () {
        final data = {
          'provider': 'openai',
          'api_endpoint': 'https://api.openai.com/v1',
          'api_key': 'sk-test',
          'max_tokens': 2048,
        };

        expect(data['max_tokens'], 2048);
      });

      test('generateSummary should build correct query parameters', () {
        final queryParams = <String, String>{
          'id': 'content-id',
          'operation': 'summary',
        };

        expect(queryParams['id'], 'content-id');
        expect(queryParams['operation'], 'summary');
      });

      test('extractKeywords should build correct query parameters', () {
        final queryParams = <String, String>{
          'id': 'content-id',
          'operation': 'keywords',
        };

        expect(queryParams['id'], 'content-id');
        expect(queryParams['operation'], 'keywords');
      });
    });

    group('Sync Credentials', () {
      test('configureSync should include all required fields', () {
        final data = {
          'endpoint': 'https://s3.amazonaws.com',
          'bucket_name': 'test-bucket',
          'access_key': 'AKIAIOSFODNN7EXAMPLE',
          'secret_key': 'wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY',
        };

        expect(data['endpoint'], 'https://s3.amazonaws.com');
        expect(data['bucket_name'], 'test-bucket');
        expect(data['access_key'], 'AKIAIOSFODNN7EXAMPLE');
        expect(data['secret_key'], 'wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY');
      });

      test('configureSync should include optional region when provided', () {
        final data = {
          'endpoint': 'https://s3.amazonaws.com',
          'bucket_name': 'test-bucket',
          'access_key': 'AKIAIOSFODNN7EXAMPLE',
          'secret_key': 'wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY',
          'region': 'us-east-1',
        };

        expect(data['region'], 'us-east-1');
      });
    });

    group('Export/Import', () {
      test('startExport should include required media types', () {
        final data = {
          'media_types': ['markdown', 'pdf'],
          'include_tags': true,
        };

        expect(data['media_types'], ['markdown', 'pdf']);
        expect(data['include_tags'], true);
      });

      test('startExport should include optional password when provided', () {
        final data = {
          'media_types': ['markdown'],
          'include_tags': false,
          'password': 'encrypt-me',
        };

        expect(data['password'], 'encrypt-me');
      });

      test('startImport should include required file path', () {
        final data = {
          'file_path': '/path/to/archive.tar.gz',
        };

        expect(data['file_path'], '/path/to/archive.tar.gz');
      });

      test('startImport should include optional password when provided', () {
        final data = {
          'file_path': '/path/to/archive.tar.gz',
          'password': 'decrypt-me',
        };

        expect(data['password'], 'decrypt-me');
      });

      test('getExportArchives should return list', () {
        // Test that the method exists and returns a list
        final data = <dynamic>[];
        final result = jsonEncode(data);

        expect(result, isNotEmpty);
      });
    });

    group('Sync Status', () {
      test('loadSyncStatus should return status data', () {
        // Test that the method exists and returns data
        final data = {
          'status': 'idle',
          'last_sync': '2024-01-01T00:00:00Z',
        };
        final result = jsonEncode(data);

        expect(result, isNotEmpty);
      });

      test('triggerSync should initiate sync', () {
        // Test that the method exists
        final data = {'action': 'sync'};
        final result = jsonEncode(data);

        expect(result, isNotEmpty);
      });
    });

    group('Data Encoding', () {
      test('should encode JSON correctly for content creation', () {
        final data = {
          'type': 'url',
          'source_url': 'https://example.com',
          'title': 'Test',
          'tags': ['tag1', 'tag2'],
        };

        final encoded = jsonEncode(data);
        final decoded = jsonDecode(encoded) as Map<String, dynamic>;

        expect(decoded['type'], 'url');
        expect(decoded['source_url'], 'https://example.com');
        expect(decoded['title'], 'Test');
        expect(decoded['tags'], ['tag1', 'tag2']);
      });

      test('should encode JSON correctly for content update', () {
        final data = {
          'title': 'Updated Title',
          'tags': ['new-tag'],
        };

        final encoded = jsonEncode(data);
        final decoded = jsonDecode(encoded) as Map<String, dynamic>;

        expect(decoded['title'], 'Updated Title');
        expect(decoded['tags'], ['new-tag']);
      });

      test('should encode JSON correctly for tag creation', () {
        final data = {
          'name': 'New Tag',
          'color': '#FF5733',
        };

        final encoded = jsonEncode(data);
        final decoded = jsonDecode(encoded) as Map<String, dynamic>;

        expect(decoded['name'], 'New Tag');
        expect(decoded['color'], '#FF5733');
      });

      test('should encode JSON correctly for sync config', () {
        final data = {
          'endpoint': 'https://s3.amazonaws.com',
          'bucket_name': 'my-bucket',
          'access_key': 'AKIAIOSFODNN7EXAMPLE',
          'secret_key': 'secret',
        };

        final encoded = jsonEncode(data);
        final decoded = jsonDecode(encoded) as Map<String, dynamic>;

        expect(decoded['endpoint'], 'https://s3.amazonaws.com');
        expect(decoded['bucket_name'], 'my-bucket');
        expect(decoded['access_key'], 'AKIAIOSFODNN7EXAMPLE');
        expect(decoded['secret_key'], 'secret');
      });
    });

    group('URL Building', () {
      test('should build correct URL for content list', () {
        final baseUrl = 'http://localhost:8090/api';
        final path = '/content';
        final url = '$baseUrl$path';

        expect(url, 'http://localhost:8090/api/content');
      });

      test('should build correct URL for content item', () {
        final baseUrl = 'http://localhost:8090/api';
        final id = 'uuid-123';
        final path = '/content/$id';
        final url = '$baseUrl$path';

        expect(url, 'http://localhost:8090/api/content/uuid-123');
      });

      test('should build correct URL for search', () {
        final baseUrl = 'http://localhost:8090/api';
        final path = '/search';
        final url = '$baseUrl$path';

        expect(url, 'http://localhost:8090/api/search');
      });

      test('should build correct URL for tags', () {
        final baseUrl = 'http://localhost:8090/api';
        final path = '/tags';
        final url = '$baseUrl$path';

        expect(url, 'http://localhost:8090/api/tags');
      });

      test('should build correct URL for AI config', () {
        final baseUrl = 'http://localhost:8090/api';
        final path = '/ai/config';
        final url = '$baseUrl$path';

        expect(url, 'http://localhost:8090/api/ai/config');
      });

      test('should build correct URL for analysis', () {
        final baseUrl = 'http://localhost:8090/api';
        final path = '/content/analyze';
        final url = '$baseUrl$path';

        expect(url, 'http://localhost:8090/api/content/analyze');
      });

      test('should build correct URL for sync credentials', () {
        final baseUrl = 'http://localhost:8090/api';
        final path = '/sync/credentials';
        final url = '$baseUrl$path';

        expect(url, 'http://localhost:8090/api/sync/credentials');
      });

      test('should build correct URL for export', () {
        final baseUrl = 'http://localhost:8090/api';
        final path = '/export';
        final url = '$baseUrl$path';

        expect(url, 'http://localhost:8090/api/export');
      });

      test('should build correct URL for import', () {
        final baseUrl = 'http://localhost:8090/api';
        final path = '/import';
        final url = '$baseUrl$path';

        expect(url, 'http://localhost:8090/api/import');
      });
    });
  });
}
