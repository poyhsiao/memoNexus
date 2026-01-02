import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:memonexus_frontend/models/content_item.dart';
import 'package:memonexus_frontend/providers/content_provider.dart';
import 'package:memonexus_frontend/services/api_client.dart';

// =====================================================
// Fake API Client for Content Provider Tests
// =====================================================

class _FakeMemoNexusAPIClient extends MemoNexusAPIClient {
  List<Map<String, dynamic>> mockContentItems = [];
  bool shouldFail = false;
  String errorMessage = 'Test error';

  @override
  Future<Map<String, dynamic>> listContentItems({
    int page = 1,
    int perPage = 20,
    String sort = 'created_at',
    String order = 'desc',
    String? mediaType,
    String? tag,
  }) async {
    if (shouldFail) {
      throw APIException(
        code: 'LIST_ERROR',
        message: errorMessage,
        statusCode: 500,
      );
    }
    return {
      'items': mockContentItems,
      'page': page,
      'per_page': perPage,
      'total_items': mockContentItems.length,
      'total_pages': 1,
      'has_more': false,
    };
  }

  @override
  Future<Map<String, dynamic>> getContentItem(String id) async {
    if (shouldFail) {
      throw APIException(
        code: 'GET_ERROR',
        message: errorMessage,
        statusCode: 500,
      );
    }
    final item = mockContentItems.firstWhere(
      (i) => i['id'] == id,
      orElse: () => {
        'id': id,
        'title': 'Test Content',
        'content_text': 'Test content body',
        'media_type': 'web',
        'tags': '',
        'created_at': 1704067200,
        'updated_at': 1704067200,
        'version': 1,
      },
    );
    return item;
  }

  @override
  Future<void> deleteContentItem(String id) async {
    if (shouldFail) {
      throw APIException(
        code: 'DELETE_ERROR',
        message: errorMessage,
        statusCode: 500,
      );
    }
    mockContentItems.removeWhere((item) => item['id'] == id);
  }

  @override
  Future<List<dynamic>> listTags() async => [];
}

void main() {
  group('ContentListState', () {
    test('should have correct default values', () {
      const state = ContentListState();

      expect(state.items, isEmpty);
      expect(state.isLoading, false);
      expect(state.error, isNull);
      expect(state.currentPage, 1);
      expect(state.hasMore, true);
      expect(state.activeFilterMediaType, isNull);
      expect(state.activeFilterTag, isNull);
    });

    test('copyWith should create new state with updated values', () {
      const original = ContentListState(
        currentPage: 2,
        hasMore: false,
      );
      final updated = original.copyWith(
        isLoading: true,
        currentPage: 3,
      );

      expect(original.isLoading, false);
      expect(original.currentPage, 2);
      expect(original.hasMore, false);
      expect(updated.isLoading, true);
      expect(updated.currentPage, 3);
      expect(updated.hasMore, false); // preserved
    });

    test('copyWith should merge items lists', () {
      final item1 = ContentItem(
        id: '1',
        title: 'Item 1',
        contentText: 'Content 1',
        mediaType: MediaType.web,
        tags: [],
        createdAt: DateTime(2024, 1, 1),
        updatedAt: DateTime(2024, 1, 1),
        version: 1,
      );
      final item2 = ContentItem(
        id: '2',
        title: 'Item 2',
        contentText: 'Content 2',
        mediaType: MediaType.web,
        tags: [],
        createdAt: DateTime(2024, 1, 1),
        updatedAt: DateTime(2024, 1, 1),
        version: 1,
      );

      final original = ContentListState(items: [item1]);
      final updated = original.copyWith(items: [item2]);

      expect(original.items.length, 1);
      expect(original.items.first.id, '1');
      expect(updated.items.length, 1);
      expect(updated.items.first.id, '2');
    });

    test('copyWith with null items should preserve original', () {
      final item = ContentItem(
        id: '1',
        title: 'Item',
        contentText: 'Content',
        mediaType: MediaType.web,
        tags: [],
        createdAt: DateTime(2024, 1, 1),
        updatedAt: DateTime(2024, 1, 1),
        version: 1,
      );

      final original = ContentListState(items: [item]);
      final updated = original.copyWith(
        isLoading: true,
      );

      expect(identical(updated.items, original.items), isTrue);
      expect(updated.items.length, 1);
    });
  });

  group('Provider Integration', () {
    test('apiClientProvider should provide MemoNexusAPIClient', () {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      final apiClient = container.read(apiClientProvider);

      expect(apiClient, isA<MemoNexusAPIClient>());
      expect(apiClient.baseUrl, 'http://localhost:8090/api');
    });

    test('contentListProvider should provide ContentListState', () async {
      final fakeApi = _FakeMemoNexusAPIClient();
      final container = ProviderContainer(
        overrides: [
          apiClientProvider.overrideWith((ref) => fakeApi),
        ],
      );
      addTearDown(container.dispose);

      // First read the provider to trigger auto-loading, then wait for it
      container.read(contentListProvider);
      await Future.delayed(const Duration(milliseconds: 10));

      final state = container.read(contentListProvider);

      expect(state, isA<ContentListState>());
      // After auto-loading completes with fake API, should not be loading
      expect(state.isLoading, false);

      // Wait for any pending async operations
      await Future.delayed(const Duration(milliseconds: 10));
    });

    test('filteredContentListProvider should filter by query', () async {
      final fakeApi = _FakeMemoNexusAPIClient();
      final container = ProviderContainer(
        overrides: [
          apiClientProvider.overrideWith((ref) => fakeApi),
        ],
      );
      addTearDown(container.dispose);

      // First read the provider to trigger auto-loading, then wait for it
      container.read(contentListProvider);
      await Future.delayed(const Duration(milliseconds: 10));

      final item1 = ContentItem(
        id: '1',
        title: 'Apple Pie',
        contentText: 'Delicious dessert',
        mediaType: MediaType.web,
        tags: ['dessert'],
        createdAt: DateTime(2024, 1, 1),
        updatedAt: DateTime(2024, 1, 1),
        version: 1,
      );
      final item2 = ContentItem(
        id: '2',
        title: 'Apple Sauce',
        contentText: 'Tasty topping',
        mediaType: MediaType.web,
        tags: ['sauce'],
        createdAt: DateTime(2024, 1, 1),
        updatedAt: DateTime(2024, 1, 1),
        version: 1,
      );
      final item3 = ContentItem(
        id: '3',
        title: 'Chocolate Cake',
        contentText: 'Sweet dessert',
        mediaType: MediaType.web,
        tags: ['dessert'],
        createdAt: DateTime(2024, 1, 1),
        updatedAt: DateTime(2024, 1, 1),
        version: 1,
      );

      // Mock the content list state with items
      container.read(contentListProvider.notifier).state = ContentListState(
        items: [item1, item2, item3],
      );

      // Test filtering
      final appleResults = container.read(filteredContentListProvider('apple'));
      expect(appleResults.length, 2);
      expect(appleResults.map((e) => e.id), contains('1'));
      expect(appleResults.map((e) => e.id), contains('2'));
      expect(appleResults.map((e) => e.id), isNot(contains('3')));

      final dessertResults = container.read(filteredContentListProvider('dessert'));
      expect(dessertResults.length, 2);
      expect(dessertResults.map((e) => e.id), contains('1'));
      expect(dessertResults.map((e) => e.id), contains('3'));
      expect(dessertResults.map((e) => e.id), isNot(contains('2')));

      final emptyQuery = container.read(filteredContentListProvider(''));
      expect(emptyQuery.length, 3); // all items returned

      // Wait for any pending async operations
      await Future.delayed(const Duration(milliseconds: 10));
    });

    test('filteredContentListProvider should search in tags', () async {
      final fakeApi = _FakeMemoNexusAPIClient();
      final container = ProviderContainer(
        overrides: [
          apiClientProvider.overrideWith((ref) => fakeApi),
        ],
      );
      addTearDown(container.dispose);

      // First read the provider to trigger auto-loading, then wait for it
      container.read(contentListProvider);
      await Future.delayed(const Duration(milliseconds: 10));

      final item1 = ContentItem(
        id: '1',
        title: 'Recipe',
        contentText: 'Content',
        mediaType: MediaType.web,
        tags: ['dessert', 'sweet'],
        createdAt: DateTime(2024, 1, 1),
        updatedAt: DateTime(2024, 1, 1),
        version: 1,
      );
      final item2 = ContentItem(
        id: '2',
        title: 'Another',
        contentText: 'Content',
        mediaType: MediaType.web,
        tags: ['savory'],
        createdAt: DateTime(2024, 1, 1),
        updatedAt: DateTime(2024, 1, 1),
        version: 1,
      );

      // Now safely set the state after auto-load is complete
      container.read(contentListProvider.notifier).state = ContentListState(
        items: [item1, item2],
      );

      final sweetResults = container.read(filteredContentListProvider('sweet'));
      expect(sweetResults.length, 1);
      expect(sweetResults.first.id, '1');

      // Wait for any pending async operations
      await Future.delayed(const Duration(milliseconds: 10));
    });

    test('filteredContentListProvider should be case-insensitive', () async {
      final fakeApi = _FakeMemoNexusAPIClient();
      final container = ProviderContainer(
        overrides: [
          apiClientProvider.overrideWith((ref) => fakeApi),
        ],
      );
      addTearDown(container.dispose);

      // First read the provider to trigger auto-loading, then wait for it
      container.read(contentListProvider);
      await Future.delayed(const Duration(milliseconds: 10));

      final item = ContentItem(
        id: '1',
        title: 'APPLE PIE',
        contentText: 'Delicious',
        mediaType: MediaType.web,
        tags: [],
        createdAt: DateTime(2024, 1, 1),
        updatedAt: DateTime(2024, 1, 1),
        version: 1,
      );

      // Now safely set the state after auto-load is complete
      container.read(contentListProvider.notifier).state = ContentListState(
        items: [item],
      );

      final results = container.read(filteredContentListProvider('apple'));
      expect(results.length, 1);
      expect(results.first.id, '1');

      final resultsCaps = container.read(filteredContentListProvider('APPLE'));
      expect(resultsCaps.length, 1);
      expect(resultsCaps.first.id, '1');

      // Wait for any pending async operations
      await Future.delayed(const Duration(milliseconds: 10));
    });
  });

  group('ContentListNotifier loadMore', () {
    test('should load more items when hasMore is true', () async {
      final fakeApi = _FakeMemoNexusAPIClient();
      final container = ProviderContainer(
        overrides: [
          apiClientProvider.overrideWith((ref) => fakeApi),
        ],
      );
      addTearDown(container.dispose);

      final notifier = container.read(contentListProvider.notifier);

      // Set up state with hasMore = true
      notifier.state = const ContentListState(
        items: [],
        isLoading: false,
        currentPage: 2,
        hasMore: true,
      );

      await notifier.loadMore();

      expect(notifier.state.isLoading, false);
      // Wait for async operations
      await Future.delayed(const Duration(milliseconds: 10));
    });

    test('should not load more when hasMore is false', () async {
      final fakeApi = _FakeMemoNexusAPIClient();
      final container = ProviderContainer(
        overrides: [
          apiClientProvider.overrideWith((ref) => fakeApi),
        ],
      );
      addTearDown(container.dispose);

      final notifier = container.read(contentListProvider.notifier);

      // Set up state with hasMore = false
      notifier.state = const ContentListState(
        items: [],
        isLoading: false,
        currentPage: 2,
        hasMore: false,
      );

      final oldPage = notifier.state.currentPage;
      final oldHasMore = notifier.state.hasMore;

      await notifier.loadMore();

      // Page and hasMore should not change (loadMore is skipped when hasMore is false)
      expect(notifier.state.currentPage, oldPage);
      expect(notifier.state.hasMore, oldHasMore);
    });

    test('should not load more when already loading', () async {
      final fakeApi = _FakeMemoNexusAPIClient();
      final container = ProviderContainer(
        overrides: [
          apiClientProvider.overrideWith((ref) => fakeApi),
        ],
      );
      addTearDown(container.dispose);

      final notifier = container.read(contentListProvider.notifier);

      // Set up state with isLoading = true
      notifier.state = const ContentListState(
        items: [],
        isLoading: true,
        currentPage: 2,
        hasMore: true,
      );

      final oldPage = notifier.state.currentPage;

      await notifier.loadMore();

      // Page should not change (loadMore is skipped when already loading)
      expect(notifier.state.currentPage, oldPage);
    });
  });

  group('ContentListNotifier refresh', () {
    test('should refresh items and reset page', () async {
      final fakeApi = _FakeMemoNexusAPIClient();
      final container = ProviderContainer(
        overrides: [
          apiClientProvider.overrideWith((ref) => fakeApi),
        ],
      );
      addTearDown(container.dispose);

      final notifier = container.read(contentListProvider.notifier);

      // Set up initial state
      notifier.state = const ContentListState(
        items: [],
        isLoading: false,
        currentPage: 3,
        hasMore: true,
      );

      await notifier.refresh();

      // Wait for async operations
      await Future.delayed(const Duration(milliseconds: 10));

      expect(notifier.state.isLoading, false);
      // refresh calls loadItems(refresh: true) which passes page=1, then sets currentPage to page + 1 = 2
      expect(notifier.state.currentPage, 2);
    });
  });

  group('ContentListNotifier setMediaTypeFilter', () {
    test('should update filter and reload when different', () async {
      final fakeApi = _FakeMemoNexusAPIClient();
      final container = ProviderContainer(
        overrides: [
          apiClientProvider.overrideWith((ref) => fakeApi),
        ],
      );
      addTearDown(container.dispose);

      final notifier = container.read(contentListProvider.notifier);

      notifier.state = const ContentListState(
        items: [],
        isLoading: false,
        activeFilterMediaType: 'web',
        currentPage: 3,
      );

      notifier.setMediaTypeFilter('pdf');

      // Wait for async loadItems to complete
      await Future.delayed(const Duration(milliseconds: 10));

      expect(notifier.state.activeFilterMediaType, 'pdf');
      // loadItems(refresh: true) passes page=1, then sets currentPage to page + 1 = 2
      expect(notifier.state.currentPage, 2);
      // mock returns has_more: false, and items.length < 20 so hasMore becomes false
      expect(notifier.state.hasMore, false);
    });

    test('should not reload when filter is same', () async {
      final fakeApi = _FakeMemoNexusAPIClient();
      final container = ProviderContainer(
        overrides: [
          apiClientProvider.overrideWith((ref) => fakeApi),
        ],
      );
      addTearDown(container.dispose);

      final notifier = container.read(contentListProvider.notifier);

      // Start with default state (no filters set)
      expect(notifier.state.activeFilterMediaType, isNull);

      // Set filter to 'web' first
      notifier.setMediaTypeFilter('web');
      await Future.delayed(const Duration(milliseconds: 10));

      expect(notifier.state.activeFilterMediaType, 'web');
      final pageAfterFirstSet = notifier.state.currentPage;

      // Call setMediaTypeFilter with same value 'web'
      notifier.setMediaTypeFilter('web');
      await Future.delayed(const Duration(milliseconds: 10));

      // Page should not change if filter is the same (no reload)
      expect(notifier.state.currentPage, pageAfterFirstSet);
    });
  });

  group('ContentListNotifier setTagFilter', () {
    test('should update tag filter and reload when different', () async {
      final fakeApi = _FakeMemoNexusAPIClient();
      final container = ProviderContainer(
        overrides: [
          apiClientProvider.overrideWith((ref) => fakeApi),
        ],
      );
      addTearDown(container.dispose);

      final notifier = container.read(contentListProvider.notifier);

      notifier.state = const ContentListState(
        items: [],
        isLoading: false,
        activeFilterTag: 'old-tag',
        currentPage: 3,
      );

      notifier.setTagFilter('new-tag');

      // Wait for async loadItems to complete
      await Future.delayed(const Duration(milliseconds: 10));

      expect(notifier.state.activeFilterTag, 'new-tag');
      // loadItems(refresh: true) passes page=1, then sets currentPage to page + 1 = 2
      expect(notifier.state.currentPage, 2);
      // mock returns has_more: false, and items.length < 20 so hasMore becomes false
      expect(notifier.state.hasMore, false);
    });

    test('should not reload when tag is same', () async {
      final fakeApi = _FakeMemoNexusAPIClient();
      final container = ProviderContainer(
        overrides: [
          apiClientProvider.overrideWith((ref) => fakeApi),
        ],
      );
      addTearDown(container.dispose);

      final notifier = container.read(contentListProvider.notifier);

      // Start with default state (no filters set)
      expect(notifier.state.activeFilterTag, isNull);

      // Set filter to 'same-tag' first
      notifier.setTagFilter('same-tag');
      await Future.delayed(const Duration(milliseconds: 10));

      expect(notifier.state.activeFilterTag, 'same-tag');
      final pageAfterFirstSet = notifier.state.currentPage;

      // Call setTagFilter with same value
      notifier.setTagFilter('same-tag');
      await Future.delayed(const Duration(milliseconds: 10));

      // Page should not change if tag is the same (no reload)
      expect(notifier.state.currentPage, pageAfterFirstSet);
    });
  });

  group('ContentListNotifier clearFilters', () {
    test('should clear filters and reload', () async {
      final fakeApi = _FakeMemoNexusAPIClient();
      final container = ProviderContainer(
        overrides: [
          apiClientProvider.overrideWith((ref) => fakeApi),
        ],
      );
      addTearDown(container.dispose);

      final notifier = container.read(contentListProvider.notifier);

      notifier.state = const ContentListState(
        items: [],
        isLoading: false,
        activeFilterMediaType: 'web',
        activeFilterTag: 'test',
        currentPage: 3,
      );

      notifier.clearFilters();

      // Wait for async loadItems to complete
      await Future.delayed(const Duration(milliseconds: 10));

      expect(notifier.state.activeFilterMediaType, isNull);
      expect(notifier.state.activeFilterTag, isNull);
      // loadItems(refresh: true) passes page=1, then sets currentPage to page + 1 = 2
      expect(notifier.state.currentPage, 2);
    });

    test('should not reload when no filters set', () async {
      final fakeApi = _FakeMemoNexusAPIClient();
      final container = ProviderContainer(
        overrides: [
          apiClientProvider.overrideWith((ref) => fakeApi),
        ],
      );
      addTearDown(container.dispose);

      final notifier = container.read(contentListProvider.notifier);

      notifier.state = const ContentListState(
        items: [],
        isLoading: false,
        currentPage: 3,
      );

      final oldState = notifier.state;

      notifier.clearFilters();

      // State should not change
      expect(identical(notifier.state, oldState), isTrue);
    });
  });

  group('ContentListNotifier deleteItem', () {
    test('should remove item from list', () async {
      final fakeApi = _FakeMemoNexusAPIClient();
      final item1 = ContentItem(
        id: '1',
        title: 'Item 1',
        contentText: 'Content 1',
        mediaType: MediaType.web,
        tags: [],
        createdAt: DateTime(2024, 1, 1),
        updatedAt: DateTime(2024, 1, 1),
        version: 1,
      );
      final item2 = ContentItem(
        id: '2',
        title: 'Item 2',
        contentText: 'Content 2',
        mediaType: MediaType.web,
        tags: [],
        createdAt: DateTime(2024, 1, 1),
        updatedAt: DateTime(2024, 1, 1),
        version: 1,
      );

      final container = ProviderContainer(
        overrides: [
          apiClientProvider.overrideWith((ref) => fakeApi),
        ],
      );
      addTearDown(container.dispose);

      final notifier = container.read(contentListProvider.notifier);

      notifier.state = ContentListState(items: [item1, item2]);

      await notifier.deleteItem('1');

      expect(notifier.state.items.length, 1);
      expect(notifier.state.items.first.id, '2');
    });

    test('should set error on delete failure', () async {
      final fakeApi = _FakeMemoNexusAPIClient();
      fakeApi.shouldFail = true;
      fakeApi.errorMessage = 'Delete failed';

      final container = ProviderContainer(
        overrides: [
          apiClientProvider.overrideWith((ref) => fakeApi),
        ],
      );
      addTearDown(container.dispose);

      final notifier = container.read(contentListProvider.notifier);

      await notifier.deleteItem('1');

      expect(notifier.state.error, isNotNull);
      expect(notifier.state.error, contains('Delete failed'));
    });
  });

  group('contentItemProvider', () {
    test('should fetch individual content item', () async {
      final fakeApi = _FakeMemoNexusAPIClient();
      final container = ProviderContainer(
        overrides: [
          apiClientProvider.overrideWith((ref) => fakeApi),
        ],
      );
      addTearDown(container.dispose);

      // For FutureProvider, we can directly await the future
      final item = await container.read(contentItemProvider('test-id').future);

      expect(item, isA<ContentItem>());
      expect(item.id, 'test-id');
    });
  });
}
