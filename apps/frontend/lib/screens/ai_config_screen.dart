// AI Configuration Screen
// T141: AI configuration screen with provider selection, API endpoint, and key input

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/ai_config_provider.dart';

class AIConfigScreen extends ConsumerStatefulWidget {
  const AIConfigScreen({super.key});

  @override
  ConsumerState<AIConfigScreen> createState() => _AIConfigScreenState();
}

class _AIConfigScreenState extends ConsumerState<AIConfigScreen> {
  final _formKey = GlobalKey<FormState>();
  final _apiKeyController = TextEditingController();
  final _endpointController = TextEditingController();
  final _modelNameController = TextEditingController();

  String _selectedProvider = 'openai';
  int _maxTokens = 1000;
  bool _obscureKey = true;

  @override
  void initState() {
    super.initState();
    // Load current config on init
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(aiConfigProvider.notifier).loadConfig();
    });
  }

  @override
  void dispose() {
    _apiKeyController.dispose();
    _endpointController.dispose();
    _modelNameController.dispose();
    super.dispose();
  }

  void _loadCurrentConfig(AIConfigState config) {
    if (config.provider != null) {
      _selectedProvider = config.provider!;
      _endpointController.text = config.apiEndpoint ?? '';
      _modelNameController.text = config.modelName ?? '';
      _maxTokens = config.maxTokens ?? 1000;
    }
  }

  Future<void> _saveConfig() async {
    if (!_formKey.currentState!.validate()) return;

    final success = await ref.read(aiConfigProvider.notifier).setAIConfig(
      provider: _selectedProvider,
      apiEndpoint: _endpointController.text.trim(),
      apiKey: _apiKeyController.text.trim(),
      modelName: _modelNameController.text.trim().isEmpty
          ? null
          : _modelNameController.text.trim(),
      maxTokens: _maxTokens,
    );

    if (!mounted) return;

    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(success ? 'AI configuration saved' : 'Failed to save configuration'),
        backgroundColor: success ? Colors.green : Colors.red,
      ),
    );
  }

  Future<void> _disableAI() async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Disable AI'),
        content: const Text('Are you sure you want to disable AI analysis?'),
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
      await ref.read(aiConfigProvider.notifier).disableAI();
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('AI analysis disabled')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('AI Configuration'),
        actions: [
          Consumer(
            builder: (context, ref, _) {
              final config = ref.watch(aiConfigProvider);
              if (config.isEnabled) {
                return TextButton.icon(
                  onPressed: _disableAI,
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
          final config = ref.watch(aiConfigProvider);

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
                        ref.read(aiConfigProvider.notifier).loadConfig(),
                    child: const Text('Retry'),
                  ),
                ],
              ),
            );
          }

          // Initialize form with current config
          if (config.isEnabled && _apiKeyController.text.isEmpty) {
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
                  _StatusBanner(isEnabled: config.isEnabled),
                  const SizedBox(height: 24),

                  // Provider selection
                  _ProviderDropdown(
                    value: _selectedProvider,
                    onChanged: (value) {
                      if (value == null) return;
                      setState(() {
                        _selectedProvider = value;
                        // Update defaults when provider changes
                        final defaults = defaultModels[value];
                        if (defaults != null) {
                          _modelNameController.text = '';
                        }
                      });
                    },
                  ),
                  const SizedBox(height: 16),

                  // API Endpoint
                  _EndpointField(
                    controller: _endpointController,
                    provider: _selectedProvider,
                  ),
                  const SizedBox(height: 16),

                  // API Key
                  _APIKeyField(
                    controller: _apiKeyController,
                    obscureKey: _obscureKey,
                    onToggle: () => setState(() => _obscureKey = !_obscureKey),
                  ),
                  const SizedBox(height: 16),

                  // Model name
                  _ModelNameField(
                    controller: _modelNameController,
                    provider: _selectedProvider,
                  ),
                  const SizedBox(height: 16),

                  // Max tokens
                  _MaxTokensSlider(
                    value: _maxTokens,
                    onChanged: (value) => setState(() => _maxTokens = value),
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
                ],
              ),
            ),
          );
        },
      ),
    );
  }
}

// =====================================================
// Widgets
// =====================================================

class _StatusBanner extends StatelessWidget {
  final bool isEnabled;

  const _StatusBanner({required this.isEnabled});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: isEnabled ? Colors.green.shade50 : Colors.orange.shade50,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(
          color: isEnabled ? Colors.green.shade200 : Colors.orange.shade200,
        ),
      ),
      child: Row(
        children: [
          Icon(
            isEnabled ? Icons.check_circle : Icons.info_outline,
            color: isEnabled ? Colors.green : Colors.orange,
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Text(
              isEnabled
                  ? 'AI analysis is enabled'
                  : 'AI analysis is disabled - configure below to enable',
              style: TextStyle(
                color: isEnabled ? Colors.green.shade900 : Colors.orange.shade900,
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
      initialValue: value,
      decoration: const InputDecoration(
        labelText: 'AI Provider',
        border: OutlineInputBorder(),
        helperText: 'Select your AI service provider',
      ),
      items: providerNames.entries.map((entry) {
        return DropdownMenuItem(
          value: entry.key,
          child: Row(
            children: [
              if (entry.key == 'ollama') ...[
                const Icon(Icons.computer, size: 20),
                const SizedBox(width: 8),
              ] else if (entry.key == 'claude') ...[
                const Icon(Icons.psychology, size: 20),
                const SizedBox(width: 8),
              ] else ...[
                const Icon(Icons.smart_toy, size: 20),
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

  String _defaultEndpointFor(String provider) {
    switch (provider) {
      case 'openai':
        return 'https://api.openai.com/v1';
      case 'claude':
        return 'https://api.anthropic.com';
      case 'ollama':
        return 'http://localhost:11434';
      default:
        return '';
    }
  }

  @override
  Widget build(BuildContext context) {
    return TextFormField(
      controller: controller,
      decoration: InputDecoration(
        labelText: 'API Endpoint',
        border: const OutlineInputBorder(),
        hintText: _defaultEndpointFor(provider),
        helperText: provider == 'ollama'
            ? 'Local Ollama instance URL'
            : 'API endpoint URL',
      ),
      validator: (value) {
        if (value == null || value.trim().isEmpty) {
          return 'API endpoint is required';
        }
        return null;
      },
    );
  }
}

class _APIKeyField extends StatelessWidget {
  final TextEditingController controller;
  final bool obscureKey;
  final VoidCallback onToggle;

  const _APIKeyField({
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
        labelText: 'API Key',
        border: const OutlineInputBorder(),
        helperText: 'Your API key will be encrypted and stored locally',
        suffixIcon: IconButton(
          icon: Icon(obscureKey ? Icons.visibility : Icons.visibility_off),
          onPressed: onToggle,
        ),
      ),
      validator: (value) {
        if (value == null || value.trim().isEmpty) {
          return 'API key is required';
        }
        return null;
      },
    );
  }
}

class _ModelNameField extends StatelessWidget {
  final TextEditingController controller;
  final String provider;

  const _ModelNameField({
    required this.controller,
    required this.provider,
  });

  @override
  Widget build(BuildContext context) {
    final defaultModel = defaultModels[provider];

    return TextFormField(
      controller: controller,
      decoration: InputDecoration(
        labelText: 'Model Name (Optional)',
        border: const OutlineInputBorder(),
        hintText: defaultModel,
        helperText: 'Leave empty for default model',
      ),
    );
  }
}

class _MaxTokensSlider extends StatelessWidget {
  final int value;
  final ValueChanged<int> onChanged;

  const _MaxTokensSlider({
    required this.value,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            const Text('Max Tokens'),
            Text(value.toString(), style: const TextStyle(fontWeight: FontWeight.bold)),
          ],
        ),
        Slider(
          value: value.toDouble(),
          min: 100,
          max: 4000,
          divisions: 39,
          onChanged: (value) => onChanged(value.toInt()),
        ),
        Text(
          'Maximum tokens for AI responses (100-4000)',
          style: Theme.of(context).textTheme.bodySmall,
        ),
      ],
    );
  }
}
