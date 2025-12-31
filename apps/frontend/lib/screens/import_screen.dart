// Import Screen - T196: User Story 5
// Screen for importing encrypted export archives
//
// ★ Insight ─────────────────────────────────────
// 1. File picker allows users to select archive files
//    from device storage with type filtering.
// 2. Import progress stages provide visual feedback
//    for long-running operations (decrypting, extracting).
// 3. Success/error states guide users through completion
//    or troubleshooting.
// ─────────────────────────────────────────────────

import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/export_provider.dart';

class ImportScreen extends ConsumerStatefulWidget {
  const ImportScreen({super.key});

  @override
  ConsumerState<ImportScreen> createState() => _ImportScreenState();
}

class _ImportScreenState extends ConsumerState<ImportScreen> {
  final _formKey = GlobalKey<FormState>();
  final _passwordController = TextEditingController();
  String? _selectedFilePath;

  @override
  void dispose() {
    _passwordController.dispose();
    super.dispose();
  }

  Future<void> _pickFile() async {
    final result = await FilePicker.platform.pickFiles(
      type: FileType.custom,
      allowedExtensions: ['tar.gz', 'gz'],
      dialogTitle: 'Select Export Archive',
    );

    if (result != null && result.files.single.path != null) {
      setState(() {
        _selectedFilePath = result.files.single.path;
      });
    }
  }

  Future<void> _startImport(WidgetRef ref) async {
    if (!_formKey.currentState!.validate()) return;
    if (_selectedFilePath == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please select an archive file')),
      );
      return;
    }

    final password = _passwordController.text.trim();
    if (password.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please enter the archive password')),
      );
      return;
    }

    await ref.read(importProvider.notifier).startImport(
      filePath: _selectedFilePath!,
      password: password,
    );
  }

  @override
  Widget build(BuildContext context) {
    final importState = ref.watch(importProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Import Data'),
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // Import Progress
          if (importState.isRunning ||
              importState.status == ImportStatus.completed ||
              importState.isFailed)
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    if (importState.isRunning) ...[
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
                                  importState.currentStage,
                                  style: Theme.of(context)
                                      .textTheme
                                      .titleSmall,
                                ),
                                const SizedBox(height: 8),
                                LinearProgressIndicator(
                                  value: importState.progressPercent / 100,
                                ),
                              ],
                            ),
                          ),
                          Text('${importState.progressPercent.toStringAsFixed(0)}%'),
                        ],
                      ),
                      const SizedBox(height: 16),
                      TextButton.icon(
                        onPressed: () =>
                            ref.read(importProvider.notifier).cancelImport(),
                        icon: const Icon(Icons.cancel, size: 18),
                        label: const Text('Cancel'),
                      ),
                    ],
                    if (importState.isSuccess) ...[
                      Row(
                        children: [
                          const Icon(Icons.check_circle,
                              color: Colors.green, size: 24),
                          const SizedBox(width: 16),
                          Expanded(
                            child: Column(
                              crossAxisAlignment:
                                  CrossAxisAlignment.start,
                              children: [
                                Text(
                                  'Import completed successfully!',
                                  style: Theme.of(context)
                                      .textTheme
                                      .titleMedium,
                                ),
                                Text(
                                  '${importState.importedItems} items imported',
                                  style: Theme.of(context)
                                      .textTheme
                                      .bodySmall
                                      ?.copyWith(
                                        color: Colors.grey[600],
                                      ),
                                ),
                                if (importState.skippedItems > 0)
                                  Text(
                                    '${importState.skippedItems} items skipped (duplicates)',
                                    style: Theme.of(context)
                                        .textTheme
                                        .bodySmall
                                        ?.copyWith(
                                          color: Colors.orange[700],
                                        ),
                                  ),
                              ],
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 16),
                      FilledButton.icon(
                        onPressed: () =>
                            ref.read(importProvider.notifier).reset(),
                        icon: const Icon(Icons.check),
                        label: const Text('Done'),
                      ),
                    ],
                    if (importState.isFailed) ...[
                      Row(
                        children: [
                          const Icon(Icons.error,
                              color: Colors.red, size: 24),
                          const SizedBox(width: 16),
                          Expanded(
                            child: Column(
                              crossAxisAlignment:
                                  CrossAxisAlignment.start,
                              children: [
                                Text(
                                  'Import failed',
                                  style: Theme.of(context)
                                      .textTheme
                                      .titleMedium,
                                ),
                                if (importState.errorMessage != null)
                                  Text(
                                    importState.errorMessage!,
                                    style: Theme.of(context)
                                        .textTheme
                                        .bodySmall
                                        ?.copyWith(
                                          color: Colors.red[700],
                                        ),
                                  ),
                              ],
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 16),
                      OutlinedButton.icon(
                        onPressed: () =>
                            ref.read(importProvider.notifier).reset(),
                        icon: const Icon(Icons.refresh),
                        label: const Text('Try Again'),
                      ),
                    ],
                  ],
                ),
              ),
            ),
          if (importState.status == ImportStatus.idle) ...[
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Form(
                  key: _formKey,
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Import Export Archive',
                        style:
                            Theme.of(context).textTheme.titleMedium,
                      ),
                      const SizedBox(height: 16),
                      Text(
                        'Restore your knowledge base from a previously '
                        'created encrypted export archive.',
                        style: Theme.of(context)
                            .textTheme
                            .bodyMedium
                            ?.copyWith(color: Colors.grey[700]),
                      ),
                      const SizedBox(height: 24),
                      InkWell(
                        onTap: _pickFile,
                        child: InputDecorator(
                          decoration: const InputDecoration(
                            labelText: 'Archive File',
                            hintText: 'Select archive file',
                            prefixIcon: Icon(Icons.folder_open),
                            border: OutlineInputBorder(),
                          ),
                          child: Text(
                            _selectedFilePath?.split('/').last ??
                                'No file selected',
                          ),
                        ),
                      ),
                      const SizedBox(height: 16),
                      TextFormField(
                        controller: _passwordController,
                        decoration: const InputDecoration(
                          labelText: 'Archive Password',
                          hintText: 'Enter the archive password',
                          prefixIcon: Icon(Icons.lock),
                          border: OutlineInputBorder(),
                        ),
                        obscureText: true,
                        validator: (value) {
                          if (value == null || value.isEmpty) {
                            return 'Please enter the password';
                          }
                          return null;
                        },
                      ),
                      const SizedBox(height: 24),
                      SizedBox(
                        width: double.infinity,
                        child: FilledButton.icon(
                          onPressed: () => _startImport(ref),
                          icon: const Icon(Icons.file_download),
                          label: const Text('Import Archive'),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ],
        ],
      ),
    );
  }
}
