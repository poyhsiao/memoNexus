# Manual Testing Guide

## Overview

This document describes manual testing tasks that require human intervention to complete. These tasks cannot be automated and must be validated by a human tester.

---

## T224: Application Launch Time Verification

**Objective**: Verify application launches in under 2 seconds with 10,000 content items.

**Prerequisites**:
- Application built and installed
- Database populated with 10,000+ content items

**Testing Steps**:
1. Populate database with 10,000+ content items
   ```bash
   # Use the bulk import feature or sync from existing data
   ```
2. Completely close the application
3. Start application and measure launch time
4. Verify time from app start to fully rendered UI < 2 seconds

**Success Criteria**:
- ✓ Launch time < 2 seconds with 10K items
- ✓ No visible lag or freeze during startup
- ✓ UI renders smoothly after launch

**Tools**:
- macOS: Use Activity Monitor to measure startup time
- Windows: Use Task Manager or Performance Monitor
- Manual stopwatch method acceptable

---

## T234: Quickstart Guide Validation

**Objective**: Follow `quickstart.md` from scratch to verify onboarding experience.

**Location**: [specs/001-memo-core/quickstart.md](quickstart.md)

**Prerequisites**:
- Clean environment (no previous installations)
- Fresh clone of the repository

**Testing Steps**:
1. Start with a clean machine/environment
2. Follow every step in quickstart.md exactly as written
3. Document any issues, missing steps, or confusing instructions
4. Verify all features work as described

**Success Criteria**:
- ✓ Can build and run application from scratch
- ✓ All steps are clear and unambiguous
- ✓ All features described in guide work correctly
- ✓ No dependencies or prerequisites are missing

**Notes to Document**:
- Any ambiguous instructions
- Missing environment setup steps
- Commands that fail or produce errors
- Feature gaps vs. documentation

---

## T235: Accessibility Testing

**Objective**: Verify application is accessible to users with disabilities.

**Testing Platform**: Test on each target platform (macOS, Windows, Linux, Android, iOS)

### 1. Keyboard Navigation

**Steps**:
1. Unplug mouse/trackpad
2. Navigate entire application using only keyboard (Tab, Arrow keys, Enter, Esc)
3. Verify all interactive elements are accessible
4. Check focus indicators are visible

**Success Criteria**:
- ✓ All buttons, links, form fields accessible via keyboard
- ✓ Logical tab order (left-to-right, top-to-bottom)
- ✓ Visible focus indicator on all elements
- ✓ Esc key closes modals/dropdowns
- ✓ Arrow keys navigate lists (content list, tags)

### 2. Screen Reader Testing

**Steps**:
1. Enable screen reader (VoiceOver on macOS, NVDA on Windows, TalkBack on Android/iOS)
2. Navigate through the application
3. Verify all content is announced correctly

**Success Criteria**:
- ✓ All buttons and links have descriptive labels
- ✓ Form fields have associated labels
- ✓ Images have alt text or are marked decorative
- ✓ Status messages are announced
- ✓ Error messages are announced with context

### 3. Color Contrast

**Steps**:
1. Use contrast checker tool or browser extension
2. Verify all text meets WCAG AA standards
3. Check UI elements (buttons, inputs, links)

**Success Criteria**:
- ✓ Normal text: Minimum 4.5:1 contrast ratio
- ✓ Large text (18pt+): Minimum 3:1 contrast ratio
- ✓ UI components: Minimum 3:1 contrast ratio

**Tools**:
- [WebAIM Contrast Checker](https://webaim.org/resources/contrastchecker/)
- macOS: Digital Color Meter
- Chrome: Axe DevTools extension

### 4. Text Scaling

**Steps**:
1. Increase system text size (150%, 200%)
2. Verify UI remains usable
3. Check for text truncation or overflow

**Success Criteria**:
- ✓ No text is truncated at 200% scaling
- ✓ Layout remains intact
- ✓ All text remains readable
- ✓ No horizontal scrolling required

---

## T236: Performance Benchmarking

**Objective**: Verify application meets performance requirements.

### 1. Search Performance (< 100ms for 10K items)

**Setup**:
- Database with 10,000+ content items
- Full-text search enabled (FTS5)

**Testing Steps**:
1. Populate database with 10,000 diverse items
2. Perform various search queries:
   - Single word: "machine learning"
   - Phrase: "artificial intelligence"
   - Tag-based: "#python"
   - Date range: last 30 days
3. Measure search response time for each query
4. Repeat 10 times per query type

**Success Criteria**:
- ✓ 95th percentile response time < 100ms
- ✓ No query exceeds 200ms
- ✓ Results display instantly after query

**Measurement**:
- Use built-in logging (search duration in logs)
- Verify in apps/frontend/lib/screens/search_screen.dart

### 2. Application Launch Time (< 2 seconds)

**Testing Steps**:
1. Cold start: Launch after system reboot
2. Warm start: Launch after recently closing
3. Measure 5 launches, take average

**Success Criteria**:
- ✓ Average launch time < 2 seconds
- ✓ No individual launch > 3 seconds

### 3. List Rendering Performance (< 500ms for 1,000 items)

**Testing Steps**:
1. Create list view with 1,000 items
2. Scroll through entire list
3. Verify no dropped frames
4. Check virtual scrolling is working

**Success Criteria**:
- ✓ Initial render < 500ms
- ✓ Smooth scrolling at 60fps
- ✓ No visible lag when loading more items

**Measurement**:
- Use Flutter DevTools Performance overlay
- Check frame rendering time
- Verify virtualized list implementation

---

## General Testing Notes

### Environment
- Test on real hardware, not just emulators
- Test on slow hardware (minimum spec machine)
- Test with large datasets (10K+ items)

### Documentation
- Document any issues found in GitHub issues
- Include screenshots/videos of failures
- Note system specifications (OS version, hardware specs)

### Automation Candidates
If any manual test can be automated in the future, document:
- Why it currently requires manual testing
- What would be needed to automate it
- Expected complexity of automation

---

## Completion Checklist

After completing all manual tests:

- [ ] T224: Application launch time verified < 2s
- [ ] T234: Quickstart guide validated from scratch
- [ ] T235: Accessibility testing completed (all 4 categories)
- [ ] T236: Performance benchmarks met (search, launch, render)

**Update tasks.md**:
```markdown
- [x] T224 Verify application launch time (<2 seconds with 10K items)
- [x] T234 Run quickstart.md validation (follow guide from scratch)
- [x] T235 Manual accessibility testing
- [x] T236 Performance testing (search <100ms for 10K items, launch <2s, list render <500ms)
```
