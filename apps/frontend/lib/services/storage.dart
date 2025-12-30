// Local Storage Service for mobile testing (uses sqflite)
// Desktop platforms use the API client instead (api_client.dart)

import 'package:sqflite/sqflite.dart';
import 'package:path/path.dart';

class MemoNexusStorage {
  static Database? _db;
  static const String _dbName = 'memonexus.db';
  static const int _dbVersion = 1;

  static Future<Database> get db async {
    if (_db != null) return _db!;
    _db = await _initDB();
    return _db!;
  }

  static Future<Database> _initDB() async {
    final dbPath = await getDatabasesPath();
    final path = join(dbPath, _dbName);

    return await openDatabase(
      path,
      version: _dbVersion,
      onCreate: _onCreate,
      onConfigure: _onConfigure,
    );
  }

  static Future<void> _onConfigure(Database db) async {
    // Enable foreign keys
    await db.execute('PRAGMA foreign_keys = ON');
    // Enable WAL mode
    await db.execute('PRAGMA journal_mode = WAL');
  }

  static Future<void> _onCreate(Database db, int version) async {
    // Create content_items table
    await db.execute('''
      CREATE TABLE content_items (
        id TEXT PRIMARY KEY CHECK(length(id) = 36),
        title TEXT NOT NULL CHECK(length(title) > 0),
        content_text TEXT NOT NULL DEFAULT '',
        source_url TEXT,
        media_type TEXT NOT NULL CHECK(media_type IN ('web', 'image', 'video', 'pdf', 'markdown')),
        tags TEXT DEFAULT '',
        summary TEXT,
        is_deleted INTEGER NOT NULL DEFAULT 0 CHECK(is_deleted IN (0, 1)),
        created_at INTEGER NOT NULL,
        updated_at INTEGER NOT NULL,
        version INTEGER NOT NULL DEFAULT 1,
        content_hash TEXT
      )
    ''');

    // Create tags table
    await db.execute('''
      CREATE TABLE tags (
        id TEXT PRIMARY KEY CHECK(length(id) = 36),
        name TEXT NOT NULL UNIQUE CHECK(length(name) > 0),
        color TEXT DEFAULT '#3B82F6',
        is_deleted INTEGER NOT NULL DEFAULT 0,
        created_at INTEGER NOT NULL,
        updated_at INTEGER NOT NULL
      )
    ''');

    // Create content_tags relationship table
    await db.execute('''
      CREATE TABLE content_tags (
        content_id TEXT NOT NULL CHECK(length(content_id) = 36),
        tag_id TEXT NOT NULL CHECK(length(tag_id) = 36),
        assigned_at INTEGER NOT NULL,
        PRIMARY KEY (content_id, tag_id),
        FOREIGN KEY (content_id) REFERENCES content_items(id) ON DELETE CASCADE,
        FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
      )
    ''');

    // Create indexes
    await db.execute(
        'CREATE INDEX idx_content_items_created_at ON content_items(created_at DESC)');
    await db.execute(
        'CREATE INDEX idx_content_items_is_deleted ON content_items(is_deleted)');
    await db.execute('CREATE INDEX idx_tags_name ON tags(name)');
    await db.execute(
        'CREATE INDEX idx_content_tags_content_id ON content_tags(content_id)');
    await db.execute(
        'CREATE INDEX idx_content_tags_tag_id ON content_tags(tag_id)');
  }

  // =====================================================
  // Content Items
  // =====================================================

  static Future<Map<String, dynamic>> createContentItem({
    required String id,
    required String title,
    required String contentText,
    String? sourceUrl,
    required String mediaType,
    required String tags,
    String? summary,
    String? contentHash,
  }) async {
    final db = await MemoNexusStorage.db;
    final now = DateTime.now().millisecondsSinceEpoch ~/ 1000;

    await db.insert(
      'content_items',
      {
        'id': id,
        'title': title,
        'content_text': contentText,
        'source_url': sourceUrl,
        'media_type': mediaType,
        'tags': tags,
        'summary': summary,
        'is_deleted': 0,
        'created_at': now,
        'updated_at': now,
        'version': 1,
        'content_hash': contentHash,
      },
    );

    return await getContentItem(id);
  }

  static Future<Map<String, dynamic>> getContentItem(String id) async {
    final db = await MemoNexusStorage.db;
    final results = await db.query(
      'content_items',
      where: 'id = ? AND is_deleted = 0',
      whereArgs: [id],
    );

    if (results.isEmpty) {
      throw Exception('Content item not found: $id');
    }

    return results.first;
  }

  static Future<List<Map<String, dynamic>>> listContentItems({
    int limit = 20,
    int offset = 0,
    String? mediaType,
  }) async {
    final db = await MemoNexusStorage.db;

    String? where;
    List<dynamic>? whereArgs;

    if (mediaType != null) {
      where = 'is_deleted = 0 AND media_type = ?';
      whereArgs = [mediaType];
    } else {
      where = 'is_deleted = 0';
    }

    final results = await db.query(
      'content_items',
      where: where,
      whereArgs: whereArgs,
      orderBy: 'created_at DESC',
      limit: limit,
      offset: offset,
    );

    return results;
  }

  static Future<Map<String, dynamic>> updateContentItem(
    String id, {
    String? title,
    String? contentText,
    String? tags,
    String? summary,
    String? contentHash,
  }) async {
    final db = await MemoNexusStorage.db;
    final now = DateTime.now().millisecondsSinceEpoch ~/ 1000;

    // Get current version
    final current = await getContentItem(id);
    final version = (current['version'] as int) + 1;

    final updates = <String, dynamic>{
      'updated_at': now,
      'version': version,
    };

    if (title != null) updates['title'] = title;
    if (contentText != null) updates['content_text'] = contentText;
    if (tags != null) updates['tags'] = tags;
    if (summary != null) updates['summary'] = summary;
    if (contentHash != null) updates['content_hash'] = contentHash;

    await db.update(
      'content_items',
      updates,
      where: 'id = ? AND is_deleted = 0',
      whereArgs: [id],
    );

    return await getContentItem(id);
  }

  static Future<void> deleteContentItem(String id) async {
    final db = await MemoNexusStorage.db;
    final now = DateTime.now().millisecondsSinceEpoch ~/ 1000;

    await db.update(
      'content_items',
      {
        'is_deleted': 1,
        'updated_at': now,
      },
      where: 'id = ?',
      whereArgs: [id],
    );
  }

  // =====================================================
  // Tags
  // =====================================================

  static Future<Map<String, dynamic>> createTag({
    required String id,
    required String name,
    String color = '#3B82F6',
  }) async {
    final db = await MemoNexusStorage.db;
    final now = DateTime.now().millisecondsSinceEpoch ~/ 1000;

    await db.insert(
      'tags',
      {
        'id': id,
        'name': name,
        'color': color,
        'is_deleted': 0,
        'created_at': now,
        'updated_at': now,
      },
    );

    return await getTag(id);
  }

  static Future<Map<String, dynamic>> getTag(String id) async {
    final db = await MemoNexusStorage.db;
    final results = await db.query(
      'tags',
      where: 'id = ? AND is_deleted = 0',
      whereArgs: [id],
    );

    if (results.isEmpty) {
      throw Exception('Tag not found: $id');
    }

    return results.first;
  }

  static Future<List<Map<String, dynamic>>> listTags() async {
    final db = await MemoNexusStorage.db;

    final results = await db.query(
      'tags',
      where: 'is_deleted = 0',
      orderBy: 'name',
    );

    return results;
  }

  static Future<Map<String, dynamic>> updateTag(
    String id, {
    String? name,
    String? color,
  }) async {
    final db = await MemoNexusStorage.db;
    final now = DateTime.now().millisecondsSinceEpoch ~/ 1000;

    final updates = <String, dynamic>{
      'updated_at': now,
    };

    if (name != null) updates['name'] = name;
    if (color != null) updates['color'] = color;

    await db.update(
      'tags',
      updates,
      where: 'id = ? AND is_deleted = 0',
      whereArgs: [id],
    );

    return await getTag(id);
  }

  static Future<void> deleteTag(String id) async {
    final db = await MemoNexusStorage.db;
    final now = DateTime.now().millisecondsSinceEpoch ~/ 1000;

    await db.update(
      'tags',
      {
        'is_deleted': 1,
        'updated_at': now,
      },
      where: 'id = ?',
      whereArgs: [id],
    );
  }

  // =====================================================
  // Search (Full-Text)
  // =====================================================

  static Future<List<Map<String, dynamic>>> search({
    required String query,
    int limit = 20,
    String? mediaType,
  }) async {
    final db = await MemoNexusStorage.db;

    // Note: FTS5 is not available in sqflite by default
    // For mobile testing, use LIKE pattern matching
    final pattern = '%${query.toLowerCase()}%';

    String where = '''
      is_deleted = 0 AND (
        LOWER(title) LIKE ? OR
        LOWER(content_text) LIKE ? OR
        LOWER(tags) LIKE ?
      )
    ''';

    final whereArgs = <dynamic>[pattern, pattern, pattern];

    if (mediaType != null) {
      where += ' AND media_type = ?';
      whereArgs.add(mediaType);
    }

    final results = await db.query(
      'content_items',
      where: where,
      whereArgs: whereArgs,
      orderBy: 'created_at DESC',
      limit: limit,
    );

    return results;
  }

  // =====================================================
  // Content-Tag Relationships
  // =====================================================

  static Future<void> assignTagToContent({
    required String contentId,
    required String tagId,
  }) async {
    final db = await MemoNexusStorage.db;
    final now = DateTime.now().millisecondsSinceEpoch ~/ 1000;

    await db.insert(
      'content_tags',
      {
        'content_id': contentId,
        'tag_id': tagId,
        'assigned_at': now,
      },
      conflictAlgorithm: ConflictAlgorithm.replace,
    );
  }

  static Future<void> removeTagFromContent({
    required String contentId,
    required String tagId,
  }) async {
    final db = await MemoNexusStorage.db;

    await db.delete(
      'content_tags',
      where: 'content_id = ? AND tag_id = ?',
      whereArgs: [contentId, tagId],
    );
  }

  static Future<List<Map<String, dynamic>>> getTagsForContent(
      String contentId) async {
    final db = await MemoNexusStorage.db;

    final results = await db.rawQuery('''
      SELECT t.*
      FROM tags t
      INNER JOIN content_tags ct ON ct.tag_id = t.id
      WHERE ct.content_id = ? AND t.is_deleted = 0
      ORDER BY t.name
    ''', [contentId]);

    return results;
  }

  // =====================================================
  // Database Management
  // =====================================================

  static Future<void> close() async {
    if (_db != null) {
      await _db!.close();
      _db = null;
    }
  }

  static Future<void> clear() async {
    final db = await MemoNexusStorage.db;

    await db.delete('content_tags');
    await db.delete('content_items');
    await db.delete('tags');
  }
}
