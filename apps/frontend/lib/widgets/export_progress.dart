// Export Progress Widget - T197: User Story 5
// Displays real-time export progress with stage information
//
// ★ Insight ─────────────────────────────────────
// 1. ConsumerWidget from Riverpod automatically rebuilds
//    when the export state changes, eliminating manual
//    setState calls.
// 2. LinearProgressIndicator provides smooth animation
//    for progress updates without additional animation code.
// ─────────────────────────────────────────────────

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/export_provider.dart';

class ExportProgressWidget extends ConsumerWidget {
  const ExportProgressWidget({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final exportState = ref.watch(exportProvider);

    if (!exportState.isRunning && exportState.status != ExportStatus.completed) {
      return const SizedBox.shrink();
    }

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            Row(
              children: [
                const SizedBox(
                  width: 20,
                  height: 20,
                  child: CircularProgressIndicator(strokeWidth: 2),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        exportState.currentStage,
                        style: Theme.of(context).textTheme.titleSmall,
                      ),
                      if (exportState.currentFile.isNotEmpty)
                        Padding(
                          padding: const EdgeInsets.only(top: 4),
                          child: Text(
                            exportState.currentFile,
                            style: Theme.of(context)
                                .textTheme
                                .bodySmall
                                ?.copyWith(color: Colors.grey[600]),
                          ),
                        ),
                    ],
                  ),
                ),
                Text('${exportState.progressPercent.toStringAsFixed(0)}%'),
              ],
            ),
            const SizedBox(height: 16),
            LinearProgressIndicator(
              value: exportState.progressPercent / 100,
              backgroundColor: Colors.grey[200],
            ),
            if (exportState.totalItems > 0) ...[
              const SizedBox(height: 8),
              Text(
                '${exportState.processedItems} / ${exportState.totalItems} items',
                style: Theme.of(context).textTheme.bodySmall,
              ),
            ],
            if (exportState.status == ExportStatus.completed) ...[
              const SizedBox(height: 16),
              Row(
                children: [
                  const Icon(Icons.check_circle, color: Colors.green, size: 20),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Export completed!',
                          style: Theme.of(context).textTheme.titleSmall,
                        ),
                        if (exportState.filePath != null)
                          Text(
                            exportState.filePath!,
                            style: Theme.of(context)
                                .textTheme
                                .bodySmall
                                ?.copyWith(color: Colors.grey[600]),
                          ),
                      ],
                    ),
                  ),
                ],
              ),
            ],
            if (exportState.isRunning) ...[
              const SizedBox(height: 16),
              TextButton.icon(
                onPressed: () => ref.read(exportProvider.notifier).cancelExport(),
                icon: const Icon(Icons.cancel, size: 18),
                label: const Text('Cancel'),
              ),
            ],
          ],
        ),
      ),
    );
  }
}
