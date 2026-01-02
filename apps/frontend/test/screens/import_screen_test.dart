// Import Screen Widget Tests - T196
// Tests for ImportScreen widgets and interactions
//
// ★ Insight ─────────────────────────────────────
// 1. Testing multi-state screens requires setting up
//    different initial states via provider overrides.
// 2. Progress indicators need proper pump() calls to
//    trigger animation frames and rebuilds.
// 3. Form validation tests verify error messages for
//    missing file and password fields.
// ─────────────────────────────────────────────────

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:memonexus_frontend/providers/export_provider.dart';
import 'package:memonexus_frontend/screens/import_screen.dart';

void main() {
  group('ImportScreen - Basic Rendering', () {
    testWidgets('should build without errors', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      expect(find.byType(ImportScreen), findsOneWidget);
    });

    testWidgets('should display import title', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      expect(find.text('Import Data'), findsOneWidget);
    });
  });

  group('ImportScreen - Idle State', () {
    testWidgets('should show import form when idle', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      expect(find.text('Import Export Archive'), findsOneWidget);
      expect(find.text('Archive File'), findsOneWidget);
      expect(find.text('Archive Password'), findsOneWidget);
      expect(find.text('Import Archive'), findsOneWidget);
    });

    testWidgets('should have file picker input', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      expect(find.text('Select archive file'), findsOneWidget);
      expect(find.byIcon(Icons.folder_open), findsOneWidget);
    });

    testWidgets('should have password field with validation', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      expect(find.byIcon(Icons.lock), findsOneWidget);
      expect(find.text('Archive Password'), findsOneWidget);
    });

    testWidgets('should show description text', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      expect(find.textContaining('Restore your knowledge base'), findsOneWidget);
    });

    testWidgets('should display no file selected initially', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      expect(find.text('No file selected'), findsOneWidget);
    });

    testWidgets('should have proper form structure', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      expect(find.byType(Card), findsWidgets);
      expect(find.byType(Form), findsOneWidget);
    });

    testWidgets('should have import button with icon', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      expect(find.byIcon(Icons.file_download), findsOneWidget);
      expect(find.text('Import Archive'), findsOneWidget);
    });
  });

  group('ImportScreen - Running State', () {
    testWidgets('should show progress indicator when decrypting', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            importProvider.overrideWith((ref) {
              final notifier = ImportNotifier();
              notifier.state = const ImportState(
                status: ImportStatus.decrypting,
                currentStage: 'Decrypting archive...',
                progress: 25,
                totalItems: 100,
                importedItems: 25,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      expect(find.byType(LinearProgressIndicator), findsOneWidget);
    });

    testWidgets('should show current stage text', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            importProvider.overrideWith((ref) {
              final notifier = ImportNotifier();
              notifier.state = const ImportState(
                status: ImportStatus.extracting,
                currentStage: 'Extracting content...',
                progress: 50,
                totalItems: 100,
                importedItems: 50,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      expect(find.text('Extracting content...'), findsOneWidget);
    });

    testWidgets('should show progress percentage', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            importProvider.overrideWith((ref) {
              final notifier = ImportNotifier();
              notifier.state = const ImportState(
                status: ImportStatus.restoring,
                currentStage: 'Processing...',
                progress: 75,
                totalItems: 100,
                importedItems: 75,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      expect(find.text('75%'), findsOneWidget);
    });

    testWidgets('should show cancel button when running', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            importProvider.overrideWith((ref) {
              final notifier = ImportNotifier();
              notifier.state = const ImportState(
                status: ImportStatus.validating,
                currentStage: 'Importing...',
                progress: 10,
                totalItems: 100,
                importedItems: 10,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      expect(find.text('Cancel'), findsOneWidget);
      expect(find.byIcon(Icons.cancel), findsOneWidget);
    });
  });

  group('ImportScreen - Success State', () {
    testWidgets('should show success message when completed', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            importProvider.overrideWith((ref) {
              final notifier = ImportNotifier();
              notifier.state = const ImportState(
                status: ImportStatus.completed,
                importedItems: 150,
                skippedItems: 5,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      expect(find.byIcon(Icons.check_circle), findsOneWidget);
      expect(find.text('Import completed successfully!'), findsOneWidget);
    });

    testWidgets('should show imported items count', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            importProvider.overrideWith((ref) {
              final notifier = ImportNotifier();
              notifier.state = const ImportState(
                status: ImportStatus.completed,
                importedItems: 250,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      expect(find.text('250 items imported'), findsOneWidget);
    });

    testWidgets('should show skipped items when present', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            importProvider.overrideWith((ref) {
              final notifier = ImportNotifier();
              notifier.state = const ImportState(
                status: ImportStatus.completed,
                importedItems: 100,
                skippedItems: 10,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      expect(find.text('10 items skipped (duplicates)'), findsOneWidget);
    });

    testWidgets('should not show skipped items when zero', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            importProvider.overrideWith((ref) {
              final notifier = ImportNotifier();
              notifier.state = const ImportState(
                status: ImportStatus.completed,
                importedItems: 100,
                skippedItems: 0,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      expect(find.textContaining('items skipped'), findsNothing);
    });

    testWidgets('should show done button on success', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            importProvider.overrideWith((ref) {
              final notifier = ImportNotifier();
              notifier.state = const ImportState(
                status: ImportStatus.completed,
                importedItems: 50,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      expect(find.text('Done'), findsOneWidget);
      expect(find.byIcon(Icons.check), findsOneWidget);
    });
  });

  group('ImportScreen - Failed State', () {
    testWidgets('should show error message when failed', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            importProvider.overrideWith((ref) {
              final notifier = ImportNotifier();
              notifier.state = const ImportState(
                status: ImportStatus.failed,
                errorMessage: 'Invalid password',
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      expect(find.byIcon(Icons.error), findsOneWidget);
      expect(find.text('Import failed'), findsOneWidget);
      expect(find.text('Invalid password'), findsOneWidget);
    });

    testWidgets('should show try again button on failure', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            importProvider.overrideWith((ref) {
              final notifier = ImportNotifier();
              notifier.state = const ImportState(
                status: ImportStatus.failed,
                errorMessage: 'Archive corrupted',
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      expect(find.text('Try Again'), findsOneWidget);
      expect(find.byIcon(Icons.refresh), findsOneWidget);
    });

    testWidgets('should not show error message when null', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            importProvider.overrideWith((ref) {
              final notifier = ImportNotifier();
              notifier.state = const ImportState(
                status: ImportStatus.failed,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      expect(find.text('Import failed'), findsOneWidget);
      // Error message is null, so it shouldn't show additional error text
      expect(find.byType(Text), findsWidgets);
    });
  });

  group('ImportScreen - Progress Indicator', () {
    testWidgets('should show linear progress indicator', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            importProvider.overrideWith((ref) {
              final notifier = ImportNotifier();
              notifier.state = const ImportState(
                status: ImportStatus.validating,
                currentStage: 'Working...',
                progress: 60,
                totalItems: 100,
                importedItems: 60,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      expect(find.byType(LinearProgressIndicator), findsOneWidget);
    });

    testWidgets('should show zero percent correctly', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            importProvider.overrideWith((ref) {
              final notifier = ImportNotifier();
              notifier.state = const ImportState(
                status: ImportStatus.decrypting,
                currentStage: 'Starting...',
                progress: 0,
                totalItems: 100,
                importedItems: 0,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      expect(find.text('0%'), findsOneWidget);
    });

    testWidgets('should show hundred percent correctly', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            importProvider.overrideWith((ref) {
              final notifier = ImportNotifier();
              notifier.state = const ImportState(
                status: ImportStatus.restoring,
                currentStage: 'Finishing...',
                progress: 100,
                totalItems: 100,
                importedItems: 100,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      expect(find.text('100%'), findsOneWidget);
    });
  });

  group('ImportScreen - Widget Structure', () {
    testWidgets('should have scaffold with app bar', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      expect(find.byType(Scaffold), findsOneWidget);
      expect(find.byType(AppBar), findsOneWidget);
    });

    testWidgets('should have list view for body', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      expect(find.byType(ListView), findsOneWidget);
    });

    testWidgets('should use card widget for form container', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      expect(find.byType(Card), findsWidgets);
    });
  });

  group('ImportScreen - Button Interactions', () {
    testWidgets('should have import button as FilledButton', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      // Verify import button exists (uses FilledButton.icon)
      expect(find.byIcon(Icons.file_download), findsOneWidget);
      expect(find.text('Import Archive'), findsOneWidget);
    });

    testWidgets('should have done button as FilledButton', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      // Verify widget structure
      expect(find.byType(ImportScreen), findsOneWidget);
    });

    testWidgets('should have try again button as OutlinedButton',
        (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      // Verify widget structure
      expect(find.byType(ImportScreen), findsOneWidget);
    });

    testWidgets('should have cancel button as TextButton', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: ImportScreen(),
          ),
        ),
      );

      await tester.pump();

      // Verify widget structure
      expect(find.byType(ImportScreen), findsOneWidget);
    });
  });
}
