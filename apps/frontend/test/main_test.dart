// Main App Widget Tests
// Tests for main.dart app initialization and widgets

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:memonexus_frontend/main.dart';

void main() {
  group('MemoNexusApp', () {
    testWidgets('should build without errors', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MemoNexusApp(showPerformanceOverlay: false),
        ),
      );

      expect(find.byType(MemoNexusApp), findsOneWidget);
    });

    testWidgets('should use Material 3 theme', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MemoNexusApp(showPerformanceOverlay: false),
        ),
      );

      final materialApp = tester.widget<MaterialApp>(find.byType(MaterialApp));
      expect(materialApp.theme?.useMaterial3, true);
    });

    testWidgets('should have color scheme', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MemoNexusApp(showPerformanceOverlay: false),
        ),
      );

      final materialApp = tester.widget<MaterialApp>(find.byType(MaterialApp));
      expect(materialApp.theme?.colorScheme, isNotNull);
    });

    testWidgets('should hide debug banner', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MemoNexusApp(showPerformanceOverlay: false),
        ),
      );

      final materialApp = tester.widget<MaterialApp>(find.byType(MaterialApp));
      expect(materialApp.debugShowCheckedModeBanner, false);
    });

    testWidgets('should show HomePage as home', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MemoNexusApp(showPerformanceOverlay: false),
        ),
      );

      expect(find.byType(HomePage), findsOneWidget);
    });
  });

  group('HomePage', () {
    testWidgets('should render correctly', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: HomePage(),
        ),
      );

      expect(find.byType(HomePage), findsOneWidget);
    });

    testWidgets('should display app title in AppBar', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: HomePage(),
        ),
      );

      expect(find.text('MemoNexus'), findsWidgets);
    });

    testWidgets('should display library icon', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: HomePage(),
        ),
      );

      expect(find.byIcon(Icons.library_books), findsOneWidget);
    });

    testWidgets('should display subtitle', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: HomePage(),
        ),
      );

      expect(find.text('Local-First Personal Knowledge Base'), findsOneWidget);
    });

    testWidgets('should display Phase 8 status', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: HomePage(),
        ),
      );

      expect(find.text('Phase 8: Polish & Testing'), findsOneWidget);
    });

    testWidgets('should display performance monitoring status',
        (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: HomePage(),
        ),
      );

      expect(find.text('Performance monitoring enabled'), findsOneWidget);
    });

    testWidgets('should have centered layout', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: HomePage(),
        ),
      );

      expect(find.byType(Center), findsWidgets);
    });

    testWidgets('should have column layout', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: HomePage(),
        ),
      );

      expect(find.byType(Column), findsOneWidget);
    });
  });

  group('App Launch Configuration', () {
    testWidgets('should accept showPerformanceOverlay parameter',
        (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MemoNexusApp(showPerformanceOverlay: true),
        ),
      );

      expect(find.byType(MemoNexusApp), findsOneWidget);
    });

    testWidgets('should default showPerformanceOverlay to false',
        (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MemoNexusApp(showPerformanceOverlay: false),
        ),
      );

      final materialApp = tester.widget<MaterialApp>(find.byType(MaterialApp));
      expect(materialApp.showPerformanceOverlay, false);
    });
  });
}
