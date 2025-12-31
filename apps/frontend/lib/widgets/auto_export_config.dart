// Auto Export Config Widget - T199: User Story 5
// Configures automatic export scheduling
//
// ★ Insight ─────────────────────────────────────
// 1. SwitchListTile provides Material Design toggle with
//    label and subtitle in a single widget.
// 2. DropdownButton for enum selection provides type-safe
//    option selection with built-in validation.
// ─────────────────────────────────────────────────

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/export_provider.dart';

class AutoExportConfigWidget extends ConsumerWidget {
  const AutoExportConfigWidget({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final config = ref.watch(autoExportConfigProvider);

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Automatic Export',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 16),
            SwitchListTile(
              title: const Text('Enable automatic export'),
              subtitle: Text(
                config.enabled
                    ? 'Automatic exports are enabled'
                    : 'Automatic exports are disabled',
              ),
              value: config.enabled,
              onChanged: (value) {
                ref.read(autoExportConfigProvider.notifier).setEnabled(value);
              },
            ),
            if (config.enabled) ...[
              const Divider(),
              ListTile(
                title: const Text('Export interval'),
                subtitle: Text(_intervalLabel(config.interval)),
                trailing: DropdownButton<AutoExportInterval>(
                  value: config.interval,
                  items: AutoExportInterval.values.map((interval) {
                    return DropdownMenuItem(
                      value: interval,
                      child: Text(_intervalLabel(interval)),
                    );
                  }).toList(),
                  onChanged: (value) {
                    if (value != null) {
                      ref
                          .read(autoExportConfigProvider.notifier)
                          .setInterval(value);
                    }
                  },
                ),
              ),
              ListTile(
                title: const Text('Retention count'),
                subtitle: Text('Keep last ${config.retentionCount} exports'),
                trailing: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    IconButton(
                      icon: const Icon(Icons.remove),
                      onPressed: config.retentionCount > 1
                          ? () => ref
                              .read(autoExportConfigProvider.notifier)
                              .setRetentionCount(config.retentionCount - 1)
                          : null,
                    ),
                    Text('${config.retentionCount}'),
                    IconButton(
                      icon: const Icon(Icons.add),
                      onPressed: config.retentionCount < 100
                          ? () => ref
                              .read(autoExportConfigProvider.notifier)
                              .setRetentionCount(config.retentionCount + 1)
                          : null,
                    ),
                  ],
                ),
              ),
              SwitchListTile(
                title: const Text('Include media files'),
                subtitle: const Text(
                  'Include images and videos in exports (increases size)',
                ),
                value: config.includeMedia,
                onChanged: (value) {
                  ref.read(autoExportConfigProvider.notifier).setIncludeMedia(value);
                },
              ),
              if (config.lastExportAt != null)
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                  child: Text(
                    'Last export: ${_formatDate(config.lastExportAt!)}',
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: Colors.grey[600],
                        ),
                  ),
                ),
              if (config.nextExportAt != null)
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                  child: Text(
                    'Next export: ${_formatDate(config.nextExportAt!)}',
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: Colors.grey[600],
                        ),
                  ),
                ),
            ],
          ],
        ),
      ),
    );
  }

  String _intervalLabel(AutoExportInterval interval) {
    switch (interval) {
      case AutoExportInterval.manual:
        return 'Manual only';
      case AutoExportInterval.daily:
        return 'Daily';
      case AutoExportInterval.weekly:
        return 'Weekly';
      case AutoExportInterval.monthly:
        return 'Monthly';
    }
  }

  String _formatDate(DateTime date) {
    return '${date.day}/${date.month}/${date.year} ${date.hour}:${date.minute.toString().padLeft(2, '0')}';
  }
}
