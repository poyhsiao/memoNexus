// Tag Picker Widget
// Multi-select tag picker with tag creation support

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/capture_provider.dart';

class TagPickerWidget extends ConsumerStatefulWidget {
  final List<String> selectedTags;
  final ValueChanged<List<String>> onTagsChanged;

  const TagPickerWidget({
    super.key,
    required this.selectedTags,
    required this.onTagsChanged,
  });

  @override
  ConsumerState<TagPickerWidget> createState() => _TagPickerWidgetState();
}

class _TagPickerWidgetState extends ConsumerState<TagPickerWidget> {
  final TextEditingController _newTagController = TextEditingController();
  bool _showCreateField = false;

  @override
  void dispose() {
    _newTagController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final availableTags = ref.watch(availableTagsProvider);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Selected tags display
        if (widget.selectedTags.isNotEmpty) ...[
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: widget.selectedTags.map((tag) {
              return Chip(
                label: Text(tag),
                onDeleted: () {
                  final updated = List<String>.from(widget.selectedTags)
                    ..remove(tag);
                  widget.onTagsChanged(updated);
                },
                deleteIcon: const Icon(Icons.close, size: 18),
              );
            }).toList(),
          ),
          const SizedBox(height: 12),
        ],
        // Available tags
        Text(
          'Select Tags',
          style: Theme.of(context).textTheme.titleSmall,
        ),
        const SizedBox(height: 8),
        if (availableTags.isEmpty) ...[
          const Text('No tags created yet'),
          const SizedBox(height: 8),
        ] else ...[
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: availableTags.map((tag) {
              final isSelected = widget.selectedTags.contains(tag);
              return FilterChip(
                label: Text(tag),
                selected: isSelected,
                onSelected: (selected) {
                  final updated = List<String>.from(widget.selectedTags);
                  if (selected) {
                    updated.add(tag);
                  } else {
                    updated.remove(tag);
                  }
                  widget.onTagsChanged(updated);
                },
                selectedColor: Theme.of(context)
                    .colorScheme
                    .primary
                    .withValues(alpha: 0.2),
                checkmarkColor: Theme.of(context).colorScheme.primary,
              );
            }).toList(),
          ),
          const SizedBox(height: 12),
        ],
        // Create new tag
        if (!_showCreateField)
          TextButton.icon(
            onPressed: () => setState(() => _showCreateField = true),
            icon: const Icon(Icons.add),
            label: const Text('Create new tag'),
          )
        else
          Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _newTagController,
                  decoration: const InputDecoration(
                    hintText: 'Tag name',
                    border: OutlineInputBorder(),
                    isDense: true,
                  ),
                  onSubmitted: (value) => _createNewTag(value),
                ),
              ),
              const SizedBox(width: 8),
              IconButton(
                onPressed: () => _createNewTag(_newTagController.text),
                icon: const Icon(Icons.check),
              ),
              IconButton(
                onPressed: () => setState(() => _showCreateField = false),
                icon: const Icon(Icons.close),
              ),
            ],
          ),
      ],
    );
  }

  void _createNewTag(String tagName) {
    if (tagName.trim().isEmpty) return;

    final trimmed = tagName.trim();
    if (!widget.selectedTags.contains(trimmed)) {
      final updated = List<String>.from(widget.selectedTags)..add(trimmed);
      widget.onTagsChanged(updated);
    }

    _newTagController.clear();
    setState(() => _showCreateField = false);
  }
}

// =====================================================
// Simplified Tag Picker for Dialog
// =====================================================

class TagPickerDialog extends ConsumerStatefulWidget {
  final List<String> initialTags;

  const TagPickerDialog({
    super.key,
    this.initialTags = const [],
  });

  @override
  ConsumerState<TagPickerDialog> createState() => _TagPickerDialogState();
}

class _TagPickerDialogState extends ConsumerState<TagPickerDialog> {
  late List<String> _selectedTags;

  @override
  void initState() {
    super.initState();
    _selectedTags = List.from(widget.initialTags);
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('Select Tags'),
      content: SizedBox(
        width: double.maxFinite,
        child: TagPickerWidget(
          selectedTags: _selectedTags,
          onTagsChanged: (tags) => setState(() => _selectedTags = tags),
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('Cancel'),
        ),
        FilledButton(
          onPressed: () => Navigator.pop(context, _selectedTags),
          child: const Text('Done'),
        ),
      ],
    );
  }
}
