// Content List Widget
// Displays a scrollable list of content items with thumbnails, titles, tags, and timestamps
// T220: Optimized with itemExtent for virtual scrolling (render within 500ms for 1,000 items)
// T203: Keyboard navigation support - arrow keys to navigate, Enter to select

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/content_item.dart';
import '../providers/content_provider.dart';
import '../screens/content_detail_screen.dart';

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
  final FocusNode _focusNode = FocusNode();
  int _focusedIndex = -1;

  @override
  void initState() {
    super.initState();
    _scrollController.addListener(_onScroll);
  }

  @override
  void dispose() {
    _scrollController.removeListener(_onScroll);
    _scrollController.dispose();
    _focusNode.dispose();
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

  // T203: Handle keyboard navigation
  KeyEventResult _handleKeyEvent(KeyEvent event) {
    final filteredItems = ref.read(
      filteredContentListProvider(widget.searchQuery ?? ''),
    );

    if (filteredItems.isEmpty) {
      return KeyEventResult.ignored;
    }

    // Handle arrow key navigation and Enter selection
    if (event is KeyDownEvent) {
      if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
        setState(() {
          if (_focusedIndex < filteredItems.length - 1) {
            _focusedIndex++;
            _scrollToIndex(_focusedIndex);
          }
        });
        return KeyEventResult.handled;
      } else if (event.logicalKey == LogicalKeyboardKey.arrowUp) {
        setState(() {
          if (_focusedIndex > 0) {
            _focusedIndex--;
            _scrollToIndex(_focusedIndex);
          } else if (_focusedIndex == -1) {
            _focusedIndex = 0;
          }
        });
        return KeyEventResult.handled;
      } else if (event.logicalKey == LogicalKeyboardKey.enter && _focusedIndex >= 0) {
        final item = filteredItems[_focusedIndex];
        Navigator.push(
          context,
          MaterialPageRoute(
            builder: (context) => ContentDetailScreen(itemId: item.id),
          ),
        );
        return KeyEventResult.handled;
      }
    }

    return KeyEventResult.ignored;
  }

  // T203: Scroll to the focused index
  void _scrollToIndex(int index) {
    const itemHeight = 132.0; // Matches itemExtent
    final targetOffset = index * itemHeight;
    final currentOffset = _scrollController.offset;
    final viewportHeight = _scrollController.position.viewportDimension;

    // Scroll if the focused item is not visible
    if (targetOffset < currentOffset) {
      _scrollController.animateTo(
        targetOffset,
        duration: const Duration(milliseconds: 200),
        curve: Curves.easeInOut,
      );
    } else if (targetOffset + itemHeight > currentOffset + viewportHeight) {
      _scrollController.animateTo(
        targetOffset + itemHeight - viewportHeight,
        duration: const Duration(milliseconds: 200),
        curve: Curves.easeInOut,
      );
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

    // T203: Wrap with KeyboardListener and Focus for keyboard navigation
    return KeyboardListener(
      focusNode: _focusNode,
      onKeyEvent: _handleKeyEvent,
      child: RefreshIndicator(
        onRefresh: () async {
          await ref.read(contentListProvider.notifier).refresh();
        },
        // T220: Use ListView.builder with itemExtent for virtual scrolling
        // The separator is included in each item to maintain fixed item extent
        child: ListView.builder(
          controller: _scrollController,
          padding: const EdgeInsets.all(16),
          itemCount: filteredItems.length,
          // T220: Fixed item extent enables virtual scrolling optimization
          // Flutter can calculate scroll position without measuring each item
          // Estimated height: tile (~120px) + separator (12px) = 132px
          itemExtent: 132.0,
          itemBuilder: (context, index) {
            final item = filteredItems[index];
            // Include separator at the bottom of each item (except last)
            final showSeparator = index < filteredItems.length - 1;
            final isFocused = index == _focusedIndex;

            return Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                ContentListTile(
                  item: item,
                  isFocused: isFocused,
                  onTap: () {
                    setState(() {
                      _focusedIndex = index;
                    });
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (context) => ContentDetailScreen(itemId: item.id),
                      ),
                    );
                  },
                ),
                if (showSeparator) const SizedBox(height: 12),
              ],
            );
          },
        ),
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
// T203: Supports keyboard focus indicator
// T204: Includes semantic labels for screen readers
// =====================================================

class ContentListTile extends StatelessWidget {
  final ContentItem item;
  final VoidCallback onTap;
  final bool isFocused;

  const ContentListTile({
    super.key,
    required this.item,
    required this.onTap,
    this.isFocused = false,
  });

  @override
  Widget build(BuildContext context) {
    // T203: Visible focus indicator - thicker border when focused
    final borderSide = isFocused
        ? BorderSide(
            color: Theme.of(context).colorScheme.primary,
            width: 2.5,
          )
        : BorderSide.none;

    // T203: Focused background has slight tint
    final backgroundColor = isFocused
        ? Theme.of(context).colorScheme.primaryContainer.withValues(alpha: 0.3)
        : null;

    return Material(
      elevation: isFocused ? 4 : 2,
      borderRadius: BorderRadius.circular(12),
      color: backgroundColor,
      // T206: Focus management - visible border indicator
      child: Container(
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(12),
          border: Border.fromBorderSide(borderSide),
        ),
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(12),
          // T204: Semantics for screen readers
          child: Semantics(
            button: true,
            selected: isFocused,
            label: 'Content: ${item.title}, Type: ${item.mediaType.name}, Tags: ${item.tags.join(", ")}',
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
        ),
      ),
    );
  }

  Widget _buildThumbnail(BuildContext context) {
    return Container(
      width: 56,
      height: 56,
      decoration: BoxDecoration(
        color: _getMediaTypeColor(context).withValues(alpha: 0.1),
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
