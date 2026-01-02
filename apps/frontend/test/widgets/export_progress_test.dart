// Export Progress Widget Tests - T197: User Story 5
// Tests for ExportProgressWidget showing export stages
//
// ★ Insight ─────────────────────────────────────
// 1. Provider overrides allow testing different states
//    without needing actual API calls or async operations.
// 2. Testing conditional rendering requires verifying
//    widget presence/absence based on state changes.
// ─────────────────────────────────────────────────

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:memonexus_frontend/providers/export_provider.dart';
import 'package:memonexus_frontend/widgets/export_progress.dart';

void main() {
  group('ExportProgressWidget - Idle State', () {
    testWidgets('should show nothing when idle and not completed',
        (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            exportProvider.overrideWith((ref) {
              final notifier = ExportNotifier();
              notifier.state = const ExportState(status: ExportStatus.idle);
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: ExportProgressWidget(),
            ),
          ),
        ),
      );

      // Widget should be hidden (SizedBox.shrink)
      expect(find.byType(ExportProgressWidget), findsOneWidget);
      expect(find.byType(Card), findsNothing);
    });

    testWidgets('should show nothing when failed', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            exportProvider.overrideWith((ref) {
              final notifier = ExportNotifier();
              notifier.state = const ExportState(
                status: ExportStatus.failed,
                errorMessage: 'Test error',
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: ExportProgressWidget(),
            ),
          ),
        ),
      );

      // Widget should be hidden when failed
      expect(find.byType(Card), findsNothing);
    });
  });

  group('ExportProgressWidget - Running State', () {
    testWidgets('should show progress when running', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            exportProvider.overrideWith((ref) {
              final notifier = ExportNotifier();
              notifier.state = const ExportState(
                status: ExportStatus.preparing,
                currentStage: 'Preparing export',
                totalItems: 100,
                processedItems: 10,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: ExportProgressWidget(),
            ),
          ),
        ),
      );

      expect(find.byType(Card), findsOneWidget);
      expect(find.text('Preparing export'), findsOneWidget);
      expect(find.text('10%'), findsOneWidget);
      expect(find.byType(LinearProgressIndicator), findsOneWidget);
      expect(find.byType(CircularProgressIndicator), findsOneWidget);
    });

    testWidgets('should show encrypting stage', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            exportProvider.overrideWith((ref) {
              final notifier = ExportNotifier();
              notifier.state = const ExportState(
                status: ExportStatus.encrypting,
                currentStage: 'Encrypting data',
                currentFile: 'database.sqlite',
                totalItems: 100,
                processedItems: 30,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: ExportProgressWidget(),
            ),
          ),
        ),
      );

      expect(find.text('Encrypting data'), findsOneWidget);
      expect(find.text('database.sqlite'), findsOneWidget);
      expect(find.text('30%'), findsOneWidget);
    });

    testWidgets('should show compressing stage', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            exportProvider.overrideWith((ref) {
              final notifier = ExportNotifier();
              notifier.state = const ExportState(
                status: ExportStatus.compressing,
                currentStage: 'Compressing',
                currentFile: 'archive.tar.gz',
                totalItems: 100,
                processedItems: 50,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: ExportProgressWidget(),
            ),
          ),
        ),
      );

      expect(find.text('Compressing'), findsOneWidget);
      expect(find.text('archive.tar.gz'), findsOneWidget);
      expect(find.text('50%'), findsOneWidget);
    });

    testWidgets('should show uploading stage', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            exportProvider.overrideWith((ref) {
              final notifier = ExportNotifier();
              notifier.state = const ExportState(
                status: ExportStatus.uploading,
                currentStage: 'Uploading',
                totalItems: 100,
                processedItems: 80,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: ExportProgressWidget(),
            ),
          ),
        ),
      );

      expect(find.text('Uploading'), findsOneWidget);
      expect(find.text('80%'), findsOneWidget);
    });

    testWidgets('should show items count when totalItems > 0',
        (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            exportProvider.overrideWith((ref) {
              final notifier = ExportNotifier();
              notifier.state = const ExportState(
                status: ExportStatus.preparing,
                currentStage: 'Processing',
                totalItems: 100,
                processedItems: 45,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: ExportProgressWidget(),
            ),
          ),
        ),
      );

      expect(find.text('45 / 100 items'), findsOneWidget);
    });

    testWidgets('should show cancel button when running', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            exportProvider.overrideWith((ref) {
              final notifier = ExportNotifier();
              notifier.state = const ExportState(
                status: ExportStatus.preparing,
                totalItems: 100,
                processedItems: 10,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: ExportProgressWidget(),
            ),
          ),
        ),
      );

      expect(find.text('Cancel'), findsOneWidget);
      expect(find.byIcon(Icons.cancel), findsOneWidget);
    });

    testWidgets('should cancel export when cancel button pressed',
        (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            exportProvider.overrideWith((ref) {
              final notifier = ExportNotifier();
              notifier.state = const ExportState(
                status: ExportStatus.preparing,
                totalItems: 100,
                processedItems: 10,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: ExportProgressWidget(),
            ),
          ),
        ),
      );

      // Verify cancel button exists before tap
      expect(find.text('Cancel'), findsOneWidget);

      await tester.tap(find.text('Cancel'));
      await tester.pump();

      // After cancel, widget should be hidden (state becomes idle)
      expect(find.byType(Card), findsNothing);
    });
  });

  group('ExportProgressWidget - Completed State', () {
    testWidgets('should show completion message', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            exportProvider.overrideWith((ref) {
              final notifier = ExportNotifier();
              notifier.state = const ExportState(
                status: ExportStatus.completed,
                totalItems: 100,
                processedItems: 100,
                filePath: '/exports/backup.tar.gz',
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: ExportProgressWidget(),
            ),
          ),
        ),
      );

      expect(find.text('Export completed!'), findsOneWidget);
      expect(find.byIcon(Icons.check_circle), findsOneWidget);
      expect(find.text('/exports/backup.tar.gz'), findsOneWidget);
      expect(find.text('100%'), findsOneWidget);
    });

    testWidgets('should not show cancel button when completed',
        (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            exportProvider.overrideWith((ref) {
              final notifier = ExportNotifier();
              notifier.state = const ExportState(
                status: ExportStatus.completed,
                progress: 100,
                totalItems: 100,
                processedItems: 100,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: ExportProgressWidget(),
            ),
          ),
        ),
      );

      expect(find.text('Cancel'), findsNothing);
    });

    testWidgets('should show file path when completed', (tester) async {
      const testPath = '/exports/memonexus_20250102.tar.gz';

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            exportProvider.overrideWith((ref) {
              final notifier = ExportNotifier();
              notifier.state = const ExportState(
                status: ExportStatus.completed,
                filePath: testPath,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: ExportProgressWidget(),
            ),
          ),
        ),
      );

      expect(find.text(testPath), findsOneWidget);
    });
  });

  group('ExportProgressWidget - Edge Cases', () {
    testWidgets('should handle zero totalItems', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            exportProvider.overrideWith((ref) {
              final notifier = ExportNotifier();
              notifier.state = const ExportState(
                status: ExportStatus.preparing,
                totalItems: 0,
                processedItems: 0,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: ExportProgressWidget(),
            ),
          ),
        ),
      );

      // Should not show items count when totalItems is 0
      expect(find.textContaining('/'), findsNothing);
    });

    testWidgets('should handle empty currentFile', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            exportProvider.overrideWith((ref) {
              final notifier = ExportNotifier();
              notifier.state = const ExportState(
                status: ExportStatus.preparing,
                currentStage: 'Preparing',
                currentFile: '',
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: ExportProgressWidget(),
            ),
          ),
        ),
      );

      // Should show stage but not file name
      expect(find.text('Preparing'), findsOneWidget);
    });

    testWidgets('should handle null filePath when completed',
        (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            exportProvider.overrideWith((ref) {
              final notifier = ExportNotifier();
              notifier.state = const ExportState(
                status: ExportStatus.completed,
                filePath: null,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: ExportProgressWidget(),
            ),
          ),
        ),
      );

      // Should show completion but no file path
      expect(find.text('Export completed!'), findsOneWidget);
    });
  });
}
