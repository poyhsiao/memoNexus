import 'package:uuid/uuid.dart';

// MediaType enum for content items
enum MediaType {
  web,
  image,
  video,
  pdf,
  markdown,
}

// ContentItem model
class ContentItem {
  final String id;
  final String title;
  final String contentText;
  final String? sourceUrl;
  final MediaType mediaType;
  final List<String> tags;
  final String? summary;
  final bool isDeleted;
  final DateTime createdAt;
  final DateTime updatedAt;
  final int version;
  final String? contentHash;

  ContentItem({
    String? id,
    required this.title,
    required this.contentText,
    this.sourceUrl,
    required this.mediaType,
    required this.tags,
    this.summary,
    this.isDeleted = false,
    required this.createdAt,
    required this.updatedAt,
    required this.version,
    this.contentHash,
  }) : id = id ?? const Uuid().v4();

  factory ContentItem.fromJson(Map<String, dynamic> json) {
    return ContentItem(
      id: json['id'] as String,
      title: json['title'] as String,
      contentText: json['content_text'] as String,
      sourceUrl: json['source_url'] as String?,
      mediaType: MediaType.values.firstWhere(
        (e) => e.name == json['media_type'],
        orElse: () => MediaType.web,
      ),
      tags: (json['tags'] as String)
          .split(',')
          .where((t) => t.isNotEmpty)
          .toList(),
      summary: json['summary'] as String?,
      isDeleted: json['is_deleted'] == 1,
      createdAt: DateTime.fromMillisecondsSinceEpoch(json['created_at'] * 1000),
      updatedAt: DateTime.fromMillisecondsSinceEpoch(json['updated_at'] * 1000),
      version: json['version'] as int,
      contentHash: json['content_hash'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'title': title,
      'content_text': contentText,
      'source_url': sourceUrl,
      'media_type': mediaType.name,
      'tags': tags.join(','),
      'summary': summary,
      'is_deleted': isDeleted ? 1 : 0,
      'created_at': createdAt.millisecondsSinceEpoch ~/ 1000,
      'updated_at': updatedAt.millisecondsSinceEpoch ~/ 1000,
      'version': version,
      'content_hash': contentHash,
    };
  }

  ContentItem copyWith({
    String? title,
    String? contentText,
    List<String>? tags,
    String? summary,
    bool? isDeleted,
    DateTime? updatedAt,
    String? contentHash,
  }) {
    return ContentItem(
      id: id,
      title: title ?? this.title,
      contentText: contentText ?? this.contentText,
      sourceUrl: sourceUrl,
      mediaType: mediaType,
      tags: tags ?? this.tags,
      summary: summary ?? this.summary,
      isDeleted: isDeleted ?? this.isDeleted,
      createdAt: this.createdAt, // createdAt is immutable
      updatedAt: updatedAt ?? this.updatedAt,
      version: version + 1,
      contentHash: contentHash ?? this.contentHash,
    );
  }
}

// Tag model
class Tag {
  final String id;
  final String name;
  final String color;
  final bool isDeleted;
  final DateTime createdAt;
  final DateTime updatedAt;

  Tag({
    String? id,
    required this.name,
    this.color = '#3B82F6',
    this.isDeleted = false,
    required this.createdAt,
    required this.updatedAt,
  }) : id = id ?? const Uuid().v4();

  factory Tag.fromJson(Map<String, dynamic> json) {
    return Tag(
      id: json['id'] as String,
      name: json['name'] as String,
      color: json['color'] as String? ?? '#3B82F6',
      isDeleted: json['is_deleted'] == 1,
      createdAt: DateTime.fromMillisecondsSinceEpoch(json['created_at'] * 1000),
      updatedAt: DateTime.fromMillisecondsSinceEpoch(json['updated_at'] * 1000),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      'color': color,
      'is_deleted': isDeleted ? 1 : 0,
      'created_at': createdAt.millisecondsSinceEpoch ~/ 1000,
      'updated_at': updatedAt.millisecondsSinceEpoch ~/ 1000,
    };
  }
}

// SearchResult model
class SearchResult {
  final ContentItem item;
  final double relevance;
  final List<String> matchedTerms;

  SearchResult({
    required this.item,
    required this.relevance,
    required this.matchedTerms,
  });

  factory SearchResult.fromJson(Map<String, dynamic> json) {
    return SearchResult(
      item: ContentItem.fromJson(json['item'] as Map<String, dynamic>),
      relevance: (json['relevance'] as num).toDouble(),
      matchedTerms: (json['matched_terms'] as List<dynamic>)
          .map((e) => e as String)
          .toList(),
    );
  }
}
