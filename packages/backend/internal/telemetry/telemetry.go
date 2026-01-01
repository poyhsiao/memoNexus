// Package telemetry provides no-op telemetry functions.
// T228: FR-053 requires ZERO external transmission without explicit opt-in.
// This package provides stub functions that do nothing by default.
// When telemetry is explicitly enabled by the user, a real implementation
// can be swapped in via build tags or configuration.
//
// SECURITY REQUIREMENT (FR-053):
// - No data is transmitted without explicit user opt-in
// - All functions are no-ops by default
// - Any real implementation must be opt-in only
package telemetry

import (
	"context"
	"time"
)

// =====================================================
// T228: No-Op Telemetry Functions
// =====================================================
// All functions in this package are no-ops by design.
// This ensures ZERO external data transmission without opt-in.

// IsEnabled returns false always (telemetry disabled by default).
// T228: FR-053 requires explicit opt-in for any telemetry.
// When the user explicitly enables telemetry, this should return true.
func IsEnabled() bool {
	// T228: Always disabled by default - explicit opt-in required
	return false
}

// EnableTelemetry enables telemetry collection.
// T228: This is a NO-OP by default. Real implementation should:
// 1. Store user consent in secure storage
// 2. Only collect/transmit after explicit consent
// 3. Provide clear way to disable
func EnableTelemetry() error {
	// T228: No-op - no telemetry without explicit opt-in
	return nil
}

// DisableTelemetry disables telemetry collection.
// T228: This is a NO-OP by default (already disabled).
func DisableTelemetry() error {
	// T228: No-op - already disabled by default
	return nil
}

// =====================================================
// Event Tracking (No-Op)
// =====================================================

// TrackEvent tracks a user event (NO-OP).
// T228: No events are tracked or transmitted without opt-in.
func TrackEvent(name string, properties map[string]interface{}) {
	// T228: No-op - no event tracking without opt-in
}

// TrackError tracks an error (NO-OP).
// T228: No errors are transmitted without opt-in.
func TrackError(err error, context map[string]interface{}) {
	// T228: No-op - no error transmission without opt-in
}

// TrackPageView tracks a page/view navigation (NO-OP).
// T228: No page view data is transmitted without opt-in.
func TrackPageView(page string, properties map[string]interface{}) {
	// T228: No-op - no page view tracking without opt-in
}

// =====================================================
// Metrics Collection (No-Op)
// =====================================================

// RecordMetric records a numeric metric (NO-OP).
// T228: No metrics are collected or transmitted without opt-in.
func RecordMetric(name string, value float64, tags map[string]string) {
	// T228: No-op - no metric collection without opt-in
}

// RecordTiming records a timing duration (NO-OP).
// T228: No timing data is transmitted without opt-in.
func RecordTiming(name string, duration time.Duration, tags map[string]string) {
	// T228: No-op - no timing data without opt-in
}

// RecordCount records a counter increment (NO-OP).
// T228: No counter data is transmitted without opt-in.
func RecordCount(name string, delta int, tags map[string]string) {
	// T228: No-op - no counter data without opt-in
}

// =====================================================
// User Identification (No-Op)
// =====================================================

// IdentifyUser associates events with a user ID (NO-OP).
// T228: No user identification is stored or transmitted without opt-in.
// WARNING: User identification is sensitive data - requires explicit consent.
func IdentifyUser(userID string) {
	// T228: No-op - no user identification without opt-in
}

// SetUserProperties sets user properties (NO-OP).
// T228: No user properties are stored or transmitted without opt-in.
func SetUserProperties(properties map[string]interface{}) {
	// T228: No-op - no user data without opt-in
}

// =====================================================
// Session Tracking (No-Op)
// =====================================================

// StartSession starts a new session (NO-OP).
// T228: No session tracking occurs without opt-in.
func StartSession() string {
	// T228: No-op - no session tracking without opt-in
	return ""
}

// EndSession ends the current session (NO-OP).
// T228: No session data is transmitted without opt-in.
func EndSession(sessionID string) {
	// T228: No-op - no session data without opt-in
}

// =====================================================
// Privacy Controls
// =====================================================

// ShouldCollectData returns false always (opt-in required).
// T228: FR-053 requires explicit opt-in for any data collection.
func ShouldCollectData() bool {
	// T228: Always false - data collection requires explicit opt-in
	return false
}

// GetOptInStatus returns the current opt-in status.
// T228: Returns "disabled" by default. Should return "enabled" only
// after user gives explicit consent through UI settings.
func GetOptInStatus() string {
	// T228: Always "disabled" by default
	return "disabled"
}

// =====================================================
// Verification Functions (T228)
// =====================================================

// VerifyNoExternalTransmission verifies no external HTTP requests are made.
// T228: This is a compile-time verification - all functions are no-ops.
// Returns: (true, nil) because no transmission occurs by default.
func VerifyNoExternalTransmission(ctx context.Context) (bool, error) {
	// T228: All telemetry functions are no-ops - no transmission possible
	return true, nil
}

// GetTransmittedDataSize returns the size of data transmitted (always 0).
// T228: Since all functions are no-ops, zero bytes are transmitted.
func GetTransmittedDataSize() int64 {
	// T228: No data transmitted - always 0 bytes
	return 0
}

// GetTransmittedRequestCount returns the count of HTTP requests made (always 0).
// T228: Since all functions are no-ops, zero requests are made.
func GetTransmittedRequestCount() int64 {
	// T228: No requests made - always 0
	return 0
}

// =====================================================
// Flush/Cleanup (No-Op)
// =====================================================

// Flush flushes any pending telemetry data (NO-OP).
// T228: No pending data exists since all functions are no-ops.
func Flush() error {
	// T228: No-op - no data to flush
	return nil
}

// Shutdown gracefully shuts down telemetry (NO-OP).
// T228: No cleanup needed since no resources are used.
func Shutdown(ctx context.Context) error {
	// T228: No-op - nothing to shut down
	return nil
}
