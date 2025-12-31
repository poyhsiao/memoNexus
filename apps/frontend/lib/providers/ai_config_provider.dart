// AI Config Provider for Riverpod state management
// Manages AI service configuration (OpenAI, Claude, Ollama) with encrypted storage
// T144: AIConfig Riverpod provider

import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../services/api_client.dart';
import 'content_provider.dart' show apiClientProvider;

// =====================================================
// AI Config State
// =====================================================

class AIConfigState {
  final bool isEnabled;
  final String? provider; // openai, claude, ollama
  final String? apiEndpoint;
  final String? modelName;
  final int? maxTokens;
  final bool isLoading;
  final String? error;

  const AIConfigState({
    this.isEnabled = false,
    this.provider,
    this.apiEndpoint,
    this.modelName,
    this.maxTokens,
    this.isLoading = false,
    this.error,
  });

  AIConfigState copyWith({
    bool? isEnabled,
    String? provider,
    String? apiEndpoint,
    String? modelName,
    int? maxTokens,
    bool? isLoading,
    String? error,
  }) {
    return AIConfigState(
      isEnabled: isEnabled ?? this.isEnabled,
      provider: provider ?? this.provider,
      apiEndpoint: apiEndpoint ?? this.apiEndpoint,
      modelName: modelName ?? this.modelName,
      maxTokens: maxTokens ?? this.maxTokens,
      isLoading: isLoading ?? this.isLoading,
      error: error,
    );
  }

  bool get hasConfig => provider != null && apiEndpoint != null;
}

// =====================================================
// AI Config Notifier
// =====================================================

class AIConfigNotifier extends StateNotifier<AIConfigState> {
  final MemoNexusAPIClient _api;

  AIConfigNotifier(this._api) : super(const AIConfigState());

  /// Load current AI configuration
  Future<void> loadConfig() async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final data = await _api.getAIConfig();

      state = AIConfigState(
        isEnabled: data['enabled'] as bool? ?? false,
        provider: data['provider'] as String?,
        apiEndpoint: data['api_endpoint'] as String?,
        modelName: data['model_name'] as String?,
        maxTokens: data['max_tokens'] as int?,
        isLoading: false,
      );
    } catch (e) {
      state = state.copyWith(
        isEnabled: false,
        isLoading: false,
        error: e.toString(),
      );
    }
  }

  /// Configure AI service with credentials
  Future<bool> setAIConfig({
    required String provider,
    required String apiEndpoint,
    required String apiKey,
    String? modelName,
    int? maxTokens,
  }) async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      await _api.setAIConfig(
        provider: provider,
        apiEndpoint: apiEndpoint,
        apiKey: apiKey,
        modelName: modelName,
        maxTokens: maxTokens,
      );

      // Reload config to get the saved state
      await loadConfig();
      return true;
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: e.toString(),
      );
      return false;
    }
  }

  /// Disable AI service
  Future<void> disableAI() async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      await _api.disableAI();
      state = const AIConfigState();
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: e.toString(),
      );
    }
  }

  /// Clear error state
  void clearError() {
    state = state.copyWith(error: null);
  }
}

// =====================================================
// Providers
// =====================================================

/// AI config provider - main AI configuration state and operations
final aiConfigProvider =
    StateNotifierProvider<AIConfigNotifier, AIConfigState>((ref) {
  final api = ref.watch(apiClientProvider);
  final notifier = AIConfigNotifier(api);

  // Load config on initialization
  notifier.loadConfig();

  return notifier;
});

/// Provider name for API endpoint display
final providerNames = <String, String>{
  'openai': 'OpenAI',
  'claude': 'Anthropic Claude',
  'ollama': 'Ollama (Local)',
};

/// Default model names for each provider
final defaultModels = <String, String>{
  'openai': 'gpt-4',
  'claude': 'claude-3-opus-20240229',
  'ollama': 'llama2',
};
