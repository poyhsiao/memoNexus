// Keyword Suggestions Widget
// T143: Displays auto-generated keyword suggestions for tagging

import 'package:flutter/material.dart';

/// Keyword suggestions widget that displays AI/TextRank extracted keywords
/// as clickable chips that can be converted to tags.
class KeywordSuggestions extends StatelessWidget {
  /// List of suggested keywords
  final List<String> keywords;

  /// Callback when a keyword is selected to add as a tag
  final ValueChanged<String> onKeywordSelected;

  /// Callback to refresh/regenerate keywords
  final VoidCallback? onRefresh;

  /// Whether keywords are currently loading
  final bool isLoading;

  /// The method used to extract keywords (ai, textrank, tfidf)
  final String method;

  const KeywordSuggestions({
    super.key,
    required this.keywords,
    required this.onKeywordSelected,
    this.onRefresh,
    this.isLoading = false,
    this.method = 'tfidf',
  });

  @override
  Widget build(BuildContext context) {
    if (isLoading) {
      return const _KeywordLoadingView();
    }

    if (keywords.isEmpty) {
      return _KeywordEmptyView(
        onRefresh: onRefresh,
        method: method,
      );
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _buildHeader(context),
        const SizedBox(height: 12),
        _buildKeywordsGrid(context),
      ],
    );
  }

  Widget _buildHeader(BuildContext context) {
    return Row(
      children: [
        Icon(
          _getMethodIcon(),
          size: 20,
          color: _getMethodColor(),
        ),
        const SizedBox(width: 8),
        Text(
          'Keyword Suggestions',
          style: Theme.of(context).textTheme.titleMedium?.copyWith(
                fontWeight: FontWeight.w600,
              ),
        ),
        const Spacer(),
        // Method badge
        _MethodBadge(method: method),
        if (onRefresh != null) ...[
          const SizedBox(width: 8),
          IconButton(
            icon: const Icon(Icons.refresh, size: 18),
            onPressed: onRefresh,
            tooltip: 'Refresh keywords',
            visualDensity: VisualDensity.compact,
          ),
        ],
      ],
    );
  }

  Widget _buildKeywordsGrid(BuildContext context) {
    return Wrap(
      spacing: 8,
      runSpacing: 8,
      children: keywords
          .map((keyword) => _KeywordChip(
                keyword: keyword,
                onTap: () => onKeywordSelected(keyword),
              ))
          .toList(),
    );
  }

  IconData _getMethodIcon() {
    switch (method.toLowerCase()) {
      case 'ai':
        return Icons.psychology;
      case 'textrank':
        return Icons.graphic_eq;
      default:
        return Icons.text_fields;
    }
  }

  Color _getMethodColor() {
    switch (method.toLowerCase()) {
      case 'ai':
        return Colors.deepPurple;
      case 'textrank':
        return Colors.blue;
      default:
        return Colors.green;
    }
  }
}

// =====================================================
// Sub-widgets
// =====================================================

class _KeywordChip extends StatefulWidget {
  final String keyword;
  final VoidCallback onTap;

  const _KeywordChip({
    required this.keyword,
    required this.onTap,
  });

  @override
  State<_KeywordChip> createState() => _KeywordChipState();
}

class _KeywordChipState extends State<_KeywordChip> {
  bool _isPressed = false;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTapDown: (_) => setState(() => _isPressed = true),
      onTapUp: (_) => setState(() => _isPressed = false),
      onTapCancel: () => setState(() => _isPressed = false),
      onTap: widget.onTap,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 150),
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        decoration: BoxDecoration(
          color: _isPressed
              ? Colors.blue.withValues(alpha: 0.2)
              : Colors.blue.shade50,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: _isPressed
                ? Colors.blue.withValues(alpha: 0.5)
                : Colors.blue.shade200,
          ),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              Icons.add_circle_outline,
              size: 14,
              color: _isPressed ? Colors.blue.shade700 : Colors.blue.shade400,
            ),
            const SizedBox(width: 6),
            Text(
              widget.keyword,
              style: TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.w500,
                color: _isPressed ? Colors.blue.shade900 : Colors.blue.shade700,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _MethodBadge extends StatelessWidget {
  final String method;

  const _MethodBadge({required this.method});

  @override
  Widget build(BuildContext context) {
    final (label, color) = _getMethodInfo();

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withValues(alpha: 0.5)),
      ),
      child: Text(
        label,
        style: TextStyle(
          fontSize: 11,
          fontWeight: FontWeight.w600,
          color: color,
        ),
      ),
    );
  }

  (String, Color) _getMethodInfo() {
    switch (method.toLowerCase()) {
      case 'ai':
        return ('AI', Colors.deepPurple);
      case 'textrank':
        return ('TextRank', Colors.blue);
      case 'tfidf':
        return ('TF-IDF', Colors.green);
      default:
        return ('Extractive', Colors.grey);
    }
  }
}

// =====================================================
// Loading and Empty States
// =====================================================

class _KeywordLoadingView extends StatelessWidget {
  const _KeywordLoadingView();

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            const SizedBox(
              width: 16,
              height: 16,
              child: CircularProgressIndicator(strokeWidth: 2),
            ),
            const SizedBox(width: 12),
            Text(
              'Extracting keywords...',
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: Theme.of(context).colorScheme.onSurfaceVariant,
                  ),
            ),
          ],
        ),
      ],
    );
  }
}

class _KeywordEmptyView extends StatelessWidget {
  final VoidCallback? onRefresh;
  final String method;

  const _KeywordEmptyView({
    required this.onRefresh,
    required this.method,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: Theme.of(context).colorScheme.outline.withValues(alpha: 0.3),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(
                Icons.tag_outlined,
                size: 20,
                color: Theme.of(context).colorScheme.onSurfaceVariant,
              ),
              const SizedBox(width: 8),
              Text(
                'No Keyword Suggestions',
                style: Theme.of(context).textTheme.titleMedium?.copyWith(
                      color: Theme.of(context).colorScheme.onSurfaceVariant,
                    ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(
            'Extract keywords using ${_getMethodName()} to quickly tag your content.',
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: Theme.of(context).colorScheme.onSurfaceVariant,
                ),
          ),
          if (onRefresh != null) ...[
            const SizedBox(height: 12),
            FilledButton.tonalIcon(
              onPressed: onRefresh,
              icon: const Icon(Icons.auto_awesome, size: 16),
              label: const Text('Extract Keywords'),
            ),
          ],
        ],
      ),
    );
  }

  String _getMethodName() {
    switch (method.toLowerCase()) {
      case 'ai':
        return 'AI';
      case 'textrank':
        return 'TextRank';
      default:
        return 'TF-IDF';
    }
  }
}

// =====================================================
// Selected Keywords Widget
// =====================================================

/// Widget showing keywords that have been selected/converted to tags
class SelectedKeywords extends StatelessWidget {
  final List<String> keywords;
  final ValueChanged<String> onRemove;

  const SelectedKeywords({
    super.key,
    required this.keywords,
    required this.onRemove,
  });

  @override
  Widget build(BuildContext context) {
    if (keywords.isEmpty) {
      return const SizedBox.shrink();
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            const Icon(
              Icons.check_circle_outline,
              size: 18,
              color: Colors.green,
            ),
            const SizedBox(width: 8),
            Text(
              'Selected (${keywords.length})',
              style: Theme.of(context).textTheme.labelLarge?.copyWith(
                    color: Colors.green.shade700,
                    fontWeight: FontWeight.w600,
                  ),
            ),
          ],
        ),
        const SizedBox(height: 8),
        Wrap(
          spacing: 8,
          runSpacing: 8,
          children: keywords
              .map((keyword) => Chip(
                    label: Text(keyword),
                    onDeleted: () => onRemove(keyword),
                    deleteIconColor: Colors.red.shade400,
                    backgroundColor: Colors.green.shade50,
                    side: BorderSide(color: Colors.green.shade200),
                  ))
              .toList(),
        ),
      ],
    );
  }
}
