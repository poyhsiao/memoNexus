// App Theme Tests
// Tests for application theme configuration

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:memonexus_frontend/theme/app_theme.dart';

void main() {
  group('AppTheme', () {
    test('should have light theme defined', () {
      final theme = AppTheme.lightTheme;

      expect(theme, isNotNull);
      expect(theme.brightness, Brightness.light);
    });

    test('should have dark theme defined', () {
      final theme = AppTheme.darkTheme;

      expect(theme, isNotNull);
      expect(theme.brightness, Brightness.dark);
    });

    test('should have consistent color scheme', () {
      final lightTheme = AppTheme.lightTheme;
      final darkTheme = AppTheme.darkTheme;

      expect(lightTheme.colorScheme, isNotNull);
      expect(darkTheme.colorScheme, isNotNull);
    });

    test('should have defined text theme', () {
      final lightTheme = AppTheme.lightTheme;
      final darkTheme = AppTheme.darkTheme;

      expect(lightTheme.textTheme, isNotNull);
      expect(darkTheme.textTheme, isNotNull);
    });

    test('should use material 3', () {
      final lightTheme = AppTheme.lightTheme;
      final darkTheme = AppTheme.darkTheme;

      expect(lightTheme.useMaterial3, isTrue);
      expect(darkTheme.useMaterial3, isTrue);
    });
  });
}
