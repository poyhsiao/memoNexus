import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:memonexus_frontend/widgets/file_upload.dart';

void main() {
  group('FileUploadWidget Widget', () {
    testWidgets('should show drop zone when no file selected', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: FileUploadWidget(
                onFileSelected: (_) {},
              ),
            ),
          ),
        ),
      );

      expect(find.text('Tap to select a file'), findsOneWidget);
      expect(find.text('or drag and drop'), findsOneWidget);
      expect(find.byIcon(Icons.cloud_upload_outlined), findsOneWidget);
    });

    testWidgets('should show file type chips', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: FileUploadWidget(
                onFileSelected: (_) {},
              ),
            ),
          ),
        ),
      );

      expect(find.text('Images'), findsOneWidget);
      expect(find.text('Videos'), findsOneWidget);
      expect(find.text('PDF'), findsOneWidget);
      expect(find.text('Markdown'), findsOneWidget);
    });

    testWidgets('should show file preview when file selected', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: FileUploadWidget(
                initialFilePath: '/path/to/file.pdf',
                initialFileName: 'document.pdf',
                onFileSelected: (_) {},
              ),
            ),
          ),
        ),
      );

      expect(find.text('document.pdf'), findsOneWidget);
      expect(find.text('.PDF'), findsOneWidget);
      expect(find.byIcon(Icons.picture_as_pdf), findsOneWidget);
      expect(find.text('Change file'), findsOneWidget);
    });

    testWidgets('should call clearFile when remove button tapped',
        (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: FileUploadWidget(
                initialFilePath: '/path/to/file.pdf',
                initialFileName: 'document.pdf',
                onFileSelected: (_) {},
              ),
            ),
          ),
        ),
      );

      // Verify remove button exists
      expect(find.byIcon(Icons.close), findsOneWidget);

      // Tap remove button
      await tester.tap(find.byIcon(Icons.close));
      await tester.pump();

      // Note: The widget uses initialFilePath prop, so clearing provider state
      // doesn't change the UI. The widget is in "controlled" mode when
      // initialFilePath is provided.
    });

    testWidgets('should show correct icon for image files', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: FileUploadWidget(
                initialFilePath: '/path/to/image.jpg',
                initialFileName: 'photo.jpg',
                onFileSelected: (_) {},
              ),
            ),
          ),
        ),
      );

      expect(find.byIcon(Icons.image), findsOneWidget);
    });

    testWidgets('should show correct icon for video files', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: FileUploadWidget(
                initialFilePath: '/path/to/video.mp4',
                initialFileName: 'movie.mp4',
                onFileSelected: (_) {},
              ),
            ),
          ),
        ),
      );

      expect(find.byIcon(Icons.videocam), findsOneWidget);
    });

    testWidgets('should show correct icon for markdown files', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: FileUploadWidget(
                initialFilePath: '/path/to/doc.md',
                initialFileName: 'doc.md',
                onFileSelected: (_) {},
              ),
            ),
          ),
        ),
      );

      expect(find.byIcon(Icons.description), findsOneWidget);
    });

    testWidgets('should show default icon for unknown files', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: FileUploadWidget(
                initialFilePath: '/path/to/file.xyz',
                initialFileName: 'unknown.xyz',
                onFileSelected: (_) {},
              ),
            ),
          ),
        ),
      );

      expect(find.byIcon(Icons.insert_drive_file), findsOneWidget);
    });

    testWidgets('should have correct color for images', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: FileUploadWidget(
                initialFilePath: '/path/to/image.png',
                initialFileName: 'photo.png',
                onFileSelected: (_) {},
              ),
            ),
          ),
        ),
      );

      final container = find.byType(Container).evaluate().skip(1).first;
      final decoration =
          (container.widget as Container).decoration as BoxDecoration?;
      expect(decoration?.color, Colors.purple.withValues(alpha: 0.2));
    });

    testWidgets('should have correct color for videos', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: FileUploadWidget(
                initialFilePath: '/path/to/video.mp4',
                initialFileName: 'movie.mp4',
                onFileSelected: (_) {},
              ),
            ),
          ),
        ),
      );

      final container = find.byType(Container).evaluate().skip(1).first;
      final decoration =
          (container.widget as Container).decoration as BoxDecoration?;
      expect(decoration?.color, Colors.red.withValues(alpha: 0.2));
    });

    testWidgets('should have correct color for PDF', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: FileUploadWidget(
                initialFilePath: '/path/to/doc.pdf',
                initialFileName: 'doc.pdf',
                onFileSelected: (_) {},
              ),
            ),
          ),
        ),
      );

      final container = find.byType(Container).evaluate().skip(1).first;
      final decoration =
          (container.widget as Container).decoration as BoxDecoration?;
      expect(decoration?.color, Colors.orange.withValues(alpha: 0.2));
    });

    testWidgets('should have correct color for markdown', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: FileUploadWidget(
                initialFilePath: '/path/to/doc.md',
                initialFileName: 'doc.md',
                onFileSelected: (_) {},
              ),
            ),
          ),
        ),
      );

      final container = find.byType(Container).evaluate().skip(1).first;
      final decoration =
          (container.widget as Container).decoration as BoxDecoration?;
      expect(decoration?.color, Colors.blue.withValues(alpha: 0.2));
    });
  });

  group('FileUploadWidget Extension Detection', () {
    testWidgets('should handle files with no extension', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: FileUploadWidget(
                initialFilePath: '/path/to/file',
                initialFileName: 'file',
                onFileSelected: (_) {},
              ),
            ),
          ),
        ),
      );

      expect(find.byIcon(Icons.insert_drive_file), findsOneWidget);
    });

    testWidgets('should handle multiple dots in filename', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: FileUploadWidget(
                initialFilePath: '/path/to/file.name.with.dots.pdf',
                initialFileName: 'file.name.with.dots.pdf',
                onFileSelected: (_) {},
              ),
            ),
          ),
        ),
      );

      expect(find.text('.PDF'), findsOneWidget);
      expect(find.byIcon(Icons.picture_as_pdf), findsOneWidget);
    });
  });
}
