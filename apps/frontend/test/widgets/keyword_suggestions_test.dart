import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:memonexus_frontend/widgets/keyword_suggestions.dart';

void main() {
  group('KeywordSuggestions Widget', () {
    testWidgets('should show loading state when isLoading is true',
        (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: KeywordSuggestions(
              keywords: const [],
              onKeywordSelected: (_) {},
              isLoading: true,
            ),
          ),
        ),
      );

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      expect(find.text('Extracting keywords...'), findsOneWidget);
    });

    testWidgets('should show empty state when keywords is empty',
        (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: KeywordSuggestions(
              keywords: const [],
              onKeywordSelected: (_) {},
              method: 'tfidf',
            ),
          ),
        ),
      );

      expect(find.text('No Keyword Suggestions'), findsOneWidget);
      expect(
        find.textContaining('Extract keywords using TF-IDF'),
        findsOneWidget,
      );
    });

    testWidgets('should show keywords when available', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: KeywordSuggestions(
              keywords: const ['flutter', 'dart', 'mobile'],
              onKeywordSelected: (_) {},
            ),
          ),
        ),
      );

      expect(find.text('Keyword Suggestions'), findsOneWidget);
      expect(find.text('flutter'), findsOneWidget);
      expect(find.text('dart'), findsOneWidget);
      expect(find.text('mobile'), findsOneWidget);
    });

    testWidgets('should call onKeywordSelected when keyword tapped',
        (tester) async {
      String? selectedKeyword;

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: KeywordSuggestions(
              keywords: const ['flutter', 'dart'],
              onKeywordSelected: (keyword) => selectedKeyword = keyword,
            ),
          ),
        ),
      );

      await tester.tap(find.text('flutter'));
      await tester.pump();

      expect(selectedKeyword, 'flutter');
    });

    testWidgets('should show refresh button when onRefresh provided',
        (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: KeywordSuggestions(
              keywords: const ['flutter'],
              onKeywordSelected: (_) {},
              onRefresh: () {},
            ),
          ),
        ),
      );

      expect(find.byIcon(Icons.refresh), findsOneWidget);
    });

    testWidgets('should show correct method badge for AI', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: KeywordSuggestions(
              keywords: const ['flutter'],
              onKeywordSelected: (_) {},
              method: 'ai',
            ),
          ),
        ),
      );

      expect(find.text('AI'), findsOneWidget);
      expect(find.byIcon(Icons.psychology), findsOneWidget);
    });

    testWidgets('should show correct method badge for TextRank',
        (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: KeywordSuggestions(
              keywords: const ['flutter'],
              onKeywordSelected: (_) {},
              method: 'textrank',
            ),
          ),
        ),
      );

      expect(find.text('TextRank'), findsOneWidget);
      expect(find.byIcon(Icons.graphic_eq), findsOneWidget);
    });

    testWidgets('should show correct method badge for TF-IDF',
        (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: KeywordSuggestions(
              keywords: const ['flutter'],
              onKeywordSelected: (_) {},
              method: 'tfidf',
            ),
          ),
        ),
      );

      expect(find.text('TF-IDF'), findsOneWidget);
      expect(find.byIcon(Icons.text_fields), findsOneWidget);
    });

    testWidgets('should show empty state with Extract button when onRefresh',
        (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: KeywordSuggestions(
              keywords: const [],
              onKeywordSelected: (_) {},
              onRefresh: () {},
              method: 'ai',
            ),
          ),
        ),
      );

      expect(find.text('No Keyword Suggestions'), findsOneWidget);
      expect(
        find.textContaining('Extract keywords using AI'),
        findsOneWidget,
      );
      expect(find.text('Extract Keywords'), findsOneWidget);
      expect(find.byIcon(Icons.auto_awesome), findsOneWidget);
    });
  });

  group('SelectedKeywords Widget', () {
    testWidgets('should be empty when no keywords', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SelectedKeywords(
              keywords: const [],
              onRemove: (_) {},
            ),
          ),
        ),
      );

      expect(find.byType(SizedBox), findsOneWidget);
    });

    testWidgets('should show selected keywords', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SelectedKeywords(
              keywords: const ['flutter', 'dart'],
              onRemove: (_) {},
            ),
          ),
        ),
      );

      expect(find.text('Selected (2)'), findsOneWidget);
      expect(find.text('flutter'), findsOneWidget);
      expect(find.text('dart'), findsOneWidget);
      expect(find.byType(Chip), findsNWidgets(2));
    });

    testWidgets('should call onRemove when chip deleted', (tester) async {
      String? removedKeyword;

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SelectedKeywords(
              keywords: const ['flutter'],
              onRemove: (keyword) => removedKeyword = keyword,
            ),
          ),
        ),
      );

      // Find the delete button (Chip uses onDeleted)
      final chip = tester.widget<Chip>(find.byType(Chip));
      chip.onDeleted?.call();

      expect(removedKeyword, 'flutter');
    });
  });
}
