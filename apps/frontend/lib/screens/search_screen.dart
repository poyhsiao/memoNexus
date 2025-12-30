// Search Screen
// T120: Main search screen with query input, filters, and results list
// T124: Navigation to content detail screen

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/content_item.dart';
import '../providers/search_provider.dart';
import '../widgets/search_result_item.dart';
import '../widgets/search_filters.dart';
import 'content_detail_screen.dart';

class SearchScreen extends ConsumerStatefulWidget {
  const SearchScreen({super.key});

  @override
  ConsumerState<SearchScreen> createState() => _SearchScreenState();
}

class _SearchScreenState extends ConsumerState<SearchScreen> {
  final TextEditingController _searchController = TextEditingController();
  final FocusNode _searchFocusNode = FocusNode();

  @override
  void dispose() {
    _searchController.dispose();
    _searchFocusNode.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final searchState = ref.watch(searchProvider);
    final recentSearches = ref.watch(recentSearchesProvider);

    return Scaffold(
      body: SafeArea(
        child: Column(
          children: [
            // Search header
            _buildSearchHeader(context, searchState),

            // Quick filter chips
            const QuickFilterChips(),

            // Content
            Expanded(
              child: _buildContent(context, searchState, recentSearches),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSearchHeader(BuildContext context, SearchState state) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        border: Border(
          bottom: BorderSide(
            color: Theme.of(context).dividerColor,
            width: 1,
          ),
        ),
      ),
      child: Row(
        children: [
          // Back button
          if (Navigator.canPop(context))
            IconButton(
              icon: const Icon(Icons.arrow_back),
              onPressed: () => Navigator.pop(context),
            ),

          // Search field
          Expanded(
            child: TextField(
              controller: _searchController,
              focusNode: _searchFocusNode,
              autofocus: true,
              decoration: InputDecoration(
                hintText: 'Search content...',
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(28),
                ),
                filled: true,
                fillColor: Theme.of(context).colorScheme.surfaceContainerHighest,
                prefixIcon: const Icon(Icons.search),
                suffixIcon: state.query.isNotEmpty
                    ? IconButton(
                        icon: const Icon(Icons.clear),
                        onPressed: () {
                          _searchController.clear();
                          ref.read(searchProvider.notifier).clear();
                        },
                      )
                    : null,
              ),
              onSubmitted: (query) {
                if (query.trim().isNotEmpty) {
                  _performSearch(context, query);
                }
              },
              onChanged: (value) {
                // Real-time search debouncing would go here
                // For now, search on submit
              },
            ),
          ),

          // Filter button
          if (state.hasFilters)
            IconButton(
              icon: Icon(
                Icons.filter_list,
                color: Theme.of(context).colorScheme.primary,
              ),
              onPressed: () {
                _showFilterBottomSheet(context);
              },
            )
          else
            IconButton(
              icon: const Icon(Icons.filter_list_outlined),
              onPressed: () {
                _showFilterBottomSheet(context);
              },
            ),
        ],
      ),
    );
  }

  Widget _buildContent(
    BuildContext context,
    SearchState state,
    List<String> recentSearches,
  ) {
    if (state.isLoading) {
      return const Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            CircularProgressIndicator(),
            SizedBox(height: 16),
            Text('Searching...'),
          ],
        ),
      );
    }

    if (state.hasError) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.error_outline, size: 64, color: Colors.red),
            const SizedBox(height: 16),
            Text(
              'Search Error',
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: 8),
            Text(
              state.error!,
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: Colors.grey[600],
                  ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 16),
            FilledButton.icon(
              onPressed: () {
                if (state.query.isNotEmpty) {
                  ref.read(searchProvider.notifier).search(state.query);
                }
              },
              icon: const Icon(Icons.refresh),
              label: const Text('Retry'),
            ),
          ],
        ),
      );
    }

    if (!state.hasQuery) {
      // Show recent searches
      if (recentSearches.isNotEmpty) {
        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Padding(
              padding: const EdgeInsets.all(16),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    'Recent Searches',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  TextButton.icon(
                    onPressed: () => clearRecentSearches(ref),
                    icon: const Icon(Icons.clear_all, size: 18),
                    label: const Text('Clear'),
                  ),
                ],
              ),
            ),
            Expanded(
              child: ListView.builder(
                itemCount: recentSearches.length,
                itemBuilder: (context, index) {
                  final query = recentSearches[index];
                  return ListTile(
                    leading: const Icon(Icons.history),
                    title: Text(query),
                    trailing: const Icon(Icons.arrow_forward_ios, size: 16),
                    onTap: () {
                      _searchController.text = query;
                      _performSearch(context, query);
                    },
                  );
                },
              ),
            ),
          ],
        );
      }

      // Empty state with tips
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.search,
              size: 80,
              color: Colors.grey[400],
            ),
            const SizedBox(height: 16),
            Text(
              'Search Your Knowledge Base',
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: 8),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 32),
              child: Text(
                'Enter keywords to search through titles, content, and tags',
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      color: Colors.grey[600],
                    ),
                textAlign: TextAlign.center,
              ),
            ),
            const SizedBox(height: 32),
            _buildSearchTips(context),
          ],
        ),
      );
    }

    if (state.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.search_off,
              size: 64,
              color: Colors.grey[400],
            ),
            const SizedBox(height: 16),
            Text(
              'No Results Found',
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: 8),
            Text(
              'Try different keywords or adjust filters',
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: Colors.grey[600],
                  ),
            ),
          ],
        ),
      );
    }

    // Show results
    return Column(
      children: [
        // Results count
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
          alignment: Alignment.centerLeft,
          child: Text(
            '${state.total} result${state.total == 1 ? '' : 's'} for "${state.query}"',
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: Colors.grey[600],
                ),
          ),
        ),

        // Results list
        Expanded(
          child: ListView.builder(
            itemCount: state.results.length,
            itemBuilder: (context, index) {
              final result = state.results[index];
              return SearchResultItemWidget(
                result: result,
                onTap: () => _openContentDetail(context, result.item),
                onLongPress: () => _showContentOptions(context, result.item),
              );
            },
          ),
        ),
      ],
    );
  }

  Widget _buildSearchTips(BuildContext context) {
    final tips = [
      'Use quotes for exact phrases: "machine learning"',
      'Combine terms: flutter database',
      'Filter by media type using the chips above',
    ];

    return Container(
      padding: const EdgeInsets.all(16),
      margin: const EdgeInsets.symmetric(horizontal: 32),
      decoration: BoxDecoration(
        color: Colors.blue.shade50,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.blue.shade200),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.lightbulb, color: Colors.blue.shade700, size: 20),
              const SizedBox(width: 8),
              Text(
                'Search Tips',
                style: Theme.of(context).textTheme.labelLarge?.copyWith(
                      color: Colors.blue.shade700,
                    ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          ...tips.map((tip) => Padding(
                padding: const EdgeInsets.only(bottom: 4),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text('• ', style: TextStyle(color: Colors.blue.shade700)),
                    Expanded(
                      child: Text(
                        tip,
                        style: TextStyle(
                          color: Colors.blue.shade900,
                          fontSize: 13,
                        ),
                      ),
                    ),
                  ],
                ),
              )),
        ],
      ),
    );
  }

  void _performSearch(BuildContext context, String query) {
    if (query.trim().isEmpty) return;
    addRecentSearch(ref, query);
    ref.read(searchProvider.notifier).search(query);
  }

  void _showFilterBottomSheet(BuildContext context) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      builder: (context) => DraggableScrollableSheet(
        initialChildSize: 0.6,
        minChildSize: 0.4,
        maxChildSize: 0.9,
        expand: false,
        builder: (context, scrollController) => Column(
          children: [
            AppBar(
              title: const Text('Filters'),
              leading: IconButton(
                icon: const Icon(Icons.close),
                onPressed: () => Navigator.pop(context),
              ),
            ),
            Expanded(
              child: ListView(
                controller: scrollController,
                padding: const EdgeInsets.only(bottom: 16),
                children: [
                  SearchFiltersWidget(
                    onFilterChanged: () {
                      // Auto-close and re-search if has query
                      Navigator.pop(context);
                      final searchState = ref.read(searchProvider);
                      if (searchState.hasQuery) {
                        ref.read(searchProvider.notifier).search(searchState.query);
                      }
                    },
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _openContentDetail(BuildContext context, ContentItem item) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => ContentDetailScreen(itemId: item.id),
      ),
    );
  }

  void _showContentOptions(BuildContext context, ContentItem item) {
    showModalBottomSheet(
      context: context,
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.visibility),
              title: const Text('View Details'),
              onTap: () {
                Navigator.pop(context);
                _openContentDetail(context, item);
              },
            ),
            ListTile(
              leading: const Icon(Icons.share),
              title: const Text('Share'),
              onTap: () {
                Navigator.pop(context);
                // TODO: Implement share
              },
            ),
          ],
        ),
      ),
    );
  }
}
