// Content List Widget
// Displays a scrollable list of content items with thumbnails, titles, tags, and timestamps

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/content_item.dart';
import '../providers/content_provider.dart';
import 'content_detail_screen.dart';

class ContentListWidget extends ConsumerStatefulWidget {
  final String? searchQuery;

  const ContentListWidget({
    super.key,
    this.searchQuery,
  });

  @override
  ConsumerState<ContentListWidget> createState() => _ContentListWidgetState();
}

class _ContentListWidgetState extends ConsumerState<ContentListWidget> {
  final ScrollController _scrollController = ScrollController();

  @override
  void initState() {
    super.initState();
    _scrollController.addListener(_onScroll);
  }

  @override
  void dispose() {
    _scrollController.removeListener(_onScroll);
    _scrollController.dispose();
    super.dispose();
  }

  void _onScroll() {
    if (_scrollController.position.pixels >=
        _scrollController.position.maxScrollExtent * 0.8) {
      final notifier = ref.read(contentListProvider.notifier);
      if (!notifier.state.isLoading) {
        notifier.loadMore();
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final filteredItems = ref.watch(
      filteredContentListProvider(widget.searchQuery ?? ''),
    );

    if (filteredItems.isEmpty) {
      return _buildEmptyState(context);
    }

    return RefreshIndicator(
      onRefresh: () async {
        await ref.read(contentListProvider.notifier).refresh();
      },
      child: ListView.separated(
        controller: _scrollController,
        padding: const EdgeInsets.all(16),
        itemCount: filteredItems.length,
        separatorBuilder: (context, index) => const SizedBox(height: 12),
        itemBuilder: (context, index) {
          final item = filteredItems[index];
          return ContentListTile(
            item: item,
            onTap: () {
              Navigator.push(
                context,
                MaterialPageRoute(
                  builder: (context) => ContentDetailScreen(itemId: item.id),
                ),
              );
            },
          );
        },
      ),
    );
  }

  Widget _buildEmptyState(BuildContext context) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.folder_open_outlined,
            size: 80,
            color: Colors.grey[400],
          ),
          const SizedBox(height: 16),
          Text(
            widget.searchQuery?.isEmpty ?? true
                ? 'No content yet'
                : 'No results for "${widget.searchQuery}"',
            style: Theme.of(context).textTheme.titleLarge?.copyWith(
                  color: Colors.grey[600],
                ),
          ),
          const SizedBox(height: 8),
          Text(
            'Tap the + button to add your first item',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: Colors.grey[500],
                ),
          ),
        ],
      ),
    );
  }
}

// =====================================================
// Content List Tile
// =====================================================

class ContentListTile extends StatelessWidget {
  final ContentItem item;
  final VoidCallback onTap;

  const ContentListTile({
    super.key,
    required this.item,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return Material(
      elevation: 2,
      borderRadius: BorderRadius.circular(12),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Row(
            children: [
              _buildThumbnail(context),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      item.title,
                      style: Theme.of(context).textTheme.titleMedium?.copyWith(
                            fontWeight: FontWeight.w600,
                          ),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 4),
                    Text(
                      item.contentText,
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                            color: Colors.grey[600],
                          ),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 8),
                    _buildMetadata(context),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildThumbnail(BuildContext context) {
    return Container(
      width: 56,
      height: 56,
      decoration: BoxDecoration(
        color: _getMediaTypeColor(context).withOpacity(0.1),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Icon(
        _getMediaTypeIcon(),
        color: _getMediaTypeColor(context),
        size: 28,
      ),
    );
  }

  IconData _getMediaTypeIcon() {
    switch (item.mediaType) {
      case MediaType.image:
        return Icons.image;
      case MediaType.video:
        return Icons.videocam;
      case MediaType.pdf:
        return Icons.picture_as_pdf;
      case MediaType.markdown:
        return Icons.description;
      case MediaType.web:
      default:
        return Icons.link;
    }
  }

  Color _getMediaTypeColor(BuildContext context) {
    switch (item.mediaType) {
      case MediaType.image:
        return Colors.purple;
      case MediaType.video:
        return Colors.red;
      case MediaType.pdf:
        return Colors.orange;
      case MediaType.markdown:
        return Colors.blue;
      case MediaType.web:
      default:
        return Theme.of(context).colorScheme.primary;
    }
  }

  Widget _buildMetadata(BuildContext context) {
    return Row(
      children: [
        ...item.tags.take(3).map((tag) {
          return Padding(
            padding: const EdgeInsets.only(right: 6),
            child: Chip(
              label: Text(
                tag,
                style: const TextStyle(fontSize: 10),
              ),
              padding: EdgeInsets.zero,
              materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
              visualDensity: VisualDensity.compact,
            ),
          );
        }),
        const Spacer(),
        Text(
          _formatTimestamp(item.updatedAt),
          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: Colors.grey[500],
                fontSize: 11,
              ),
        ),
      ],
    );
  }

  String _formatTimestamp(DateTime timestamp) {
    final now = DateTime.now();
    final difference = now.difference(timestamp);

    if (difference.inDays > 0) {
      return '${difference.inDays}d ago';
    } else if (difference.inHours > 0) {
      return '${difference.inHours}h ago';
    } else if (difference.inMinutes > 0) {
      return '${difference.inMinutes}m ago';
    } else {
      return 'Just now';
    }
  }
}
