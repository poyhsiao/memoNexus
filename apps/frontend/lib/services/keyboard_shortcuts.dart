// Keyboard Shortcuts Service - T208: Configure keyboard shortcuts documentation
//
// ★ Insight ─────────────────────────────────────
// 1. LogicalKeyboardKey provides platform-independent key codes
//    that work across macOS, Windows, and Linux.
// 2. SingleActivator allows creating key combinations with
//    modifiers (Ctrl, Alt, Shift, Meta).
// 3. ShortcutsIntent widget binds keyboard events to actions.
// ─────────────────────────────────────────────────

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

/// Keyboard shortcut definition
class Shortcut {
  final String label;
  final String description;
  final List<LogicalKeyboardKey> keys;

  const Shortcut({
    required this.label,
    required this.description,
    required this.keys,
  });

  /// Format the key combination for display
  String get formattedKeys {
    return keys.map((k) => _formatLogicalKey(k)).join(' + ');
  }

  String _formatLogicalKey(LogicalKeyboardKey key) {
    // Map common keys to platform-specific labels
    if (key == LogicalKeyboardKey.control) {
      return 'Ctrl';
    } else if (key == LogicalKeyboardKey.meta) {
      return 'Cmd'; // macOS
    } else if (key == LogicalKeyboardKey.alt) {
      return 'Alt';
    } else if (key == LogicalKeyboardKey.shift) {
      return 'Shift';
    } else if (key == LogicalKeyboardKey.keyN) {
      return 'N';
    } else if (key == LogicalKeyboardKey.keyO) {
      return 'O';
    } else if (key == LogicalKeyboardKey.keyS) {
      return 'S';
    } else if (key == LogicalKeyboardKey.keyF) {
      return 'F';
    } else if (key == LogicalKeyboardKey.keyH) {
      return 'H';
    } else if (key == LogicalKeyboardKey.comma) {
      return ',';
    } else if (key == LogicalKeyboardKey.slash) {
      return '/';
    } else if (key == LogicalKeyboardKey.escape) {
      return 'Esc';
    }
    return key.keyLabel;
  }
}

/// Predefined shortcuts for the application
class AppShortcuts {
  // File menu shortcuts
  static const Shortcut newNote = Shortcut(
    label: 'New Note',
    description: 'Create a new content item',
    keys: [
      LogicalKeyboardKey.meta,
      LogicalKeyboardKey.keyN,
    ],
  );

  static const Shortcut open = Shortcut(
    label: 'Open Import',
    description: 'Open import dialog',
    keys: [
      LogicalKeyboardKey.meta,
      LogicalKeyboardKey.keyO,
    ],
  );

  static const Shortcut save = Shortcut(
    label: 'Export',
    description: 'Export data',
    keys: [
      LogicalKeyboardKey.meta,
      LogicalKeyboardKey.keyS,
    ],
  );

  // Edit shortcuts
  static const Shortcut find = Shortcut(
    label: 'Find',
    description: 'Search content',
    keys: [
      LogicalKeyboardKey.meta,
      LogicalKeyboardKey.keyF,
    ],
  );

  // Navigation shortcuts
  static const Shortcut home = Shortcut(
    label: 'Go Home',
    description: 'Navigate to home screen',
    keys: [
      LogicalKeyboardKey.meta,
      LogicalKeyboardKey.shift,
      LogicalKeyboardKey.keyH,
    ],
  );

  static const Shortcut settings = Shortcut(
    label: 'Settings',
    description: 'Open settings',
    keys: [
      LogicalKeyboardKey.meta,
      LogicalKeyboardKey.comma,
    ],
  );

  // Utility shortcuts
  static const Shortcut help = Shortcut(
    label: 'Help',
    description: 'Show keyboard shortcuts',
    keys: [
      LogicalKeyboardKey.control,
      LogicalKeyboardKey.shift,
      LogicalKeyboardKey.slash,
    ],
  );

  static const Shortcut escape = Shortcut(
    label: 'Close/Back',
    description: 'Close dialog or go back',
    keys: [
      LogicalKeyboardKey.escape,
    ],
  );

  /// All shortcuts categorized by context
  static const Map<String, List<Shortcut>> byCategory = {
    'File': [
      newNote,
      open,
      save,
    ],
    'Edit': [
      find,
    ],
    'Navigation': [
      home,
      settings,
    ],
    'Utility': [
      help,
      escape,
    ],
  };

  /// Get all shortcuts as a flat list
  static List<Shortcut> get all => byCategory.values.expand((e) => e).toList();

  /// Create default shortcut mappings
  static Map<ShortcutActivator, Intent> get defaultMappings {
    return {
      // File menu
      const SingleActivator(LogicalKeyboardKey.keyN, meta: true):
          const NewNoteIntent(),
      const SingleActivator(LogicalKeyboardKey.keyO, meta: true):
          const OpenImportIntent(),
      const SingleActivator(LogicalKeyboardKey.keyS, meta: true):
          const ExportIntent(),

      // Edit
      const SingleActivator(LogicalKeyboardKey.keyF, meta: true):
          const FindIntent(),

      // Navigation
      const SingleActivator(LogicalKeyboardKey.keyH, meta: true, shift: true):
          const GoHomeIntent(),
      const SingleActivator(LogicalKeyboardKey.comma, meta: true):
          const SettingsIntent(),

      // Utility
      const SingleActivator(LogicalKeyboardKey.escape):
          const CloseIntent(),
    };
  }
}

// Intents for each action

class NewNoteIntent extends Intent {
  const NewNoteIntent();
}

class OpenImportIntent extends Intent {
  const OpenImportIntent();
}

class ExportIntent extends Intent {
  const ExportIntent();
}

class FindIntent extends Intent {
  const FindIntent();
}

class GoHomeIntent extends Intent {
  const GoHomeIntent();
}

class SettingsIntent extends Intent {
  const SettingsIntent();
}

class CloseIntent extends Intent {
  const CloseIntent();
}

/// Actions for handling shortcuts
typedef ShortcutCallback = void Function();

/// Shortcut action handler
class ShortcutActions {
  final ShortcutCallback? onNewNote;
  final ShortcutCallback? onOpenImport;
  final ShortcutCallback? onExport;
  final ShortcutCallback? onFind;
  final ShortcutCallback? onGoHome;
  final ShortcutCallback? onSettings;
  final ShortcutCallback? onClose;

  const ShortcutActions({
    this.onNewNote,
    this.onOpenImport,
    this.onExport,
    this.onFind,
    this.onGoHome,
    this.onSettings,
    this.onClose,
  });

  /// Create type-safe action mappings
  Map<Type, Action<Intent>> get toActionsMap {
    return {
      NewNoteIntent: CallbackAction<NewNoteIntent>(onInvoke: (_) => onNewNote?.call()),
      OpenImportIntent: CallbackAction<OpenImportIntent>(onInvoke: (_) => onOpenImport?.call()),
      ExportIntent: CallbackAction<ExportIntent>(onInvoke: (_) => onExport?.call()),
      FindIntent: CallbackAction<FindIntent>(onInvoke: (_) => onFind?.call()),
      GoHomeIntent: CallbackAction<GoHomeIntent>(onInvoke: (_) => onGoHome?.call()),
      SettingsIntent: CallbackAction<SettingsIntent>(onInvoke: (_) => onSettings?.call()),
      CloseIntent: CallbackAction<CloseIntent>(onInvoke: (_) => onClose?.call()),
    };
  }
}

/// Helper widget to show shortcuts dialog
class ShortcutsDialog extends StatelessWidget {
  const ShortcutsDialog({super.key});

  static Future<void> show(BuildContext context) {
    return showDialog<void>(
      context: context,
      builder: (_) => const ShortcutsDialog(),
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return AlertDialog(
      title: const Text('Keyboard Shortcuts'),
      content: SizedBox(
        width: 500,
        child: ListView.builder(
          shrinkWrap: true,
          itemCount: AppShortcuts.byCategory.length,
          itemBuilder: (context, index) {
            final category = AppShortcuts.byCategory.keys.elementAt(index);
            final shortcuts = AppShortcuts.byCategory[category]!;

            return Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Padding(
                  padding: const EdgeInsets.only(top: 16, bottom: 8),
                  child: Text(
                    category,
                    style: theme.textTheme.titleSmall?.copyWith(
                      color: theme.colorScheme.primary,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
                ...shortcuts.map(
                  (shortcut) => Padding(
                    padding: const EdgeInsets.symmetric(vertical: 4),
                    child: Row(
                      children: [
                        Expanded(
                          child: Text(shortcut.description),
                        ),
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 8,
                            vertical: 4,
                          ),
                          decoration: BoxDecoration(
                            color: theme.colorScheme.surfaceContainerHighest,
                            borderRadius: BorderRadius.circular(4),
                          ),
                          child: Text(
                            shortcut.formattedKeys,
                            style: theme.textTheme.labelSmall,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            );
          },
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('Close'),
        ),
      ],
    );
  }
}
