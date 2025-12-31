// Export Archive List Widget - T198: User Story 5
// Displays list of export archives with delete functionality
//
// ★ Insight ─────────────────────────────────────
// 1. AsyncValue from Riverpod provides loading/error/data
//    states for asynchronous operations with built-in
//    when/switch expressions for pattern matching.
// 2. ListTile provides standardized list item layout with
//    leading/trailing widgets and tap handling.
// ─────────────────────────────────────────────────

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/export_provider.dart';

class ExportArchiveListWidget extends ConsumerWidget {
  const ExportArchiveListWidget({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final archivesAsync = ref.watch(exportArchivesProvider);

    return archivesAsync.when(
      data: (archives) {
        if (archives.isEmpty) {
          return Card(
            child: Padding(
              padding: const EdgeInsets.all(32),
              child: Center(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(
                      Icons.archive_outlined,
                      size: 48,
                      color: Colors.grey[400],
                    ),
                    const SizedBox(height: 16),
                    Text(
                      'No export archives yet',
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                    const SizedBox(height: 8),
                    Text(
                      'Create your first export to back up your knowledge base',
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                            color: Colors.grey[600],
                          ),
                    ),
                  ],
                ),
              ),
            ),
          );
        }

        return Card(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Padding(
                padding: const EdgeInsets.all(16),
                child: Text(
                  'Export Archives (${archives.length})',
                  style: Theme.of(context).textTheme.titleMedium,
                ),
              ),
              const Divider(height: 1),
              ListView.builder(
                shrinkWrap: true,
                physics: const NeverScrollableScrollPhysics(),
                itemCount: archives.length,
                itemBuilder: (context, index) {
                  final archive = archives[index];
                  return _ArchiveListTile(
                    archive: archive,
                    onDelete: () => _confirmDelete(
                      context,
                      ref,
                      archive.id,
                    ),
                  );
                },
              ),
            ],
          ),
        );
      },
      loading: () => const Card(
        child: Padding(
          padding: EdgeInsets.all(32),
          child: Center(child: CircularProgressIndicator()),
        ),
      ),
      error: (error, _) => Card(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Text('Failed to load archives: $error'),
        ),
      ),
    );
  }

  Future<void> _confirmDelete(
    BuildContext context,
    WidgetRef ref,
    String archiveId,
  ) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete Archive'),
        content: const Text(
          'Are you sure you want to delete this export archive? '
          'This action cannot be undone.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Delete', style: TextStyle(color: Colors.red)),
          ),
        ],
      ),
    );

    if (confirmed == true && context.mounted) {
      await ref.read(exportArchivesProvider.notifier).deleteArchive(archiveId);
    }
  }
}

class _ArchiveListTile extends StatelessWidget {
  final ExportArchive archive;
  final VoidCallback onDelete;

  const _ArchiveListTile({
    required this.archive,
    required this.onDelete,
  });

  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: CircleAvatar(
        backgroundColor: archive.isEncrypted
            ? Colors.green[100]
            : Colors.grey[200],
        child: Icon(
          archive.isEncrypted ? Icons.lock : Icons.lock_open,
          size: 20,
          color: archive.isEncrypted ? Colors.green[700] : Colors.grey[600],
        ),
      ),
      title: Text(
        archive.filePath.split('/').last,
        style: const TextStyle(fontWeight: FontWeight.w500),
      ),
      subtitle: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('${archive.itemCount} items • ${archive.formattedSize}'),
          Text(
            archive.formattedDate,
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: Colors.grey[600],
                ),
          ),
        ],
      ),
      trailing: IconButton(
        icon: const Icon(Icons.delete_outline, color: Colors.red),
        tooltip: 'Delete archive',
        onPressed: onDelete,
      ),
    );
  }
}
