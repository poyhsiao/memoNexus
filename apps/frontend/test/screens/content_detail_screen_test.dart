import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:memonexus_frontend/models/content_item.dart';
import 'package:memonexus_frontend/providers/content_provider.dart';
import 'package:memonexus_frontend/screens/content_detail_screen.dart';
import 'package:memonexus_frontend/services/api_client.dart';

// =====================================================
// Fake API Client for Content Detail
// =====================================================

class _FakeContentAPIClient extends MemoNexusAPIClient {
  bool shouldFail = false;
  String errorMessage = 'Test error';
  ContentItem? mockItem;
  List<String> mockKeywords = [];
  Map<String, dynamic> mockSummaryResult = {'method': 'tfidf'};
  List<String> updatedTags = [];

  @override
  Future<Map<String, dynamic>> getContentItem(String id) async {
    if (shouldFail) {
      throw APIException(
        code: 'CONTENT_ERROR',
        message: errorMessage,
        statusCode: 500,
      );
    }
    if (mockItem != null) {
      return mockItem!.toJson();
    }
    final now = DateTime.now().millisecondsSinceEpoch ~/ 1000;
    return {
      'id': id,
      'title': 'Test Content',
      'content_text': 'Test content body',
      'media_type': 'web',
      'tags': 'test,demo',
      'source_url': 'https://example.com',
      'created_at': now,
      'updated_at': now,
      'version': 1,
      'is_deleted': 0,
    };
  }

  @override
  Future<Map<String, dynamic>> updateContentItem(
    String id, {
    String? title,
    String? contentText,
    List<String>? tags,
  }) async {
    if (shouldFail) {
      throw APIException(
        code: 'UPDATE_ERROR',
        message: errorMessage,
        statusCode: 500,
      );
    }
    updatedTags = tags ?? [];
    final now = DateTime.now().millisecondsSinceEpoch ~/ 1000;
    return (mockItem ?? ContentItem(
      id: id,
      title: 'Test',
      contentText: 'Content',
      mediaType: MediaType.web,
      tags: const [],
      createdAt: DateTime.fromMillisecondsSinceEpoch(now * 1000),
      updatedAt: DateTime.fromMillisecondsSinceEpoch(now * 1000),
      version: 1,
    )).toJson();
  }

  @override
  Future<List<String>> extractKeywords(String id) async {
    if (shouldFail) {
      throw APIException(
        code: 'KEYWORD_ERROR',
        message: errorMessage,
        statusCode: 500,
      );
    }
    return mockKeywords;
  }

  @override
  Future<Map<String, dynamic>> generateSummary(String id) async {
    if (shouldFail) {
      throw APIException(
        code: 'SUMMARY_ERROR',
        message: errorMessage,
        statusCode: 500,
      );
    }
    return mockSummaryResult;
  }

  @override
  Future<List<Map<String, dynamic>>> listTags() async => [];
}

void main() {
  group('ContentDetailScreen Widget', () {
    late _FakeContentAPIClient fakeApi;
    final now = DateTime.now().millisecondsSinceEpoch ~/ 1000;

    setUp(() {
      fakeApi = _FakeContentAPIClient();
    });

    Widget buildTestWidget() {
      return ProviderScope(
        overrides: [
          apiClientProvider.overrideWith((ref) => fakeApi),
        ],
        child: const MaterialApp(
          home: ContentDetailScreen(itemId: 'test-id'),
        ),
      );
    }

    testWidgets('should show loading state initially', (tester) async {
      await tester.pumpWidget(buildTestWidget());

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      expect(find.text('Content Details'), findsOneWidget);
    });

    testWidgets('should display content when loaded', (tester) async {
      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();

      expect(find.text('Test Content'), findsOneWidget);
      expect(find.text('Content'), findsOneWidget);
      expect(find.text('Tags'), findsOneWidget);
      expect(find.text('Technical Metadata'), findsOneWidget);
    });

    testWidgets('should show error state on API failure', (tester) async {
      fakeApi.shouldFail = true;
      fakeApi.errorMessage = 'Network error';

      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.error_outline), findsOneWidget);
      expect(find.textContaining('Error loading content'), findsOneWidget);
      expect(find.text('Retry'), findsOneWidget);
    });

    testWidgets('should show metadata chips', (tester) async {
      fakeApi.mockItem = ContentItem(
        id: '1',
        title: 'Test',
        contentText: 'Content',
        mediaType: MediaType.web,
        tags: const [],
        createdAt: DateTime.fromMillisecondsSinceEpoch(now * 1000),
        updatedAt: DateTime.fromMillisecondsSinceEpoch(now * 1000),
        version: 1,
        contentHash: 'abc123',
      );

      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();

      expect(find.text('WEB'), findsOneWidget);
      expect(find.text('SHA-256'), findsOneWidget);
    });

    testWidgets('should show tags section', (tester) async {
      fakeApi.mockItem = ContentItem(
        id: '1',
        title: 'Test',
        contentText: 'Content',
        mediaType: MediaType.web,
        tags: ['flutter', 'dart'],
        createdAt: DateTime.fromMillisecondsSinceEpoch(now * 1000),
        updatedAt: DateTime.fromMillisecondsSinceEpoch(now * 1000),
        version: 1,
      );

      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();

      expect(find.text('Tags'), findsOneWidget);
      expect(find.text('flutter'), findsOneWidget);
      expect(find.text('dart'), findsOneWidget);
      expect(find.text('Edit Tags'), findsOneWidget);
    });

    testWidgets('should show no tags when item has no tags', (tester) async {
      fakeApi.mockItem = ContentItem(
        id: '1',
        title: 'Test',
        contentText: 'Content',
        mediaType: MediaType.web,
        tags: [],
        createdAt: DateTime.fromMillisecondsSinceEpoch(now * 1000),
        updatedAt: DateTime.fromMillisecondsSinceEpoch(now * 1000),
        version: 1,
      );

      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();

      expect(find.text('No tags'), findsOneWidget);
    });

    testWidgets('should enter edit mode when Edit Tags tapped', (tester) async {
      fakeApi.mockItem = ContentItem(
        id: '1',
        title: 'Test',
        contentText: 'Content',
        mediaType: MediaType.web,
        tags: ['test'],
        createdAt: DateTime.fromMillisecondsSinceEpoch(now * 1000),
        updatedAt: DateTime.fromMillisecondsSinceEpoch(now * 1000),
        version: 1,
      );

      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();

      await tester.tap(find.text('Edit Tags'));
      await tester.pump();

      expect(find.text('Cancel'), findsOneWidget);
      expect(find.text('Save Tags'), findsOneWidget);
    });

    testWidgets('should exit edit mode on Cancel', (tester) async {
      fakeApi.mockItem = ContentItem(
        id: '1',
        title: 'Test',
        contentText: 'Content',
        mediaType: MediaType.web,
        tags: ['test'],
        createdAt: DateTime.fromMillisecondsSinceEpoch(now * 1000),
        updatedAt: DateTime.fromMillisecondsSinceEpoch(now * 1000),
        version: 1,
      );

      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();

      await tester.tap(find.text('Edit Tags'));
      await tester.pump();
      await tester.tap(find.text('Cancel'));
      await tester.pump();

      expect(find.text('Edit Tags'), findsOneWidget);
    });

    testWidgets('should show summary section', (tester) async {
      fakeApi.mockItem = ContentItem(
        id: '1',
        title: 'Test',
        contentText: 'Content',
        mediaType: MediaType.web,
        tags: const [],
        summary: 'This is a summary',
        createdAt: DateTime.fromMillisecondsSinceEpoch(now * 1000),
        updatedAt: DateTime.fromMillisecondsSinceEpoch(now * 1000),
        version: 1,
      );

      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();

      expect(find.text('This is a summary'), findsOneWidget);
      expect(find.text('TF-IDF'), findsOneWidget);
    });

    testWidgets('should show keyword suggestions section', (tester) async {
      fakeApi.mockItem = ContentItem(
        id: '1',
        title: 'Test',
        contentText: 'Content',
        mediaType: MediaType.web,
        tags: const [],
        createdAt: DateTime.fromMillisecondsSinceEpoch(now * 1000),
        updatedAt: DateTime.fromMillisecondsSinceEpoch(now * 1000),
        version: 1,
      );

      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();

      // When keywords are empty, shows "No Keyword Suggestions" empty state
      expect(find.text('No Keyword Suggestions'), findsOneWidget);
    });

    testWidgets('should show content section', (tester) async {
      fakeApi.mockItem = ContentItem(
        id: '1',
        title: 'Test',
        contentText: 'Full content body text here',
        mediaType: MediaType.web,
        tags: const [],
        createdAt: DateTime.fromMillisecondsSinceEpoch(now * 1000),
        updatedAt: DateTime.fromMillisecondsSinceEpoch(now * 1000),
        version: 1,
      );

      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();

      expect(find.text('Content'), findsOneWidget);
      expect(find.text('Full content body text here'), findsOneWidget);
    });

    testWidgets('should show source URL when available', (tester) async {
      fakeApi.mockItem = ContentItem(
        id: '1',
        title: 'Test',
        contentText: 'Content',
        mediaType: MediaType.web,
        tags: const [],
        sourceUrl: 'https://example.com/article',
        createdAt: DateTime.fromMillisecondsSinceEpoch(now * 1000),
        updatedAt: DateTime.fromMillisecondsSinceEpoch(now * 1000),
        version: 1,
      );

      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();

      expect(find.text('Source'), findsOneWidget);
      expect(find.text('https://example.com/article'), findsOneWidget);
      expect(find.byIcon(Icons.link), findsOneWidget);
    });

    testWidgets('should show technical metadata tile', (tester) async {
      fakeApi.mockItem = ContentItem(
        id: 'item-123',
        title: 'Test',
        contentText: 'Content',
        mediaType: MediaType.web,
        tags: const [],
        summary: 'Test summary',
        contentHash: 'hash123',
        createdAt: DateTime.fromMillisecondsSinceEpoch(now * 1000),
        updatedAt: DateTime.fromMillisecondsSinceEpoch(now * 1000),
        version: 1,
      );

      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();

      expect(find.text('Technical Metadata'), findsOneWidget);
    });

    testWidgets('should have delete button in app bar', (tester) async {
      fakeApi.mockItem = ContentItem(
        id: '1',
        title: 'Test Content',
        contentText: 'Content',
        mediaType: MediaType.web,
        tags: const [],
        createdAt: DateTime.fromMillisecondsSinceEpoch(now * 1000),
        updatedAt: DateTime.fromMillisecondsSinceEpoch(now * 1000),
        version: 1,
      );

      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.delete), findsOneWidget);
    });
  });
}
