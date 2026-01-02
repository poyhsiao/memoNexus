import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:memonexus_frontend/providers/capture_provider.dart';
import 'package:memonexus_frontend/widgets/tag_picker.dart';

void main() {
  group('TagPickerWidget Widget', () {
    testWidgets('should show selected tags when provided', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            availableTagsProvider.overrideWith((ref) => const ['tag1', 'tag2', 'tag3']),
          ],
          child: MaterialApp(
            home: Scaffold(
              body: TagPickerWidget(
                selectedTags: const ['tag1'],
                onTagsChanged: (_) {},
              ),
            ),
          ),
        ),
      );

      expect(find.text('Select Tags'), findsOneWidget);
      expect(find.text('tag1'), findsWidgets);
      expect(find.byType(Chip), findsOneWidget);
    });

    testWidgets('should show available tags when not empty', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            availableTagsProvider.overrideWith((ref) => ['tag1', 'tag2']),
          ],
          child: MaterialApp(
            home: Scaffold(
              body: TagPickerWidget(
                selectedTags: const [],
                onTagsChanged: (_) {},
              ),
            ),
          ),
        ),
      );

      expect(find.text('Select Tags'), findsOneWidget);
      expect(find.byType(FilterChip), findsNWidgets(2));
    });

    testWidgets('should show empty state when no available tags', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            availableTagsProvider.overrideWith((ref) => const []),
          ],
          child: MaterialApp(
            home: Scaffold(
              body: TagPickerWidget(
                selectedTags: const [],
                onTagsChanged: (_) {},
              ),
            ),
          ),
        ),
      );

      expect(find.text('No tags created yet'), findsOneWidget);
    });

    testWidgets('should call onTagsChanged when tag is selected',
        (tester) async {
      List<String> selectedTags = [];

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            availableTagsProvider.overrideWith((ref) => ['tag1', 'tag2']),
          ],
          child: MaterialApp(
            home: Scaffold(
              body: TagPickerWidget(
                selectedTags: const [],
                onTagsChanged: (tags) => selectedTags = tags,
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('tag1'));
      await tester.pump();

      expect(selectedTags, ['tag1']);
    });

    testWidgets('should remove tag when chip delete is pressed', (tester) async {
      List<String> selectedTags = ['tag1', 'tag2'];

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            availableTagsProvider.overrideWith((ref) => ['tag1', 'tag2', 'tag3']),
          ],
          child: MaterialApp(
            home: Scaffold(
              body: TagPickerWidget(
                selectedTags: selectedTags,
                onTagsChanged: (tags) => selectedTags = tags,
              ),
            ),
          ),
        ),
      );

      // Find the first chip and tap its delete button
      final chip = tester.widget<Chip>(find.byType(Chip).first);
      chip.onDeleted?.call();
      await tester.pump();

      expect(selectedTags, ['tag2']);
    });

    testWidgets('should deselect tag when filter chip is tapped again',
        (tester) async {
      List<String> selectedTags = ['tag1'];

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            availableTagsProvider.overrideWith((ref) => ['tag1', 'tag2']),
          ],
          child: MaterialApp(
            home: Scaffold(
              body: TagPickerWidget(
                selectedTags: selectedTags,
                onTagsChanged: (tags) => selectedTags = tags,
              ),
            ),
          ),
        ),
      );

      // Tap the selected tag to deselect it
      await tester.tap(find.byType(FilterChip).first);
      await tester.pump();

      expect(selectedTags, isEmpty);
    });

    testWidgets('should show create button initially', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            availableTagsProvider.overrideWith((ref) => const []),
          ],
          child: MaterialApp(
            home: Scaffold(
              body: TagPickerWidget(
                selectedTags: const [],
                onTagsChanged: (_) {},
              ),
            ),
          ),
        ),
      );

      expect(find.text('Create new tag'), findsOneWidget);
      expect(find.byIcon(Icons.add), findsOneWidget);
    });

    testWidgets('should show input field when create button tapped',
        (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            availableTagsProvider.overrideWith((ref) => const []),
          ],
          child: MaterialApp(
            home: Scaffold(
              body: TagPickerWidget(
                selectedTags: const [],
                onTagsChanged: (_) {},
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Create new tag'));
      await tester.pump();

      expect(find.text('Tag name'), findsOneWidget);
      expect(find.byType(TextField), findsOneWidget);
      expect(find.byIcon(Icons.check), findsOneWidget);
      expect(find.byIcon(Icons.close), findsOneWidget);
    });

    testWidgets('should create new tag when submitted via text field',
        (tester) async {
      List<String> selectedTags = [];

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            availableTagsProvider.overrideWith((ref) => const []),
          ],
          child: MaterialApp(
            home: Scaffold(
              body: TagPickerWidget(
                selectedTags: const [],
                onTagsChanged: (tags) => selectedTags = tags,
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Create new tag'));
      await tester.pump();

      await tester.enterText(find.byType(TextField), 'newtag');
      await tester.testTextInput.receiveAction(TextInputAction.done);
      await tester.pumpAndSettle();

      expect(selectedTags, ['newtag']);
      // The input field is closed, verify create button is back
      expect(find.text('Create new tag'), findsOneWidget);
    });

    testWidgets('should create new tag when check button pressed',
        (tester) async {
      List<String> selectedTags = [];

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            availableTagsProvider.overrideWith((ref) => const []),
          ],
          child: MaterialApp(
            home: Scaffold(
              body: TagPickerWidget(
                selectedTags: const [],
                onTagsChanged: (tags) => selectedTags = tags,
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Create new tag'));
      await tester.pump();

      await tester.enterText(find.byType(TextField), 'mytag');
      await tester.tap(find.byIcon(Icons.check));
      await tester.pump();

      expect(selectedTags, ['mytag']);
    });

    testWidgets('should cancel create when close button pressed', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            availableTagsProvider.overrideWith((ref) => const []),
          ],
          child: MaterialApp(
            home: Scaffold(
              body: TagPickerWidget(
                selectedTags: const [],
                onTagsChanged: (_) {},
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Create new tag'));
      await tester.pump();

      await tester.tap(find.byIcon(Icons.close));
      await tester.pump();

      // Create button should be back
      expect(find.text('Create new tag'), findsOneWidget);
      expect(find.byType(TextField), findsNothing);
    });

    testWidgets('should not create empty tag', (tester) async {
      List<String> selectedTags = [];

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            availableTagsProvider.overrideWith((ref) => const []),
          ],
          child: MaterialApp(
            home: Scaffold(
              body: TagPickerWidget(
                selectedTags: const [],
                onTagsChanged: (tags) => selectedTags = tags,
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Create new tag'));
      await tester.pump();

      await tester.enterText(find.byType(TextField), '   ');
      await tester.tap(find.byIcon(Icons.check));
      await tester.pump();

      expect(selectedTags, isEmpty);
    });

    testWidgets('should trim whitespace from new tag', (tester) async {
      List<String> selectedTags = [];

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            availableTagsProvider.overrideWith((ref) => const []),
          ],
          child: MaterialApp(
            home: Scaffold(
              body: TagPickerWidget(
                selectedTags: const [],
                onTagsChanged: (tags) => selectedTags = tags,
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Create new tag'));
      await tester.pump();

      await tester.enterText(find.byType(TextField), '  mytag  ');
      await tester.tap(find.byIcon(Icons.check));
      await tester.pump();

      expect(selectedTags, ['mytag']);
    });

    testWidgets('should not add duplicate tag', (tester) async {
      List<String> selectedTags = ['existing'];

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            availableTagsProvider.overrideWith((ref) => const []),
          ],
          child: MaterialApp(
            home: Scaffold(
              body: TagPickerWidget(
                selectedTags: selectedTags,
                onTagsChanged: (tags) => selectedTags = tags,
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Create new tag'));
      await tester.pump();

      await tester.enterText(find.byType(TextField), 'existing');
      await tester.tap(find.byIcon(Icons.check));
      await tester.pump();

      // Should still have only one instance
      expect(selectedTags, ['existing']);
    });
  });

  group('TagPickerDialog Widget', () {
    testWidgets('should display dialog with TagPickerWidget', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            availableTagsProvider.overrideWith((ref) => const []),
          ],
          child: MaterialApp(
            home: Scaffold(
              body: ElevatedButton(
                onPressed: () {
                  showDialog(
                    context: tester.element(find.byType(ElevatedButton)),
                    builder: (context) => const TagPickerDialog(),
                  );
                },
                child: const Text('Show Dialog'),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Show Dialog'));
      await tester.pumpAndSettle();

      // Check dialog is displayed
      expect(find.byType(AlertDialog), findsOneWidget);
      // "Select Tags" appears in both dialog title and widget
      expect(find.text('Select Tags'), findsWidgets);
      expect(find.text('Cancel'), findsOneWidget);
      expect(find.text('Done'), findsOneWidget);
    });

    testWidgets('should initialize with initialTags', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            availableTagsProvider.overrideWith((ref) => const []),
          ],
          child: MaterialApp(
            home: Scaffold(
              body: ElevatedButton(
                onPressed: () {
                  showDialog(
                    context: tester.element(find.byType(ElevatedButton)),
                    builder: (context) => const TagPickerDialog(
                      initialTags: ['tag1'],
                    ),
                  );
                },
                child: const Text('Show Dialog'),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Show Dialog'));
      await tester.pumpAndSettle();

      // tag1 appears in the selected tags chip
      expect(find.byType(Chip), findsOneWidget);
    });

    testWidgets('should pop with null when Cancel tapped', (tester) async {
      String? result;

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            availableTagsProvider.overrideWith((ref) => const []),
          ],
          child: MaterialApp(
            home: Scaffold(
              body: ElevatedButton(
                onPressed: () async {
                  final context = tester.element(find.byType(ElevatedButton));
                  result = await showDialog<String>(
                    context: context,
                    builder: (context) => const TagPickerDialog(),
                  );
                },
                child: const Text('Show Dialog'),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Show Dialog'));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();

      expect(result, isNull);
    });

    testWidgets('should pop with selected tags when Done tapped',
        (tester) async {
      List<String>? result;

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            availableTagsProvider.overrideWith((ref) => ['tag1', 'tag2']),
          ],
          child: MaterialApp(
            home: Scaffold(
              body: ElevatedButton(
                onPressed: () async {
                  final context = tester.element(find.byType(ElevatedButton));
                  final tags = await showDialog<List<String>>(
                    context: context,
                    builder: (context) => const TagPickerDialog(
                      initialTags: ['tag1'],
                    ),
                  );
                  result = tags;
                },
                child: const Text('Show Dialog'),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Show Dialog'));
      await tester.pumpAndSettle();

      // Add another tag
      await tester.tap(find.byType(FilterChip).at(1));
      await tester.pump();

      await tester.tap(find.text('Done'));
      await tester.pumpAndSettle();

      expect(result, ['tag1', 'tag2']);
    });
  });
}
