// File Upload Widget
// Drag-and-drop file upload widget with file preview

import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/capture_provider.dart';

class FileUploadWidget extends ConsumerWidget {
  final String? initialFilePath;
  final String? initialFileName;
  final ValueChanged<String> onFileSelected;

  const FileUploadWidget({
    super.key,
    this.initialFilePath,
    this.initialFileName,
    required this.onFileSelected,
  });

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final filePath = initialFilePath;
    final fileName = initialFileName;

    if (filePath != null) {
      return _buildFilePreview(context, ref, filePath, fileName);
    }

    return _buildDropZone(context, ref);
  }

  Widget _buildDropZone(BuildContext context, WidgetRef ref) {
    return Container(
      height: 200,
      decoration: BoxDecoration(
        border: Border.all(
          color: Theme.of(context).colorScheme.outlineVariant,
          width: 2,
        ),
        borderRadius: BorderRadius.circular(12),
      ),
      child: InkWell(
        onTap: () => _pickFile(context, ref),
        borderRadius: BorderRadius.circular(12),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.cloud_upload_outlined,
              size: 48,
              color: Theme.of(context).colorScheme.primary,
            ),
            const SizedBox(height: 16),
            Text(
              'Tap to select a file',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 8),
            Text(
              'or drag and drop',
              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                    color: Colors.grey[600],
                  ),
            ),
            const SizedBox(height: 16),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              alignment: WrapAlignment.center,
              children: [
                _buildFileTypeChip(
                  context,
                  Icons.image,
                  'Images',
                  ['.jpg', '.jpeg', '.png', '.gif', '.webp'],
                ),
                _buildFileTypeChip(
                  context,
                  Icons.videocam,
                  'Videos',
                  ['.mp4', '.webm', '.mov', '.avi'],
                ),
                _buildFileTypeChip(
                  context,
                  Icons.picture_as_pdf,
                  'PDF',
                  ['.pdf'],
                ),
                _buildFileTypeChip(
                  context,
                  Icons.description,
                  'Markdown',
                  ['.md', '.markdown'],
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildFileTypeChip(
    BuildContext context,
    IconData icon,
    String label,
    List<String> extensions,
  ) {
    return Chip(
      avatar: Icon(icon, size: 16),
      label: Text(
        label,
        style: const TextStyle(fontSize: 11),
      ),
      padding: EdgeInsets.zero,
      materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
      visualDensity: VisualDensity.compact,
    );
  }

  Widget _buildFilePreview(
    BuildContext context,
    WidgetRef ref,
    String filePath,
    String? fileName,
  ) {
    final extension = _getFileExtension(filePath);
    final icon = _getFileIcon(extension);
    final color = _getFileColor(extension);

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        border: Border.all(
          color: color.withOpacity(0.3),
          width: 2,
        ),
        borderRadius: BorderRadius.circular(12),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                width: 48,
                height: 48,
                decoration: BoxDecoration(
                  color: color.withOpacity(0.2),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Icon(
                  icon,
                  color: color,
                  size: 28,
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      fileName ?? filePath.split('/').last,
                      style: Theme.of(context).textTheme.titleSmall,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    Text(
                      extension.toUpperCase(),
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                            color: color,
                            fontWeight: FontWeight.w500,
                          ),
                    ),
                  ],
                ),
              ),
              IconButton(
                onPressed: () {
                  ref.read(captureProvider.notifier).clearFile();
                },
                icon: const Icon(Icons.close),
                tooltip: 'Remove file',
              ),
            ],
          ),
          const SizedBox(height: 12),
          Row(
            children: [
              FilledButton.tonalIcon(
                onPressed: () => _pickFile(context, ref),
                icon: const Icon(Icons.refresh),
                label: const Text('Change file'),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Future<void> _pickFile(BuildContext context, WidgetRef ref) async {
    try {
      final result = await FilePicker.platform.pickFiles(
        type: FileType.custom,
        allowedExtensions: [
          'jpg',
          'jpeg',
          'png',
          'gif',
          'webp',
          'mp4',
          'webm',
          'mov',
          'avi',
          'pdf',
          'md',
          'markdown',
        ],
        allowMultiple: false,
      );

      if (result != null && result.files.isNotEmpty) {
        final file = result.files.first;
        final notifier = ref.read(captureProvider.notifier);

        // Handle both XFile (mobile) and File (desktop)
        if (file.path != null) {
          notifier.setFile(file.path!, file.name);
          onFileSelected(file.path!);
        }
      }
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to pick file: $e')),
        );
      }
    }
  }

  String _getFileExtension(String filePath) {
    final dotIndex = filePath.lastIndexOf('.');
    if (dotIndex < 0) return '';
    return filePath.substring(dotIndex).toLowerCase();
  }

  IconData _getFileIcon(String extension) {
    switch (extension) {
      case '.jpg':
      case '.jpeg':
      case '.png':
      case '.gif':
      case '.webp':
        return Icons.image;
      case '.mp4':
      case '.webm':
      case '.mov':
      case '.avi':
        return Icons.videocam;
      case '.pdf':
        return Icons.picture_as_pdf;
      case '.md':
      case '.markdown':
        return Icons.description;
      default:
        return Icons.insert_drive_file;
    }
  }

  Color _getFileColor(String extension) {
    switch (extension) {
      case '.jpg':
      case '.jpeg':
      case '.png':
      case '.gif':
      case '.webp':
        return Colors.purple;
      case '.mp4':
      case '.webm':
      case '.mov':
      case '.avi':
        return Colors.red;
      case '.pdf':
        return Colors.orange;
      case '.md':
      case '.markdown':
        return Colors.blue;
      default:
        return Colors.grey;
    }
  }
}
