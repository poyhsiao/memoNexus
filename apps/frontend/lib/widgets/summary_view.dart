// Summary View Widget
// T142: Displays AI/TF-IDF generated summary with metadata

import 'package:flutter/material.dart';

/// Summary view widget that displays AI or TF-IDF generated summaries.
/// Shows the summary text along with metadata about the generation method.
class SummaryView extends StatelessWidget {
  /// The summary text to display
  final String summary;

  /// The method used to generate the summary (ai, tfidf, textrank)
  final String method;

  /// The detected language of the content
  final String? language;

  /// Confidence score for the summary (0.0 - 1.0)
  final double? confidence;

  /// Whether AI was used for generation
  final bool aiUsed;

  const SummaryView({
    super.key,
    required this.summary,
    this.method = 'tfidf',
    this.language,
    this.confidence,
    this.aiUsed = false,
  });

  @override
  Widget build(BuildContext context) {
    if (summary.isEmpty) {
      return const SizedBox.shrink();
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Header with title and metadata
        _buildHeader(context),
        const SizedBox(height: 12),

        // Summary content
        _buildSummaryContent(context),
        const SizedBox(height: 12),

        // Metadata footer
        _buildMetadata(context),
      ],
    );
  }

  Widget _buildHeader(BuildContext context) {
    return Row(
      children: [
        Icon(
          aiUsed ? Icons.psychology : Icons.summarize,
          size: 20,
          color: aiUsed
              ? Colors.deepPurple
              : Theme.of(context).colorScheme.primary,
        ),
        const SizedBox(width: 8),
        Text(
          'Summary',
          style: Theme.of(context).textTheme.titleMedium?.copyWith(
                fontWeight: FontWeight.w600,
              ),
        ),
        const Spacer(),
        // Method badge
        _MethodBadge(
          method: method,
          aiUsed: aiUsed,
        ),
      ],
    );
  }

  Widget _buildSummaryContent(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: aiUsed
            ? Colors.deepPurple.shade50
            : Theme.of(context).colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: aiUsed
              ? Colors.deepPurple.shade200
              : Theme.of(context).colorScheme.outline.withValues(alpha: 0.3),
        ),
      ),
      child: SelectableText(
        summary,
        style: Theme.of(context).textTheme.bodyMedium?.copyWith(
              height: 1.5,
              fontStyle: aiUsed ? FontStyle.normal : FontStyle.italic,
            ),
      ),
    );
  }

  Widget _buildMetadata(BuildContext context) {
    return Wrap(
      spacing: 8,
      runSpacing: 8,
      children: [
        if (language != null)
          _MetadataChip(
            icon: Icons.language,
            label: _formatLanguage(language!),
          ),
        if (confidence != null && confidence! > 0)
          _MetadataChip(
            icon: Icons.signal_cellular_alt,
            label: '${(confidence! * 100).toInt()}% confidence',
          ),
        if (aiUsed)
          _MetadataChip(
            icon: Icons.smart_toy,
            label: 'AI Generated',
            color: Colors.deepPurple,
          ),
      ],
    );
  }

  String _formatLanguage(String lang) {
    switch (lang.toLowerCase()) {
      case 'en':
        return 'English';
      case 'cjk':
        return '中文/日本語/한국어';
      default:
        return lang.toUpperCase();
    }
  }
}

// =====================================================
// Sub-widgets
// =====================================================

class _MethodBadge extends StatelessWidget {
  final String method;
  final bool aiUsed;

  const _MethodBadge({
    required this.method,
    required this.aiUsed,
  });

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
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(
            _getMethodIcon(),
            size: 12,
            color: color,
          ),
          const SizedBox(width: 4),
          Text(
            label,
            style: TextStyle(
              fontSize: 11,
              fontWeight: FontWeight.w600,
              color: color,
            ),
          ),
        ],
      ),
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

class _MetadataChip extends StatelessWidget {
  final IconData icon;
  final String label;
  final Color? color;

  const _MetadataChip({
    required this.icon,
    required this.label,
    this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Chip(
      avatar: Icon(
        icon,
        size: 14,
        color: color ?? Theme.of(context).colorScheme.onSurfaceVariant,
      ),
      label: Text(
        label,
        style: TextStyle(
          fontSize: 11,
          color: color ?? Theme.of(context).colorScheme.onSurfaceVariant,
        ),
      ),
      visualDensity: VisualDensity.compact,
      backgroundColor: color?.withValues(alpha: 0.1) ??
          Theme.of(context).colorScheme.surfaceContainerHighest,
      side: BorderSide(
        color: color?.withValues(alpha: 0.3) ??
            Theme.of(context).colorScheme.outline.withValues(alpha: 0.3),
      ),
    );
  }
}

// =====================================================
// Loading and Empty States
// =====================================================

/// Loading state for summary generation
class SummaryLoadingView extends StatelessWidget {
  final String? message;

  const SummaryLoadingView({
    super.key,
    this.message,
  });

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
              message ?? 'Generating summary...',
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

/// Empty state when no summary is available
class SummaryEmptyView extends StatelessWidget {
  final VoidCallback? onGenerate;

  const SummaryEmptyView({
    super.key,
    this.onGenerate,
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
                Icons.summarize_outlined,
                size: 20,
                color: Theme.of(context).colorScheme.onSurfaceVariant,
              ),
              const SizedBox(width: 8),
              Text(
                'No Summary',
                style: Theme.of(context).textTheme.titleMedium?.copyWith(
                      color: Theme.of(context).colorScheme.onSurfaceVariant,
                    ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(
            'Generate an AI or TF-IDF summary to quickly understand the content.',
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: Theme.of(context).colorScheme.onSurfaceVariant,
                ),
          ),
          if (onGenerate != null) ...[
            const SizedBox(height: 12),
            FilledButton.tonalIcon(
              onPressed: onGenerate,
              icon: const Icon(Icons.auto_awesome, size: 16),
              label: const Text('Generate Summary'),
            ),
          ],
        ],
      ),
    );
  }
}
