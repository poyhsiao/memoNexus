import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:memonexus_frontend/providers/capture_provider.dart';
import 'package:memonexus_frontend/screens/capture_screen.dart';
import 'package:memonexus_frontend/services/api_client.dart';
import 'package:memonexus_frontend/widgets/file_upload.dart';

// =====================================================
// Fake API Client
// =====================================================

class _FakeCaptureAPIClient extends MemoNexusAPIClient {
  bool shouldFail = false;
  String? errorMessage;

  @override
  Future<Map<String, dynamic>> createContentFromURL({
    required String sourceUrl,
    String? title,
    List<String>? tags,
  }) async {
    if (shouldFail) {
      throw APIException(
        code: 'CREATE_ERROR',
        message: errorMessage ?? 'Failed to create',
        statusCode: 500,
      );
    }

    final now = DateTime.now().millisecondsSinceEpoch ~/ 1000;
    return {
      'id': 'test-id',
      'title': title ?? sourceUrl,
      'content_text': 'Test content',
      'media_type': 'web',
      'tags': tags?.join(',') ?? '',
      'created_at': now,
      'updated_at': now,
      'version': 1,
      'is_deleted': 0,
    };
  }

  @override
  Future<List<Map<String, dynamic>>> listTags() async => [];
}

void main() {
  group('CaptureScreen Widget Tests', () {
    late _FakeCaptureAPIClient fakeApi;

    setUp(() {
      fakeApi = _FakeCaptureAPIClient();
    });

    Widget buildTestWidget() {
      return ProviderScope(
        overrides: [
          captureProvider.overrideWith((ref) => CaptureNotifier(fakeApi)),
          availableTagsProvider.overrideWith((ref) => []),
        ],
        child: const MaterialApp(
          home: CaptureScreen(),
        ),
      );
    }

    testWidgets('should render capture form', (tester) async {
      await tester.pumpWidget(buildTestWidget());

      // "Add Content" appears in both AppBar and submit button
      expect(find.text('Add Content'), findsWidgets);
      expect(find.text('Content Type'), findsOneWidget);
      expect(find.byType(SegmentedButton<CaptureType>), findsOneWidget);
    });

    testWidgets('should show URL input when URL type selected',
        (tester) async {
      await tester.pumpWidget(buildTestWidget());

      // Verify URL input is shown by default (URL is default type)
      // "URL" appears in both SegmentedButton and label
      expect(find.text('URL'), findsWidgets);
      // Multiple TextFields exist (URL input and custom title input)
      expect(find.byType(TextField), findsWidgets);
    });

    testWidgets('should show title input section', (tester) async {
      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Title (Optional)'), findsOneWidget);
      expect(find.text('Custom title'), findsOneWidget);
    });

    testWidgets('should show tags section', (tester) async {
      await tester.pumpWidget(buildTestWidget());

      expect(find.text('Tags (Optional)'), findsOneWidget);
      expect(find.text('Select Tags'), findsOneWidget);
    });

    testWidgets('should show add content button', (tester) async {
      await tester.pumpWidget(buildTestWidget());

      // "Add Content" appears in both AppBar and submit button
      expect(find.text('Add Content'), findsWidgets);
      expect(find.byIcon(Icons.add), findsWidgets);
    });

    testWidgets('should submit when valid URL entered', (tester) async {
      await tester.pumpWidget(buildTestWidget());

      // Enter valid URL - find URL TextField by its hint text
      final urlTextField = find.widgetWithText(TextField, 'https://example.com/article');
      expect(urlTextField, findsOneWidget);
      await tester.enterText(urlTextField, 'https://example.com');
      await tester.pumpAndSettle();

      // Get the provider and call submit directly
      final container = ProviderScope.containerOf(
        tester.element(find.byType(CaptureScreen)),
      );
      final notifier = container.read(captureProvider.notifier);
      await notifier.submit();
      await tester.pumpAndSettle();

      // Verify state was updated - check if createdItem exists in state
      final state = container.read(captureProvider);
      expect(state.createdItem, isNotNull);
    });

    testWidgets('should disable submit when URL is invalid', (tester) async {
      await tester.pumpWidget(buildTestWidget());

      // Verify initial state - URL is empty so submit should be disabled
      final container = ProviderScope.containerOf(
        tester.element(find.byType(CaptureScreen)),
      );
      final state = container.read(captureProvider);

      // Empty URL should be invalid
      expect(state.isValid, isFalse);
    });

    testWidgets('should show error message on API failure', (tester) async {
      fakeApi.shouldFail = true;
      fakeApi.errorMessage = 'Network error';

      await tester.pumpWidget(buildTestWidget());

      final urlTextField = find.widgetWithText(TextField, 'https://example.com/article');
      await tester.enterText(urlTextField, 'https://example.com');
      await tester.pumpAndSettle();

      // Get the provider and call submit directly
      final container = ProviderScope.containerOf(
        tester.element(find.byType(CaptureScreen)),
      );
      final notifier = container.read(captureProvider.notifier);
      await notifier.submit();
      await tester.pumpAndSettle();

      final state = container.read(captureProvider);
      expect(state.error, isNotNull);
      expect(state.error, contains('Network error'));
    });

    testWidgets('should show FileUploadWidget when File type selected',
        (tester) async {
      await tester.pumpWidget(buildTestWidget());

      // Find the File button in SegmentedButton - it's a Text widget with "File"
      final fileButton = find.text('File');
      expect(fileButton, findsOneWidget);

      await tester.tap(fileButton);
      await tester.pumpAndSettle();

      // Verify FileUploadWidget is shown when file type is selected
      expect(find.byType(FileUploadWidget), findsOneWidget);
    });
  });
}
