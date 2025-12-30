// Search Filters Widget
// T122: Search filter widget with media type, tags, and date range filters

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/search_provider.dart';
import '../models/content_item.dart';

class SearchFiltersWidget extends ConsumerStatefulWidget {
  final VoidCallback? onFilterChanged;

  const SearchFiltersWidget({
    super.key,
    this.onFilterChanged,
  });

  @override
  ConsumerState<SearchFiltersWidget> createState() =>
      _SearchFiltersWidgetState();
}

class _SearchFiltersWidgetState extends ConsumerState<SearchFiltersWidget> {
  bool _isExpanded = false;

  @override
  Widget build(BuildContext context) {
    final searchState = ref.watch(searchProvider);
    final hasActiveFilters = searchState.hasFilters;

    return Card(
      margin: const EdgeInsets.all(16),
      child: Column(
        children: [
          // Header with expand/collapse
          ListTile(
            leading: Icon(
              _isExpanded ? Icons.expand_less : Icons.expand_more,
            ),
            title: Text(
              'Filters',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            trailing: hasActiveFilters
                ? Chip(
                    label: Text('${ref.read(searchProvider.notifier).activeFilterCount}'),
                    backgroundColor: Colors.blue.shade100,
                  )
                : null,
            onTap: () {
              setState(() {
                _isExpanded = !_isExpanded;
              });
            },
          ),

          // Filter options (expandable)
          if (_isExpanded) ...[
            const Divider(height: 1),
            Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Media type filter
                  _buildMediaTypeFilter(context, searchState),
                  const SizedBox(height: 16),

                  // Tags filter
                  _buildTagsFilter(context, searchState),
                  const SizedBox(height: 16),

                  // Date range filter
                  _buildDateRangeFilter(context, searchState),
                  const SizedBox(height: 16),

                  // Clear filters button
                  if (hasActiveFilters)
                    SizedBox(
                      width: double.infinity,
                      child: OutlinedButton.icon(
                        onPressed: () async {
                          await ref.read(searchProvider.notifier).clearFilters();
                          widget.onFilterChanged?.call();
                        },
                        icon: const Icon(Icons.clear_all, size: 18),
                        label: const Text('Clear All Filters'),
                      ),
                    ),
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildMediaTypeFilter(BuildContext context, SearchState state) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Media Type',
          style: Theme.of(context).textTheme.labelLarge,
        ),
        const SizedBox(height: 8),
        Wrap(
          spacing: 8,
          children: MediaType.values.map((type) {
            final isSelected = state.filterMediaType == type.name;
            return FilterChip(
              label: Text(type.name.toUpperCase()),
              selected: isSelected,
              onSelected: (selected) async {
                await ref.read(searchProvider.notifier).setMediaTypeFilter(
                      selected ? type.name : null,
                    );
                widget.onFilterChanged?.call();
              },
              selectedColor: Colors.blue.shade100,
              checkmarkColor: Colors.blue,
            );
          }).toList(),
        ),
      ],
    );
  }

  Widget _buildTagsFilter(BuildContext context, SearchState state) {
    final controller = TextEditingController(text: state.filterTags ?? '');

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Tags',
          style: Theme.of(context).textTheme.labelLarge,
        ),
        const SizedBox(height: 8),
        TextField(
          controller: controller,
          decoration: const InputDecoration(
            hintText: 'comma, separated, tags',
            border: OutlineInputBorder(),
            prefixIcon: Icon(Icons.tag),
          ),
          onSubmitted: (value) async {
            await ref.read(searchProvider.notifier).setTagsFilter(
                  value.trim().isEmpty ? null : value.trim(),
                );
            widget.onFilterChanged?.call();
          },
        ),
        const SizedBox(height: 4),
        Text(
          'Enter tags separated by commas, press Enter to apply',
          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: Colors.grey[600],
              ),
        ),
      ],
    );
  }

  Widget _buildDateRangeFilter(BuildContext context, SearchState state) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Date Range',
          style: Theme.of(context).textTheme.labelLarge,
        ),
        const SizedBox(height: 8),
        Row(
          children: [
            Expanded(
              child: OutlinedButton.icon(
                onPressed: () => _selectDateFrom(context, state),
                icon: const Icon(Icons.calendar_today, size: 18),
                label: Text(
                  state.filterDateFrom != null
                      ? _formatTimestamp(state.filterDateFrom!)
                      : 'From Date',
                ),
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: OutlinedButton.icon(
                onPressed: () => _selectDateTo(context, state),
                icon: const Icon(Icons.calendar_today, size: 18),
                label: Text(
                  state.filterDateTo != null
                      ? _formatTimestamp(state.filterDateTo!)
                      : 'To Date',
                ),
              ),
            ),
          ],
        ),
        if (state.filterDateFrom != null || state.filterDateTo != null) ...[
          const SizedBox(height: 8),
          TextButton.icon(
            onPressed: () async {
              await ref.read(searchProvider.notifier).setDateRangeFilter(null, null);
              widget.onFilterChanged?.call();
            },
            icon: const Icon(Icons.clear, size: 16),
            label: const Text('Clear date range'),
            style: TextButton.styleFrom(
              padding: EdgeInsets.zero,
              minimumSize: Size.zero,
              tapTargetSize: MaterialTapTargetSize.shrinkWrap,
            ),
          ),
        ],
      ],
    );
  }

  Future<void> _selectDateFrom(BuildContext context, SearchState state) async {
    final initialDate = state.filterDateFrom != null
        ? DateTime.fromMillisecondsSinceEpoch(state.filterDateFrom! * 1000)
        : DateTime.now();

    final picked = await showDatePicker(
      context: context,
      initialDate: initialDate,
      firstDate: DateTime(2000),
      lastDate: DateTime.now().add(const Duration(days: 365)),
    );

    if (picked != null) {
      await ref.read(searchProvider.notifier).setDateRangeFilter(
            picked.millisecondsSinceEpoch ~/ 1000,
            state.filterDateTo,
          );
      widget.onFilterChanged?.call();
    }
  }

  Future<void> _selectDateTo(BuildContext context, SearchState state) async {
    final initialDate = state.filterDateTo != null
        ? DateTime.fromMillisecondsSinceEpoch(state.filterDateTo! * 1000)
        : DateTime.now();

    final picked = await showDatePicker(
      context: context,
      initialDate: initialDate,
      firstDate: DateTime(2000),
      lastDate: DateTime.now().add(const Duration(days: 365)),
    );

    if (picked != null) {
      await ref.read(searchProvider.notifier).setDateRangeFilter(
            state.filterDateFrom,
            picked.millisecondsSinceEpoch ~/ 1000,
          );
      widget.onFilterChanged?.call();
    }
  }

  String _formatTimestamp(int timestamp) {
    final date = DateTime.fromMillisecondsSinceEpoch(timestamp * 1000);
    return '${date.year}-${date.month.toString().padLeft(2, '0')}-${date.day.toString().padLeft(2, '0')}';
  }
}

// Compact filter chips for use in app bar
class QuickFilterChips extends ConsumerWidget {
  final ValueChanged<String?>? onMediaTypeChanged;

  const QuickFilterChips({
    super.key,
    this.onMediaTypeChanged,
  });

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final searchState = ref.watch(searchProvider);

    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Row(
        children: [
          // All
          FilterChip(
            label: const Text('All'),
            selected: searchState.filterMediaType == null,
            onSelected: (selected) async {
              await ref.read(searchProvider.notifier).setMediaTypeFilter(null);
              onMediaTypeChanged?.call(null);
            },
            selectedColor: Colors.blue.shade100,
          ),
          const SizedBox(width: 8),
          // Media types
          ...MediaType.values.map((type) {
            return Padding(
              padding: const EdgeInsets.only(right: 8),
              child: FilterChip(
                label: Text(type.name.toUpperCase()),
                selected: searchState.filterMediaType == type.name,
                onSelected: (selected) async {
                  await ref.read(searchProvider.notifier).setMediaTypeFilter(
                        selected ? type.name : null,
                      );
                  onMediaTypeChanged?.call(selected ? type.name : null);
                },
                selectedColor: Colors.blue.shade100,
              ),
            );
          }),
        ],
      ),
    );
  }
}
