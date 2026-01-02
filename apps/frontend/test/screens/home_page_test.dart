import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:memonexus_frontend/main.dart';

void main() {
  group('HomePage Widget Tests', () {
    testWidgets('should render app title and description', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: HomePage(),
      ));

      // "MemoNexus" appears in both AppBar and body
      expect(find.text('MemoNexus'), findsWidgets);
      expect(find.text('Local-First Personal Knowledge Base'), findsOneWidget);
    });

    testWidgets('should show phase indicator', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: HomePage(),
      ));

      expect(find.text('Phase 8: Polish & Testing'), findsOneWidget);
      expect(find.textContaining('Performance monitoring'), findsOneWidget);
    });

    testWidgets('should display library icon', (tester) async {
      await tester.pumpWidget(const MaterialApp(
        home: HomePage(),
      ));

      expect(find.byIcon(Icons.library_books), findsOneWidget);
    });
  });
}
