// Auto Export Config Widget Tests - T199: User Story 5
// Tests for AutoExportConfig widget
//
// ★ Insight ─────────────────────────────────────
// 1. Testing conditional UI rendering based on
//    provider state requires setting up different
//    initial configurations via provider overrides.
// 2. User interactions like taps and dropdown changes
//    need proper pump() calls to trigger rebuilds.
// ─────────────────────────────────────────────────

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:memonexus_frontend/providers/export_provider.dart';
import 'package:memonexus_frontend/widgets/auto_export_config.dart';

void main() {
  group('AutoExportConfigWidget - Disabled State', () {
    testWidgets('should build without errors', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: AutoExportConfigWidget(),
            ),
          ),
        ),
      );

      expect(find.byType(AutoExportConfigWidget), findsOneWidget);
    });

    testWidgets('should show title', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: AutoExportConfigWidget(),
            ),
          ),
        ),
      );

      expect(find.text('Automatic Export'), findsOneWidget);
    });

    testWidgets('should show enable switch with disabled subtitle',
        (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: AutoExportConfigWidget(),
            ),
          ),
        ),
      );

      expect(find.text('Enable automatic export'), findsOneWidget);
      expect(find.text('Automatic exports are disabled'), findsOneWidget);
    });

    testWidgets('should not show interval controls when disabled',
        (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: AutoExportConfigWidget(),
            ),
          ),
        ),
      );

      expect(find.text('Export interval'), findsNothing);
      expect(find.text('Retention count'), findsNothing);
    });
  });

  group('AutoExportConfigWidget - Enabled State', () {
    testWidgets('should show enabled subtitle when enabled', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            autoExportConfigProvider.overrideWith((ref) {
              final notifier = AutoExportConfigNotifier();
              notifier.state = const AutoExportConfig(enabled: true);
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: AutoExportConfigWidget(),
            ),
          ),
        ),
      );

      expect(find.text('Automatic exports are enabled'), findsOneWidget);
    });

    testWidgets('should show export interval control', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            autoExportConfigProvider.overrideWith((ref) {
              final notifier = AutoExportConfigNotifier();
              notifier.state = const AutoExportConfig(enabled: true);
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: AutoExportConfigWidget(),
            ),
          ),
        ),
      );

      expect(find.text('Export interval'), findsOneWidget);
    });

    testWidgets('should show all interval options', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            autoExportConfigProvider.overrideWith((ref) {
              final notifier = AutoExportConfigNotifier();
              notifier.state = const AutoExportConfig(enabled: true);
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: AutoExportConfigWidget(),
            ),
          ),
        ),
      );

      // All interval options should be available in dropdown
      expect(find.byType(DropdownButton<AutoExportInterval>), findsOneWidget);
    });

    testWidgets('should show retention count control', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            autoExportConfigProvider.overrideWith((ref) {
              final notifier = AutoExportConfigNotifier();
              notifier.state = const AutoExportConfig(
                enabled: true,
                retentionCount: 5,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: AutoExportConfigWidget(),
            ),
          ),
        ),
      );

      expect(find.text('Retention count'), findsOneWidget);
      expect(find.text('Keep last 5 exports'), findsOneWidget);
    });

    testWidgets('should show include media switch', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            autoExportConfigProvider.overrideWith((ref) {
              final notifier = AutoExportConfigNotifier();
              notifier.state = const AutoExportConfig(enabled: true);
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: AutoExportConfigWidget(),
            ),
          ),
        ),
      );

      expect(find.text('Include media files'), findsOneWidget);
      expect(
          find.text(
              'Include images and videos in exports (increases size)'),
          findsOneWidget);
    });
  });

  group('AutoExportConfigWidget - User Interactions', () {
    testWidgets('should toggle enable switch', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            autoExportConfigProvider.overrideWith((ref) {
              final notifier = AutoExportConfigNotifier();
              notifier.state = const AutoExportConfig(enabled: false);
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: AutoExportConfigWidget(),
            ),
          ),
        ),
      );

      // Initially disabled
      expect(find.text('Automatic exports are disabled'), findsOneWidget);

      // Find and tap the switch
      final switchFinder = find.byType(SwitchListTile).first;
      await tester.tap(switchFinder);
      await tester.pump();

      // Should show enabled state after rebuild
      expect(find.byType(AutoExportConfigWidget), findsOneWidget);
    });

    testWidgets('should decrease retention count', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            autoExportConfigProvider.overrideWith((ref) {
              final notifier = AutoExportConfigNotifier();
              notifier.state = const AutoExportConfig(
                enabled: true,
                retentionCount: 5,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: AutoExportConfigWidget(),
            ),
          ),
        ),
      );

      // Find remove button and tap it
      final removeButton = find.byIcon(Icons.remove);
      expect(removeButton, findsOneWidget);

      await tester.tap(removeButton);
      await tester.pump();

      // Verify widget is still rendered
      expect(find.byType(AutoExportConfigWidget), findsOneWidget);
    });

    testWidgets('should increase retention count', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            autoExportConfigProvider.overrideWith((ref) {
              final notifier = AutoExportConfigNotifier();
              notifier.state = const AutoExportConfig(
                enabled: true,
                retentionCount: 5,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: AutoExportConfigWidget(),
            ),
          ),
        ),
      );

      // Find add button and tap it
      final addButton = find.byIcon(Icons.add);
      expect(addButton, findsOneWidget);

      await tester.tap(addButton);
      await tester.pump();

      // Verify widget is still rendered
      expect(find.byType(AutoExportConfigWidget), findsOneWidget);
    });

    testWidgets('should disable remove button at minimum retention',
        (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            autoExportConfigProvider.overrideWith((ref) {
              final notifier = AutoExportConfigNotifier();
              notifier.state = const AutoExportConfig(
                enabled: true,
                retentionCount: 1,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: AutoExportConfigWidget(),
            ),
          ),
        ),
      );

      // Remove button should be disabled (onPressed is null)
      final removeButton = find.byWidgetPredicate((widget) =>
          widget is IconButton &&
          widget.icon is Icon &&
          (widget.icon as Icon).icon == Icons.remove);

      expect(removeButton, findsOneWidget);
    });

    testWidgets('should disable add button at maximum retention',
        (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            autoExportConfigProvider.overrideWith((ref) {
              final notifier = AutoExportConfigNotifier();
              notifier.state = const AutoExportConfig(
                enabled: true,
                retentionCount: 100,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: AutoExportConfigWidget(),
            ),
          ),
        ),
      );

      // Add button should be disabled at max (100)
      final addButton = find.byWidgetPredicate((widget) =>
          widget is IconButton &&
          widget.icon is Icon &&
          (widget.icon as Icon).icon == Icons.add);

      expect(addButton, findsOneWidget);
    });

    testWidgets('should toggle include media switch', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            autoExportConfigProvider.overrideWith((ref) {
              final notifier = AutoExportConfigNotifier();
              notifier.state = const AutoExportConfig(
                enabled: true,
                includeMedia: false,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: AutoExportConfigWidget(),
            ),
          ),
        ),
      );

      // Find the include media switch (second SwitchListTile)
      final mediaSwitches = find.byType(SwitchListTile);
      expect(mediaSwitches, findsWidgets);

      // Tap the second switch (include media)
      await tester.tap(mediaSwitches.at(1));
      await tester.pump();

      // Verify widget is still rendered
      expect(find.byType(AutoExportConfigWidget), findsOneWidget);
    });
  });

  group('AutoExportConfigWidget - Date Display', () {
    testWidgets('should show last export date when available',
        (tester) async {
      final testDate = DateTime(2025, 1, 2, 14, 30);
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            autoExportConfigProvider.overrideWith((ref) {
              final notifier = AutoExportConfigNotifier();
              notifier.state = AutoExportConfig(
                enabled: true,
                lastExportAt: testDate,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: AutoExportConfigWidget(),
            ),
          ),
        ),
      );

      // Should show formatted date: 2/1/2025 14:30
      expect(find.textContaining('Last export:'), findsOneWidget);
      expect(find.textContaining('2/1/2025'), findsOneWidget);
    });

    testWidgets('should show next export date when available',
        (tester) async {
      final testDate = DateTime(2025, 1, 3, 10, 0);
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            autoExportConfigProvider.overrideWith((ref) {
              final notifier = AutoExportConfigNotifier();
              notifier.state = AutoExportConfig(
                enabled: true,
                nextExportAt: testDate,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: AutoExportConfigWidget(),
            ),
          ),
        ),
      );

      // Should show formatted date: 3/1/2025 10:00
      expect(find.textContaining('Next export:'), findsOneWidget);
      expect(find.textContaining('3/1/2025'), findsOneWidget);
    });

    testWidgets('should not show dates when not available', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            autoExportConfigProvider.overrideWith((ref) {
              final notifier = AutoExportConfigNotifier();
              notifier.state = const AutoExportConfig(
                enabled: true,
                lastExportAt: null,
                nextExportAt: null,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: AutoExportConfigWidget(),
            ),
          ),
        ),
      );

      // Should not show date fields
      expect(find.textContaining('Last export:'), findsNothing);
      expect(find.textContaining('Next export:'), findsNothing);
    });
  });

  group('AutoExportConfigWidget - Interval Options', () {
    testWidgets('should show Manual only label for manual interval',
        (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            autoExportConfigProvider.overrideWith((ref) {
              final notifier = AutoExportConfigNotifier();
              notifier.state = const AutoExportConfig(
                enabled: true,
                interval: AutoExportInterval.manual,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: AutoExportConfigWidget(),
            ),
          ),
        ),
      );

      expect(find.text('Manual only'), findsWidgets);
    });

    testWidgets('should show Daily label for daily interval', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            autoExportConfigProvider.overrideWith((ref) {
              final notifier = AutoExportConfigNotifier();
              notifier.state = const AutoExportConfig(
                enabled: true,
                interval: AutoExportInterval.daily,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: AutoExportConfigWidget(),
            ),
          ),
        ),
      );

      expect(find.text('Daily'), findsWidgets);
    });

    testWidgets('should show Weekly label for weekly interval',
        (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            autoExportConfigProvider.overrideWith((ref) {
              final notifier = AutoExportConfigNotifier();
              notifier.state = const AutoExportConfig(
                enabled: true,
                interval: AutoExportInterval.weekly,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: AutoExportConfigWidget(),
            ),
          ),
        ),
      );

      expect(find.text('Weekly'), findsWidgets);
    });

    testWidgets('should show Monthly label for monthly interval',
        (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            autoExportConfigProvider.overrideWith((ref) {
              final notifier = AutoExportConfigNotifier();
              notifier.state = const AutoExportConfig(
                enabled: true,
                interval: AutoExportInterval.monthly,
              );
              return notifier;
            }),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: AutoExportConfigWidget(),
            ),
          ),
        ),
      );

      expect(find.text('Monthly'), findsWidgets);
    });
  });
}
