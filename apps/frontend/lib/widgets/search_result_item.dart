// Search Result Item Widget
// T121: Search result item widget with highlighting and relevance display

import 'package:flutter/material.dart';
import '../models/content_item.dart';

class SearchResultItemWidget extends StatelessWidget {
  final SearchResult result;
  final VoidCallback? onTap;
  final VoidCallback? onLongPress;

  const SearchResultItemWidget({
    super.key,
    required this.result,
    this.onTap,
    this.onLongPress,
  });

  @override
  Widget build(BuildContext context) {
    final item = result.item;
    final theme = Theme.of(context);

    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: onTap,
        onLongPress: onLongPress,
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Header: media type icon and relevance score
              Row(
                children: [
                  _buildMediaTypeIcon(item.mediaType),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Text(
                      item.title,
                      style: theme.textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.w600,
                      ),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  _buildRelevanceBadge(result.relevance),
                ],
              ),
              const SizedBox(height: 8),

              // Content preview with matched terms
              if (item.contentText.isNotEmpty) ...[
                Text(
                  _getContentPreview(item.contentText, result.matchedTerms),
                  style: theme.textTheme.bodyMedium?.copyWith(
                    color: Colors.grey[700],
                  ),
                  maxLines: 3,
                  overflow: TextOverflow.ellipsis,
                ),
                const SizedBox(height: 8),
              ],

              // Footer: tags and metadata
              Wrap(
                spacing: 8,
                runSpacing: 4,
                children: [
                  if (item.tags.isNotEmpty) ...[
                    ...item.tags.take(3).map((tag) => Chip(
                          label: Text(
                            tag,
                            style: const TextStyle(fontSize: 12),
                          ),
                          visualDensity: VisualDensity.compact,
                          materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                        )),
                  ],
                  Text(
                    _formatDate(item.createdAt),
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: Colors.grey[600],
                    ),
                  ),
                  if (item.sourceUrl != null)
                    Text(
                      'URL',
                      style: theme.textTheme.bodySmall?.copyWith(
                        color: Colors.blue,
                      ),
                    ),
                ],
              ),

              // Matched terms indicator
              if (result.matchedTerms.isNotEmpty) ...[
                const SizedBox(height: 8),
                Wrap(
                  spacing: 4,
                  children: result.matchedTerms.take(5).map((term) {
                    return Chip(
                      label: Text(
                        term,
                        style: const TextStyle(fontSize: 10),
                      ),
                          backgroundColor: Colors.blue.shade50,
                          side: BorderSide(color: Colors.blue.shade200),
                          visualDensity: VisualDensity.compact,
                          materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                        );
                  }).toList(),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildMediaTypeIcon(MediaType mediaType) {
    IconData icon;
    Color color;

    switch (mediaType) {
      case MediaType.web:
        icon = Icons.language;
        color = Colors.blue;
        break;
      case MediaType.image:
        icon = Icons.image;
        color = Colors.green;
        break;
      case MediaType.video:
        icon = Icons.video_library;
        color = Colors.red;
        break;
      case MediaType.pdf:
        icon = Icons.picture_as_pdf;
        color = Colors.orange;
        break;
      case MediaType.markdown:
        icon = Icons.code;
        color = Colors.purple;
        break;
    }

    return Container(
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Icon(icon, color: color, size: 20),
    );
  }

  Widget _buildRelevanceBadge(double relevance) {
    // Convert BM25 score to percentage-like display
    // Higher relevance = better match
    final displayScore = relevance.toStringAsFixed(1);

    Color color;
    if (relevance >= 2.0) {
      color = Colors.green;
    } else if (relevance >= 1.0) {
      color = Colors.orange;
    } else {
      color = Colors.grey;
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withValues(alpha: 0.3)),
      ),
      child: Text(
        displayScore,
        style: TextStyle(
          fontSize: 12,
          fontWeight: FontWeight.w600,
          color: color,
        ),
      ),
    );
  }

  String _getContentPreview(String content, List<String> matchedTerms) {
    // Truncate content to preview length
    const maxLength = 150;
    if (content.length <= maxLength) return content;

    // Try to break at word boundary
    var preview = content.substring(0, maxLength);
    final lastSpace = preview.lastIndexOf(' ');
    if (lastSpace > maxLength * 2 ~/ 3) {
      preview = preview.substring(0, lastSpace);
    }

    return '$preview...';
  }

  String _formatDate(DateTime date) {
    final now = DateTime.now();
    final difference = now.difference(date);

    if (difference.inDays == 0) {
      return 'Today';
    } else if (difference.inDays == 1) {
      return 'Yesterday';
    } else if (difference.inDays < 7) {
      return '${difference.inDays}d ago';
    } else if (difference.inDays < 30) {
      return '${(difference.inDays / 7).floor()}w ago';
    } else if (difference.inDays < 365) {
      return '${(difference.inDays / 30).floor()}mo ago';
    } else {
      return '${(difference.inDays / 365).floor()}y ago';
    }
  }
}
