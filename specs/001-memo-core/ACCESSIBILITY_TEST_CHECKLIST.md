# Accessibility Testing Checklist (T235)

**Feature**: 001-memo-core  
**Date**: 2025-01-01  
**Platform**: Flutter (Cross-platform)

## Test Environment

- [ ] Test on macOS with VoiceOver enabled
- [ ] Test on Windows with Narrator enabled
- [ ] Test on Linux with Orca enabled
- [ ] Test on Android with TalkBack enabled
- [ ] Test on iOS with VoiceOver enabled

## 1. Keyboard Navigation

### 1.1 Tab Navigation
- [ ] All interactive elements are reachable via Tab key
- [ ] Tab order follows logical reading order (left-to-right, top-to-bottom)
- [ ] Focus indicator is clearly visible
- [ ] No keyboard traps (can navigate away from any element)

### 1.2 Shortcut Keys
- [ ] Escape key closes dialogs/modals
- [ ] Enter key activates focused buttons
- [ ] Space key toggles checkboxes and radio buttons
- [ ] Arrow keys navigate within lists and dropdowns

### 1.3 Screen-Specific Tests
- [ ] **Capture Screen**: Tab through URL input, file picker, tags, save button
- [ ] **Search Screen**: Tab through search bar, filter dropdown, results list
- [ ] **Content Detail**: Tab through content, edit button, delete button
- [ ] **Settings Screens**: Tab through all form fields

## 2. Screen Reader Compatibility

### 2.1 Element Labels
- [ ] All buttons have descriptive labels (not just "button")
- [ ] All text fields have associated labels
- [ ] Icons have accessible labels or are hidden from screen readers
- [ ] Images have alt text or are marked as decorative

### 2.2 Focus Announcements
- [ ] Screen reader announces when focus moves to new element
- [ ] Form validation errors are announced
- [ ] Success messages are announced
- [ ] Loading states are announced

### 2.3 Content Reading
- [ ] Screen reader reads content in correct order
- [ ] Dynamic content updates are announced
- [ ] Off-screen content is not read
- [ ] Semantic HTML is used (headings, lists, etc.)

## 3. Color Contrast

### 3.1 Text Contrast (WCAG AA: 4.5:1)
- [ ] Body text meets 4.5:1 contrast ratio against background
- [ ] Large text (18pt+) meets 3:1 contrast ratio
- [ ] Text on buttons meets 4.5:1 contrast ratio
- [ ] Text in form fields meets 4.5:1 contrast ratio

### 3.2 UI Element Contrast
- [ ] Icons meet 3:1 contrast ratio against background
- [ ] Borders and focus indicators are visible
- [ ] Error messages have sufficient contrast
- [ ] Disabled elements are distinguishable from enabled

### 3.3 Color Independence
- [ ] Information is not conveyed by color alone
- [ ] Links are distinguished by more than just color (underline, icon)
- [ ] Error states use additional indicators beyond red color
- [ ] Form field validation uses icons/text in addition to color

## 4. Text Scaling

### 4.1 Font Size
- [ ] App respects system font size settings
- [ ] Text is readable at 200% zoom
- [ ] UI layout doesn't break at large font sizes
- [ ] No text is truncated or overlaps

### 4.2 Text Layout
- [ ] Sufficient line height (1.5x font size for body text)
- [ ] Paragraph spacing (2x font size between paragraphs)
- [ ] Character spacing (0.12x for body text)
- [ ] Word spacing (0.16x for body text)

## 5. Touch Targets (Mobile)

### 5.1 Target Size
- [ ] All touch targets are at least 44x44 dp (Android) / 44x44 pt (iOS)
- [ ] Targets are not too close together (minimum 8dp spacing)
- [ ] Large enough targets for users with motor impairments

### 5.2 Gesture Alternatives
- [ ] All gestures have button alternatives
- [ ] Swipe actions can be performed via keyboard
- [ ] Pinch-to-zoom has button alternative
- [ ] Long-press has button alternative

## 6. Platform-Specific Tests

### 6.1 macOS
- [ ] VoiceOver reads all UI elements correctly
- [ ] Full keyboard navigation works
- [ ] System contrast preferences are respected
- [ ] System font size preferences are respected

### 6.2 iOS
- [ ] VoiceOver reads all UI elements correctly
- [ ] Touch targets meet iOS HIG guidelines (44x44 pt)
- [ ] Dynamic Type support works
- [ ] Reduce Motion preference is respected

### 6.3 Android
- [ ] TalkBack reads all UI elements correctly
- [ ] Touch targets meet Material Design guidelines (48x48 dp)
- [ ] Font scale settings work (100% - 200%)
- [ ] TalkBack focus order is logical

### 6.4 Windows
- [ ] Narrator reads all UI elements correctly
- [ ] High contrast mode works
- [ ] Keyboard navigation is complete
- [ ] DPI scaling works correctly

### 6.5 Linux
- [ ] Orca screen reader works
- [ ] High contrast mode works
- [ ] Keyboard navigation is complete
- [ ] System theme is respected

## 7. Manual Testing Scenarios

### Scenario 1: Add Content (Keyboard Only)
1. Launch app
2. Press Tab to navigate to "Add Content" button
3. Press Enter to activate
4. Tab through form fields
5. Enter content using keyboard
6. Tab to Save button and press Enter
7. Verify success message is announced

### Scenario 2: Search Content (Screen Reader)
1. Enable screen reader
2. Navigate to search bar
3. Type search query
4. Navigate through results
5. Verify each result is read aloud
6. Select result and verify detail view is announced

### Scenario 3: Navigate Content List (Voice Control)
1. Use voice commands to scroll list
2. Use voice commands to select item
3. Verify all actions can be completed via voice

### Scenario 4: Adjust Text Size
1. Set system font size to 200%
2. Launch app
3. Verify all text is readable
4. Verify no text is truncated
5. Verify no layout issues

## 8. Tools & Verification

### Automated Checks
- [ ] Run `flutter analyze` for accessibility hints
- [ ] Use `accessibility_tools` package for automated checks
- [ ] Verify contrast ratios using contrast checker tools

### Manual Checks
- [ ] Test with real screen readers (VoiceOver, TalkBack, Narrator)
- [ ] Test with keyboard only (no mouse/touch)
- [ ] Test with high contrast mode enabled
- [ ] Test with enlarged text (200%)

## 9. Known Issues & Mitigations

| Issue | Severity | Mitigation |
|-------|----------|------------|
| (List any accessibility issues found) | | |

## 10. Test Results Summary

- **Date**: ___________
- **Tester**: ___________
- **Platform(s) Tested**: ___________
- **Overall Status**: ☐ Pass ☐ Fail (with notes)

### Issues Found
1. __________________________________________________________
2. __________________________________________________________
3. __________________________________________________________

### Recommendations
1. __________________________________________________________
2. __________________________________________________________
3. __________________________________________________________

---

## References

- [WCAG 2.1 Guidelines](https://www.w3.org/WAI/WCAG21/quickref/)
- [Flutter Accessibility Guide](https://docs.flutter.dev/development/accessibility-and-localization/accessibility)
- [Material Design Accessibility](https://material.io/design/usability/accessibility.html)
- [iOS Human Interface Guidelines - Accessibility](https://developer.apple.com/design/human-interface-guidelines/accessibility)
- [Android Accessibility Guidelines](https://developer.android.com/guide/topics/ui/accessibility)
