import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:memonexus_frontend/models/content_item.dart';
import 'package:memonexus_frontend/widgets/search_result_item.dart';

void main() {
  group('SearchResultItemWidget', () {
    final testItem = ContentItem(
      id: 'test-id',
      title: 'Test Article',
      contentText: 'This is a test article content about machine learning and artificial intelligence.',
      sourceUrl: 'https://example.com/article',
      mediaType: MediaType.web,
      tags: ['ml', 'ai'],
      createdAt: DateTime.now().subtract(const Duration(days: 2)),
      updatedAt: DateTime.now(),
      version: 1,
    );

    testWidgets('should display item title', (tester) async {
      final result = SearchResult(
        item: testItem,
        relevance: 1.5,
        matchedTerms: ['machine', 'learning'],
      );

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SearchResultItemWidget(
              result: result,
            ),
          ),
        ),
      );

      expect(find.text('Test Article'), findsOneWidget);
    });

    testWidgets('should display content preview', (tester) async {
      final result = SearchResult(
        item: testItem,
        relevance: 1.0,
        matchedTerms: const [],
      );

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SearchResultItemWidget(
              result: result,
            ),
          ),
        ),
      );

      // Should show truncated content
      expect(find.byType(Text), findsWidgets);
    });

    testWidgets('should display relevance badge', (tester) async {
      final result = SearchResult(
        item: testItem,
        relevance: 2.5,
        matchedTerms: const [],
      );

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SearchResultItemWidget(
              result: result,
            ),
          ),
        ),
      );

      // Should find the relevance score
      expect(find.text('2.5'), findsOneWidget);
    });

    testWidgets('should color relevance badge green for high scores', (tester) async {
      final result = SearchResult(
        item: testItem,
        relevance: 2.5,
        matchedTerms: const [],
      );

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SearchResultItemWidget(
              result: result,
            ),
          ),
        ),
      );

      final container = tester.widget<Container>(
        find.ancestor(
          of: find.text('2.5'),
          matching: find.byType(Container),
        ),
      );

      final decoration = container.decoration as BoxDecoration;
      expect(decoration.color, Colors.green.withValues(alpha: 0.1));
    });

    testWidgets('should color relevance badge orange for medium scores', (tester) async {
      final result = SearchResult(
        item: testItem,
        relevance: 1.2,
        matchedTerms: const [],
      );

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SearchResultItemWidget(
              result: result,
            ),
          ),
        ),
      );

      final container = tester.widget<Container>(
        find.ancestor(
          of: find.text('1.2'),
          matching: find.byType(Container),
        ),
      );

      final decoration = container.decoration as BoxDecoration;
      expect(decoration.color, Colors.orange.withValues(alpha: 0.1));
    });

    testWidgets('should display matched terms', (tester) async {
      final result = SearchResult(
        item: testItem,
        relevance: 1.0,
        matchedTerms: ['machine', 'learning'],
      );

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SearchResultItemWidget(
              result: result,
            ),
          ),
        ),
      );

      expect(find.text('machine'), findsOneWidget);
      expect(find.text('learning'), findsOneWidget);
    });

    testWidgets('should limit matched terms to 5', (tester) async {
      final result = SearchResult(
        item: testItem,
        relevance: 1.0,
        matchedTerms: ['term1', 'term2', 'term3', 'term4', 'term5', 'term6'],
      );

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SearchResultItemWidget(
              result: result,
            ),
          ),
        ),
      );

      // 2 tag chips + 5 matched term chips = 7 total, but only 5 matched terms shown
      expect(find.byType(Chip), findsNWidgets(7));
      // Verify that the 6th matched term is not shown
      expect(find.text('term6'), findsNothing);
    });

    testWidgets('should display tags', (tester) async {
      final result = SearchResult(
        item: testItem,
        relevance: 1.0,
        matchedTerms: const [],
      );

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SearchResultItemWidget(
              result: result,
            ),
          ),
        ),
      );

      expect(find.text('ml'), findsOneWidget);
      expect(find.text('ai'), findsOneWidget);
    });

    testWidgets('should limit tags to 3', (tester) async {
      final item = ContentItem(
        id: 'test-id',
        title: 'Test',
        contentText: 'Content',
        mediaType: MediaType.web,
        tags: ['tag1', 'tag2', 'tag3', 'tag4', 'tag5'],
        createdAt: DateTime.now(),
        updatedAt: DateTime.now(),
        version: 1,
      );

      final result = SearchResult(
        item: item,
        relevance: 1.0,
        matchedTerms: const [],
      );

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SearchResultItemWidget(
              result: result,
            ),
          ),
        ),
      );

      // Should only show 3 tags (no matched terms)
      expect(find.byType(Chip), findsNWidgets(3));
      expect(find.text('tag1'), findsOneWidget);
      expect(find.text('tag2'), findsOneWidget);
      expect(find.text('tag3'), findsOneWidget);
      // Verify that 4th and 5th tags are not shown
      expect(find.text('tag4'), findsNothing);
      expect(find.text('tag5'), findsNothing);
    });

    testWidgets('should display URL indicator for web content', (tester) async {
      final result = SearchResult(
        item: testItem,
        relevance: 1.0,
        matchedTerms: const [],
      );

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SearchResultItemWidget(
              result: result,
            ),
          ),
        ),
      );

      expect(find.text('URL'), findsOneWidget);
    });

    testWidgets('should not display URL indicator for non-web content', (tester) async {
      final item = ContentItem(
        id: 'test-id',
        title: 'Test',
        contentText: 'Content',
        mediaType: MediaType.markdown,
        tags: const [],
        createdAt: DateTime.now(),
        updatedAt: DateTime.now(),
        version: 1,
      );

      final result = SearchResult(
        item: item,
        relevance: 1.0,
        matchedTerms: const [],
      );

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SearchResultItemWidget(
              result: result,
            ),
          ),
        ),
      );

      expect(find.text('URL'), findsNothing);
    });

    testWidgets('should call onTap when tapped', (tester) async {
      var tapped = false;
      final result = SearchResult(
        item: testItem,
        relevance: 1.0,
        matchedTerms: const [],
      );

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SearchResultItemWidget(
              result: result,
              onTap: () => tapped = true,
            ),
          ),
        ),
      );

      // Find the Card widget and tap the InkWell inside it
      final cardFinder = find.byType(Card);
      await tester.tap(cardFinder);
      expect(tapped, isTrue);
    });

    testWidgets('should call onLongPress when long pressed', (tester) async {
      var longPressed = false;
      final result = SearchResult(
        item: testItem,
        relevance: 1.0,
        matchedTerms: const [],
      );

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SearchResultItemWidget(
              result: result,
              onLongPress: () => longPressed = true,
            ),
          ),
        ),
      );

      // Find the Card widget and long press it
      final cardFinder = find.byType(Card);
      await tester.pump();
      await tester.longPressAt(tester.getCenter(cardFinder));
      expect(longPressed, isTrue);
    });
  });

  group('MediaType Icons', () {
    testWidgets('should show language icon for web', (tester) async {
      final item = ContentItem(
        id: 'test-id',
        title: 'Test',
        contentText: 'Content',
        mediaType: MediaType.web,
        tags: const [],
        createdAt: DateTime.now(),
        updatedAt: DateTime.now(),
        version: 1,
      );

      final result = SearchResult(item: item, relevance: 1.0, matchedTerms: const []);

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SearchResultItemWidget(result: result),
          ),
        ),
      );

      expect(find.byIcon(Icons.language), findsOneWidget);
    });

    testWidgets('should show image icon for images', (tester) async {
      final item = ContentItem(
        id: 'test-id',
        title: 'Test',
        contentText: 'Content',
        mediaType: MediaType.image,
        tags: const [],
        createdAt: DateTime.now(),
        updatedAt: DateTime.now(),
        version: 1,
      );

      final result = SearchResult(item: item, relevance: 1.0, matchedTerms: const []);

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SearchResultItemWidget(result: result),
          ),
        ),
      );

      expect(find.byIcon(Icons.image), findsOneWidget);
    });

    testWidgets('should show video icon for videos', (tester) async {
      final item = ContentItem(
        id: 'test-id',
        title: 'Test',
        contentText: 'Content',
        mediaType: MediaType.video,
        tags: const [],
        createdAt: DateTime.now(),
        updatedAt: DateTime.now(),
        version: 1,
      );

      final result = SearchResult(item: item, relevance: 1.0, matchedTerms: const []);

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SearchResultItemWidget(result: result),
          ),
        ),
      );

      expect(find.byIcon(Icons.video_library), findsOneWidget);
    });

    testWidgets('should show PDF icon for PDFs', (tester) async {
      final item = ContentItem(
        id: 'test-id',
        title: 'Test',
        contentText: 'Content',
        mediaType: MediaType.pdf,
        tags: const [],
        createdAt: DateTime.now(),
        updatedAt: DateTime.now(),
        version: 1,
      );

      final result = SearchResult(item: item, relevance: 1.0, matchedTerms: const []);

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SearchResultItemWidget(result: result),
          ),
        ),
      );

      expect(find.byIcon(Icons.picture_as_pdf), findsOneWidget);
    });

    testWidgets('should show code icon for markdown', (tester) async {
      final item = ContentItem(
        id: 'test-id',
        title: 'Test',
        contentText: 'Content',
        mediaType: MediaType.markdown,
        tags: const [],
        createdAt: DateTime.now(),
        updatedAt: DateTime.now(),
        version: 1,
      );

      final result = SearchResult(item: item, relevance: 1.0, matchedTerms: const []);

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SearchResultItemWidget(result: result),
          ),
        ),
      );

      expect(find.byIcon(Icons.code), findsOneWidget);
    });
  });

  group('Date Formatting', () {
    testWidgets('should show "Today" for same day', (tester) async {
      final item = ContentItem(
        id: 'test-id',
        title: 'Test',
        contentText: 'Content',
        mediaType: MediaType.web,
        tags: const [],
        createdAt: DateTime.now(),
        updatedAt: DateTime.now(),
        version: 1,
      );

      final result = SearchResult(item: item, relevance: 1.0, matchedTerms: const []);

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SearchResultItemWidget(result: result),
          ),
        ),
      );

      expect(find.text('Today'), findsOneWidget);
    });

    testWidgets('should show "Yesterday" for previous day', (tester) async {
      final item = ContentItem(
        id: 'test-id',
        title: 'Test',
        contentText: 'Content',
        mediaType: MediaType.web,
        tags: const [],
        createdAt: DateTime.now().subtract(const Duration(days: 1)),
        updatedAt: DateTime.now(),
        version: 1,
      );

      final result = SearchResult(item: item, relevance: 1.0, matchedTerms: const []);

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SearchResultItemWidget(result: result),
          ),
        ),
      );

      expect(find.text('Yesterday'), findsOneWidget);
    });

    testWidgets('should show "X days ago" for recent items', (tester) async {
      final item = ContentItem(
        id: 'test-id',
        title: 'Test',
        contentText: 'Content',
        mediaType: MediaType.web,
        tags: const [],
        createdAt: DateTime.now().subtract(const Duration(days: 3)),
        updatedAt: DateTime.now(),
        version: 1,
      );

      final result = SearchResult(item: item, relevance: 1.0, matchedTerms: const []);

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SearchResultItemWidget(result: result),
          ),
        ),
      );

      expect(find.text('3d ago'), findsOneWidget);
    });

    testWidgets('should show "Xw ago" for older items', (tester) async {
      final item = ContentItem(
        id: 'test-id',
        title: 'Test',
        contentText: 'Content',
        mediaType: MediaType.web,
        tags: const [],
        createdAt: DateTime.now().subtract(const Duration(days: 14)),
        updatedAt: DateTime.now(),
        version: 1,
      );

      final result = SearchResult(item: item, relevance: 1.0, matchedTerms: const []);

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SearchResultItemWidget(result: result),
          ),
        ),
      );

      expect(find.text('2w ago'), findsOneWidget);
    });

    testWidgets('should show "Xmo ago" for much older items', (tester) async {
      final item = ContentItem(
        id: 'test-id',
        title: 'Test',
        contentText: 'Content',
        mediaType: MediaType.web,
        tags: const [],
        createdAt: DateTime.now().subtract(const Duration(days: 60)),
        updatedAt: DateTime.now(),
        version: 1,
      );

      final result = SearchResult(item: item, relevance: 1.0, matchedTerms: const []);

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SearchResultItemWidget(result: result),
          ),
        ),
      );

      expect(find.text('2mo ago'), findsOneWidget);
    });
  });
}
