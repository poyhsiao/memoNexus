import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:memonexus_frontend/models/content_item.dart';
import 'package:memonexus_frontend/providers/search_provider.dart';
import 'package:memonexus_frontend/services/api_client.dart';

// Fake API client for testing
class FakeMemoNexusAPIClient extends MemoNexusAPIClient {
  bool shouldFail = false;
  String? errorMessage;
  Map<String, dynamic>? mockResponse;
  List<Map<String, dynamic>>? mockSearchResults;
  int searchCallCount = 0;

  @override
  Future<Map<String, dynamic>> search({
    required String query,
    int limit = 20,
    String? mediaType,
    String? tags,
    int? dateFrom,
    int? dateTo,
  }) async {
    searchCallCount++;
    if (shouldFail) {
      throw APIException(
        statusCode: 500,
        code: 'TEST_ERROR',
        message: errorMessage ?? 'Test error',
      );
    }

    if (mockSearchResults != null) {
      return {
        'results': mockSearchResults,
        'total': mockSearchResults!.length,
      };
    }

    return mockResponse ?? {'results': [], 'total': 0};
  }
}

void main() {
  group('SearchState', () {
    test('should have correct default values', () {
      const state = SearchState();

      expect(state.query, '');
      expect(state.results, isEmpty);
      expect(state.total, 0);
      expect(state.isLoading, false);
      expect(state.error, isNull);
      expect(state.filterMediaType, isNull);
      expect(state.filterTags, isNull);
      expect(state.filterDateFrom, isNull);
      expect(state.filterDateTo, isNull);
    });

    test('hasQuery should return true for non-empty query', () {
      const stateWithQuery = SearchState(query: 'test');
      const stateWithoutQuery = SearchState(query: '');
      const stateWithSpaces = SearchState(query: '   ');

      expect(stateWithQuery.hasQuery, true);
      expect(stateWithoutQuery.hasQuery, false);
      expect(stateWithSpaces.hasQuery, false);
    });

    test('hasFilters should return true when filters are set', () {
      const withMediaType = SearchState(filterMediaType: 'web');
      const withTags = SearchState(filterTags: 'test');
      const withDates = SearchState(filterDateFrom: 100);
      const noFilters = SearchState();

      expect(withMediaType.hasFilters, true);
      expect(withTags.hasFilters, true);
      // Note: hasFilters only checks mediaType and tags, not dates
      expect(withDates.hasFilters, false);
      expect(noFilters.hasFilters, false);
    });

    test('isEmpty should return true when not loading and no results', () {
      const emptyState = SearchState(results: [], isLoading: false);
      const loadingState = SearchState(results: [], isLoading: true);
      final withResults = SearchState(
        results: [SearchResult(
          item: ContentItem(
            id: '1',
            title: 'Test',
            contentText: 'Content',
            mediaType: MediaType.web,
            tags: [],
            createdAt: DateTime(2024, 1, 1),
            updatedAt: DateTime(2024, 1, 1),
            version: 1,
          ),
          relevance: 1.0,
          matchedTerms: [],
        )],
      );

      expect(emptyState.isEmpty, true);
      expect(loadingState.isEmpty, false);
      expect(withResults.isEmpty, false);
    });

    test('hasError should return true when error is set', () {
      const withError = SearchState(error: 'Test error');
      const withoutError = SearchState();

      expect(withError.hasError, true);
      expect(withoutError.hasError, false);
    });

    test('copyWith should create new state with updated values', () {
      const original = SearchState(query: 'original');
      final updated = original.copyWith(
        query: 'updated',
        isLoading: true,
      );

      expect(original.query, 'original');
      expect(original.isLoading, false);
      expect(updated.query, 'updated');
      expect(updated.isLoading, true);
    });

    test('copyWith with clearFilters should reset all filters', () {
      const original = SearchState(
        filterMediaType: 'web',
        filterTags: 'test',
        filterDateFrom: 100,
        filterDateTo: 200,
      );

      final cleared = original.copyWith(clearFilters: true);

      expect(cleared.filterMediaType, isNull);
      expect(cleared.filterTags, isNull);
      expect(cleared.filterDateFrom, isNull);
      expect(cleared.filterDateTo, isNull);
    });

    test('copyWith should preserve original state', () {
      const original = SearchState(
        query: 'test',
        filterMediaType: 'web',
        filterTags: 'tag1,tag2',
      );

      final updated = original.copyWith(query: 'new query');

      expect(original.query, 'test');
      expect(original.filterMediaType, 'web');
      expect(original.filterTags, 'tag1,tag2');

      expect(updated.query, 'new query');
      expect(updated.filterMediaType, 'web');
      expect(updated.filterTags, 'tag1,tag2');
    });
  });

  group('SearchNotifier', () {
    late FakeMemoNexusAPIClient fakeApi;

    setUp(() {
      fakeApi = FakeMemoNexusAPIClient();
    });

    test('should initialize with default state', () {
      final notifier = SearchNotifier(fakeApi);

      expect(notifier.state.query, '');
      expect(notifier.state.results, isEmpty);
      expect(notifier.state.isLoading, false);
      expect(notifier.state.error, isNull);
    });

    test('search should update state with results', () async {
      fakeApi.mockSearchResults = [
        {
          'item': {
            'id': '1',
            'title': 'Test Article',
            'content_text': 'Test content',
            'media_type': 'web',
            'tags': '',
            'created_at': 1704110400,
            'updated_at': 1704110400,
            'version': 1,
          },
          'relevance': 1.5,
          'matched_terms': ['test'],
        },
      ];

      final notifier = SearchNotifier(fakeApi);
      await notifier.search('test query');

      expect(notifier.state.query, 'test query');
      expect(notifier.state.isLoading, false);
      expect(notifier.state.results.length, 1);
      expect(notifier.state.total, 1);
      expect(notifier.state.results.first.item.id, '1');
      expect(notifier.state.results.first.item.title, 'Test Article');
      expect(notifier.state.results.first.relevance, 1.5);
      expect(notifier.state.error, isNull);
    });

    test('search should handle empty query by clearing state', () async {
      final notifier = SearchNotifier(fakeApi);
      notifier.state = notifier.state.copyWith(
        results: [SearchResult(
          item: ContentItem(
            id: '1',
            title: 'Test',
            contentText: 'Content',
            mediaType: MediaType.web,
            tags: [],
            createdAt: DateTime(2024, 1, 1),
            updatedAt: DateTime(2024, 1, 1),
            version: 1,
          ),
          relevance: 1.0,
          matchedTerms: [],
        )],
      );

      await notifier.search('   ');

      expect(notifier.state.query, '');
      expect(notifier.state.results, isEmpty);
      expect(notifier.state.isLoading, false);
    });

    test('search should handle API errors', () async {
      fakeApi.shouldFail = true;
      fakeApi.errorMessage = 'Network error';

      final notifier = SearchNotifier(fakeApi);
      await notifier.search('test');

      expect(notifier.state.isLoading, false);
      expect(notifier.state.results, isEmpty);
      expect(notifier.state.total, 0);
      expect(notifier.state.error, contains('Network error'));
    });

    test('search should show loading state during request', () async {
      fakeApi.mockResponse = {'results': [], 'total': 0};
      final notifier = SearchNotifier(fakeApi);

      // Start search but don't await
      final future = notifier.search('test');

      // Check loading state is set
      expect(notifier.state.isLoading, true);
      expect(notifier.state.query, 'test');

      // Wait for completion
      await future;
      expect(notifier.state.isLoading, false);
    });

    test('clear should reset state to default', () async {
      final notifier = SearchNotifier(fakeApi);
      notifier.state = notifier.state.copyWith(
        query: 'test',
        results: [SearchResult(
          item: ContentItem(
            id: '1',
            title: 'Test',
            contentText: 'Content',
            mediaType: MediaType.web,
            tags: [],
            createdAt: DateTime(2024, 1, 1),
            updatedAt: DateTime(2024, 1, 1),
            version: 1,
          ),
          relevance: 1.0,
          matchedTerms: [],
        )],
        total: 1,
        error: 'error',
        filterMediaType: 'web',
      );

      notifier.clear();

      expect(notifier.state.query, '');
      expect(notifier.state.results, isEmpty);
      expect(notifier.state.total, 0);
      expect(notifier.state.error, isNull);
      expect(notifier.state.filterMediaType, isNull);
    });

    test('setMediaTypeFilter should update filter and re-search if has query', () async {
      fakeApi.mockSearchResults = [];

      final notifier = SearchNotifier(fakeApi);
      notifier.state = notifier.state.copyWith(query: 'test');

      await notifier.setMediaTypeFilter('web');

      expect(notifier.state.filterMediaType, 'web');
      expect(fakeApi.searchCallCount, 1); // search was called
    });

    test('setMediaTypeFilter should convert empty string to null', () async {
      final notifier = SearchNotifier(fakeApi);

      await notifier.setMediaTypeFilter('');

      expect(notifier.state.filterMediaType, isNull);
    });

    test('setMediaTypeFilter should not re-search if no query', () async {
      fakeApi.mockSearchResults = [];

      final notifier = SearchNotifier(fakeApi);

      await notifier.setMediaTypeFilter('web');

      expect(fakeApi.searchCallCount, 0); // no search call
    });

    test('setTagsFilter should update filter and re-search if has query', () async {
      fakeApi.mockSearchResults = [];

      final notifier = SearchNotifier(fakeApi);
      notifier.state = notifier.state.copyWith(query: 'test');

      await notifier.setTagsFilter('ml,ai');

      expect(notifier.state.filterTags, 'ml,ai');
      expect(fakeApi.searchCallCount, 1);
    });

    test('setDateRangeFilter should update filter and re-search if has query', () async {
      fakeApi.mockSearchResults = [];

      final notifier = SearchNotifier(fakeApi);
      notifier.state = notifier.state.copyWith(query: 'test');

      await notifier.setDateRangeFilter(100, 200);

      expect(notifier.state.filterDateFrom, 100);
      expect(notifier.state.filterDateTo, 200);
      expect(fakeApi.searchCallCount, 1);
    });

    test('clearFilters should reset all filters and re-search if has query', () async {
      fakeApi.mockSearchResults = [];

      final notifier = SearchNotifier(fakeApi);
      notifier.state = notifier.state.copyWith(
        query: 'test',
        filterMediaType: 'web',
        filterTags: 'test',
        filterDateFrom: 100,
      );

      await notifier.clearFilters();

      expect(notifier.state.filterMediaType, isNull);
      expect(notifier.state.filterTags, isNull);
      expect(notifier.state.filterDateFrom, isNull);
      expect(notifier.state.filterDateTo, isNull);
      expect(fakeApi.searchCallCount, 1);
    });

    test('clearFilters should not re-search if no filters set', () async {
      fakeApi.mockSearchResults = [];

      final notifier = SearchNotifier(fakeApi);
      notifier.state = notifier.state.copyWith(query: 'test');

      await notifier.clearFilters();

      expect(fakeApi.searchCallCount, 0);
    });

    test('activeFilterCount should count active filters', () {
      final notifier = SearchNotifier(fakeApi);

      expect(notifier.activeFilterCount, 0);

      notifier.state = notifier.state.copyWith(filterMediaType: 'web');
      expect(notifier.activeFilterCount, 1);

      notifier.state = notifier.state.copyWith(filterTags: 'test');
      expect(notifier.activeFilterCount, 2);

      notifier.state = notifier.state.copyWith(filterDateFrom: 100);
      expect(notifier.activeFilterCount, 3);

      notifier.state = notifier.state.copyWith(filterDateTo: 200);
      expect(notifier.activeFilterCount, 3); // date range counts as 1
    });

    test('multiple search results should be parsed correctly', () async {
      fakeApi.mockSearchResults = [
        {
          'item': {
            'id': '1',
            'title': 'First',
            'content_text': 'Content 1',
            'media_type': 'web',
            'tags': 'tag1',
            'created_at': 1704110400,
            'updated_at': 1704110400,
            'version': 1,
          },
          'relevance': 2.0,
          'matched_terms': ['first'],
        },
        {
          'item': {
            'id': '2',
            'title': 'Second',
            'content_text': 'Content 2',
            'media_type': 'pdf',
            'tags': 'tag2',
            'created_at': 1704110400,
            'updated_at': 1704110400,
            'version': 1,
          },
          'relevance': 1.5,
          'matched_terms': ['second'],
        },
      ];

      final notifier = SearchNotifier(fakeApi);
      await notifier.search('test');

      expect(notifier.state.results.length, 2);
      expect(notifier.state.total, 2);
      expect(notifier.state.results[0].item.id, '1');
      expect(notifier.state.results[1].item.id, '2');
      expect(notifier.state.results[0].matchedTerms, ['first']);
      expect(notifier.state.results[1].matchedTerms, ['second']);
    });
  });

  group('Provider Integration', () {
    test('recentSearchesProvider should start with empty list', () {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      final recent = container.read(recentSearchesProvider);
      expect(recent, isEmpty);
    });

    test('recentSearchesProvider should update with new searches', () {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      // Direct state manipulation
      container.read(recentSearchesProvider.notifier).state = ['query 1', 'query 2'];

      final recent = container.read(recentSearchesProvider);
      expect(recent, ['query 1', 'query 2']);
    });

    test('recentSearchesProvider should replace entire list', () {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      container.read(recentSearchesProvider.notifier).state = ['old 1', 'old 2'];
      container.read(recentSearchesProvider.notifier).state = ['new 1', 'new 2', 'new 3'];

      final recent = container.read(recentSearchesProvider);
      expect(recent, ['new 1', 'new 2', 'new 3']);
    });

    test('recentSearchesProvider should clear to empty', () {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      container.read(recentSearchesProvider.notifier).state = ['query 1', 'query 2'];
      container.read(recentSearchesProvider.notifier).state = [];

      final recent = container.read(recentSearchesProvider);
      expect(recent, isEmpty);
    });
  });
}
