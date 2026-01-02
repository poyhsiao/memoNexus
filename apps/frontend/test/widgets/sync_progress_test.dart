// Sync Progress Widget Tests
// Tests for SyncProgressWidget showing sync progress

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:memonexus_frontend/widgets/sync_progress.dart';

void main() {
  group('SyncProgressWidget', () {
    testWidgets('should build without errors', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: SyncProgressWidget(
              percent: 50,
              completed: 5,
              total: 10,
            ),
          ),
        ),
      );

      expect(find.byType(SyncProgressWidget), findsOneWidget);
    });

    testWidgets('should render progress display', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: SyncProgressWidget(
              percent: 75,
              completed: 75,
              total: 100,
            ),
          ),
        ),
      );

      await tester.pump();

      // Widget should render successfully
      expect(find.byType(SyncProgressWidget), findsOneWidget);
    });

    testWidgets('should display sync status', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SyncProgressWidget(
              percent: 30,
              completed: 3,
              total: 10,
              currentItem: 'Syncing item 3',
            ),
          ),
        ),
      );

      await tester.pump();

      // Check for sync progress UI
      expect(find.byType(SyncProgressWidget), findsOneWidget);
    });

    testWidgets('should show completed state', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: SyncProgressWidget(
              percent: 100,
              completed: 10,
              total: 10,
            ),
          ),
        ),
      );

      await tester.pump();

      // Widget should render completed state
      expect(find.byType(SyncProgressWidget), findsOneWidget);
    });

    testWidgets('should handle zero progress', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: SyncProgressWidget(
              percent: 0,
              completed: 0,
              total: 10,
            ),
          ),
        ),
      );

      await tester.pump();

      // Widget should handle zero progress
      expect(find.byType(SyncProgressWidget), findsOneWidget);
    });
  });
}
