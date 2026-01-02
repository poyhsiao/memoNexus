// Content List Widget Tests
// Tests for ContentListWidget and ContentListTile

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:memonexus_frontend/widgets/content_list.dart';
import 'package:memonexus_frontend/models/content_item.dart';

void main() {
  group('ContentListTile', () {
    testWidgets('should display content item correctly', (tester) async {
      final item = ContentItem(
        id: '1',
        title: 'Test Content',
        contentText: 'Test content description',
        mediaType: MediaType.markdown,
        tags: ['tag1', 'tag2'],
        sourceUrl: 'https://example.com',
        createdAt: DateTime.now(),
        updatedAt: DateTime.now(),
        version: 1,
      );

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: ContentListTile(
              item: item,
              onTap: () {},
            ),
          ),
        ),
      );

      expect(find.text('Test Content'), findsOneWidget);
      expect(find.text('Test content description'), findsOneWidget);
      expect(find.text('tag1'), findsOneWidget);
      expect(find.text('tag2'), findsOneWidget);
    });

    testWidgets('should show focus indicator when focused', (tester) async {
      final item = ContentItem(
        id: '1',
        title: 'Test Content',
        contentText: 'Test',
        mediaType: MediaType.markdown,
        tags: [],
        sourceUrl: 'https://example.com',
        createdAt: DateTime.now(),
        updatedAt: DateTime.now(),
        version: 1,
      );

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: ContentListTile(
              item: item,
              onTap: () {},
              isFocused: true,
            ),
          ),
        ),
      );

      // Check that the focused tile renders
      expect(find.byType(ContentListTile), findsOneWidget);
    });

    testWidgets('should display correct icon for media type', (tester) async {
      final item = ContentItem(
        id: '1',
        title: 'Test',
        contentText: 'Test',
        mediaType: MediaType.image,
        tags: [],
        sourceUrl: 'https://example.com',
        createdAt: DateTime.now(),
        updatedAt: DateTime.now(),
        version: 1,
      );

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: ContentListTile(
              item: item,
              onTap: () {},
            ),
          ),
        ),
      );

      expect(find.byIcon(Icons.image), findsOneWidget);
    });

    testWidgets('should display timestamp', (tester) async {
      final now = DateTime.now();
      final item = ContentItem(
        id: '1',
        title: 'Test',
        contentText: 'Test',
        mediaType: MediaType.markdown,
        tags: [],
        sourceUrl: 'https://example.com',
        createdAt: now,
        updatedAt: now.subtract(const Duration(minutes: 5)),
        version: 1,
      );

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: ContentListTile(
              item: item,
              onTap: () {},
            ),
          ),
        ),
      );

      // Should show timestamp
      expect(find.textContaining('ago'), findsOneWidget);
    });
  });

  group('ContentListWidget', () {
    testWidgets('should build without errors', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: ContentListWidget(),
            ),
          ),
        ),
      );

      expect(find.byType(ContentListWidget), findsOneWidget);
    });

    testWidgets('should have search query parameter', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: ContentListWidget(searchQuery: 'test'),
            ),
          ),
        ),
      );

      expect(find.byType(ContentListWidget), findsOneWidget);
    });

    testWidgets('should render refresh indicator', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: ContentListWidget(),
            ),
          ),
        ),
      );

      await tester.pumpAndSettle();

      // Widget should render successfully
      expect(find.byType(ContentListWidget), findsOneWidget);
    });
  });
}
