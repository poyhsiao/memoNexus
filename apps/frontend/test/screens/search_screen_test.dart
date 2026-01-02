import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:memonexus_frontend/providers/search_provider.dart';
import 'package:memonexus_frontend/screens/search_screen.dart';
import 'package:memonexus_frontend/services/api_client.dart';
import 'package:memonexus_frontend/widgets/search_result_item.dart';

// =====================================================
// Fake API Client for Search
// =====================================================

class _TestNavigatorObserver extends NavigatorObserver {
  final List<Route> pushedRoutes = [];

  @override
  void didPush(Route route, Route? previousRoute) {
    pushedRoutes.add(route);
    super.didPush(route, previousRoute);
  }
}

class _FakeSearchAPIClient extends MemoNexusAPIClient {
  bool shouldFail = false;
  String? errorMessage;
  List<Map<String, dynamic>> mockResults = [];

  @override
  Future<Map<String, dynamic>> search({
    required String query,
    int limit = 20,
    String? mediaType,
    String? tags,
    int? dateFrom,
    int? dateTo,
  }) async {
    if (shouldFail) {
      throw APIException(
        code: 'SEARCH_ERROR',
        message: errorMessage ?? 'Search failed',
        statusCode: 500,
      );
    }

    return {
      'results': mockResults,
      'total': mockResults.length,
      'query': query,
    };
  }

  @override
  Future<List<Map<String, dynamic>>> listTags() async => [];
}

// =====================================================
// Search Screen Tests
// =====================================================

void main() {
  group('SearchScreen Widget Tests', () {
    late _FakeSearchAPIClient fakeApi;

    setUp(() {
      fakeApi = _FakeSearchAPIClient();
    });

    Widget buildTestWidget() {
      return ProviderScope(
        overrides: [
          searchProvider.overrideWith((ref) => SearchNotifier(fakeApi)),
        ],
        child: const MaterialApp(
          home: SearchScreen(),
        ),
      );
    }

    testWidgets('should render search header and initial empty state',
        (tester) async {
      await tester.pumpWidget(buildTestWidget());

      // Verify search header exists
      expect(find.byType(TextField), findsOneWidget);
      expect(find.text('Search content...'), findsOneWidget);

      // Verify empty state message
      expect(find.text('Search Your Knowledge Base'), findsOneWidget);
      expect(find.textContaining('Enter keywords'), findsOneWidget);
    });

    testWidgets('should show search tips on initial state', (tester) async {
      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Search Tips'), findsOneWidget);
      expect(find.textContaining('exact phrases'), findsOneWidget);
    });

    testWidgets('should perform search and show results', (tester) async {
      // Setup mock results with correct structure
      final now = DateTime.now().millisecondsSinceEpoch ~/ 1000;
      fakeApi.mockResults = [
        {
          'item': {
            'id': '1',
            'title': 'Test Article',
            'content_text': 'Test content',
            'media_type': 'web',
            'tags': '',
            'created_at': now,
            'updated_at': now,
            'version': 1,
            'is_deleted': 0,
          },
          'relevance': 0.95,
          'matched_terms': <String>[],
        },
      ];

      await tester.pumpWidget(buildTestWidget());

      // Enter and submit search
      final textField = find.byType(TextField);
      await tester.enterText(textField, 'test');
      await tester.testTextInput.receiveAction(TextInputAction.done);
      await tester.pumpAndSettle();

      // Verify results count is displayed
      expect(find.textContaining('result'), findsOneWidget);
    });

    testWidgets('should show empty results when no matches', (tester) async {
      fakeApi.mockResults = []; // Empty results

      await tester.pumpWidget(buildTestWidget());

      final textField = find.byType(TextField);
      await tester.enterText(textField, 'no results');
      await tester.testTextInput.receiveAction(TextInputAction.done);
      await tester.pumpAndSettle();

      expect(find.text('No Results Found'), findsOneWidget);
    });

    testWidgets('should show error state when search fails', (tester) async {
      fakeApi.shouldFail = true;
      fakeApi.errorMessage = 'Network error';

      await tester.pumpWidget(buildTestWidget());

      final textField = find.byType(TextField);
      await tester.enterText(textField, 'fail');
      await tester.testTextInput.receiveAction(TextInputAction.done);
      await tester.pumpAndSettle();

      expect(find.text('Search Error'), findsOneWidget);
      // The error message is displayed by state.error
      final container = ProviderScope.containerOf(
        tester.element(find.byType(SearchScreen)),
      );
      final state = container.read(searchProvider);
      expect(state.error, isNotNull);
      expect(state.error, contains('Network error'));
    });

    testWidgets('should clear search when clear button tapped', (tester) async {
      final now = DateTime.now().millisecondsSinceEpoch ~/ 1000;
      fakeApi.mockResults = [
        {
          'item': {
            'id': '1',
            'title': 'Test',
            'content_text': 'Content',
            'media_type': 'web',
            'tags': '',
            'created_at': now,
            'updated_at': now,
            'version': 1,
            'is_deleted': 0,
          },
          'relevance': 1.0,
          'matched_terms': <String>[],
        },
      ];

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            searchProvider.overrideWith((ref) => SearchNotifier(fakeApi)),
            recentSearchesProvider.overrideWith((ref) => []), // No recent searches
          ],
          child: const MaterialApp(
            home: SearchScreen(),
          ),
        ),
      );

      final textField = find.byType(TextField);
      await tester.enterText(textField, 'test');
      await tester.testTextInput.receiveAction(TextInputAction.done);
      await tester.pumpAndSettle();

      // Verify search was performed
      expect(find.textContaining('result'), findsOneWidget);

      // Tap clear button
      await tester.tap(find.byIcon(Icons.clear));
      await tester.pumpAndSettle();

      // After clearing, the search is added to recent searches
      // Since we now have a recent search, that section is shown
      // Verify query was cleared (state check)
      final container = ProviderScope.containerOf(
        tester.element(find.byType(SearchScreen)),
      );
      final state = container.read(searchProvider);
      expect(state.query, isEmpty);
      expect(find.text('Recent Searches'), findsOneWidget);
      expect(find.text('test'), findsOneWidget); // The search we just did
    });

    testWidgets('should show filter button when filters applied',
        (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            searchProvider.overrideWith((ref) {
              final notifier = SearchNotifier(fakeApi);
              notifier.state = const SearchState(filterMediaType: 'pdf');
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: SearchScreen(),
          ),
        ),
      );

      expect(find.byIcon(Icons.filter_list), findsOneWidget);
    });

    testWidgets('should show recent searches when available', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            searchProvider.overrideWith((ref) => SearchNotifier(fakeApi)),
            recentSearchesProvider.overrideWith((ref) => [
                  'flutter tutorial',
                  'dart guide',
                ]),
          ],
          child: const MaterialApp(
            home: SearchScreen(),
          ),
        ),
      );

      expect(find.text('Recent Searches'), findsOneWidget);
      expect(find.text('flutter tutorial'), findsOneWidget);
      expect(find.text('Clear'), findsOneWidget);
    });

    testWidgets('should clear recent searches when clear tapped',
        (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            searchProvider.overrideWith((ref) => SearchNotifier(fakeApi)),
            recentSearchesProvider.overrideWith((ref) => ['test query']),
          ],
          child: const MaterialApp(
            home: SearchScreen(),
          ),
        ),
      );

      await tester.tap(find.text('Clear'));
      await tester.pump();

      expect(find.text('Recent Searches'), findsNothing);
    });

    testWidgets('should display search results correctly', (tester) async {
      final now = DateTime.now().millisecondsSinceEpoch ~/ 1000;
      fakeApi.mockResults = [
        {
          'item': {
            'id': 'item-123',
            'title': 'Test Article',
            'content_text': 'Test',
            'media_type': 'web',
            'tags': '',
            'created_at': now,
            'updated_at': now,
            'version': 1,
            'is_deleted': 0,
          },
          'relevance': 1.0,
          'matched_terms': <String>[],
        },
      ];

      await tester.pumpWidget(buildTestWidget());

      final textField = find.byType(TextField);
      await tester.enterText(textField, 'test');
      await tester.testTextInput.receiveAction(TextInputAction.done);
      await tester.pumpAndSettle();

      // Verify results are displayed
      expect(find.text('Test Article'), findsOneWidget);

      // Note: Navigation test would require setting up navigation
      // The SearchResultItemWidget handles onTap callbacks
      // This test verifies results render correctly
      expect(find.byType(SearchResultItemWidget), findsWidgets);
    });

    testWidgets('should show loading state while searching', (tester) async {
      // Create a notifier and manually set loading state
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            searchProvider.overrideWith((ref) {
              final notifier = SearchNotifier(fakeApi);
              // Manually set loading state before pumping widget
              notifier.state = const SearchState(isLoading: true);
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: SearchScreen(),
          ),
        ),
      );

      await tester.pump();

      // Should show loading indicator
      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      expect(find.text('Searching...'), findsOneWidget);
    });

    testWidgets('should navigate to detail when result tapped', (tester) async {
      final now = DateTime.now().millisecondsSinceEpoch ~/ 1000;
      fakeApi.mockResults = [
        {
          'item': {
            'id': 'item-123',
            'title': 'Test Article',
            'content_text': 'Test',
            'media_type': 'web',
            'tags': '',
            'created_at': now,
            'updated_at': now,
            'version': 1,
            'is_deleted': 0,
          },
          'relevance': 1.0,
          'matched_terms': <String>[],
        },
      ];

      // Track navigation using a navigator observer
      final navigatorObserver = _TestNavigatorObserver();

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            searchProvider.overrideWith((ref) => SearchNotifier(fakeApi)),
          ],
          child: MaterialApp(
            home: const SearchScreen(),
            navigatorObservers: [navigatorObserver],
          ),
        ),
      );

      final textField = find.byType(TextField);
      await tester.enterText(textField, 'test');
      await tester.testTextInput.receiveAction(TextInputAction.done);
      await tester.pumpAndSettle();

      // Tap the result item - navigation should be attempted
      await tester.tap(find.byType(SearchResultItemWidget));
      await tester.pumpAndSettle();

      // Verify a route was pushed (navigation happened)
      expect(navigatorObserver.pushedRoutes, isNotEmpty);
    });

    testWidgets('should show content options on long press', (tester) async {
      final now = DateTime.now().millisecondsSinceEpoch ~/ 1000;
      fakeApi.mockResults = [
        {
          'item': {
            'id': 'item-123',
            'title': 'Test Article',
            'content_text': 'Test',
            'media_type': 'web',
            'tags': '',
            'created_at': now,
            'updated_at': now,
            'version': 1,
            'is_deleted': 0,
          },
          'relevance': 1.0,
          'matched_terms': <String>[],
        },
      ];

      await tester.pumpWidget(buildTestWidget());

      final textField = find.byType(TextField);
      await tester.enterText(textField, 'test');
      await tester.testTextInput.receiveAction(TextInputAction.done);
      await tester.pumpAndSettle();

      // Long press the result item
      await tester.longPress(find.byType(SearchResultItemWidget));
      await tester.pumpAndSettle();

      // Should show bottom sheet with options
      expect(find.text('View Details'), findsOneWidget);
      expect(find.text('Share'), findsOneWidget);
    });

    testWidgets('should tap recent search to perform search', (tester) async {
      fakeApi.mockResults = [
        {
          'item': {
            'id': '1',
            'title': 'Result',
            'content_text': 'Content',
            'media_type': 'web',
            'tags': '',
            'created_at': DateTime.now().millisecondsSinceEpoch ~/ 1000,
            'updated_at': DateTime.now().millisecondsSinceEpoch ~/ 1000,
            'version': 1,
            'is_deleted': 0,
          },
          'relevance': 1.0,
          'matched_terms': <String>[],
        },
      ];

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            searchProvider.overrideWith((ref) => SearchNotifier(fakeApi)),
            recentSearchesProvider.overrideWith((ref) => ['test query']),
          ],
          child: const MaterialApp(
            home: SearchScreen(),
          ),
        ),
      );

      // Tap the recent search item
      await tester.tap(find.text('test query'));
      await tester.pumpAndSettle();

      // Should perform the search and show results
      expect(find.textContaining('result'), findsOneWidget);
    });

    testWidgets('should show filter button in header', (tester) async {
      await tester.pumpWidget(buildTestWidget());

      // Should have filter button
      expect(find.byIcon(Icons.filter_list_outlined), findsOneWidget);
    });
  });

  group('SearchScreen Integration Tests', () {
    testWidgets('should integrate with search provider', (tester) async {
      final fakeApi = _FakeSearchAPIClient();
      final now = DateTime.now().millisecondsSinceEpoch ~/ 1000;
      fakeApi.mockResults = [
        {
          'item': {
            'id': '1',
            'title': 'Result',
            'content_text': 'Content',
            'media_type': 'web',
            'tags': '',
            'created_at': now,
            'updated_at': now,
            'version': 1,
            'is_deleted': 0,
          },
          'relevance': 1.0,
          'matched_terms': <String>[],
        },
      ];

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            searchProvider.overrideWith((ref) => SearchNotifier(fakeApi)),
          ],
          child: const MaterialApp(
            home: SearchScreen(),
          ),
        ),
      );

      await tester.enterText(find.byType(TextField), 'test');
      await tester.testTextInput.receiveAction(TextInputAction.done);
      await tester.pumpAndSettle();

      final container = ProviderScope.containerOf(
        tester.element(find.byType(SearchScreen)),
      );
      final state = container.read(searchProvider);

      expect(state.query, 'test');
      expect(state.results.length, 1);
      expect(state.results.first.item.id, '1');
    });
  });
}
