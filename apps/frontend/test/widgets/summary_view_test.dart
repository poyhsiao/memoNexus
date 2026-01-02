import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:memonexus_frontend/widgets/summary_view.dart';

void main() {
  group('SummaryView Widget', () {
    testWidgets('should be empty when summary is empty', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: SummaryView(
              summary: '',
            ),
          ),
        ),
      );

      expect(find.byType(SizedBox), findsOneWidget);
    });

    testWidgets('should display summary text', (tester) async {
      const testSummary = 'This is a test summary of the content.';

      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: SummaryView(
              summary: testSummary,
              method: 'tfidf',
            ),
          ),
        ),
      );

      expect(find.text(testSummary), findsOneWidget);
      expect(find.text('Summary'), findsOneWidget);
    });

    testWidgets('should show TF-IDF method badge', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: SummaryView(
              summary: 'Test summary',
              method: 'tfidf',
            ),
          ),
        ),
      );

      expect(find.text('TF-IDF'), findsOneWidget);
      expect(find.byIcon(Icons.summarize), findsOneWidget);
    });

    testWidgets('should show AI method badge when aiUsed is true',
        (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: SummaryView(
              summary: 'AI generated summary',
              method: 'ai',
              aiUsed: true,
            ),
          ),
        ),
      );

      expect(find.text('AI'), findsOneWidget);
      expect(find.byIcon(Icons.psychology), findsWidgets);
      expect(find.text('AI Generated'), findsOneWidget);
    });

    testWidgets('should show TextRank method badge', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: SummaryView(
              summary: 'Test summary',
              method: 'textrank',
            ),
          ),
        ),
      );

      expect(find.text('TextRank'), findsOneWidget);
      expect(find.byIcon(Icons.graphic_eq), findsOneWidget);
    });

    testWidgets('should show language metadata when provided',
        (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: SummaryView(
              summary: 'Test summary',
              language: 'en',
            ),
          ),
        ),
      );

      expect(find.text('English'), findsOneWidget);
      expect(find.byIcon(Icons.language), findsOneWidget);
    });

    testWidgets('should show CJK language', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: SummaryView(
              summary: 'Test summary',
              language: 'cjk',
            ),
          ),
        ),
      );

      expect(find.textContaining('中文'), findsOneWidget);
    });

    testWidgets('should show confidence metadata when provided',
        (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: SummaryView(
              summary: 'Test summary',
              confidence: 0.85,
            ),
          ),
        ),
      );

      expect(find.text('85% confidence'), findsOneWidget);
      expect(find.byIcon(Icons.signal_cellular_alt), findsOneWidget);
    });

    testWidgets('should not show confidence when zero', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: SummaryView(
              summary: 'Test summary',
              confidence: 0.0,
            ),
          ),
        ),
      );

      expect(find.text('% confidence'), findsNothing);
    });

    testWidgets('should have italic style for non-AI summaries',
        (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: SummaryView(
              summary: 'Test summary',
              aiUsed: false,
            ),
          ),
        ),
      );

      final selectableText = tester.widget<SelectableText>(
        find.byType(SelectableText),
      );
      expect(
        selectableText.style?.fontStyle,
        FontStyle.italic,
      );
    });

    testWidgets('should have normal style for AI summaries',
        (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: SummaryView(
              summary: 'Test summary',
              aiUsed: true,
            ),
          ),
        ),
      );

      final selectableText = tester.widget<SelectableText>(
        find.byType(SelectableText),
      );
      expect(
        selectableText.style?.fontStyle,
        FontStyle.normal,
      );
    });
  });

  group('SummaryEmptyView Widget', () {
    testWidgets('should display empty state message', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: SummaryEmptyView(),
          ),
        ),
      );

      expect(find.text('No Summary'), findsOneWidget);
      expect(
        find.text('Generate an AI or TF-IDF summary to quickly understand the content.'),
        findsOneWidget,
      );
    });

    testWidgets('should show generate button when onGenerate provided',
        (tester) async {
      bool generatePressed = false;

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SummaryEmptyView(
              onGenerate: () => generatePressed = true,
            ),
          ),
        ),
      );

      expect(find.text('Generate Summary'), findsOneWidget);

      await tester.tap(find.text('Generate Summary'));
      await tester.pump();

      expect(generatePressed, isTrue);
    });

    testWidgets('should not show button without onGenerate', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: SummaryEmptyView(),
          ),
        ),
      );

      expect(find.text('Generate Summary'), findsNothing);
    });
  });

  group('SummaryLoadingView Widget', () {
    testWidgets('should show loading indicator', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: SummaryLoadingView(),
          ),
        ),
      );

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      expect(find.text('Generating summary...'), findsOneWidget);
    });

    testWidgets('should show custom message', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: SummaryLoadingView(
              message: 'AI is thinking...',
            ),
          ),
        ),
      );

      expect(find.text('AI is thinking...'), findsOneWidget);
    });
  });
}
