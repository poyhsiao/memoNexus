// Notification Banner - T215: Non-blocking notification system
// Dismissable, retry button for temporary failures
//
// ★ Insight ─────────────────────────────────────
// 1. MaterialBanner shows at the top of the screen without
//    blocking user interaction (non-blocking per FR-056).
// 2. SafeArea ensures the banner doesn't overlap system UI.
// 3. AnimationController provides smooth show/hide transitions.
// ─────────────────────────────────────────────────

import 'package:flutter/material.dart';

/// Notification severity level
enum NotificationSeverity {
  info,
  success,
  warning,
  error,
}

/// Notification data model
class Notification {
  final String id;
  final String title;
  final String? message;
  final NotificationSeverity severity;
  final VoidCallback? onRetry;
  final Duration duration;

  Notification({
    required this.id,
    required this.title,
    this.message,
    this.severity = NotificationSeverity.info,
    this.onRetry,
    this.duration = const Duration(seconds: 5),
  });
}

/// Notification banner widget - displays non-blocking notifications
class NotificationBanner extends StatefulWidget {
  final Notification notification;
  final VoidCallback? onDismiss;

  const NotificationBanner({
    super.key,
    required this.notification,
    this.onDismiss,
  });

  @override
  State<NotificationBanner> createState() => _NotificationBannerState();
}

class _NotificationBannerState extends State<NotificationBanner>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;
  late Animation<Offset> _slideAnimation;
  bool _isDisappearing = false;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 300),
    );

    _slideAnimation = Tween<Offset>(
      begin: const Offset(0, -1), // Start above the screen
      end: Offset.zero,
    ).animate(CurvedAnimation(
      parent: _controller,
      curve: Curves.easeOut,
    ));

    _controller.forward();

    // Auto-dismiss after duration
    if (widget.notification.duration.inMilliseconds > 0) {
      Future.delayed(widget.notification.duration, _dismiss);
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  void _dismiss() {
    if (mounted && !_isDisappearing) {
      _isDisappearing = true;
      _controller.reverse().then((_) {
        if (mounted) {
          widget.onDismiss?.call();
        }
      });
    }
  }

  Color _getBackgroundColor(BuildContext context) {
    final theme = Theme.of(context);
    switch (widget.notification.severity) {
      case NotificationSeverity.info:
        return theme.colorScheme.primaryContainer;
      case NotificationSeverity.success:
        return theme.colorScheme.secondaryContainer;
      case NotificationSeverity.warning:
        return Colors.orange.withValues(alpha: 0.9);
      case NotificationSeverity.error:
        return theme.colorScheme.errorContainer;
    }
  }

  IconData _getIcon() {
    switch (widget.notification.severity) {
      case NotificationSeverity.info:
        return Icons.info_outline;
      case NotificationSeverity.success:
        return Icons.check_circle_outline;
      case NotificationSeverity.warning:
        return Icons.warning_outlined;
      case NotificationSeverity.error:
        return Icons.error_outline;
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return SafeArea(
      child: SlideTransition(
        position: _slideAnimation,
        child: MaterialBanner(
          backgroundColor: _getBackgroundColor(context),
          content: Row(
            children: [
              Icon(
                _getIcon(),
                color: theme.colorScheme.onPrimaryContainer,
                size: 24,
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(
                      widget.notification.title,
                      style: theme.textTheme.titleSmall?.copyWith(
                        color: theme.colorScheme.onPrimaryContainer,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    if (widget.notification.message != null) ...[
                      const SizedBox(height: 4),
                      Text(
                        widget.notification.message!,
                        style: theme.textTheme.bodySmall?.copyWith(
                          color: theme.colorScheme.onPrimaryContainer,
                        ),
                      ),
                    ],
                  ],
                ),
              ),
            ],
          ),
          actions: [
            if (widget.notification.onRetry != null)
              TextButton(
                onPressed: () {
                  widget.notification.onRetry?.call();
                  _dismiss();
                },
                child: Text(
                  'Retry',
                  style: TextStyle(color: theme.colorScheme.onPrimaryContainer),
                ),
              ),
            TextButton(
              onPressed: _dismiss,
              child: Text(
                'Dismiss',
                style: TextStyle(color: theme.colorScheme.onPrimaryContainer),
              ),
            ),
          ],
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        ),
      ),
    );
  }
}

/// Notification manager - shows notifications overlay
class NotificationManager extends StatefulWidget {
  final Widget child;

  const NotificationManager({
    super.key,
    required this.child,
  });

  @override
  State<NotificationManager> createState() => _NotificationManagerState();

  /// Show a notification
  static void show(
    BuildContext context, {
    required String title,
    String? message,
    NotificationSeverity severity = NotificationSeverity.info,
    VoidCallback? onRetry,
    Duration duration = const Duration(seconds: 5),
  }) {
    final manager = context.findAncestorStateOfType<_NotificationManagerState>();
    manager?.show(
      Notification(
        id: DateTime.now().millisecondsSinceEpoch.toString(),
        title: title,
        message: message,
        severity: severity,
        onRetry: onRetry,
        duration: duration,
      ),
    );
  }
}

class _NotificationManagerState extends State<NotificationManager> {
  final List<Notification> _notifications = [];

  void show(Notification notification) {
    setState(() {
      _notifications.add(notification);
    });
  }

  void remove(Notification notification) {
    setState(() {
      _notifications.remove(notification);
    });
  }

  @override
  Widget build(BuildContext context) {
    return Stack(
      children: [
        widget.child,
        if (_notifications.isNotEmpty)
          Positioned(
            top: 0,
            left: 0,
            right: 0,
            child: Column(
              children: _notifications
                  .map(
                    (notif) => NotificationBanner(
                      key: ValueKey(notif.id),
                      notification: notif,
                      onDismiss: () => remove(notif),
                    ),
                  )
                  .toList(),
            ),
          ),
      ],
    );
  }
}

/// Convenience methods for common notifications
class NotificationHelper {
  static void showInfo(
    BuildContext context,
    String title, [
    String? message,
  ]) {
    NotificationManager.show(context, title: title, message: message);
  }

  static void showSuccess(
    BuildContext context,
    String title, [
    String? message,
  ]) {
    NotificationManager.show(
      context,
      title: title,
      message: message,
      severity: NotificationSeverity.success,
    );
  }

  static void showWarning(
    BuildContext context,
    String title, [
    String? message,
  ]) {
    NotificationManager.show(
      context,
      title: title,
      message: message,
      severity: NotificationSeverity.warning,
    );
  }

  static void showError(
    BuildContext context,
    String title, [
    String? message,
    VoidCallback? onRetry,
  ]) {
    NotificationManager.show(
      context,
      title: title,
      message: message,
      severity: NotificationSeverity.error,
      onRetry: onRetry,
    );
  }

  static void showRetryable(
    BuildContext context,
    String title,
    VoidCallback onRetry, {
    String? message,
    Duration duration = const Duration(seconds: 10),
  }) {
    NotificationManager.show(
      context,
      title: title,
      message: message,
      severity: NotificationSeverity.warning,
      onRetry: onRetry,
      duration: duration,
    );
  }
}
