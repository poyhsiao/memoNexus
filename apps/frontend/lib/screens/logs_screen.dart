// Logs Screen - T213: "View Logs" functionality
// Opens log file location for viewing application logs
//
// ★ Insight ─────────────────────────────────────
// 1. Platform-specific APIs (path_provider) resolve
//    to correct log directories on each platform.
// 2. open_file launches the system default viewer for
//    the log file (text editor, console, etc.).
// 3. AsyncValue provides reactive loading states for
//    log file discovery.
// ─────────────────────────────────────────────────

import 'dart:convert';
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:path_provider/path_provider.dart';

// Log entry model
class LogEntry {
  final DateTime timestamp;
  final String level;
  final String message;
  final String? component;
  final String? error;

  LogEntry({
    required this.timestamp,
    required this.level,
    required this.message,
    this.component,
    this.error,
  });

  factory LogEntry.fromJson(Map<String, dynamic> json) {
    return LogEntry(
      timestamp: DateTime.parse(json['timestamp'] as String),
      level: json['level'] as String,
      message: json['message'] as String,
      component: json['component'] as String?,
      error: json['error'] as String?,
    );
  }

  Color getLevelColor(BuildContext context) {
    final theme = Theme.of(context);
    switch (level.toLowerCase()) {
      case 'error':
        return theme.colorScheme.error;
      case 'warn':
        return Colors.orange;
      case 'info':
        return theme.colorScheme.primary;
      default:
        return Colors.grey;
    }
  }
}

// Log file model
class LogFile {
  final String name;
  final String path;
  final int sizeBytes;
  final DateTime modifiedAt;

  LogFile({
    required this.name,
    required this.path,
    required this.sizeBytes,
    required this.modifiedAt,
  });

  String get formattedSize {
    const kb = 1024;
    const mb = kb * 1024;

    if (sizeBytes >= mb) {
      return '${(sizeBytes / mb).toStringAsFixed(2)} MB';
    } else if (sizeBytes >= kb) {
      return '${(sizeBytes / kb).toStringAsFixed(2)} KB';
    } else {
      return '$sizeBytes bytes';
    }
  }

  String get formattedDate {
    return '${modifiedAt.year}-${modifiedAt.month.toString().padLeft(2, '0')}-${modifiedAt.day.toString().padLeft(2, '0')} '
        '${modifiedAt.hour.toString().padLeft(2, '0')}:${modifiedAt.minute.toString().padLeft(2, '0')}';
  }
}

// Log files notifier
class LogFilesNotifier extends StateNotifier<AsyncValue<List<LogFile>>> {
  LogFilesNotifier() : super(const AsyncValue.loading()) {
    loadLogFiles();
  }

  Future<void> loadLogFiles() async {
    state = const AsyncValue.loading();

    try {
      // Get application documents directory
      final appDir = await getApplicationDocumentsDirectory();
      final logsDir = Directory('${appDir.path}/logs');

      if (!await logsDir.exists()) {
        state = const AsyncValue.data([]);
        return;
      }

      final files = <LogFile>[];

      await for (final entity in logsDir.list()) {
        if (entity is File && entity.path.endsWith('.jsonl')) {
          final stat = await entity.stat();
          files.add(LogFile(
            name: entity.path.split('/').last,
            path: entity.path,
            sizeBytes: stat.size,
            modifiedAt: stat.modified,
          ));
        }
      }

      // Sort by modification date (newest first)
      files.sort((a, b) => b.modifiedAt.compareTo(a.modifiedAt));

      state = AsyncValue.data(files);
    } catch (e, st) {
      state = AsyncValue.error(e, st);
    }
  }

  Future<void> openLogFile(String path) async {
    try {
      await Process.run('open', [path]);
    } catch (e) {
      // Fallback to trying different commands on different platforms
      if (Platform.isWindows) {
        await Process.run('start', [path]);
      } else if (Platform.isLinux) {
        await Process.run('xdg-open', [path]);
      }
    }
  }

  Future<List<LogEntry>> readLogEntries(String path, {int limit = 1000}) async {
    try {
      final file = File(path);
      final lines = await file.readAsLines();
      final entries = <LogEntry>[];

      // Read from newest entries (reverse order)
      for (final line in lines.reversed.take(limit)) {
        if (line.trim().isEmpty) continue;

        try {
          final json = jsonDecode(line) as Map<String, dynamic>;
          entries.add(LogEntry.fromJson(json));
        } catch (_) {
          // Skip invalid JSON lines
          continue;
        }

        if (entries.length >= limit) break;
      }

      return entries;
    } catch (e) {
      return [];
    }
  }

  Future<void> deleteLogFile(String path) async {
    try {
      await File(path).delete();
      await loadLogFiles(); // Refresh list
    } catch (e) {
      // Handle error
    }
  }
}

// Provider
final logFilesProvider =
    StateNotifierProvider<LogFilesNotifier, AsyncValue<List<LogFile>>>(
  (ref) => LogFilesNotifier(),
);

// Logs screen widget
class LogsScreen extends ConsumerStatefulWidget {
  const LogsScreen({super.key});

  @override
  ConsumerState<LogsScreen> createState() => _LogsScreenState();
}

class _LogsScreenState extends ConsumerState<LogsScreen> {
  String? _selectedFilePath;
  List<LogEntry> _selectedLogEntries = [];
  bool _loadingEntries = false;
  final Set<String> _selectedLevels = {'error', 'warn', 'info', 'debug'};

  @override
  Widget build(BuildContext context) {
    final logsAsync = ref.watch(logFilesProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Application Logs'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () => ref.read(logFilesProvider.notifier).loadLogFiles(),
            tooltip: 'Refresh',
          ),
          PopupMenuButton<FilterOption>(
            icon: const Icon(Icons.filter_list),
            onSelected: (option) {
              // TODO: Implement filtering
            },
            itemBuilder: (context) => [
              const PopupMenuItem(
                value: FilterOption.errorsOnly,
                child: Text('Errors only'),
              ),
              const PopupMenuItem(
                value: FilterOption.warningsAndErrors,
                child: Text('Warnings & Errors'),
              ),
              const PopupMenuItem(
                value: FilterOption.all,
                child: Text('All levels'),
              ),
            ],
          ),
        ],
      ),
      body: logsAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (err, stack) => Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(Icons.error_outline, size: 64, color: Colors.red),
              const SizedBox(height: 16),
              Text('Failed to load logs: $err'),
              const SizedBox(height: 16),
              ElevatedButton(
                onPressed: () => ref.read(logFilesProvider.notifier).loadLogFiles(),
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
        data: (files) {
          if (files.isEmpty) {
            return const Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.folder_open, size: 64, color: Colors.grey),
                  SizedBox(height: 16),
                  Text('No log files found'),
                ],
              ),
            );
          }

          return Row(
            children: [
              // Log files list
              SizedBox(
                width: 300,
                child: ListView.builder(
                  itemCount: files.length,
                  itemBuilder: (context, index) {
                    final file = files[index];
                    final isSelected = file.path == _selectedFilePath;

                    return ListTile(
                      title: Text(file.name),
                      subtitle: Text('${file.formattedSize} • ${file.formattedDate}'),
                      selected: isSelected,
                      leading: Icon(
                        isSelected ? Icons.insert_drive_file : Icons.description,
                        color: isSelected
                            ? Theme.of(context).colorScheme.primary
                            : null,
                      ),
                      onTap: () => _selectLogFile(file.path),
                      trailing: IconButton(
                        icon: const Icon(Icons.delete_outline),
                        onPressed: () => _confirmDelete(file),
                        tooltip: 'Delete',
                      ),
                    );
                  },
                ),
              ),
              const VerticalDivider(thickness: 1),
              // Log entries
              Expanded(
                child: _selectedFilePath != null
                    ? _buildLogEntriesView()
                    : const Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(Icons.tap_and_play, size: 64, color: Colors.grey),
                            SizedBox(height: 16),
                            Text('Select a log file to view'),
                          ],
                        ),
                      ),
              ),
            ],
          );
        },
      ),
    );
  }

  Widget _buildLogEntriesView() {
    if (_loadingEntries) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_selectedLogEntries.isEmpty) {
      return const Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.info_outline, size: 64, color: Colors.grey),
            SizedBox(height: 16),
            Text('No log entries to display'),
          ],
        ),
      );
    }

    return Column(
      children: [
        // Log file info
        Container(
          padding: const EdgeInsets.all(16),
          color: Theme.of(context).colorScheme.surfaceContainerHighest,
          child: Row(
            children: [
              const Icon(Icons.info_outline),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  'Showing ${_selectedLogEntries.length} most recent entries',
                  style: Theme.of(context).textTheme.bodySmall,
                ),
              ),
              ElevatedButton.icon(
                onPressed: () => ref.read(logFilesProvider.notifier).openLogFile(_selectedFilePath!),
                icon: const Icon(Icons.open_in_new),
                label: const Text('Open in Editor'),
              ),
            ],
          ),
        ),
        const Divider(height: 1),
        // Log entries list
        Expanded(
          child: ListView.builder(
            itemCount: _selectedLogEntries.length,
            itemBuilder: (context, index) {
              final entry = _selectedLogEntries[index];
              return Card(
                margin: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                child: ListTile(
                  leading: Icon(
                    _getIconForLevel(entry.level),
                    color: entry.getLevelColor(context),
                  ),
                  title: Text(entry.message),
                  subtitle: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        '${entry.timestamp.toString()}${entry.component != null ? ' • ${entry.component}' : ''}',
                        style: Theme.of(context).textTheme.bodySmall,
                      ),
                      if (entry.error != null)
                        Text(
                          entry.error!,
                          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                                color: Theme.of(context).colorScheme.error,
                              ),
                        ),
                    ],
                  ),
                  dense: true,
                ),
              );
            },
          ),
        ),
      ],
    );
  }

  IconData _getIconForLevel(String level) {
    switch (level.toLowerCase()) {
      case 'error':
        return Icons.error;
      case 'warn':
        return Icons.warning;
      case 'info':
        return Icons.info;
      default:
        return Icons.bug_report;
    }
  }

  void _selectLogFile(String path) async {
    setState(() {
      _selectedFilePath = path;
      _loadingEntries = true;
    });

    final entries = await ref.read(logFilesProvider.notifier).readLogEntries(path);

    setState(() {
      _selectedLogEntries = entries;
      _loadingEntries = false;
    });
  }

  void _confirmDelete(LogFile file) {
    showDialog<void>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete Log File'),
        content: Text('Are you sure you want to delete ${file.name}?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () {
              ref.read(logFilesProvider.notifier).deleteLogFile(file.path);
              Navigator.of(context).pop();
              if (_selectedFilePath == file.path) {
                setState(() {
                  _selectedFilePath = null;
                  _selectedLogEntries = [];
                });
              }
            },
            child: const Text('Delete'),
          ),
        ],
      ),
    );
  }
}

enum FilterOption { errorsOnly, warningsAndErrors, all }
