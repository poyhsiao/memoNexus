// Capture Screen
// Content capture screen with URL input and file upload support

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/content_item.dart';
import '../providers/capture_provider.dart';
import '../widgets/file_upload.dart';
import '../widgets/tag_picker.dart';

class CaptureScreen extends ConsumerStatefulWidget {
  final ContentItem? initialItem;

  const CaptureScreen({
    super.key,
    this.initialItem,
  });

  @override
  ConsumerState<CaptureScreen> createState() => _CaptureScreenState();
}

class _CaptureScreenState extends ConsumerState<CaptureScreen> {
  final TextEditingController _urlController = TextEditingController();
  final TextEditingController _titleController = TextEditingController();

  @override
  void dispose() {
    _urlController.dispose();
    _titleController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final captureState = ref.watch(captureProvider);

    // Check if capture was successful
    if (captureState.createdItem != null) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        Navigator.pop(context, captureState.createdItem);
      });
    }

    return Scaffold(
      appBar: AppBar(
        title: const Text('Add Content'),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Content type selector
            _buildContentTypeSelector(context, captureState),
            const SizedBox(height: 24),

            // URL input (only for URL type)
            if (captureState.type == CaptureType.url) ...[
              _buildUrlInput(context, captureState),
              const SizedBox(height: 24),
            ],

            // File upload (only for file type)
            if (captureState.type == CaptureType.file) ...[
              FileUploadWidget(
                initialFilePath: captureState.filePath,
                initialFileName: captureState.fileName,
                onFileSelected: (path) {},
              ),
              const SizedBox(height: 24),
            ],

            // Optional title
            _buildTitleInput(context),
            const SizedBox(height: 24),

            // Tags
            _buildTagsSection(context, captureState),
            const SizedBox(height: 32),

            // Error message
            if (captureState.error != null) ...[
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.red.shade50,
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: Colors.red.shade200),
                ),
                child: Row(
                  children: [
                    const Icon(Icons.error, color: Colors.red, size: 20),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Text(
                        captureState.error!,
                        style: const TextStyle(color: Colors.red),
                      ),
                    ),
                    IconButton(
                      onPressed: () {
                        ref.read(captureProvider.notifier).clearError();
                      },
                      icon: const Icon(Icons.close, size: 20),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 16),
            ],

            // Submit button
            SizedBox(
              width: double.infinity,
              child: FilledButton.icon(
                onPressed: captureState.isSubmitting ||
                        !captureState.isValid
                    ? null
                    : _handleSubmit,
                icon: captureState.isSubmitting
                    ? const SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          color: Colors.white,
                        ),
                      )
                    : const Icon(Icons.add),
                label: Text(
                  captureState.isSubmitting
                      ? 'Adding...'
                      : 'Add Content',
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildContentTypeSelector(
    BuildContext context,
    CaptureState captureState,
  ) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Content Type',
          style: Theme.of(context).textTheme.titleMedium,
        ),
        const SizedBox(height: 8),
        SegmentedButton<CaptureType>(
          segments: const [
            ButtonSegment(
              value: CaptureType.url,
              label: Text('URL'),
              icon: Icon(Icons.link),
            ),
            ButtonSegment(
              value: CaptureType.file,
              label: Text('File'),
              icon: Icon(Icons.upload_file),
            ),
          ],
          selected: {captureState.type},
          onSelectionChanged: (Set<CaptureType> selection) {
            ref.read(captureProvider.notifier).setType(selection.first);
          },
        ),
      ],
    );
  }

  Widget _buildUrlInput(BuildContext context, CaptureState captureState) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'URL',
          style: Theme.of(context).textTheme.titleMedium,
        ),
        const SizedBox(height: 8),
        TextField(
          controller: _urlController,
          decoration: InputDecoration(
            hintText: 'https://example.com/article',
            border: const OutlineInputBorder(),
            prefixIcon: const Icon(Icons.link),
            errorText: captureState.url.isNotEmpty &&
                    !captureState.isValid &&
                    captureState.type == CaptureType.url
                ? 'Please enter a valid URL'
                : null,
          ),
          keyboardType: TextInputType.url,
          onChanged: (value) {
            ref.read(captureProvider.notifier).setUrl(value);
          },
        ),
        const SizedBox(height: 8),
        Text(
          'Enter a URL to fetch and parse content from the web',
          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: Colors.grey[600],
              ),
        ),
      ],
    );
  }

  Widget _buildTitleInput(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Title (Optional)',
          style: Theme.of(context).textTheme.titleMedium,
        ),
        const SizedBox(height: 8),
        TextField(
          controller: _titleController,
          decoration: const InputDecoration(
            hintText: 'Custom title',
            border: OutlineInputBorder(),
            prefixIcon: Icon(Icons.title),
          ),
          onChanged: (value) {
            ref.read(captureProvider.notifier).setCustomTitle(value);
          },
        ),
        const SizedBox(height: 8),
        Text(
          'Leave empty to use auto-detected title',
          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: Colors.grey[600],
              ),
        ),
      ],
    );
  }

  Widget _buildTagsSection(
    BuildContext context,
    CaptureState captureState,
  ) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Tags (Optional)',
          style: Theme.of(context).textTheme.titleMedium,
        ),
        const SizedBox(height: 8),
        TagPickerWidget(
          selectedTags: captureState.selectedTags,
          onTagsChanged: (tags) {
            ref.read(captureProvider.notifier).setTags(tags);
          },
        ),
      ],
    );
  }

  void _handleSubmit() async {
    final notifier = ref.read(captureProvider.notifier);
    await notifier.submit();

    // Success is handled by checking createdItem in build()
    if (mounted && notifier.state.error != null) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(notifier.state.error!),
          action: SnackBarAction(
            label: 'Dismiss',
            onPressed: () {
              notifier.clearError();
            },
          ),
        ),
      );
    }
  }
}
