// Sync Status Widget Tests - T170
// Tests for SyncStatusWidget and all state displays
//
// ★ Insight ─────────────────────────────────────
// 1. Provider override pattern allows testing different
//    sync states (loading, failed, syncing, idle).
// 2. Private widgets (_LoadingIndicator, _FailedStatus,
//    _SyncStatusCard, _StatusIcon) are tested through
//    the main widget's state changes.
// 3. Time formatting tests cover all branches of
//    _formatLastSync() from "Just now" to date format.
// ─────────────────────────────────────────────────

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:memonexus_frontend/providers/sync_provider.dart';
import 'package:memonexus_frontend/widgets/sync_status.dart';

void main() {
  group('SyncStatusWidget - Loading State', () {
    testWidgets('should show loading indicator when isLoading', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            syncStatusProvider.overrideWith((ref) {
              final notifier = FakeSyncStatusNotifier();
              notifier.state = const SyncStatusState(isLoading: true);
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: SyncStatusWidget(),
            ),
          ),
        ),
      );

      await tester.pump();

      expect(find.text('Loading sync status...'), findsOneWidget);
      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      expect(find.byType(Container), findsWidgets);
    });

    testWidgets('should have grey background when loading', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            syncStatusProvider.overrideWith((ref) {
              final notifier = FakeSyncStatusNotifier();
              notifier.state = const SyncStatusState(isLoading: true);
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: SyncStatusWidget(),
            ),
          ),
        ),
      );

      await tester.pump();

      final container = tester.widget<Container>(
        find.byType(Container).first,
      );
      final decoration = container.decoration as BoxDecoration;
      expect(decoration.color, Colors.grey.shade100);
    });
  });

  group('SyncStatusWidget - Failed State', () {
    testWidgets('should show error message when failed', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            syncStatusProvider.overrideWith((ref) {
              final notifier = FakeSyncStatusNotifier();
              notifier.state = const SyncStatusState(
                status: 'failed',
                error: 'Network error',
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: SyncStatusWidget(),
            ),
          ),
        ),
      );

      await tester.pump();

      expect(find.text('Sync failed'), findsOneWidget);
      expect(find.text('Network error'), findsOneWidget);
      expect(find.byIcon(Icons.error_outline), findsOneWidget);
    });

    testWidgets('should show retry button when failed', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            syncStatusProvider.overrideWith((ref) {
              final notifier = FakeSyncStatusNotifier();
              notifier.state = const SyncStatusState(
                status: 'failed',
                error: 'Connection lost',
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: SyncStatusWidget(),
            ),
          ),
        ),
      );

      await tester.pump();

      // TextButton.icon creates a TextButton, so we check for the icon and text
      expect(find.text('Retry'), findsOneWidget);
      expect(find.byIcon(Icons.refresh), findsOneWidget);
      // The button exists as an icon button
      expect(find.byType(Icon), findsWidgets);
    });

    testWidgets('should have red background when failed', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            syncStatusProvider.overrideWith((ref) {
              final notifier = FakeSyncStatusNotifier();
              notifier.state = const SyncStatusState(
                status: 'failed',
                error: 'Error',
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: SyncStatusWidget(),
            ),
          ),
        ),
      );

      await tester.pump();

      final container = tester.widget<Container>(
        find.byType(Container).first,
      );
      final decoration = container.decoration as BoxDecoration;
      expect(decoration.color, Colors.red.shade50);
    });

    testWidgets('should use default error message when error is null',
        (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            syncStatusProvider.overrideWith((ref) {
              final notifier = FakeSyncStatusNotifier();
              notifier.state = const SyncStatusState(status: 'failed');
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: SyncStatusWidget(),
            ),
          ),
        ),
      );

      await tester.pump();

      // When error is null, widget shows "Sync failed" twice (title + message)
      expect(find.text('Sync failed'), findsNWidgets(2));
    });
  });

  group('SyncStatusWidget - Syncing State', () {
    testWidgets('should show syncing status when isSyncing', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            syncStatusProvider.overrideWith((ref) {
              final notifier = FakeSyncStatusNotifier();
              notifier.state = const SyncStatusState(status: 'syncing');
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: SyncStatusWidget(),
            ),
          ),
        ),
      );

      await tester.pump();

      expect(find.text('Syncing...'), findsOneWidget);
      expect(find.byType(CircularProgressIndicator), findsOneWidget);
    });

    testWidgets('should have blue background when syncing', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            syncStatusProvider.overrideWith((ref) {
              final notifier = FakeSyncStatusNotifier();
              notifier.state = const SyncStatusState(status: 'syncing');
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: SyncStatusWidget(),
            ),
          ),
        ),
      );

      await tester.pump();

      final container = tester.widget<Container>(
        find.byType(Container).first,
      );
      final decoration = container.decoration as BoxDecoration;
      expect(decoration.color, Colors.blue.shade50);
    });

    testWidgets('should show last sync time when syncing', (tester) async {
      final lastSync = DateTime.now().subtract(const Duration(minutes: 5));
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            syncStatusProvider.overrideWith((ref) {
              final notifier = FakeSyncStatusNotifier();
              notifier.state = SyncStatusState(
                status: 'syncing',
                lastSync: lastSync.millisecondsSinceEpoch,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: SyncStatusWidget(),
            ),
          ),
        ),
      );

      await tester.pump();

      expect(find.textContaining('Last sync:'), findsOneWidget);
    });
  });

  group('SyncStatusWidget - Idle State', () {
    testWidgets('should show sync enabled when idle', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            syncStatusProvider.overrideWith((ref) {
              final notifier = FakeSyncStatusNotifier();
              notifier.state = const SyncStatusState(status: 'idle');
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: SyncStatusWidget(),
            ),
          ),
        ),
      );

      await tester.pump();

      expect(find.text('Sync enabled'), findsOneWidget);
      expect(find.byIcon(Icons.cloud_done), findsOneWidget);
    });

    testWidgets('should have green background when idle', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            syncStatusProvider.overrideWith((ref) {
              final notifier = FakeSyncStatusNotifier();
              notifier.state = const SyncStatusState(status: 'idle');
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: SyncStatusWidget(),
            ),
          ),
        ),
      );

      await tester.pump();

      final container = tester.widget<Container>(
        find.byType(Container).first,
      );
      final decoration = container.decoration as BoxDecoration;
      expect(decoration.color, Colors.green.shade50);
    });

    testWidgets('should show refresh button when idle', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            syncStatusProvider.overrideWith((ref) {
              final notifier = FakeSyncStatusNotifier();
              notifier.state = const SyncStatusState(status: 'idle');
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: SyncStatusWidget(),
            ),
          ),
        ),
      );

      await tester.pump();

      expect(find.byIcon(Icons.cloud_sync), findsOneWidget);
      expect(find.byType(IconButton), findsOneWidget);
    });

    testWidgets('should show pending changes when > 0', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            syncStatusProvider.overrideWith((ref) {
              final notifier = FakeSyncStatusNotifier();
              notifier.state = const SyncStatusState(
                status: 'idle',
                pendingChanges: 5,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: SyncStatusWidget(),
            ),
          ),
        ),
      );

      await tester.pump();

      expect(find.text('5 pending change(s)'), findsOneWidget);
    });

    testWidgets('should not show pending changes when 0', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            syncStatusProvider.overrideWith((ref) {
              final notifier = FakeSyncStatusNotifier();
              notifier.state = const SyncStatusState(
                status: 'idle',
                pendingChanges: 0,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: SyncStatusWidget(),
            ),
          ),
        ),
      );

      await tester.pump();

      expect(find.textContaining('pending change'), findsNothing);
    });
  });

  group('SyncStatusWidget - Last Sync Formatting', () {
    testWidgets('should show Never when lastSync is null', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            syncStatusProvider.overrideWith((ref) {
              final notifier = FakeSyncStatusNotifier();
              notifier.state = const SyncStatusState(status: 'idle');
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: SyncStatusWidget(),
            ),
          ),
        ),
      );

      await tester.pump();

      expect(find.text('Last sync: Never'), findsOneWidget);
    });

    testWidgets('should show "Just now" for < 60 seconds', (tester) async {
      final lastSync = DateTime.now().subtract(const Duration(seconds: 30));
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            syncStatusProvider.overrideWith((ref) {
              final notifier = FakeSyncStatusNotifier();
              notifier.state = SyncStatusState(
                status: 'idle',
                lastSync: lastSync.millisecondsSinceEpoch,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: SyncStatusWidget(),
            ),
          ),
        ),
      );

      await tester.pump();

      expect(find.text('Last sync: Just now'), findsOneWidget);
    });

    testWidgets('should show "Xm ago" for < 60 minutes', (tester) async {
      final lastSync = DateTime.now().subtract(const Duration(minutes: 5));
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            syncStatusProvider.overrideWith((ref) {
              final notifier = FakeSyncStatusNotifier();
              notifier.state = SyncStatusState(
                status: 'idle',
                lastSync: lastSync.millisecondsSinceEpoch,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: SyncStatusWidget(),
            ),
          ),
        ),
      );

      await tester.pump();

      expect(find.text('Last sync: 5m ago'), findsOneWidget);
    });

    testWidgets('should show "Xh ago" for < 24 hours', (tester) async {
      final lastSync = DateTime.now().subtract(const Duration(hours: 3));
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            syncStatusProvider.overrideWith((ref) {
              final notifier = FakeSyncStatusNotifier();
              notifier.state = SyncStatusState(
                status: 'idle',
                lastSync: lastSync.millisecondsSinceEpoch,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: SyncStatusWidget(),
            ),
          ),
        ),
      );

      await tester.pump();

      expect(find.text('Last sync: 3h ago'), findsOneWidget);
    });

    testWidgets('should show "Xd ago" for < 7 days', (tester) async {
      final lastSync = DateTime.now().subtract(const Duration(days: 2));
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            syncStatusProvider.overrideWith((ref) {
              final notifier = FakeSyncStatusNotifier();
              notifier.state = SyncStatusState(
                status: 'idle',
                lastSync: lastSync.millisecondsSinceEpoch,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: SyncStatusWidget(),
            ),
          ),
        ),
      );

      await tester.pump();

      expect(find.text('Last sync: 2d ago'), findsOneWidget);
    });

    testWidgets('should show date format for >= 7 days', (tester) async {
      final lastSync = DateTime.now().subtract(const Duration(days: 10));
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            syncStatusProvider.overrideWith((ref) {
              final notifier = FakeSyncStatusNotifier();
              notifier.state = SyncStatusState(
                status: 'idle',
                lastSync: lastSync.millisecondsSinceEpoch,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: SyncStatusWidget(),
            ),
          ),
        ),
      );

      await tester.pump();

      // Format: "MMM D, HH:MM" e.g., "Dec 25, 14:30"
      expect(find.textContaining('Last sync:'), findsOneWidget);
      // The date should contain comma and colon separators
      expect(find.textContaining(','), findsOneWidget);
      expect(find.textContaining(':'), findsOneWidget);
    });
  });

  group('SyncStatusWidget - Status Icons', () {
    testWidgets('should show CircularProgressIndicator when syncing',
        (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            syncStatusProvider.overrideWith((ref) {
              final notifier = FakeSyncStatusNotifier();
              notifier.state = const SyncStatusState(status: 'syncing');
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: SyncStatusWidget(),
            ),
          ),
        ),
      );

      await tester.pump();

      expect(find.byType(CircularProgressIndicator), findsWidgets);
    });

    testWidgets('should show cloud_done icon when idle', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            syncStatusProvider.overrideWith((ref) {
              final notifier = FakeSyncStatusNotifier();
              notifier.state = const SyncStatusState(status: 'idle');
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: SyncStatusWidget(),
            ),
          ),
        ),
      );

      await tester.pump();

      expect(find.byIcon(Icons.cloud_done), findsOneWidget);
    });

    testWidgets('should show sync icon for non-syncing non-idle state',
        (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            syncStatusProvider.overrideWith((ref) {
              final notifier = FakeSyncStatusNotifier();
              // Create a state that's neither syncing nor idle
              notifier.state = const SyncStatusState(status: 'pending');
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: SyncStatusWidget(),
            ),
          ),
        ),
      );

      await tester.pump();

      expect(find.byIcon(Icons.sync), findsOneWidget);
    });

    testWidgets('should show error icon when failed', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            syncStatusProvider.overrideWith((ref) {
              final notifier = FakeSyncStatusNotifier();
              notifier.state = const SyncStatusState(
                status: 'failed',
                error: 'Error',
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: SyncStatusWidget(),
            ),
          ),
        ),
      );

      await tester.pump();

      expect(find.byIcon(Icons.error_outline), findsOneWidget);
    });
  });

  group('SyncStatusWidget - Basic Structure', () {
    testWidgets('should build without errors', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: SyncStatusWidget(),
            ),
          ),
        ),
      );

      expect(find.byType(SyncStatusWidget), findsOneWidget);
    });

    testWidgets('should render in Scaffold', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: SyncStatusWidget(),
            ),
          ),
        ),
      );

      await tester.pumpAndSettle();

      expect(find.byType(SyncStatusWidget), findsOneWidget);
    });

    testWidgets('should render successfully', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: SyncStatusWidget(),
            ),
          ),
        ),
      );

      await tester.pumpAndSettle();

      expect(find.byType(SyncStatusWidget), findsOneWidget);
    });
  });
}

// =====================================================
// Fake Sync Status Notifier for testing
// =====================================================

class FakeSyncStatusNotifier extends StateNotifier<SyncStatusState>
    implements SyncStatusNotifier {
  FakeSyncStatusNotifier() : super(const SyncStatusState());

  @override
  Future<void> loadStatus() async {
    // Fake implementation - no-op for tests
  }

  @override
  Future<bool> triggerSync() async {
    // Fake implementation - returns true for tests
    return true;
  }

  @override
  void clearError() {
    // Fake implementation - no-op for tests
  }
}
