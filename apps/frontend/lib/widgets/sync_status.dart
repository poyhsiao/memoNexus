// Sync Status Widget
// T170: Sync status widget showing status indicator, last synced, and pending count

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/sync_provider.dart';

class SyncStatusWidget extends ConsumerWidget {
  const SyncStatusWidget({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final status = ref.watch(syncStatusProvider);

    if (status.isLoading) {
      return const _LoadingIndicator();
    }

    if (status.hasFailed) {
      return _FailedStatus(
        error: status.error ?? 'Sync failed',
        onRetry: () => ref.read(syncStatusProvider.notifier).loadStatus(),
      );
    }

    return _SyncStatusCard(
      isSyncing: status.isSyncing,
      isIdle: status.isIdle,
      lastSync: status.lastSync,
      pendingChanges: status.pendingChanges,
      onRefresh: () => ref.read(syncStatusProvider.notifier).triggerSync(),
    );
  }
}

// =====================================================
// Internal Widgets
// =====================================================

class _LoadingIndicator extends StatelessWidget {
  const _LoadingIndicator();

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.grey.shade100,
        borderRadius: BorderRadius.circular(8),
      ),
      child: const Row(
        children: [
          SizedBox(
            width: 16,
            height: 16,
            child: CircularProgressIndicator(strokeWidth: 2),
          ),
          SizedBox(width: 12),
          Text('Loading sync status...'),
        ],
      ),
    );
  }
}

class _FailedStatus extends StatelessWidget {
  final String error;
  final VoidCallback onRetry;

  const _FailedStatus({
    required this.error,
    required this.onRetry,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.red.shade50,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.red.shade200),
      ),
      child: Row(
        children: [
          Icon(Icons.error_outline, color: Colors.red.shade700),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  'Sync failed',
                  style: TextStyle(
                    color: Colors.red.shade900,
                    fontWeight: FontWeight.w500,
                  ),
                ),
                Text(
                  error,
                  style: TextStyle(color: Colors.red.shade700, fontSize: 12),
                ),
              ],
            ),
          ),
          TextButton.icon(
            onPressed: onRetry,
            icon: const Icon(Icons.refresh, size: 18),
            label: const Text('Retry'),
          ),
        ],
      ),
    );
  }
}

class _SyncStatusCard extends StatelessWidget {
  final bool isSyncing;
  final bool isIdle;
  final int? lastSync;
  final int pendingChanges;
  final VoidCallback onRefresh;

  const _SyncStatusCard({
    required this.isSyncing,
    required this.isIdle,
    this.lastSync,
    required this.pendingChanges,
    required this.onRefresh,
  });

  String _formatLastSync() {
    if (lastSync == null) return 'Never';

    final now = DateTime.now();
    final syncTime = DateTime.fromMillisecondsSinceEpoch(lastSync!);
    final diff = now.difference(syncTime);

    if (diff.inSeconds < 60) {
      return 'Just now';
    } else if (diff.inMinutes < 60) {
      return '${diff.inMinutes}m ago';
    } else if (diff.inHours < 24) {
      return '${diff.inHours}h ago';
    } else if (diff.inDays < 7) {
      return '${diff.inDays}d ago';
    } else {
      // Format: Jan 15, 14:30
      final months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
      return '${months[syncTime.month - 1]} ${syncTime.day}, ${syncTime.hour.toString().padLeft(2, '0')}:${syncTime.minute.toString().padLeft(2, '0')}';
    }
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: isSyncing
            ? Colors.blue.shade50
            : isIdle
                ? Colors.green.shade50
                : Colors.orange.shade50,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(
          color: isSyncing
              ? Colors.blue.shade200
              : isIdle
                  ? Colors.green.shade200
                  : Colors.orange.shade200,
        ),
      ),
      child: Row(
        children: [
          _StatusIcon(isSyncing: isSyncing, isIdle: isIdle),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  _getStatusText(),
                  style: TextStyle(
                    color: _getStatusColor(),
                    fontWeight: FontWeight.w500,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  'Last sync: ${_formatLastSync()}',
                  style: TextStyle(
                    color: _getStatusColor().withOpacity(0.8),
                    fontSize: 12,
                  ),
                ),
                if (pendingChanges > 0)
                  Text(
                    '$pendingChanges pending change(s)',
                    style: TextStyle(
                      color: Colors.orange.shade700,
                      fontSize: 11,
                    ),
                  ),
              ],
            ),
          ),
          if (isIdle)
            IconButton(
              onPressed: onRefresh,
              icon: const Icon(Icons.cloud_sync),
              tooltip: 'Sync now',
            ),
        ],
      ),
    );
  }

  String _getStatusText() {
    if (isSyncing) return 'Syncing...';
    if (isIdle) return 'Sync enabled';
    return 'Sync';
  }

  Color _getStatusColor() {
    if (isSyncing) return Colors.blue.shade900;
    if (isIdle) return Colors.green.shade900;
    return Colors.orange.shade900;
  }
}

class _StatusIcon extends StatelessWidget {
  final bool isSyncing;
  final bool isIdle;

  const _StatusIcon({
    required this.isSyncing,
    required this.isIdle,
  });

  @override
  Widget build(BuildContext context) {
    if (isSyncing) {
      return const SizedBox(
        width: 20,
        height: 20,
        child: CircularProgressIndicator(strokeWidth: 2),
      );
    }

    if (isIdle) {
      return Icon(Icons.cloud_done, color: Colors.green.shade700);
    }

    return Icon(Icons.sync, color: Colors.orange.shade700);
  }
}
