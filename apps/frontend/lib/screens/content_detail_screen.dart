// Content Detail Screen
// Displays full content with metadata and allows tag editing

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/content_item.dart';
import '../providers/content_provider.dart';
import '../widgets/tag_picker.dart';
import '../widgets/summary_view.dart';
import '../widgets/keyword_suggestions.dart';

class ContentDetailScreen extends ConsumerStatefulWidget {
  final String itemId;

  const ContentDetailScreen({
    super.key,
    required this.itemId,
  });

  @override
  ConsumerState<ContentDetailScreen> createState() =>
      _ContentDetailScreenState();
}

class _ContentDetailScreenState extends ConsumerState<ContentDetailScreen> {
  bool _isEditingTags = false;
  List<String> _editedTags = [];

  // Summary and keyword analysis state
  List<String> _extractedKeywords = [];
  bool _isExtractingKeywords = false;

  @override
  Widget build(BuildContext context) {
    final itemAsync = ref.watch(contentItemProvider(widget.itemId));

    return Scaffold(
      appBar: AppBar(
        title: const Text('Content Details'),
        actions: itemAsync.when(
          data: (item) => [
            IconButton(
              onPressed: () => _showDeleteConfirmation(context, item),
              icon: const Icon(Icons.delete),
              tooltip: 'Delete',
            ),
          ],
          loading: () => const [],
          error: (_, __) => const [],
        ),
      ),
      body: itemAsync.when(
        data: (item) => _buildContent(context, item),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (error, stack) => Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(Icons.error_outline, size: 48, color: Colors.red),
              const SizedBox(height: 16),
              Text('Error loading content: $error'),
              const SizedBox(height: 16),
              FilledButton.icon(
                onPressed: () => ref.refresh(contentItemProvider(widget.itemId)),
                icon: const Icon(Icons.refresh),
                label: const Text('Retry'),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildContent(BuildContext context, ContentItem item) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Title
          Text(
            item.title,
            style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
          ),
          const SizedBox(height: 8),

          // Metadata
          _buildMetadataRow(context, item),
          const SizedBox(height: 24),

          // Tags section
          _buildTagsSection(context, item),
          const SizedBox(height: 24),

          // Summary section (T142)
          _buildSummarySection(context, item),
          const SizedBox(height: 24),

          // Keyword suggestions section (T143)
          _buildKeywordSuggestionsSection(context, item),
          const SizedBox(height: 24),

          // Content
          _buildContentSection(context, item),
          const SizedBox(height: 24),

          // Source URL
          if (item.sourceUrl != null && item.sourceUrl!.isNotEmpty) ...[
            _buildSourceUrlSection(context, item.sourceUrl!),
            const SizedBox(height: 24),
          ],

          // Technical metadata
          _buildTechnicalMetadata(context, item),
        ],
      ),
    );
  }

  Widget _buildMetadataRow(BuildContext context, ContentItem item) {
    return Wrap(
      spacing: 12,
      runSpacing: 8,
      children: [
        Chip(
          avatar: const Icon(Icons.public, size: 16),
          label: Text(item.mediaType.name.toUpperCase()),
          visualDensity: VisualDensity.compact,
        ),
        Chip(
          avatar: const Icon(Icons.schedule, size: 16),
          label: Text(_formatDate(item.createdAt)),
          visualDensity: VisualDensity.compact,
        ),
        if (item.contentHash != null)
          const Chip(
            avatar: Icon(Icons.fingerprint, size: 16),
            label: Text(
              'SHA-256',
              style: TextStyle(fontSize: 11),
            ),
            visualDensity: VisualDensity.compact,
          ),
      ],
    );
  }

  Widget _buildTagsSection(BuildContext context, ContentItem item) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Text(
              'Tags',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const Spacer(),
          ],
        ),
        const SizedBox(height: 8),
        if (_isEditingTags)
          TagPickerWidget(
            selectedTags: _editedTags,
            onTagsChanged: (tags) => setState(() => _editedTags = tags),
          )
        else
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: item.tags.isEmpty
                ? [
                    Text(
                      'No tags',
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                            color: Colors.grey[600],
                          ),
                    ),
                  ]
                : item.tags
                    .map((tag) => Chip(
                          label: Text(tag),
                          visualDensity: VisualDensity.compact,
                        ))
                    .toList(),
          ),
        const SizedBox(height: 8),
        Row(
          children: [
            if (_isEditingTags) ...[
              FilledButton.tonal(
                onPressed: () {
                  setState(() {
                    _isEditingTags = false;
                    _editedTags = [];
                  });
                },
                child: const Text('Cancel'),
              ),
              const SizedBox(width: 8),
              FilledButton(
                onPressed: () => _saveTags(item),
                child: const Text('Save Tags'),
              ),
            ] else ...[
              OutlinedButton.icon(
                onPressed: () {
                  setState(() {
                    _isEditingTags = true;
                    _editedTags = List.from(item.tags);
                  });
                },
                icon: const Icon(Icons.edit, size: 16),
                label: const Text('Edit Tags'),
              ),
            ],
          ],
        ),
      ],
    );
  }

  Widget _buildSummarySection(BuildContext context, ContentItem item) {
    // If summary exists, display it
    if (item.summary != null && item.summary!.isNotEmpty) {
      return SummaryView(
        summary: item.summary!,
        method: 'tfidf', // Default method for stored summaries
        aiUsed: false,
      );
    }

    // Show empty state with option to generate
    return SummaryEmptyView(
      onGenerate: () => _generateSummary(context, item),
    );
  }

  Widget _buildKeywordSuggestionsSection(BuildContext context, ContentItem item) {
    if (_extractedKeywords.isNotEmpty) {
      return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          KeywordSuggestions(
            keywords: _extractedKeywords,
            onKeywordSelected: (keyword) => _addKeywordAsTag(context, item, keyword),
            onRefresh: () => _extractKeywords(context, item),
            method: 'tfidf',
            isLoading: _isExtractingKeywords,
          ),
          if (_editedTags.isNotEmpty) ...[
            const SizedBox(height: 16),
            SelectedKeywords(
              keywords: _editedTags,
              onRemove: (keyword) {
                setState(() {
                  _editedTags.remove(keyword);
                });
              },
            ),
          ],
        ],
      );
    }

    return KeywordSuggestions(
      keywords: const [],
      onKeywordSelected: (keyword) => _addKeywordAsTag(context, item, keyword),
      onRefresh: () => _extractKeywords(context, item),
      isLoading: _isExtractingKeywords,
      method: 'tfidf',
    );
  }

  Widget _buildContentSection(BuildContext context, ContentItem item) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Content',
          style: Theme.of(context).textTheme.titleMedium,
        ),
        const SizedBox(height: 8),
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: Theme.of(context).colorScheme.surfaceContainerHighest,
            borderRadius: BorderRadius.circular(12),
          ),
          child: SelectableText(
            item.contentText,
            style: Theme.of(context).textTheme.bodyMedium,
          ),
        ),
      ],
    );
  }

  Widget _buildSourceUrlSection(BuildContext context, String url) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Source',
          style: Theme.of(context).textTheme.titleMedium,
        ),
        const SizedBox(height: 8),
        InkWell(
          onTap: () {
            // TODO: Open URL in browser
          },
          child: Row(
            children: [
              const Icon(Icons.link, size: 20),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  url,
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        color: Theme.of(context).colorScheme.primary,
                        decoration: TextDecoration.underline,
                      ),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildTechnicalMetadata(BuildContext context, ContentItem item) {
    return ExpansionTile(
      title: Text(
        'Technical Metadata',
        style: Theme.of(context).textTheme.titleMedium,
      ),
      children: [
        _buildMetadataTile('ID', item.id),
        _buildMetadataTile('Version', 'v${item.version}'),
        _buildMetadataTile('Created', _formatDateTime(item.createdAt)),
        _buildMetadataTile('Updated', _formatDateTime(item.updatedAt)),
        if (item.contentHash != null)
          _buildMetadataTile('Content Hash', item.contentHash!),
        if (item.summary != null && item.summary!.isNotEmpty)
          _buildMetadataTile('AI Summary', item.summary!),
        _buildMetadataTile('Deleted', item.isDeleted ? 'Yes' : 'No'),
      ],
    );
  }

  Widget _buildMetadataTile(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 120,
            child: Text(
              '$label:',
              style: const TextStyle(fontWeight: FontWeight.w500),
            ),
          ),
          Expanded(
            child: SelectableText(
              value,
              style: const TextStyle(fontFamily: 'monospace'),
            ),
          ),
        ],
      ),
    );
  }

  Future<void> _saveTags(ContentItem item) async {
    try {
      final api = ref.read(apiClientProvider);
      await api.updateContentItem(
        item.id,
        tags: _editedTags,
      );

      // Refresh the item to get updated data
      // ignore: unused_result
      ref.refresh(contentItemProvider(widget.itemId));

      if (mounted) {
        setState(() {
          _isEditingTags = false;
          _editedTags = [];
        });
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Tags updated')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to update tags: $e')),
        );
      }
    }
  }

  // =====================================================
  // AI Analysis Methods (T142-T143)
  // =====================================================

  Future<void> _generateSummary(BuildContext context, ContentItem item) async {
    try {
      final api = ref.read(apiClientProvider);
      final result = await api.generateSummary(item.id);

      if (mounted) {
        // Refresh the item to get the updated summary
        // ignore: unused_result
        ref.refresh(contentItemProvider(widget.itemId));
        if (!context.mounted) return;
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(result['method'] == 'ai'
                ? 'AI summary generated'
                : 'Summary created using ${result['method']}'),
            backgroundColor: Colors.green,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        if (!context.mounted) return;
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to generate summary: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  Future<void> _extractKeywords(BuildContext context, ContentItem item) async {
    setState(() {
      _isExtractingKeywords = true;
    });

    try {
      final api = ref.read(apiClientProvider);
      final keywords = await api.extractKeywords(item.id);

      if (mounted) {
        setState(() {
          _extractedKeywords = keywords;
          _isExtractingKeywords = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _isExtractingKeywords = false;
        });
        if (!context.mounted) return;
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to extract keywords: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  void _addKeywordAsTag(BuildContext context, ContentItem item, String keyword) {
    // Check if keyword already exists in tags
    if (item.tags.contains(keyword)) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('"$keyword" is already a tag'),
          duration: const Duration(seconds: 1),
        ),
      );
      return;
    }

    // Add to edited tags
    setState(() {
      if (!_editedTags.contains(keyword)) {
        _editedTags.add(keyword);
      }
    });

    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text('"$keyword" added to tags'),
        action: SnackBarAction(
          label: 'Save',
          onPressed: () => _saveTags(item),
        ),
      ),
    );
  }

  void _showDeleteConfirmation(BuildContext context, ContentItem item) {
    showDialog(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: const Text('Delete Content'),
        content: Text('Are you sure you want to delete "${item.title}"?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogContext),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () async {
              Navigator.pop(dialogContext);
              try {
                await ref.read(contentListProvider.notifier).deleteItem(item.id);
                if (mounted && context.mounted) {
                  Navigator.pop(context);
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(content: Text('Content deleted')),
                  );
                }
              } catch (e) {
                if (mounted && context.mounted) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    SnackBar(content: Text('Failed to delete: $e')),
                  );
                }
              }
            },
            style: FilledButton.styleFrom(
              backgroundColor: Colors.red,
            ),
            child: const Text('Delete'),
          ),
        ],
      ),
    );
  }

  String _formatDate(DateTime date) {
    return '${date.year}-${date.month.toString().padLeft(2, '0')}-${date.day.toString().padLeft(2, '0')}';
  }

  String _formatDateTime(DateTime date) {
    return '${date.toLocal()}'.split('.')[0];
  }
}
