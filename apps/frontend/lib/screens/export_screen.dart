// Export Screen - T195: User Story 5
// Screen for initiating and managing data exports
//
// ★ Insight ─────────────────────────────────────
// 1. Scaffold provides basic app structure with app bar
//    and body, following Material Design guidelines.
// 2. Form validation using GlobalKey ensures data
//    integrity before submission.
// 3. TextField with obscureText provides secure password
//    input for export encryption.
// ─────────────────────────────────────────────────

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/export_provider.dart';
import '../widgets/auto_export_config.dart';
import '../widgets/export_archive_list.dart';
import '../widgets/export_progress.dart';

class ExportScreen extends ConsumerStatefulWidget {
  const ExportScreen({super.key});

  @override
  ConsumerState<ExportScreen> createState() => _ExportScreenState();
}

class _ExportScreenState extends ConsumerState<ExportScreen> {
  final _formKey = GlobalKey<FormState>();
  final _passwordController = TextEditingController();
  var _includeMedia = false;

  @override
  void dispose() {
    _passwordController.dispose();
    super.dispose();
  }

  Future<void> _startExport(WidgetRef ref) async {
    if (!_formKey.currentState!.validate()) return;

    final password = _passwordController.text.trim();
    if (password.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please enter a password')),
      );
      return;
    }

    await ref.read(exportProvider.notifier).startExport(
      password: password,
      includeMedia: _includeMedia,
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Export Data'),
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // Export Progress
          const ExportProgressWidget(),
          const SizedBox(height: 16),

          // Manual Export Section
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Form(
                key: _formKey,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Export Knowledge Base',
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                    const SizedBox(height: 16),
                    Text(
                      'Create an encrypted backup of all your content, tags, and media files.',
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                            color: Colors.grey[700],
                          ),
                    ),
                    const SizedBox(height: 16),
                    TextFormField(
                      controller: _passwordController,
                      decoration: const InputDecoration(
                        labelText: 'Export Password',
                        hintText: 'Enter a strong password',
                        prefixIcon: Icon(Icons.lock),
                        border: OutlineInputBorder(),
                      ),
                      obscureText: true,
                      validator: (value) {
                        if (value == null || value.isEmpty) {
                          return 'Please enter a password';
                        }
                        if (value.length < 8) {
                          return 'Password must be at least 8 characters';
                        }
                        return null;
                      ),
                    ),
                    const SizedBox(height: 16),
                    SwitchListTile(
                      title: const Text('Include media files'),
                      subtitle: const Text(
                        'Include images and videos (increases export size)',
                      ),
                      value: _includeMedia,
                      onChanged: (value) {
                        setState(() {
                          _includeMedia = value;
                        });
                      },
                    ),
                    const SizedBox(height: 16),
                    SizedBox(
                      width: double.infinity,
                      child: FilledButton.icon(
                        onPressed: () => _startExport(ref),
                        icon: const Icon(Icons.file_upload),
                        label: const Text('Start Export'),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
          const SizedBox(height: 16),

          // Auto Export Configuration
          const AutoExportConfigWidget(),
          const SizedBox(height: 16),

          // Export Archives List
          const ExportArchiveListWidget(),
        ],
      ),
    );
  }
}
