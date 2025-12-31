// Sync Configuration Screen
// T169: Sync configuration screen with S3 endpoint, bucket, and credentials input

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/sync_provider.dart';

class SyncConfigScreen extends ConsumerStatefulWidget {
  const SyncConfigScreen({super.key});

  @override
  ConsumerState<SyncConfigScreen> createState() => _SyncConfigScreenState();
}

class _SyncConfigScreenState extends ConsumerState<SyncConfigScreen> {
  final _formKey = GlobalKey<FormState>();
  final _endpointController = TextEditingController();
  final _bucketNameController = TextEditingController();
  final _accessKeyController = TextEditingController();
  final _secretKeyController = TextEditingController();
  final _regionController = TextEditingController();

  String _selectedProvider = 'aws';
  bool _obscureSecretKey = true;

  @override
  void initState() {
    super.initState();
    // Load current config on init
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(syncConfigProvider.notifier).loadConfig();
    });
  }

  @override
  void dispose() {
    _endpointController.dispose();
    _bucketNameController.dispose();
    _accessKeyController.dispose();
    _secretKeyController.dispose();
    _regionController.dispose();
    super.dispose();
  }

  void _loadCurrentConfig(SyncConfigState config) {
    if (config.isConfigured) {
      _endpointController.text = config.endpoint ?? '';
      _bucketNameController.text = config.bucketName ?? '';
      _regionController.text = config.region ?? '';

      // Detect provider from endpoint
      _detectProviderFromEndpoint(config.endpoint ?? '');
    }
  }

  void _detectProviderFromEndpoint(String endpoint) {
    if (endpoint.contains('s3.amazonaws.com') || endpoint.contains('s3.')) {
      _selectedProvider = 'aws';
    } else if (endpoint.contains('r2.cloudflarestorage.com')) {
      _selectedProvider = 'cloudflare';
    } else if (endpoint.contains('minio') || endpoint.contains(':9000')) {
      _selectedProvider = 'minio';
    } else {
      _selectedProvider = 'custom';
    }
  }

  Future<void> _saveConfig() async {
    if (!_formKey.currentState!.validate()) return;

    final success = await ref.read(syncConfigProvider.notifier).configureSync(
      endpoint: _endpointController.text.trim(),
      bucketName: _bucketNameController.text.trim(),
      accessKey: _accessKeyController.text.trim(),
      secretKey: _secretKeyController.text.trim(),
      region: _regionController.text.trim().isEmpty ? null : _regionController.text.trim(),
    );

    if (!mounted) return;

    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(success ? 'Sync configured successfully' : 'Failed to configure sync'),
        backgroundColor: success ? Colors.green : Colors.red,
      ),
    );
  }

  Future<void> _disableSync() async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Disable Sync'),
        content: const Text('Are you sure you want to disable cloud sync? Your data will remain on this device.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Disable', style: TextStyle(color: Colors.red)),
          ),
        ],
      ),
    );

    if (confirm == true) {
      await ref.read(syncConfigProvider.notifier).disableSync();
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Cloud sync disabled')),
        );
        // Clear form
        _endpointController.clear();
        _bucketNameController.clear();
        _accessKeyController.clear();
        _secretKeyController.clear();
        _regionController.clear();
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Sync Configuration'),
        actions: [
          Consumer(
            builder: (context, ref, _) {
              final config = ref.watch(syncConfigProvider);
              if (config.isConfigured) {
                return TextButton.icon(
                  onPressed: _disableSync,
                  icon: const Icon(Icons.delete_outline),
                  label: const Text('Disable'),
                );
              }
              return const SizedBox.shrink();
            },
          ),
        ],
      ),
      body: Consumer(
        builder: (context, ref, _) {
          final config = ref.watch(syncConfigProvider);

          if (config.isLoading) {
            return const Center(child: CircularProgressIndicator());
          }

          if (config.error != null) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(Icons.error_outline, size: 48, color: Colors.red),
                  const SizedBox(height: 16),
                  Text('Error: ${config.error}'),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: () =>
                        ref.read(syncConfigProvider.notifier).loadConfig(),
                    child: const Text('Retry'),
                  ),
                ],
              ),
            );
          }

          // Initialize form with current config
          if (config.isConfigured && _endpointController.text.isEmpty) {
            _loadCurrentConfig(config);
          }

          return SingleChildScrollView(
            padding: const EdgeInsets.all(16),
            child: Form(
              key: _formKey,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Status banner
                  _StatusBanner(isConfigured: config.isConfigured),
                  const SizedBox(height: 24),

                  // Provider selection
                  _ProviderDropdown(
                    value: _selectedProvider,
                    onChanged: (value) {
                      if (value == null) return;
                      setState(() {
                        _selectedProvider = value;
                        _updateDefaultsForProvider(value);
                      });
                    },
                  ),
                  const SizedBox(height: 16),

                  // Endpoint
                  _EndpointField(
                    controller: _endpointController,
                    provider: _selectedProvider,
                  ),
                  const SizedBox(height: 16),

                  // Bucket name
                  _BucketNameField(controller: _bucketNameController),
                  const SizedBox(height: 16),

                  // Region (for AWS)
                  if (_selectedProvider == 'aws' || _selectedProvider == 'custom') ...[
                    _RegionField(controller: _regionController),
                    const SizedBox(height: 16),
                  ],

                  // Access key
                  _AccessKeyField(controller: _accessKeyController),
                  const SizedBox(height: 16),

                  // Secret key
                  _SecretKeyField(
                    controller: _secretKeyController,
                    obscureKey: _obscureSecretKey,
                    onToggle: () => setState(() => _obscureSecretKey = !_obscureSecretKey),
                  ),
                  const SizedBox(height: 32),

                  // Save button
                  SizedBox(
                    width: double.infinity,
                    child: ElevatedButton(
                      onPressed: _saveConfig,
                      style: ElevatedButton.styleFrom(
                        padding: const EdgeInsets.symmetric(vertical: 16),
                      ),
                      child: const Text('Save Configuration'),
                    ),
                  ),

                  const SizedBox(height: 16),

                  // Info card
                  _InfoCard(provider: _selectedProvider),
                ],
              ),
            ),
          );
        },
      ),
    );
  }

  void _updateDefaultsForProvider(String provider) {
    switch (provider) {
      case 'aws':
        _endpointController.text = 's3.amazonaws.com';
        _regionController.text = 'us-east-1';
        break;
      case 'cloudflare':
        _endpointController.clear();
        _bucketNameController.clear();
        _regionController.clear();
        break;
      case 'minio':
        _endpointController.text = 'localhost:9000';
        _regionController.text = 'us-east-1';
        break;
      case 'custom':
        _endpointController.clear();
        _regionController.clear();
        break;
    }
  }
}

// =====================================================
// Widgets
// =====================================================

class _StatusBanner extends StatelessWidget {
  final bool isConfigured;

  const _StatusBanner({required this.isConfigured});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: isConfigured ? Colors.green.shade50 : Colors.blue.shade50,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(
          color: isConfigured ? Colors.green.shade200 : Colors.blue.shade200,
        ),
      ),
      child: Row(
        children: [
          Icon(
            isConfigured ? Icons.cloud_done : Icons.cloud_upload,
            color: isConfigured ? Colors.green : Colors.blue,
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Text(
              isConfigured
                  ? 'Cloud sync is enabled'
                  : 'Configure cloud sync to backup and sync your knowledge base',
              style: TextStyle(
                color: isConfigured ? Colors.green.shade900 : Colors.blue.shade900,
                fontWeight: FontWeight.w500,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _ProviderDropdown extends StatelessWidget {
  final String? value;
  final ValueChanged<String?>? onChanged;

  const _ProviderDropdown({
    this.value,
    this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    return DropdownButtonFormField<String>(
      value: value,
      decoration: const InputDecoration(
        labelText: 'Storage Provider',
        border: OutlineInputBorder(),
        helperText: 'Select your S3-compatible storage provider',
      ),
      items: storageProviderNames.entries.map((entry) {
        return DropdownMenuItem(
          value: entry.key,
          child: Row(
            children: [
              if (entry.key == 'aws') ...[
                const Icon(Icons.cloud, size: 20),
                const SizedBox(width: 8),
              ] else if (entry.key == 'cloudflare') ...[
                const Icon(Icons.cloud_queue, size: 20),
                const SizedBox(width: 8),
              ] else if (entry.key == 'minio') ...[
                const Icon(Icons.storage, size: 20),
                const SizedBox(width: 8),
              ] else ...[
                const Icon(Icons.settings, size: 20),
                const SizedBox(width: 8),
              ],
              Text(entry.value),
            ],
          ),
        );
      }).toList(),
      onChanged: onChanged,
    );
  }
}

class _EndpointField extends StatelessWidget {
  final TextEditingController controller;
  final String provider;

  const _EndpointField({
    required this.controller,
    required this.provider,
  });

  String _defaultHintFor(String provider) {
    switch (provider) {
      case 'aws':
        return 's3.amazonaws.com or s3.<region>.amazonaws.com';
      case 'cloudflare':
        return '<account-id>.r2.cloudflarestorage.com';
      case 'minio':
        return 'localhost:9000 or minio.example.com';
      default:
        return 'your-s3-endpoint.com';
    }
  }

  @override
  Widget build(BuildContext context) {
    return TextFormField(
      controller: controller,
      decoration: InputDecoration(
        labelText: 'S3 Endpoint',
        border: const OutlineInputBorder(),
        hintText: _defaultHintFor(provider),
        helperText: provider == 'cloudflare'
            ? 'Your Cloudflare Account ID + .r2.cloudflarestorage.com'
            : 'S3-compatible API endpoint',
      ),
      validator: (value) {
        if (value == null || value.trim().isEmpty) {
          return 'Endpoint is required';
        }
        return null;
      },
    );
  }
}

class _BucketNameField extends StatelessWidget {
  final TextEditingController controller;

  const _BucketNameField({required this.controller});

  @override
  Widget build(BuildContext context) {
    return TextFormField(
      controller: controller,
      decoration: const InputDecoration(
        labelText: 'Bucket Name',
        border: OutlineInputBorder(),
        hintText: 'my-memonexus-backup',
        helperText: 'Name of your S3 bucket',
      ),
      validator: (value) {
        if (value == null || value.trim().isEmpty) {
          return 'Bucket name is required';
        }
        return null;
      },
    );
  }
}

class _RegionField extends StatelessWidget {
  final TextEditingController controller;

  const _RegionField({required this.controller});

  @override
  Widget build(BuildContext context) {
    return TextFormField(
      controller: controller,
      decoration: const InputDecoration(
        labelText: 'Region (Optional)',
        border: OutlineInputBorder(),
        hintText: 'us-east-1',
        helperText: 'AWS region for your bucket (required for AWS S3)',
      ),
    );
  }
}

class _AccessKeyField extends StatelessWidget {
  final TextEditingController controller;

  const _AccessKeyField({required this.controller});

  @override
  Widget build(BuildContext context) {
    return TextFormField(
      controller: controller,
      decoration: const InputDecoration(
        labelText: 'Access Key ID',
        border: OutlineInputBorder(),
        helperText: 'Your S3 access key ID',
      ),
      validator: (value) {
        if (value == null || value.trim().isEmpty) {
          return 'Access key is required';
        }
        return null;
      },
    );
  }
}

class _SecretKeyField extends StatelessWidget {
  final TextEditingController controller;
  final bool obscureKey;
  final VoidCallback onToggle;

  const _SecretKeyField({
    required this.controller,
    required this.obscureKey,
    required this.onToggle,
  });

  @override
  Widget build(BuildContext context) {
    return TextFormField(
      controller: controller,
      obscureText: obscureKey,
      decoration: InputDecoration(
        labelText: 'Secret Access Key',
        border: const OutlineInputBorder(),
        helperText: 'Your secret key will be encrypted and stored locally',
        suffixIcon: IconButton(
          icon: Icon(obscureKey ? Icons.visibility : Icons.visibility_off),
          onPressed: onToggle,
        ),
      ),
      validator: (value) {
        if (value == null || value.trim().isEmpty) {
          return 'Secret key is required';
        }
        return null;
      },
    );
  }
}

class _InfoCard extends StatelessWidget {
  final String provider;

  const _InfoCard({required this.provider});

  @override
  Widget build(BuildContext context) {
    String infoText;
    IconData icon;

    switch (provider) {
      case 'aws':
        infoText = 'AWS S3 provides reliable cloud storage. You\'ll need an AWS account with S3 access and appropriate IAM credentials.';
        icon = Icons.info_outline;
        break;
      case 'cloudflare':
        infoText = 'Cloudflare R2 offers zero-egress fee storage. You\'ll need a Cloudflare account with R2 enabled and an API token created.';
        icon = Icons.info_outline;
        break;
      case 'minio':
        infoText = 'MinIO is self-hosted S3-compatible storage. Great for local networks and complete data control. You\'ll need a running MinIO server.';
        icon = Icons.info_outline;
        break;
      default:
        infoText = 'Configure any S3-compatible storage service. Ensure your provider supports the standard S3 API.';
        icon = Icons.info_outline;
    }

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.blue.shade50,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.blue.shade200),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(icon, color: Colors.blue, size: 20),
          const SizedBox(width: 12),
          Expanded(
            child: Text(
              infoText,
              style: TextStyle(color: Colors.blue.shade900, fontSize: 12),
            ),
          ),
        ],
      ),
    );
  }
}
