// Capture Provider for Riverpod state management
// Manages content capture from URLs and file uploads

import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/content_item.dart';
import '../services/api_client.dart';
import 'content_provider.dart';

// =====================================================
// Capture State
// =====================================================

enum CaptureType { url, file }

class CaptureState {
  final CaptureType type;
  final String url;
  final String? filePath;
  final String? fileName;
  final List<String> selectedTags;
  final String customTitle;
  final bool isSubmitting;
  final String? error;
  final ContentItem? createdItem;

  const CaptureState({
    this.type = CaptureType.url,
    this.url = '',
    this.filePath,
    this.fileName,
    this.selectedTags = const [],
    this.customTitle = '',
    this.isSubmitting = false,
    this.error,
    this.createdItem,
  });

  CaptureState copyWith({
    CaptureType? type,
    String? url,
    String? filePath,
    String? fileName,
    List<String>? selectedTags,
    String? customTitle,
    bool? isSubmitting,
    String? error,
    ContentItem? createdItem,
  }) {
    return CaptureState(
      type: type ?? this.type,
      url: url ?? this.url,
      filePath: filePath ?? this.filePath,
      fileName: fileName ?? this.fileName,
      selectedTags: selectedTags ?? this.selectedTags,
      customTitle: customTitle ?? this.customTitle,
      isSubmitting: isSubmitting ?? this.isSubmitting,
      error: error,
      createdItem: createdItem ?? this.createdItem,
    );
  }

  bool get isValid {
    if (type == CaptureType.url) {
      return url.isNotEmpty && _isValidUrl(url);
    } else {
      return filePath != null && filePath!.isNotEmpty;
    }
  }

  bool _isValidUrl(String urlString) {
    try {
      final uri = Uri.parse(urlString);
      return uri.hasScheme && (uri.scheme == 'http' || uri.scheme == 'https');
    } catch (_) {
      return false;
    }
  }
}

// =====================================================
// Capture Notifier
// =====================================================

class CaptureNotifier extends StateNotifier<CaptureState> {
  final MemoNexusAPIClient _api;

  CaptureNotifier(this._api) : super(const CaptureState());

  void setType(CaptureType type) {
    state = state.copyWith(type: type);
  }

  void setUrl(String url) {
    state = state.copyWith(url: url, error: null);
  }

  void setFile(String path, String name) {
    state = state.copyWith(filePath: path, fileName: name, error: null);
  }

  void clearFile() {
    state = state.copyWith(filePath: null, fileName: null);
  }

  void setCustomTitle(String title) {
    state = state.copyWith(customTitle: title);
  }

  void toggleTag(String tag) {
    final tags = List<String>.from(state.selectedTags);
    if (tags.contains(tag)) {
      tags.remove(tag);
    } else {
      tags.add(tag);
    }
    state = state.copyWith(selectedTags: tags);
  }

  void setTags(List<String> tags) {
    state = state.copyWith(selectedTags: tags);
  }

  Future<void> submit() async {
    if (!state.isValid || state.isSubmitting) return;

    state = state.copyWith(isSubmitting: true, error: null);

    try {
      ContentItem item;

      if (state.type == CaptureType.url) {
        final data = await _api.createContentFromURL(
          sourceUrl: state.url,
          title: state.customTitle.isNotEmpty ? state.customTitle : null,
          tags: state.selectedTags.isNotEmpty ? state.selectedTags : null,
        );
        item = ContentItem.fromJson(data);
      } else {
        final data = await _api.createContentFromFile(
          filePath: state.filePath!,
          title: state.customTitle.isNotEmpty ? state.customTitle : null,
          tags: state.selectedTags.isNotEmpty ? state.selectedTags : null,
        );
        item = ContentItem.fromJson(data);
      }

      state = state.copyWith(
        isSubmitting: false,
        createdItem: item,
        error: null,
      );
    } catch (e) {
      state = state.copyWith(
        isSubmitting: false,
        error: e.toString(),
      );
    }
  }

  void reset() {
    state = const CaptureState();
  }

  void clearError() {
    state = state.copyWith(error: null);
  }
}

// =====================================================
// Providers
// =====================================================

final captureProvider =
    StateNotifierProvider<CaptureNotifier, CaptureState>((ref) {
  final api = ref.watch(apiClientProvider);
  return CaptureNotifier(api);
});

// Tags list provider for tag picker
final tagsListProvider = FutureProvider<List<Tag>>((ref) async {
  final api = ref.watch(apiClientProvider);
  final data = await api.listTags();
  return data.map((json) => Tag.fromJson(json as Map<String, dynamic>)).toList();
});

// Available tags provider (auto-refresh)
final availableTagsProvider = Provider<List<String>>((ref) {
  final tagsAsync = ref.watch(tagsListProvider);
  return tagsAsync.value?.map((tag) => tag.name).toList() ?? [];
});
