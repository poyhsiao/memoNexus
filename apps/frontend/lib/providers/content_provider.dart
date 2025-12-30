// Content Provider for Riverpod state management
// Manages content item listing, filtering, and CRUD operations

import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/content_item.dart';
import '../services/api_client.dart';

// =====================================================
// API Client Provider
// =====================================================

final apiClientProvider = Provider<MemoNexusAPIClient>((ref) {
  return MemoNexusAPIClient(
    baseUrl: 'http://localhost:8090/api',
    token: 'X-Local-Token',
  );
});

// =====================================================
// Content List State
// =====================================================

class ContentListState {
  final List<ContentItem> items;
  final bool isLoading;
  final String? error;
  final int currentPage;
  final bool hasMore;
  final String? activeFilterMediaType;
  final String? activeFilterTag;

  const ContentListState({
    this.items = const [],
    this.isLoading = false,
    this.error,
    this.currentPage = 1,
    this.hasMore = true,
    this.activeFilterMediaType,
    this.activeFilterTag,
  });

  ContentListState copyWith({
    List<ContentItem>? items,
    bool? isLoading,
    String? error,
    int? currentPage,
    bool? hasMore,
    String? activeFilterMediaType,
    String? activeFilterTag,
  }) {
    return ContentListState(
      items: items ?? this.items,
      isLoading: isLoading ?? this.isLoading,
      error: error ?? this.error,
      currentPage: currentPage ?? this.currentPage,
      hasMore: hasMore ?? this.hasMore,
      activeFilterMediaType: activeFilterMediaType ?? this.activeFilterMediaType,
      activeFilterTag: activeFilterTag ?? this.activeFilterTag,
    );
  }
}

// =====================================================
// Content List Notifier
// =====================================================

class ContentListNotifier extends StateNotifier<ContentListState> {
  final MemoNexusAPIClient _api;

  ContentListNotifier(this._api) : super(const ContentListState()) {
    loadItems();
  }

  Future<void> loadItems({bool refresh = false}) async {
    if (state.isLoading) return;

    final page = refresh ? 1 : state.currentPage;
    state = state.copyWith(
      isLoading: true,
      error: null,
    );

    try {
      final data = await _api.listContentItems(
        page: page,
        perPage: 20,
        sort: 'created_at',
        order: 'desc',
        mediaType: state.activeFilterMediaType,
        tag: state.activeFilterTag,
      );

      final itemsJson = data['items'] as List<dynamic>;
      final items = itemsJson
          .map((json) => ContentItem.fromJson(json as Map<String, dynamic>))
          .toList();

      state = state.copyWith(
        items: refresh ? items : [...state.items, ...items],
        isLoading: false,
        currentPage: page + 1,
        hasMore: items.length == 20,
      );
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: e.toString(),
      );
    }
  }

  Future<void> loadMore() async {
    if (state.hasMore && !state.isLoading) {
      await loadItems();
    }
  }

  Future<void> refresh() async {
    await loadItems(refresh: true);
  }

  void setMediaTypeFilter(String? mediaType) {
    if (state.activeFilterMediaType != mediaType) {
      state = state.copyWith(
        activeFilterMediaType: mediaType,
        currentPage: 1,
        hasMore: true,
      );
      loadItems(refresh: true);
    }
  }

  void setTagFilter(String? tag) {
    if (state.activeFilterTag != tag) {
      state = state.copyWith(
        activeFilterTag: tag,
        currentPage: 1,
        hasMore: true,
      );
      loadItems(refresh: true);
    }
  }

  void clearFilters() {
    if (state.activeFilterMediaType != null || state.activeFilterTag != null) {
      state = const ContentListState();
      loadItems(refresh: true);
    }
  }

  Future<void> deleteItem(String id) async {
    try {
      await _api.deleteContentItem(id);
      state = state.copyWith(
        items: state.items.where((item) => item.id != id).toList(),
      );
    } catch (e) {
      state = state.copyWith(error: e.toString());
    }
  }
}

// =====================================================
// Providers
// =====================================================

final contentListProvider =
    StateNotifierProvider<ContentListNotifier, ContentListState>((ref) {
  final api = ref.watch(apiClientProvider);
  return ContentListNotifier(api);
});

// Individual content item provider
final contentItemProvider =
    FutureProvider.family<ContentItem, String>((ref, id) async {
  final api = ref.watch(apiClientProvider);
  final data = await api.getContentItem(id);
  return ContentItem.fromJson(data);
});

// Filtered content list (for search results)
final filteredContentListProvider =
    Provider.family<List<ContentItem>, String>((ref, query) {
  final contentState = ref.watch(contentListProvider);
  final items = contentState.items;

  if (query.isEmpty) return items;

  final lowerQuery = query.toLowerCase();
  return items.where((item) {
    return item.title.toLowerCase().contains(lowerQuery) ||
        item.contentText.toLowerCase().contains(lowerQuery) ||
        item.tags.any((tag) => tag.toLowerCase().contains(lowerQuery));
  }).toList();
});
