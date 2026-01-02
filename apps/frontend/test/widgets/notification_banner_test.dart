import 'package:flutter/material.dart' hide Notification;
import 'package:flutter_test/flutter_test.dart';
import 'package:memonexus_frontend/widgets/notification_banner.dart';

void main() {
  group('NotificationSeverity', () {
    test('should have all required severity levels', () {
      expect(NotificationSeverity.values, contains(NotificationSeverity.info));
      expect(NotificationSeverity.values, contains(NotificationSeverity.success));
      expect(NotificationSeverity.values, contains(NotificationSeverity.warning));
      expect(NotificationSeverity.values, contains(NotificationSeverity.error));
    });
  });

  group('Notification', () {
    test('should create notification with required fields', () {
      final notification = Notification(
        id: 'test-id',
        title: 'Test Title',
      );

      expect(notification.id, 'test-id');
      expect(notification.title, 'Test Title');
      expect(notification.message, isNull);
      expect(notification.severity, NotificationSeverity.info);
      expect(notification.onRetry, isNull);
      expect(notification.duration, const Duration(seconds: 5));
    });

    test('should create notification with all fields', () {
      final onRetry = () {};
      final notification = Notification(
        id: 'test-id',
        title: 'Test Title',
        message: 'Test message',
        severity: NotificationSeverity.error,
        onRetry: onRetry,
        duration: const Duration(seconds: 10),
      );

      expect(notification.id, 'test-id');
      expect(notification.title, 'Test Title');
      expect(notification.message, 'Test message');
      expect(notification.severity, NotificationSeverity.error);
      expect(notification.onRetry, onRetry);
      expect(notification.duration, const Duration(seconds: 10));
    });

    test('should support zero duration (sticky notification)', () {
      final notification = Notification(
        id: 'test-id',
        title: 'Sticky',
        duration: Duration.zero,
      );

      expect(notification.duration, Duration.zero);
    });
  });

  group('NotificationBanner', () {
    testWidgets('should display notification title', (tester) async {
      final notification = Notification(
        id: 'test-id',
        title: 'Test Notification',
        duration: Duration.zero,
      );

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: NotificationBanner(
              notification: notification,
            ),
          ),
        ),
      );

      // Wait for slide-in animation
      await tester.pump(const Duration(milliseconds: 350));

      expect(find.text('Test Notification'), findsOneWidget);
    });

    testWidgets('should display notification message', (tester) async {
      final notification = Notification(
        id: 'test-id',
        title: 'Title',
        message: 'This is a test message',
        duration: Duration.zero,
      );

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: NotificationBanner(
              notification: notification,
            ),
          ),
        ),
      );

      await tester.pump(const Duration(milliseconds: 350));

      expect(find.text('This is a test message'), findsOneWidget);
    });

    testWidgets('should display dismiss button', (tester) async {
      final notification = Notification(
        id: 'test-id',
        title: 'Title',
        duration: Duration.zero,
      );

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: NotificationBanner(
              notification: notification,
            ),
          ),
        ),
      );

      await tester.pump(const Duration(milliseconds: 350));

      expect(find.text('Dismiss'), findsOneWidget);
    });

    testWidgets('should display retry button when callback provided', (tester) async {
      var retryCalled = false;
      final notification = Notification(
        id: 'test-id',
        title: 'Title',
        onRetry: () => retryCalled = true,
        duration: Duration.zero,
      );

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: NotificationBanner(
              notification: notification,
            ),
          ),
        ),
      );

      await tester.pump(const Duration(milliseconds: 350));

      expect(find.text('Retry'), findsOneWidget);

      // Tap retry button
      await tester.tap(find.text('Retry'));
      await tester.pump();
      expect(retryCalled, isTrue);
    });

    testWidgets('should not display retry button without callback', (tester) async {
      final notification = Notification(
        id: 'test-id',
        title: 'Title',
        duration: Duration.zero,
      );

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: NotificationBanner(
              notification: notification,
            ),
          ),
        ),
      );

      await tester.pump(const Duration(milliseconds: 350));

      expect(find.text('Retry'), findsNothing);
    });

    testWidgets('should call onDismiss when dismissed', (tester) async {
      var dismissCalled = false;
      final notification = Notification(
        id: 'test-id',
        title: 'Title',
        duration: Duration.zero,
      );

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: NotificationBanner(
              notification: notification,
              onDismiss: () => dismissCalled = true,
            ),
          ),
        ),
      );

      // Wait for slide-in animation to complete
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 400));

      // Tap dismiss button
      await tester.tap(find.text('Dismiss'));

      // Wait for slide-out animation to complete
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 400));

      expect(dismissCalled, isTrue);
    });

    testWidgets('should use appropriate colors for each severity', (tester) async {
      for (final severity in NotificationSeverity.values) {
        final notification = Notification(
          id: 'test-id',
          title: 'Title',
          severity: severity,
          duration: Duration.zero,
        );

        await tester.pumpWidget(
          MaterialApp(
            home: Scaffold(
              body: NotificationBanner(
                notification: notification,
              ),
            ),
          ),
        );

        await tester.pump(const Duration(milliseconds: 350));

        expect(find.byType(MaterialBanner), findsOneWidget);
      }
    });
  });

  group('NotificationHelper', () {
    testWidgets('showInfo should display info notification', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: NotificationManager(
            child: Scaffold(
              body: Builder(
                builder: (context) => ElevatedButton(
                  onPressed: () => NotificationManager.show(
                    context,
                    title: 'Info Title',
                    message: 'Info message',
                    duration: Duration.zero,
                  ),
                  child: const Text('Show'),
                ),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Show'));
      // Wait for slide-in animation to complete
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 400));

      expect(find.text('Info Title'), findsOneWidget);
      expect(find.text('Info message'), findsOneWidget);
    });

    testWidgets('showSuccess should display success notification', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: NotificationManager(
            child: Scaffold(
              body: Builder(
                builder: (context) => ElevatedButton(
                  onPressed: () => NotificationManager.show(
                    context,
                    title: 'Success Title',
                    message: 'Success message',
                    severity: NotificationSeverity.success,
                    duration: Duration.zero,
                  ),
                  child: const Text('Show'),
                ),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Show'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 400));

      expect(find.text('Success Title'), findsOneWidget);
    });

    testWidgets('showWarning should display warning notification', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: NotificationManager(
            child: Scaffold(
              body: Builder(
                builder: (context) => ElevatedButton(
                  onPressed: () => NotificationManager.show(
                    context,
                    title: 'Warning Title',
                    message: 'Warning message',
                    severity: NotificationSeverity.warning,
                    duration: Duration.zero,
                  ),
                  child: const Text('Show'),
                ),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Show'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 400));

      expect(find.text('Warning Title'), findsOneWidget);
    });

    testWidgets('showError should display error notification', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: NotificationManager(
            child: Scaffold(
              body: Builder(
                builder: (context) => ElevatedButton(
                  onPressed: () => NotificationManager.show(
                    context,
                    title: 'Error Title',
                    message: 'Error message',
                    severity: NotificationSeverity.error,
                    duration: Duration.zero,
                  ),
                  child: const Text('Show'),
                ),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Show'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 400));

      expect(find.text('Error Title'), findsOneWidget);
    });

    testWidgets('showRetryable should include retry button', (tester) async {
      var retryCalled = false;

      await tester.pumpWidget(
        MaterialApp(
          home: NotificationManager(
            child: Scaffold(
              body: Builder(
                builder: (context) => ElevatedButton(
                  onPressed: () => NotificationHelper.showRetryable(
                    context,
                    'Retry Title',
                    () => retryCalled = true,
                    message: 'Retry message',
                    duration: Duration.zero,
                  ),
                  child: const Text('Show'),
                ),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Show'));
      // Wait for slide-in animation to complete
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 400));

      expect(find.text('Retry Title'), findsOneWidget);
      expect(find.text('Retry'), findsOneWidget);

      await tester.tap(find.text('Retry'));
      await tester.pump();
      expect(retryCalled, isTrue);
    });

    testWidgets('should stack multiple notifications', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: NotificationManager(
            child: Scaffold(
              body: Builder(
                builder: (context) => Column(
                  children: [
                    ElevatedButton(
                      onPressed: () => NotificationManager.show(
                        context,
                        title: 'First',
                        duration: Duration.zero,
                      ),
                      child: const Text('Show 1'),
                    ),
                    ElevatedButton(
                      onPressed: () => NotificationManager.show(
                        context,
                        title: 'Second',
                        severity: NotificationSeverity.success,
                        duration: Duration.zero,
                      ),
                      child: const Text('Show 2'),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Show 1'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 400));

      await tester.tap(find.text('Show 2'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 400));

      expect(find.text('First'), findsOneWidget);
      expect(find.text('Second'), findsOneWidget);
    });
  });
}
