// Search Provider for Riverpod state management
// Manages search query, filters, and results with FTS5 integration
// T123: SearchResults Riverpod provider

import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/content_item.dart';
import '../services/api_client.dart';
import 'content_provider.dart' show apiClientProvider;

// =====================================================
// Search State
// =====================================================

class SearchState {
  final String query;
  final List<SearchResult> results;
  final int total;
  final bool isLoading;
  final String? error;
  final String? filterMediaType;
  final String? filterTags;
  final int? filterDateFrom;
  final int? filterDateTo;

  const SearchState({
    this.query = '',
    this.results = const [],
    this.total = 0,
    this.isLoading = false,
    this.error,
    this.filterMediaType,
    this.filterTags,
    this.filterDateFrom,
    this.filterDateTo,
  });

  SearchState copyWith({
    String? query,
    List<SearchResult>? results,
    int? total,
    bool? isLoading,
    String? error,
    String? filterMediaType,
    String? filterTags,
    int? filterDateFrom,
    int? filterDateTo,
    bool clearFilters = false,
  }) {
    return SearchState(
      query: query ?? this.query,
      results: results ?? this.results,
      total: total ?? this.total,
      isLoading: isLoading ?? this.isLoading,
      error: error,
      filterMediaType: clearFilters ? null : (filterMediaType ?? this.filterMediaType),
      filterTags: clearFilters ? null : (filterTags ?? this.filterTags),
      filterDateFrom: clearFilters ? null : (filterDateFrom ?? this.filterDateFrom),
      filterDateTo: clearFilters ? null : (filterDateTo ?? this.filterDateTo),
    );
  }

  bool get hasQuery => query.trim().isNotEmpty;
  bool get hasFilters => filterMediaType != null || filterTags != null;
  bool get isEmpty => !isLoading && results.isEmpty;
  bool get hasError => error != null;
}

// =====================================================
// Search Notifier
// =====================================================

class SearchNotifier extends StateNotifier<SearchState> {
  final MemoNexusAPIClient _api;

  SearchNotifier(this._api) : super(const SearchState());

  /// Execute search with current query and filters
  Future<void> search(String query) async {
    if (query.trim().isEmpty) {
      state = const SearchState();
      return;
    }

    state = state.copyWith(
      query: query,
      isLoading: true,
      error: null,
    );

    try {
      final data = await _api.search(
        query: query,
        limit: 20,
        mediaType: state.filterMediaType,
        tags: state.filterTags,
        dateFrom: state.filterDateFrom,
        dateTo: state.filterDateTo,
      );

      final resultsJson = data['results'] as List<dynamic>;
      final results = resultsJson
          .map((json) => SearchResult.fromJson(json as Map<String, dynamic>))
          .toList();

      final total = data['total'] as int? ?? results.length;

      state = state.copyWith(
        results: results,
        total: total,
        isLoading: false,
      );
    } catch (e) {
      state = state.copyWith(
        results: [],
        total: 0,
        isLoading: false,
        error: e.toString(),
      );
    }
  }

  /// Clear search results and reset state
  void clear() {
    state = const SearchState();
  }

  /// Set media type filter and re-run search if has query
  Future<void> setMediaTypeFilter(String? mediaType) async {
    final newFilter = mediaType?.isEmpty ?? true ? null : mediaType;
    if (state.filterMediaType != newFilter) {
      state = state.copyWith(filterMediaType: newFilter);
      if (state.hasQuery) {
        await search(state.query);
      }
    }
  }

  /// Set tags filter and re-run search if has query
  Future<void> setTagsFilter(String? tags) async {
    final newFilter = tags?.isEmpty ?? true ? null : tags;
    if (state.filterTags != newFilter) {
      state = state.copyWith(filterTags: newFilter);
      if (state.hasQuery) {
        await search(state.query);
      }
    }
  }

  /// Set date range filter and re-run search if has query
  Future<void> setDateRangeFilter(int? from, int? to) async {
    if (state.filterDateFrom != from || state.filterDateTo != to) {
      state = state.copyWith(
        filterDateFrom: from,
        filterDateTo: to,
      );
      if (state.hasQuery) {
        await search(state.query);
      }
    }
  }

  /// Clear all filters and re-run search if has query
  Future<void> clearFilters() async {
    if (state.hasFilters) {
      state = state.copyWith(clearFilters: true);
      if (state.hasQuery) {
        await search(state.query);
      }
    }
  }

  /// Get active filter count
  int get activeFilterCount {
    var count = 0;
    if (state.filterMediaType != null) count++;
    if (state.filterTags != null) count++;
    if (state.filterDateFrom != null || state.filterDateTo != null) count++;
    return count;
  }
}

// =====================================================
// Providers
// =====================================================

/// Search provider - main search state and operations
final searchProvider =
    StateNotifierProvider<SearchNotifier, SearchState>((ref) {
  final api = ref.watch(apiClientProvider);
  return SearchNotifier(api);
});

/// Recent searches provider (stored in memory for session)
final recentSearchesProvider = StateProvider<List<String>>((ref) => []);

/// Add query to recent searches
void addRecentSearch(WidgetRef ref, String query) {
  final recent = ref.read(recentSearchesProvider);
  final updated = [query, ...recent]..toSet().toList();
  ref.read(recentSearchesProvider.notifier).state = updated.take(10).toList();
}

/// Clear recent searches
void clearRecentSearches(WidgetRef ref) {
  ref.read(recentSearchesProvider.notifier).state = [];
}
