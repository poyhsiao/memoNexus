// API Client for desktop platforms (communicates with embedded PocketBase via REST)
// Mobile platforms use FFI bridge instead (ffi_bridge.dart)

import 'dart:async';
import 'dart:convert';
import 'package:http/http.dart' as http;

class MemoNexusAPIClient {
  final String baseUrl;
  final String token;

  MemoNexusAPIClient({
    this.baseUrl = 'http://localhost:8090/api',
    this.token = 'X-Local-Token',
  });

  Map<String, String> get _headers => {
        'Content-Type': 'application/json',
        'X-Local-Token': token,
      };

  // =====================================================
  // Content Items
  // =====================================================

  Future<Map<String, dynamic>> listContentItems({
    int page = 1,
    int perPage = 20,
    String sort = 'created_at',
    String order = 'desc',
    String? mediaType,
    String? tag,
  }) async {
    final queryParams = <String, String>{
      'page': page.toString(),
      'per_page': perPage.toString(),
      'sort': sort,
      'order': order,
    };
    if (mediaType != null) queryParams['media_type'] = mediaType;
    if (tag != null) queryParams['tag'] = tag;

    final uri = Uri.parse('$baseUrl/content')
        .replace(queryParameters: queryParams);

    final response = await http.get(uri, headers: _headers);
    _checkError(response);

    final data = jsonDecode(response.body) as Map<String, dynamic>;
    return data;
  }

  Future<Map<String, dynamic>> createContentFromURL({
    required String sourceUrl,
    String? title,
    List<String>? tags,
  }) async {
    final response = await http.post(
      Uri.parse('$baseUrl/content'),
      headers: _headers,
      body: jsonEncode({
        'type': 'url',
        'source_url': sourceUrl,
        if (title != null) 'title': title,
        if (tags != null) 'tags': tags,
      }),
    );
    _checkError(response);

    final data = jsonDecode(response.body) as Map<String, dynamic>;
    return data;
  }

  Future<Map<String, dynamic>> createContentFromFile({
    required String filePath,
    String? title,
    List<String>? tags,
  }) async {
    final response = await http.post(
      Uri.parse('$baseUrl/content'),
      headers: _headers,
      body: jsonEncode({
        'type': 'file',
        'file_path': filePath,
        if (title != null) 'title': title,
        if (tags != null) 'tags': tags,
      }),
    );
    _checkError(response);

    final data = jsonDecode(response.body) as Map<String, dynamic>;
    return data;
  }

  Future<Map<String, dynamic>> getContentItem(String id) async {
    final response = await http.get(
      Uri.parse('$baseUrl/content/$id'),
      headers: _headers,
    );
    _checkError(response);

    final data = jsonDecode(response.body) as Map<String, dynamic>;
    return data;
  }

  Future<Map<String, dynamic>> updateContentItem(
    String id, {
    String? title,
    List<String>? tags,
  }) async {
    final response = await http.put(
      Uri.parse('$baseUrl/content/$id'),
      headers: _headers,
      body: jsonEncode({
        if (title != null) 'title': title,
        if (tags != null) 'tags': tags,
      }),
    );
    _checkError(response);

    final data = jsonDecode(response.body) as Map<String, dynamic>;
    return data;
  }

  Future<void> deleteContentItem(String id) async {
    final response = await http.delete(
      Uri.parse('$baseUrl/content/$id'),
      headers: _headers,
    );
    _checkError(response);
  }

  // =====================================================
  // Search
  // =====================================================

  Future<Map<String, dynamic>> search({
    required String query,
    int limit = 20,
    String? mediaType,
    String? tags,
    int? dateFrom,
    int? dateTo,
  }) async {
    final queryParams = <String, String>{
      'q': query,
      'limit': limit.toString(),
    };
    if (mediaType != null) queryParams['media_type'] = mediaType;
    if (tags != null) queryParams['tags'] = tags;
    if (dateFrom != null) queryParams['date_from'] = dateFrom.toString();
    if (dateTo != null) queryParams['date_to'] = dateTo.toString();

    final uri = Uri.parse('$baseUrl/search')
        .replace(queryParameters: queryParams);

    final response = await http.get(uri, headers: _headers);
    _checkError(response);

    final data = jsonDecode(response.body) as Map<String, dynamic>;
    return data;
  }

  // =====================================================
  // Tags
  // =====================================================

  Future<List<dynamic>> listTags() async {
    final response = await http.get(
      Uri.parse('$baseUrl/tags'),
      headers: _headers,
    );
    _checkError(response);

    final data = jsonDecode(response.body) as List;
    return data;
  }

  Future<Map<String, dynamic>> createTag({
    required String name,
    String color = '#3B82F6',
  }) async {
    final response = await http.post(
      Uri.parse('$baseUrl/tags'),
      headers: _headers,
      body: jsonEncode({
        'name': name,
        'color': color,
      }),
    );
    _checkError(response);

    final data = jsonDecode(response.body) as Map<String, dynamic>;
    return data;
  }

  Future<Map<String, dynamic>> updateTag(
    String id, {
    String? name,
    String? color,
  }) async {
    final response = await http.put(
      Uri.parse('$baseUrl/tags/$id'),
      headers: _headers,
      body: jsonEncode({
        if (name != null) 'name': name,
        if (color != null) 'color': color,
      }),
    );
    _checkError(response);

    final data = jsonDecode(response.body) as Map<String, dynamic>;
    return data;
  }

  Future<void> deleteTag(String id) async {
    final response = await http.delete(
      Uri.parse('$baseUrl/tags/$id'),
      headers: _headers,
    );
    _checkError(response);
  }

  // =====================================================
  // AI Configuration
  // =====================================================

  Future<Map<String, dynamic>> getAIConfig() async {
    final response = await http.get(
      Uri.parse('$baseUrl/ai/config'),
      headers: _headers,
    );
    _checkError(response);

    final data = jsonDecode(response.body) as Map<String, dynamic>;
    return data;
  }

  Future<Map<String, dynamic>> setAIConfig({
    required String provider,
    required String apiEndpoint,
    required String apiKey,
    String? modelName,
    int? maxTokens,
  }) async {
    final response = await http.post(
      Uri.parse('$baseUrl/ai/config'),
      headers: _headers,
      body: jsonEncode({
        'provider': provider,
        'api_endpoint': apiEndpoint,
        'api_key': apiKey,
        if (modelName != null) 'model_name': modelName,
        if (maxTokens != null) 'max_tokens': maxTokens,
      }),
    );
    _checkError(response);

    final data = jsonDecode(response.body) as Map<String, dynamic>;
    return data;
  }

  Future<void> disableAI() async {
    final response = await http.delete(
      Uri.parse('$baseUrl/ai/config'),
      headers: _headers,
    );
    _checkError(response);
  }

  Future<Map<String, dynamic>> generateSummary(String id) async {
    final response = await http.post(
      Uri.parse('$baseUrl/content/$id/summary'),
      headers: _headers,
    );
    _checkError(response);

    final data = jsonDecode(response.body) as Map<String, dynamic>;
    return data;
  }

  // =====================================================
  // Sync Credentials
  // =====================================================

  Future<Map<String, dynamic>> getSyncCredentials() async {
    final response = await http.get(
      Uri.parse('$baseUrl/sync/credentials'),
      headers: _headers,
    );
    _checkError(response);

    final data = jsonDecode(response.body) as Map<String, dynamic>;
    return data;
  }

  Future<Map<String, dynamic>> configureSync({
    required String endpoint,
    required String bucketName,
    required String accessKey,
    required String secretKey,
    String? region,
  }) async {
    final response = await http.post(
      Uri.parse('$baseUrl/sync/credentials'),
      headers: _headers,
      body: jsonEncode({
        'endpoint': endpoint,
        'bucket_name': bucketName,
        'access_key': accessKey,
        'secret_key': secretKey,
        if (region != null) 'region': region,
      }),
    );
    _checkError(response);

    final data = jsonDecode(response.body) as Map<String, dynamic>;
    return data;
  }

  Future<void> disableSync() async {
    final response = await http.delete(
      Uri.parse('$baseUrl/sync/credentials'),
      headers: _headers,
    );
    _checkError(response);
  }

  Future<Map<String, dynamic>> getSyncStatus() async {
    final response = await http.get(
      Uri.parse('$baseUrl/sync/status'),
      headers: _headers,
    );
    _checkError(response);

    final data = jsonDecode(response.body) as Map<String, dynamic>;
    return data;
  }

  Future<Map<String, dynamic>> triggerSync() async {
    final response = await http.post(
      Uri.parse('$baseUrl/sync/now'),
      headers: _headers,
    );
    _checkError(response);

    final data = jsonDecode(response.body) as Map<String, dynamic>;
    return data;
  }

  // =====================================================
  // Export/Import
  // =====================================================

  Future<Map<String, dynamic>> exportData({
    required String password,
    bool includeMedia = true,
  }) async {
    final response = await http.post(
      Uri.parse('$baseUrl/export'),
      headers: _headers,
      body: jsonEncode({
        'password': password,
        'include_media': includeMedia,
      }),
    );
    _checkError(response);

    final data = jsonDecode(response.body) as Map<String, dynamic>;
    return data;
  }

  Future<Map<String, dynamic>> importData({
    required String archivePath,
    required String password,
  }) async {
    final response = await http.post(
      Uri.parse('$baseUrl/import'),
      headers: _headers,
      body: jsonEncode({
        'archive_path': archivePath,
        'password': password,
      }),
    );
    _checkError(response);

    final data = jsonDecode(response.body) as Map<String, dynamic>;
    return data;
  }

  // =====================================================
  // Health Check
  // =====================================================

  Future<Map<String, dynamic>> healthCheck() async {
    final response = await http.get(
      Uri.parse('$baseUrl/health'),
      headers: _headers,
    );
    _checkError(response);

    final data = jsonDecode(response.body) as Map<String, dynamic>;
    return data;
  }

  // =====================================================
  // Error Handling
  // =====================================================

  void _checkError(http.Response response) {
    if (response.statusCode >= 400) {
      final body = jsonDecode(response.body);
      throw APIException(
        statusCode: response.statusCode,
        code: body['code'] as String? ?? 'UNKNOWN_ERROR',
        message: body['error'] as String? ?? 'Unknown error',
      );
    }
  }
}

class APIException implements Exception {
  final int statusCode;
  final String code;
  final String message;

  APIException({
    required this.statusCode,
    required this.code,
    required this.message,
  });

  @override
  String toString() => '[$code] $message (HTTP $statusCode)';
}
