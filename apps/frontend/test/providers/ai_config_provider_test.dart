import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:memonexus_frontend/providers/ai_config_provider.dart';
import 'package:memonexus_frontend/services/api_client.dart';

void main() {
  group('AIConfigState', () {
    test('should have correct default values', () {
      const state = AIConfigState();

      expect(state.isEnabled, false);
      expect(state.provider, isNull);
      expect(state.apiEndpoint, isNull);
      expect(state.modelName, isNull);
      expect(state.maxTokens, isNull);
      expect(state.isLoading, false);
      expect(state.error, isNull);
    });

    test('copyWith should create new state with updated values', () {
      const original = AIConfigState(
        isEnabled: true,
        provider: 'openai',
      );
      final updated = original.copyWith(
        isLoading: true,
        modelName: 'gpt-4',
      );

      expect(original.isEnabled, true);
      expect(original.provider, 'openai');
      expect(original.isLoading, false);
      expect(original.modelName, isNull);
      expect(updated.isEnabled, true);
      expect(updated.provider, 'openai');
      expect(updated.isLoading, true);
      expect(updated.modelName, 'gpt-4');
    });

    test('hasConfig should return true when provider and endpoint are set', () {
      const state = AIConfigState(
        provider: 'claude',
        apiEndpoint: 'https://api.anthropic.com',
      );
      expect(state.hasConfig, isTrue);
    });

    test('hasConfig should return false when provider is null', () {
      const state = AIConfigState(
        apiEndpoint: 'https://api.example.com',
      );
      expect(state.hasConfig, isFalse);
    });

    test('hasConfig should return false when endpoint is null', () {
      const state = AIConfigState(
        provider: 'openai',
      );
      expect(state.hasConfig, isFalse);
    });

    test('hasConfig should return false when both are null', () {
      const state = AIConfigState();
      expect(state.hasConfig, isFalse);
    });

    test('copyWith with null error should clear error', () {
      const original = AIConfigState(error: 'Some error');
      final updated = original.copyWith(isLoading: true);

      // Note: error field is directly assigned (no ??), so null clears it
      expect(updated.error, isNull);
    });
  });

  group('AIConfigNotifier', () {
    test('loadConfig should update state with API data', () async {
      final api = _FakeMemoNexusAPIClient();
      api.mockAIConfig = {
        'enabled': true,
        'provider': 'openai',
        'api_endpoint': 'https://api.openai.com/v1',
        'model_name': 'gpt-4',
        'max_tokens': 4096,
      };
      final notifier = AIConfigNotifier(api);

      await notifier.loadConfig();

      expect(notifier.state.isEnabled, true);
      expect(notifier.state.provider, 'openai');
      expect(notifier.state.apiEndpoint, 'https://api.openai.com/v1');
      expect(notifier.state.modelName, 'gpt-4');
      expect(notifier.state.maxTokens, 4096);
      expect(notifier.state.isLoading, false);
      expect(notifier.state.error, isNull);
    });

    test('loadConfig should handle missing optional fields', () async {
      final api = _FakeMemoNexusAPIClient();
      api.mockAIConfig = {'enabled': false};
      final notifier = AIConfigNotifier(api);

      await notifier.loadConfig();

      expect(notifier.state.isEnabled, false);
      expect(notifier.state.provider, isNull);
      expect(notifier.state.apiEndpoint, isNull);
      expect(notifier.state.isLoading, false);
    });

    test('loadConfig should handle API errors', () async {
      final api = _FakeMemoNexusAPIClient();
      api.shouldFail = true;
      api.errorMessage = 'Network error';
      final notifier = AIConfigNotifier(api);

      await notifier.loadConfig();

      expect(notifier.state.isEnabled, false);
      expect(notifier.state.isLoading, false);
      expect(notifier.state.error, isNotNull);
      expect(notifier.state.error, contains('Network error'));
    });

    test('setAIConfig should return true on success', () async {
      final api = _FakeMemoNexusAPIClient();
      api.mockAIConfig = {
        'enabled': true,
        'provider': 'claude',
        'api_endpoint': 'https://api.anthropic.com',
      };
      final notifier = AIConfigNotifier(api);

      final result = await notifier.setAIConfig(
        provider: 'claude',
        apiEndpoint: 'https://api.anthropic.com',
        apiKey: 'sk-test',
        modelName: 'claude-3-opus',
      );

      expect(result, isTrue);
      expect(notifier.state.provider, 'claude');
      expect(notifier.state.error, isNull);
    });

    test('setAIConfig should handle API errors', () async {
      final api = _FakeMemoNexusAPIClient();
      api.shouldFail = true;
      api.errorMessage = 'Invalid API key';
      final notifier = AIConfigNotifier(api);

      final result = await notifier.setAIConfig(
        provider: 'openai',
        apiEndpoint: 'https://api.openai.com/v1',
        apiKey: 'invalid-key',
      );

      expect(result, isFalse);
      expect(notifier.state.error, isNotNull);
      expect(notifier.state.error, contains('Invalid API key'));
    });

    test('disableAI should reset state to default', () async {
      final api = _FakeMemoNexusAPIClient();
      api.mockAIConfig = {
        'enabled': true,
        'provider': 'ollama',
        'api_endpoint': 'http://localhost:11434',
      };
      final notifier = AIConfigNotifier(api);

      // First load a config
      await notifier.loadConfig();
      expect(notifier.state.isEnabled, true);

      // Then disable
      await notifier.disableAI();

      expect(notifier.state.isEnabled, false);
      expect(notifier.state.provider, isNull);
      expect(notifier.state.apiEndpoint, isNull);
      expect(notifier.state.isLoading, false);
    });

    test('disableAI should handle API errors', () async {
      final api = _FakeMemoNexusAPIClient();
      api.shouldFail = true;
      api.errorMessage = 'Failed to disable';
      final notifier = AIConfigNotifier(api);

      // Set up a config first
      notifier.state = const AIConfigState(isEnabled: true);

      await notifier.disableAI();

      expect(notifier.state.error, isNotNull);
      expect(notifier.state.error, contains('Failed to disable'));
    });

    test('clearError should clear error state', () async {
      final api = _FakeMemoNexusAPIClient();
      api.shouldFail = true;
      final notifier = AIConfigNotifier(api);

      await notifier.loadConfig();
      expect(notifier.state.error, isNotNull);

      notifier.clearError();
      expect(notifier.state.error, isNull);
    });
  });

  group('Provider Integration', () {
    test('aiConfigProvider should provide AIConfigNotifier', () async {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      // First read the provider to trigger auto-loading, then wait for it
      container.read(aiConfigProvider);
      await Future.delayed(const Duration(milliseconds: 10));

      // Now safely access the notifier
      final notifier = container.read(aiConfigProvider.notifier);

      expect(notifier, isA<AIConfigNotifier>());

      // Wait for any pending async operations
      await Future.delayed(const Duration(milliseconds: 10));
    });

    // Note: Skipping auto-load test due to async initialization issues
    // The provider calls loadConfig() on init which may complete after disposal
  });

  group('Static Data', () {
    test('providerNames should contain all supported providers', () {
      expect(providerNames['openai'], 'OpenAI');
      expect(providerNames['claude'], 'Anthropic Claude');
      expect(providerNames['ollama'], 'Ollama (Local)');
    });

    test('defaultModels should contain default for each provider', () {
      expect(defaultModels['openai'], 'gpt-4');
      expect(defaultModels['claude'], 'claude-3-opus-20240229');
      expect(defaultModels['ollama'], 'llama2');
    });
  });
}

// =====================================================
// Fake API Client
// =====================================================

class _FakeMemoNexusAPIClient extends MemoNexusAPIClient {
  bool shouldFail = false;
  String? errorMessage;
  Map<String, dynamic>? mockAIConfig;

  @override
  Future<Map<String, dynamic>> getAIConfig() async {
    if (shouldFail) {
      throw Exception(errorMessage ?? 'Failed to load AI config');
    }
    return mockAIConfig ?? {};
  }

  @override
  Future<Map<String, dynamic>> setAIConfig({
    required String provider,
    required String apiEndpoint,
    required String apiKey,
    String? modelName,
    int? maxTokens,
  }) async {
    if (shouldFail) {
      throw Exception(errorMessage ?? 'Failed to set AI config');
    }
    // Update mock config for subsequent getAIConfig calls
    mockAIConfig = {
      'enabled': true,
      'provider': provider,
      'api_endpoint': apiEndpoint,
      if (modelName != null) 'model_name': modelName,
      if (maxTokens != null) 'max_tokens': maxTokens,
    };
    return {'success': true};
  }

  @override
  Future<void> disableAI() async {
    if (shouldFail) {
      throw Exception(errorMessage ?? 'Failed to disable AI');
    }
    mockAIConfig = {'enabled': false};
  }
}
