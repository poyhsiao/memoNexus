// FFI Bridge for mobile platforms (Android/iOS)
// Uses Dart FFI to call Go Core shared library directly
// Desktop platforms use REST/WebSocket instead (api_client.dart)

import 'dart:convert';
import 'dart:ffi';
import 'dart:io';
import 'package:ffi/ffi.dart';

// =====================================================
// Native Library Loading
// =====================================================

class MemoNexusFFI {
  static DynamicLibrary? _lib;
  static late MemoNexusFFIBindings _bindings;

  // Initialize the FFI bridge
  static void init() {
    if (_lib != null) return;

    try {
      // Load platform-specific library
      if (Platform.isAndroid) {
        _lib = DynamicLibrary.open('libmemonexus.so');
      } else if (Platform.isIOS) {
        _lib = DynamicLibrary.open('memonexus.framework/memonexus');
      } else {
        throw UnsupportedError('FFI bridge is only supported on mobile platforms');
      }

      _bindings = MemoNexusFFIBindings(_lib!);

      // Initialize Go runtime
      _bindings.init();

      // ignore: avoid_print
      print('[FFI] MemoNexus Core loaded successfully');
    } catch (e) {
      // ignore: avoid_print
      print('[FFI] Failed to load library: $e');
      rethrow;
    }
  }

  // Cleanup resources
  static void cleanup() {
    if (_lib != null) {
      _bindings.cleanup();
      _lib = null;
    }
  }

  // Get last error message
  static String getLastError() {
    final errorPtr = _bindings.getLastError();
    if (errorPtr == nullptr) return 'Unknown error';
    final errorStr = errorPtr.toDartString();
    // Go runtime will free this string
    return errorStr;
  }

  // =====================================================
  // Content Operations
  // =====================================================

  static Map<String, dynamic> createContent({
    required String title,
    required String contentText,
    String? sourceUrl,
    required String mediaType,
    List<String>? tags,
    String? contentHash,
  }) {
    final titlePtr = title.toNativeUtf8();
    final contentTextPtr = contentText.toNativeUtf8();
    final sourceUrlPtr = sourceUrl?.toNativeUtf8() ?? nullptr;
    final mediaTypePtr = mediaType.toNativeUtf8();
    final tagsStr = tags?.join(',') ?? '';
    final tagsPtr = tagsStr.toNativeUtf8();
    final contentHashPtr = contentHash?.toNativeUtf8() ?? nullptr;

    try {
      final resultPtr = _bindings.contentCreate(
        titlePtr.cast(),
        contentTextPtr.cast(),
        sourceUrlPtr.cast(),
        mediaTypePtr.cast(),
        tagsPtr.cast(),
        contentHashPtr.cast(),
      );

      if (resultPtr == nullptr) {
        throw FFIException(getLastError());
      }

      final resultJson = resultPtr.toDartString();
      // Go runtime will free the result string
      final result = jsonDecode(resultJson) as Map<String, dynamic>;
      return result;
    } finally {
      // Free allocated strings
      calloc.free(titlePtr.cast());
      calloc.free(contentTextPtr.cast());
      if (sourceUrlPtr != nullptr) calloc.free(sourceUrlPtr.cast());
      calloc.free(mediaTypePtr.cast());
      calloc.free(tagsPtr.cast());
      if (contentHashPtr != nullptr) calloc.free(contentHashPtr.cast());
    }
  }

  static List<dynamic> listContentItems({
    int limit = 20,
    int offset = 0,
  }) {
    final resultPtr = _bindings.contentList(limit, offset);

    if (resultPtr == nullptr) {
      throw FFIException(getLastError());
    }

    final resultJson = resultPtr.toDartString();
    // Go runtime will free the result string
    final result = jsonDecode(resultJson) as Map<String, dynamic>;
    return result['items'] as List<dynamic>;
  }

  static Map<String, dynamic> getContentItem(String id) {
    final idPtr = id.toNativeUtf8();

    try {
      final resultPtr = _bindings.contentGet(idPtr.cast());

      if (resultPtr == nullptr) {
        throw FFIException(getLastError());
      }

      final resultJson = resultPtr.toDartString();
      // Go runtime will free the result string
      final result = jsonDecode(resultJson) as Map<String, dynamic>;
      return result;
    } finally {
      calloc.free(idPtr.cast());
    }
  }

  static Map<String, dynamic> updateContentItem(
    String id, {
    String? title,
    List<String>? tags,
  }) {
    final idPtr = id.toNativeUtf8();
    final titlePtr = title?.toNativeUtf8() ?? nullptr;
    final tagsStr = tags?.join(',') ?? '';
    final tagsPtr = tagsStr.toNativeUtf8();

    try {
      final resultPtr = _bindings.contentUpdate(
        idPtr.cast(),
        titlePtr.cast(),
        tagsPtr.cast(),
      );

      if (resultPtr == nullptr) {
        throw FFIException(getLastError());
      }

      final resultJson = resultPtr.toDartString();
      // Go runtime will free the result string
      final result = jsonDecode(resultJson) as Map<String, dynamic>;
      return result;
    } finally {
      calloc.free(idPtr.cast());
      if (titlePtr != nullptr) calloc.free(titlePtr.cast());
      calloc.free(tagsPtr.cast());
    }
  }

  static void deleteContentItem(String id) {
    final idPtr = id.toNativeUtf8();

    try {
      final result = _bindings.contentDelete(idPtr.cast());

      if (result != 0) {
        throw FFIException(getLastError());
      }
    } finally {
      calloc.free(idPtr.cast());
    }
  }

  // =====================================================
  // Search Operations
  // =====================================================

  static List<dynamic> search({
    required String query,
    int limit = 20,
  }) {
    final queryPtr = query.toNativeUtf8();

    try {
      final resultPtr = _bindings.search(queryPtr.cast(), limit);

      if (resultPtr == nullptr) {
        throw FFIException(getLastError());
      }

      final resultJson = resultPtr.toDartString();
      // Go runtime will free the result string
      final result = jsonDecode(resultJson) as Map<String, dynamic>;
      return result['results'] as List<dynamic>;
    } finally {
      calloc.free(queryPtr.cast());
    }
  }

  // =====================================================
  // Tag Operations
  // =====================================================

  static List<dynamic> listTags() {
    final resultPtr = _bindings.tagList();

    if (resultPtr == nullptr) {
      throw FFIException(getLastError());
    }

    final resultJson = resultPtr.toDartString();
    // Go runtime will free the result string
    return jsonDecode(resultJson) as List<dynamic>;
  }

  static Map<String, dynamic> createTag({
    required String name,
    String color = '#3B82F6',
  }) {
    final namePtr = name.toNativeUtf8();
    final colorPtr = color.toNativeUtf8();

    try {
      final resultPtr = _bindings.tagCreate(namePtr.cast(), colorPtr.cast());

      if (resultPtr == nullptr) {
        throw FFIException(getLastError());
      }

      final resultJson = resultPtr.toDartString();
      // Go runtime will free the result string
      final result = jsonDecode(resultJson) as Map<String, dynamic>;
      return result;
    } finally {
      calloc.free(namePtr.cast());
      calloc.free(colorPtr.cast());
    }
  }
}

// =====================================================
// FFI Bindings
// =====================================================

class MemoNexusFFIBindings {
  final DynamicLibrary _lib;

  MemoNexusFFIBindings(this._lib);

  void init() {
    _lib.lookupFunction<Void Function(), void Function()>('Init')();
  }

  void cleanup() {
    _lib.lookupFunction<Void Function(), void Function()>('Cleanup')();
  }

  Pointer<Utf8> getLastError() {
    return _lib.lookupFunction<Pointer<Utf8> Function(), Pointer<Utf8> Function()>('GetLastError')();
  }

  // Content operations
  Pointer<Utf8> contentCreate(
    Pointer<Utf8> title,
    Pointer<Utf8> contentText,
    Pointer<Utf8> sourceUrl,
    Pointer<Utf8> mediaType,
    Pointer<Utf8> tags,
    Pointer<Utf8> contentHash,
  ) {
    return _lib.lookupFunction<
      Pointer<Utf8> Function(
        Pointer<Utf8>,
        Pointer<Utf8>,
        Pointer<Utf8>,
        Pointer<Utf8>,
        Pointer<Utf8>,
        Pointer<Utf8>,
      ), Pointer<Utf8> Function(
        Pointer<Utf8>,
        Pointer<Utf8>,
        Pointer<Utf8>,
        Pointer<Utf8>,
        Pointer<Utf8>,
        Pointer<Utf8>,
      )>('ContentCreate')(
      title,
      contentText,
      sourceUrl,
      mediaType,
      tags,
      contentHash,
    );
  }

  Pointer<Utf8> contentList(int limit, int offset) {
    return _lib.lookupFunction<
      Pointer<Utf8> Function(Int32, Int32), Pointer<Utf8> Function(int, int)>('ContentList')(limit, offset);
  }

  Pointer<Utf8> contentGet(Pointer<Utf8> id) {
    return _lib.lookupFunction<
      Pointer<Utf8> Function(Pointer<Utf8>), Pointer<Utf8> Function(Pointer<Utf8>)>('ContentGet')(id);
  }

  Pointer<Utf8> contentUpdate(
    Pointer<Utf8> id,
    Pointer<Utf8> title,
    Pointer<Utf8> tags,
  ) {
    return _lib.lookupFunction<
      Pointer<Utf8> Function(Pointer<Utf8>, Pointer<Utf8>, Pointer<Utf8>),
      Pointer<Utf8> Function(Pointer<Utf8>, Pointer<Utf8>, Pointer<Utf8>)>('ContentUpdate')(
      id,
      title,
      tags,
    );
  }

  int contentDelete(Pointer<Utf8> id) {
    return _lib.lookupFunction<
      Int32 Function(Pointer<Utf8>), int Function(Pointer<Utf8>)>('ContentDelete')(id);
  }

  // Search operations
  Pointer<Utf8> search(Pointer<Utf8> query, int limit) {
    return _lib.lookupFunction<
      Pointer<Utf8> Function(Pointer<Utf8>, Int32), Pointer<Utf8> Function(Pointer<Utf8>, int)>('Search')(query, limit);
  }

  // Tag operations
  Pointer<Utf8> tagList() {
    return _lib.lookupFunction<Pointer<Utf8> Function(), Pointer<Utf8> Function()>('TagList')();
  }

  Pointer<Utf8> tagCreate(Pointer<Utf8> name, Pointer<Utf8> color) {
    return _lib.lookupFunction<
      Pointer<Utf8> Function(Pointer<Utf8>, Pointer<Utf8>), Pointer<Utf8> Function(Pointer<Utf8>, Pointer<Utf8>)>('TagCreate')(name, color);
  }
}

// =====================================================
// Exceptions
// =====================================================

class FFIException implements Exception {
  final String message;

  FFIException(this.message);

  @override
  String toString() => '[FFI] $message';
}
