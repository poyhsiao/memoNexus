import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'dart:developer' as developer;

// Performance tracking for T224: Launch time <2s with 10K items
final _appLaunchStartTime = DateTime.now();

void main() {
  // Enable performance overlay for development
  bool showPerformanceOverlay = false;
  assert(() {
    showPerformanceOverlay = true;
    return true;
  }());

  runApp(ProviderScope(
    observers: [
      _PerformanceObserver(),
    ],
    child: MemoNexusApp(
      showPerformanceOverlay: showPerformanceOverlay,
    ),
  ));

  // Log application launch time for T224 verification
  final launchDuration = DateTime.now().difference(_appLaunchStartTime);
  developer.log(
    'Application launched in ${launchDuration.inMilliseconds}ms '
    '(target: <2000ms per constitution)',
    name: 'performance',
  );

  if (launchDuration.inMilliseconds >= 2000) {
    developer.log(
      'WARNING: Launch time exceeds 2 second constitutional requirement!',
      name: 'performance',
      level: 1000, // WARNING level
    );
  }
}

/// Performance observer for tracking provider lifecycle and initialization
class _PerformanceObserver extends ProviderObserver {
  @override
  void didAddProvider(
    ProviderBase<Object?> provider,
    Object? value,
    ProviderContainer container,
  ) {
    developer.log(
      'Provider added: ${provider.name ?? provider.runtimeType}',
      name: 'performance',
    );
  }

  @override
  void didDisposeProvider(
    ProviderBase<Object?> provider,
    ProviderContainer container,
  ) {
    developer.log(
      'Provider disposed: ${provider.name ?? provider.runtimeType}',
      name: 'performance',
    );
  }

  @override
  void providerDidFail(
    ProviderBase<Object?> provider,
    Object error,
    StackTrace stackTrace,
    ProviderContainer container,
  ) {
    developer.log(
      'Provider failed: ${provider.name ?? provider.runtimeType} - $error',
      name: 'performance',
      level: 2000, // ERROR level
      error: error,
      stackTrace: stackTrace,
    );
  }
}

class MemoNexusApp extends StatelessWidget {
  final bool showPerformanceOverlay;

  const MemoNexusApp({
    super.key,
    required this.showPerformanceOverlay,
  });

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'MemoNexus',
      debugShowCheckedModeBanner: false,
      // T224: Performance overlay for development profiling
      showPerformanceOverlay: showPerformanceOverlay,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.blue),
        useMaterial3: true,
      ),
      home: const HomePage(),
    );
  }
}

class HomePage extends StatelessWidget {
  const HomePage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('MemoNexus'),
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
      ),
      body: const Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.library_books, size: 80, color: Colors.blue),
            SizedBox(height: 24),
            Text(
              'MemoNexus',
              style: TextStyle(fontSize: 32, fontWeight: FontWeight.bold),
            ),
            SizedBox(height: 16),
            Text(
              'Local-First Personal Knowledge Base',
              style: TextStyle(fontSize: 16, color: Colors.grey),
            ),
            SizedBox(height: 48),
            Text(
              'Phase 8: Polish & Testing',
              style: TextStyle(fontSize: 14, color: Colors.green),
            ),
            SizedBox(height: 8),
            Text(
              'Performance monitoring enabled',
              style: TextStyle(fontSize: 12, color: Colors.grey),
            ),
          ],
        ),
      ),
    );
  }
}
