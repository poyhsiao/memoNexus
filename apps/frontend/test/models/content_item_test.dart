import 'package:flutter_test/flutter_test.dart';
import 'package:memonexus_frontend/models/content_item.dart';

void main() {
  group('ContentItem', () {
    test('should create ContentItem with required fields', () {
      final now = DateTime.now();
      final item = ContentItem(
        title: 'Test Title',
        contentText: 'Test Content',
        mediaType: MediaType.web,
        tags: const ['test', 'example'],
        createdAt: now,
        updatedAt: now,
        version: 1,
      );

      expect(item.title, 'Test Title');
      expect(item.contentText, 'Test Content');
      expect(item.mediaType, MediaType.web);
      expect(item.tags, ['test', 'example']);
      expect(item.isDeleted, false);
      expect(item.version, 1);
    });

    test('should generate UUID if not provided', () {
      final now = DateTime.now();
      final item1 = ContentItem(
        title: 'Test',
        contentText: 'Content',
        mediaType: MediaType.web,
        tags: const [],
        createdAt: now,
        updatedAt: now,
        version: 1,
      );

      final item2 = ContentItem(
        title: 'Test',
        contentText: 'Content',
        mediaType: MediaType.web,
        tags: const [],
        createdAt: now,
        updatedAt: now,
        version: 1,
      );

      expect(item1.id, isNotEmpty);
      expect(item2.id, isNotEmpty);
      expect(item1.id, isNot(item2.id));
    });

    test('should serialize to JSON correctly', () {
      final now = DateTime.utc(2024, 1, 1, 12, 0, 0);
      final item = ContentItem(
        id: 'test-id',
        title: 'Test Title',
        contentText: 'Test Content',
        sourceUrl: 'https://example.com',
        mediaType: MediaType.web,
        tags: const ['test', 'example'],
        summary: 'Test Summary',
        isDeleted: false,
        createdAt: now,
        updatedAt: now,
        version: 1,
        contentHash: 'abc123',
      );

      final json = item.toJson();

      expect(json['id'], 'test-id');
      expect(json['title'], 'Test Title');
      expect(json['content_text'], 'Test Content');
      expect(json['source_url'], 'https://example.com');
      expect(json['media_type'], 'web');
      expect(json['tags'], 'test,example');
      expect(json['summary'], 'Test Summary');
      expect(json['is_deleted'], 0);
      expect(json['version'], 1);
      expect(json['content_hash'], 'abc123');
    });

    test('should deserialize from JSON correctly', () {
      final json = {
        'id': 'test-id',
        'title': 'Test Title',
        'content_text': 'Test Content',
        'source_url': 'https://example.com',
        'media_type': 'web',
        'tags': 'test,example',
        'summary': 'Test Summary',
        'is_deleted': 0,
        'created_at': 1704110400,
        'updated_at': 1704110400,
        'version': 1,
        'content_hash': 'abc123',
      };

      final item = ContentItem.fromJson(json);

      expect(item.id, 'test-id');
      expect(item.title, 'Test Title');
      expect(item.contentText, 'Test Content');
      expect(item.sourceUrl, 'https://example.com');
      expect(item.mediaType, MediaType.web);
      expect(item.tags, ['test', 'example']);
      expect(item.summary, 'Test Summary');
      expect(item.isDeleted, false);
      expect(item.version, 1);
      expect(item.contentHash, 'abc123');
    });

    test('should handle empty tags in JSON', () {
      final json = {
        'id': 'test-id',
        'title': 'Test',
        'content_text': 'Content',
        'media_type': 'web',
        'tags': '',
        'created_at': 1704110400,
        'updated_at': 1704110400,
        'version': 1,
      };

      final item = ContentItem.fromJson(json);

      expect(item.tags, isEmpty);
    });

    test('should handle tags with extra commas in JSON', () {
      final json = {
        'id': 'test-id',
        'title': 'Test',
        'content_text': 'Content',
        'media_type': 'web',
        'tags': 'test,,example',
        'created_at': 1704110400,
        'updated_at': 1704110400,
        'version': 1,
      };

      final item = ContentItem.fromJson(json);

      expect(item.tags, ['test', 'example']);
    });

    test('should deserialize is_deleted correctly', () {
      final now = DateTime.utc(2024, 1, 1);
      final json = {
        'id': 'test-id',
        'title': 'Test',
        'content_text': 'Content',
        'media_type': 'web',
        'tags': '',
        'is_deleted': 1,
        'created_at': 1704110400,
        'updated_at': 1704110400,
        'version': 1,
      };

      final item = ContentItem.fromJson(json);

      expect(item.isDeleted, true);
    });

    test('should copy with updated fields', () {
      final now = DateTime.now();
      final item = ContentItem(
        id: 'test-id',
        title: 'Original Title',
        contentText: 'Original Content',
        mediaType: MediaType.web,
        tags: const ['original'],
        createdAt: now,
        updatedAt: now,
        version: 1,
      );

      final updated = item.copyWith(
        title: 'Updated Title',
        tags: const ['updated'],
      );

      expect(updated.id, 'test-id');
      expect(updated.title, 'Updated Title');
      expect(updated.contentText, 'Original Content');
      expect(updated.tags, ['updated']);
      expect(updated.version, 2); // version should increment
    });

    test('should support all MediaType values', () {
      final types = MediaType.values;
      expect(types.length, greaterThan(0));
      expect(types.contains(MediaType.web), true);
      expect(types.contains(MediaType.image), true);
      expect(types.contains(MediaType.video), true);
      expect(types.contains(MediaType.pdf), true);
      expect(types.contains(MediaType.markdown), true);
    });
  });

  group('Tag', () {
    test('should create Tag with required fields', () {
      final now = DateTime.now();
      final tag = Tag(
        name: 'Test Tag',
        createdAt: now,
        updatedAt: now,
      );

      expect(tag.name, 'Test Tag');
      expect(tag.color, '#3B82F6'); // default color
      expect(tag.isDeleted, false);
    });

    test('should generate UUID if not provided', () {
      final now = DateTime.now();
      final tag1 = Tag(
        name: 'Test',
        createdAt: now,
        updatedAt: now,
      );

      final tag2 = Tag(
        name: 'Test',
        createdAt: now,
        updatedAt: now,
      );

      expect(tag1.id, isNotEmpty);
      expect(tag2.id, isNotEmpty);
      expect(tag1.id, isNot(tag2.id));
    });

    test('should serialize to JSON correctly', () {
      final now = DateTime.utc(2024, 1, 1);
      final tag = Tag(
        id: 'tag-id',
        name: 'Test Tag',
        color: '#FF5733',
        isDeleted: false,
        createdAt: now,
        updatedAt: now,
      );

      final json = tag.toJson();

      expect(json['id'], 'tag-id');
      expect(json['name'], 'Test Tag');
      expect(json['color'], '#FF5733');
      expect(json['is_deleted'], 0);
    });

    test('should deserialize from JSON correctly', () {
      final json = {
        'id': 'tag-id',
        'name': 'Test Tag',
        'color': '#FF5733',
        'is_deleted': 0,
        'created_at': 1704110400,
        'updated_at': 1704110400,
      };

      final tag = Tag.fromJson(json);

      expect(tag.id, 'tag-id');
      expect(tag.name, 'Test Tag');
      expect(tag.color, '#FF5733');
      expect(tag.isDeleted, false);
    });

    test('should use default color when not provided in JSON', () {
      final json = {
        'id': 'tag-id',
        'name': 'Test Tag',
        'is_deleted': 0,
        'created_at': 1704110400,
        'updated_at': 1704110400,
      };

      final tag = Tag.fromJson(json);

      expect(tag.color, '#3B82F6');
    });
  });

  group('SearchResult', () {
    test('should create SearchResult with required fields', () {
      final now = DateTime.now();
      final item = ContentItem(
        title: 'Test',
        contentText: 'Content',
        mediaType: MediaType.web,
        tags: const [],
        createdAt: now,
        updatedAt: now,
        version: 1,
      );

      final result = SearchResult(
        item: item,
        relevance: 0.95,
        matchedTerms: const ['test', 'search'],
      );

      expect(result.item, item);
      expect(result.relevance, 0.95);
      expect(result.matchedTerms, ['test', 'search']);
    });

    test('should deserialize from JSON correctly', () {
      final json = {
        'item': {
          'id': 'test-id',
          'title': 'Test',
          'content_text': 'Content',
          'media_type': 'web',
          'tags': '',
          'created_at': 1704110400,
          'updated_at': 1704110400,
          'version': 1,
        },
        'relevance': 0.85,
        'matched_terms': ['test', 'match'],
      };

      final result = SearchResult.fromJson(json);

      expect(result.item.title, 'Test');
      expect(result.relevance, 0.85);
      expect(result.matchedTerms, ['test', 'match']);
    });
  });
}
