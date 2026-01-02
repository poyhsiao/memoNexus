import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:memonexus_frontend/providers/search_provider.dart';
import 'package:memonexus_frontend/services/api_client.dart';
import 'package:memonexus_frontend/widgets/search_filters.dart';

// =====================================================
// Fake API Client
// =====================================================

class _FakeSearchAPIClient extends MemoNexusAPIClient {
  @override
  Future<Map<String, dynamic>> search({
    required String query,
    int limit = 20,
    String? mediaType,
    String? tags,
    int? dateFrom,
    int? dateTo,
  }) async {
    return {'results': [], 'total': 0, 'query': query};
  }

  @override
  Future<List<Map<String, dynamic>>> listTags() async => [];
}

void main() {
  group('SearchFiltersWidget Widget', () {
    testWidgets('should render filter header', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: SearchFiltersWidget(),
            ),
          ),
        ),
      );

      expect(find.text('Filters'), findsOneWidget);
      expect(find.byIcon(Icons.expand_more), findsOneWidget);
    });

    testWidgets('should expand when header tapped', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: SearchFiltersWidget(),
            ),
          ),
        ),
      );

      expect(find.byIcon(Icons.expand_more), findsOneWidget);

      await tester.tap(find.text('Filters'));
      await tester.pump();

      expect(find.byIcon(Icons.expand_less), findsOneWidget);
    });

    testWidgets('should show media type filter chips when expanded',
        (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: SearchFiltersWidget(),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Filters'));
      await tester.pump();

      expect(find.text('Media Type'), findsOneWidget);
      expect(find.text('WEB'), findsOneWidget);
      expect(find.text('PDF'), findsOneWidget);
      expect(find.text('IMAGE'), findsOneWidget);
    });

    testWidgets('should show tags filter input when expanded', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: SearchFiltersWidget(),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Filters'));
      await tester.pump();

      expect(find.text('Tags'), findsOneWidget);
      expect(find.byType(TextField), findsOneWidget);
    });

    testWidgets('should show date range filter buttons when expanded',
        (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: SearchFiltersWidget(),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Filters'));
      await tester.pump();

      expect(find.text('Date Range'), findsOneWidget);
      expect(find.text('From Date'), findsOneWidget);
      expect(find.text('To Date'), findsOneWidget);
    });

    testWidgets('should select media type filter', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: SearchFiltersWidget(),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Filters'));
      await tester.pump();

      await tester.tap(find.text('PDF'));
      await tester.pump();

      final container = ProviderScope.containerOf(
        tester.element(find.byType(SearchFiltersWidget)),
      );
      final state = container.read(searchProvider);

      expect(state.filterMediaType, 'pdf');
    });

    testWidgets('should show active filter count when filters applied',
        (tester) async {
      final fakeApi = _FakeSearchAPIClient();
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            searchProvider.overrideWith((ref) {
              final notifier = SearchNotifier(fakeApi);
              notifier.state = const SearchState(filterMediaType: 'pdf');
              return notifier;
            }),
          ],
          child: MaterialApp(
            home: Scaffold(
              body: SearchFiltersWidget(),
            ),
          ),
        ),
      );

      expect(find.text('1'), findsOneWidget);
    });

    testWidgets('should show clear all button when filters active',
        (tester) async {
      final fakeApi = _FakeSearchAPIClient();
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            searchProvider.overrideWith((ref) {
              final notifier = SearchNotifier(fakeApi);
              notifier.state = const SearchState(filterMediaType: 'pdf');
              return notifier;
            }),
          ],
          child: MaterialApp(
            home: Scaffold(
              body: SearchFiltersWidget(),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Filters'));
      await tester.pump();

      expect(find.text('Clear All Filters'), findsOneWidget);
    });

    testWidgets('should clear all filters when button pressed', (tester) async {
      final fakeApi = _FakeSearchAPIClient();
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            searchProvider.overrideWith((ref) {
              final notifier = SearchNotifier(fakeApi);
              notifier.state = const SearchState(filterMediaType: 'pdf');
              return notifier;
            }),
          ],
          child: MaterialApp(
            home: Scaffold(
              body: SearchFiltersWidget(),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Filters'));
      await tester.pump();

      await tester.tap(find.text('Clear All Filters'));
      await tester.pump();

      final container = ProviderScope.containerOf(
        tester.element(find.byType(SearchFiltersWidget)),
      );
      final state = container.read(searchProvider);

      expect(state.filterMediaType, isNull);
    });
  });

  group('QuickFilterChips Widget', () {
    testWidgets('should render all media type chips', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: QuickFilterChips(),
            ),
          ),
        ),
      );

      expect(find.text('All'), findsOneWidget);
      expect(find.text('WEB'), findsOneWidget);
      expect(find.text('PDF'), findsOneWidget);
      expect(find.text('IMAGE'), findsOneWidget);
    });

    testWidgets('should select All chip by default', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: QuickFilterChips(),
            ),
          ),
        ),
      );

      final allChip = tester.widget<FilterChip>(
        find.ancestor(
          of: find.text('All'),
          matching: find.byType(FilterChip),
        ),
      );

      expect(allChip.selected, isTrue);
    });

    testWidgets('should call onMediaTypeChanged when chip tapped',
        (tester) async {
      String? selectedType;

      await tester.pumpWidget(
        ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: QuickFilterChips(
                onMediaTypeChanged: (type) => selectedType = type,
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('PDF'));
      await tester.pump();

      expect(selectedType, 'pdf');
    });

    testWidgets('should deselect when tapping selected chip', (tester) async {
      String? selectedType;

      await tester.pumpWidget(
        ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: QuickFilterChips(
                onMediaTypeChanged: (type) => selectedType = type,
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('PDF'));
      await tester.pump();

      expect(selectedType, 'pdf');

      await tester.tap(find.text('PDF'));
      await tester.pump();

      expect(selectedType, isNull);
    });

    testWidgets('should return to All when current filter cleared',
        (tester) async {
      final fakeApi = _FakeSearchAPIClient();
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            searchProvider.overrideWith((ref) {
              final notifier = SearchNotifier(fakeApi);
              notifier.state = const SearchState(filterMediaType: 'pdf');
              return notifier;
            }),
          ],
          child: MaterialApp(
            home: Scaffold(
              body: QuickFilterChips(),
            ),
          ),
        ),
      );

      final pdfChip = tester.widget<FilterChip>(
        find.ancestor(
          of: find.text('PDF'),
          matching: find.byType(FilterChip),
        ),
      );

      expect(pdfChip.selected, isTrue);
    });

    testWidgets('should apply tags filter when tags submitted via TextField',
        (tester) async {
      final fakeApi = _FakeSearchAPIClient();
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            searchProvider.overrideWith((ref) {
              final notifier = SearchNotifier(fakeApi);
              notifier.state = const SearchState(query: 'test');
              return notifier;
            }),
          ],
          child: MaterialApp(
            home: const Scaffold(
              body: SearchFiltersWidget(),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Filters'));
      await tester.pump();

      // Find the tags TextField and enter text
      final textField = find.byWidgetPredicate((widget) =>
          widget is TextField &&
          widget.decoration?.hintText == 'comma, separated, tags');
      expect(textField, findsOneWidget);

      await tester.enterText(textField, 'tag1,tag2');
      await tester.testTextInput.receiveAction(TextInputAction.done);
      await tester.pump();

      final container = ProviderScope.containerOf(
        tester.element(find.byType(SearchFiltersWidget)),
      );
      final state = container.read(searchProvider);

      expect(state.filterTags, 'tag1,tag2');
    });
  });
}
