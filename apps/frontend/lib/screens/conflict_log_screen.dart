// Conflict Log Screen
// T172: Conflict log viewer showing detected conflicts with resolutions

import 'package:flutter/material.dart';

class ConflictLog {
  final String id;
  final String itemId;
  final DateTime localTimestamp;
  final DateTime remoteTimestamp;
  final String resolution; // last_write_wins, manual, etc.
  final DateTime resolvedAt;

  ConflictLog({
    required this.id,
    required this.itemId,
    required this.localTimestamp,
    required this.remoteTimestamp,
    required this.resolution,
    required this.resolvedAt,
  });

  factory ConflictLog.fromJson(Map<String, dynamic> json) {
    return ConflictLog(
      id: json['id'] as String,
      itemId: json['item_id'] as String,
      localTimestamp: DateTime.fromMillisecondsSinceEpoch(json['local_timestamp'] as int),
      remoteTimestamp: DateTime.fromMillisecondsSinceEpoch(json['remote_timestamp'] as int),
      resolution: json['resolution'] as String? ?? 'last_write_wins',
      resolvedAt: DateTime.fromMillisecondsSinceEpoch(json['resolved_at'] as int),
    );
  }
}

class ConflictLogScreen extends StatefulWidget {
  const ConflictLogScreen({super.key});

  @override
  State<ConflictLogScreen> createState() => _ConflictLogScreenState();
}

class _ConflictLogScreenState extends State<ConflictLogScreen> {
  bool _isLoading = true;
  List<ConflictLog> _conflicts = [];
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadConflicts();
  }

  Future<void> _loadConflicts() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    // TODO: Implement API call to fetch conflicts
    // For now, simulate with empty list
    try {
      await Future.delayed(const Duration(milliseconds: 500));
      setState(() {
        _conflicts = [];
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _error = e.toString();
        _isLoading = false;
      });
    }
  }

  String _formatDateTime(DateTime dt) {
    return '${dt.month}/${dt.day}/${dt.year} ${dt.hour.toString().padLeft(2, '0')}:${dt.minute.toString().padLeft(2, '0')}';
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Sync Conflicts'),
        actions: [
          IconButton(
            onPressed: _loadConflicts,
            icon: const Icon(Icons.refresh),
            tooltip: 'Refresh',
          ),
        ],
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_error != null) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.error_outline, size: 48, color: Colors.red),
            const SizedBox(height: 16),
            Text('Error: $_error'),
            const SizedBox(height: 16),
            ElevatedButton(
              onPressed: _loadConflicts,
              child: const Text('Retry'),
            ),
          ],
        ),
      );
    }

    if (_conflicts.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.check_circle_outline, size: 64, color: Colors.green.shade200),
            const SizedBox(height: 16),
            Text(
              'No conflicts detected',
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: 8),
            Text(
              'Your sync is running smoothly',
              style: TextStyle(color: Colors.grey.shade600),
            ),
          ],
        ),
      );
    }

    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: _conflicts.length,
      itemBuilder: (context, index) {
        final conflict = _conflicts[index];
        return _ConflictCard(
          conflict: conflict,
          formatDateTime: _formatDateTime,
        );
      },
    );
  }
}

class _ConflictCard extends StatelessWidget {
  final ConflictLog conflict;
  final String Function(DateTime) formatDateTime;

  const _ConflictCard({
    required this.conflict,
    required this.formatDateTime,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: ExpansionTile(
        leading: _ResolutionIcon(resolution: conflict.resolution),
        title: Text(
          'Item: ${conflict.itemId}',
          style: const TextStyle(fontWeight: FontWeight.w500),
        ),
        subtitle: Text(
          'Resolved: ${formatDateTime(conflict.resolvedAt)}',
          style: TextStyle(color: Colors.grey.shade600, fontSize: 12),
        ),
        children: [
          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _TimestampRow(
                  label: 'Local version',
                  timestamp: conflict.localTimestamp,
                  formatDateTime: formatDateTime,
                  icon: Icons.computer,
                  iconColor: Colors.blue,
                ),
                const SizedBox(height: 8),
                _TimestampRow(
                  label: 'Remote version',
                  timestamp: conflict.remoteTimestamp,
                  formatDateTime: formatDateTime,
                  icon: Icons.cloud,
                  iconColor: Colors.orange,
                ),
                const SizedBox(height: 12),
                _ResolutionBadge(resolution: conflict.resolution),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _ResolutionIcon extends StatelessWidget {
  final String resolution;

  const _ResolutionIcon({required this.resolution});

  @override
  Widget build(BuildContext context) {
    IconData icon;
    Color color;

    switch (resolution) {
      case 'last_write_wins':
        icon = Icons.history;
        color = Colors.blue;
        break;
      case 'manual':
        icon = Icons.edit;
        color = Colors.purple;
        break;
      default:
        icon = Icons.check_circle;
        color = Colors.green;
    }

    return Icon(icon, color: color);
  }
}

class _TimestampRow extends StatelessWidget {
  final String label;
  final DateTime timestamp;
  final String Function(DateTime) formatDateTime;
  final IconData icon;
  final Color iconColor;

  const _TimestampRow({
    required this.label,
    required this.timestamp,
    required this.formatDateTime,
    required this.icon,
    required this.iconColor,
  });

  @override
  Widget build(BuildContext context) {
    final isNewer = DateTime.now().difference(timestamp).inHours < 1;

    return Row(
      children: [
        Icon(icon, size: 16, color: iconColor),
        const SizedBox(width: 8),
        Text(
          '$label: ',
          style: const TextStyle(fontWeight: FontWeight.w500),
        ),
        Text(
          formatDateTime(timestamp),
          style: TextStyle(
            color: isNewer ? Colors.green.shade700 : Colors.grey.shade700,
          ),
        ),
        if (isNewer) ...[
          const SizedBox(width: 4),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
            decoration: BoxDecoration(
              color: Colors.green.shade100,
              borderRadius: BorderRadius.circular(4),
            ),
            child: Text(
              'Recent',
              style: TextStyle(
                fontSize: 10,
                color: Colors.green.shade900,
                fontWeight: FontWeight.w500,
              ),
            ),
          ),
        ],
      ],
    );
  }
}

class _ResolutionBadge extends StatelessWidget {
  final String resolution;

  const _ResolutionBadge({required this.resolution});

  String _getResolutionLabel() {
    switch (resolution) {
      case 'last_write_wins':
        return 'Auto-resolved: Latest version kept';
      case 'manual':
        return 'Manually resolved';
      default:
        return 'Resolved';
    }
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: resolution == 'manual'
            ? Colors.purple.shade50
            : Colors.blue.shade50,
        borderRadius: BorderRadius.circular(6),
        border: Border.all(
          color: resolution == 'manual'
              ? Colors.purple.shade200
              : Colors.blue.shade200,
        ),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(
            resolution == 'manual' ? Icons.check : Icons.info_outline,
            size: 16,
            color: resolution == 'manual'
                ? Colors.purple.shade700
                : Colors.blue.shade700,
          ),
          const SizedBox(width: 6),
          Text(
            _getResolutionLabel(),
            style: TextStyle(
              color: resolution == 'manual'
                  ? Colors.purple.shade900
                  : Colors.blue.shade900,
              ),
          ),
        ],
      ),
    );
  }
}
