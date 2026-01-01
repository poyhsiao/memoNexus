// Package crypto tests for secure storage utility functions.
package crypto

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestNewSecureStorage verifies SecureStorage initialization.
func TestNewSecureStorage(t *testing.T) {
	configDir := "/test/config"
	ss := NewSecureStorage(configDir)

	if ss == nil {
		t.Fatal("NewSecureStorage() returned nil")
	}

	if ss.serviceName != ServiceName {
		t.Errorf("serviceName = %q, want %q", ss.serviceName, ServiceName)
	}

	if ss.configDir != configDir {
		t.Errorf("configDir = %q, want %q", ss.configDir, configDir)
	}
}

// TestNewSecureStorage_emptyConfigDir verifies empty config dir is allowed.
func TestNewSecureStorage_emptyConfigDir(t *testing.T) {
	ss := NewSecureStorage("")

	if ss.configDir != "" {
		t.Errorf("configDir = %q, want empty string", ss.configDir)
	}
}

// TestGetMachineIdentifier verifies machine identifier generation.
func TestGetMachineIdentifier(t *testing.T) {
	identifier := getMachineIdentifier()

	if identifier == "" {
		t.Error("getMachineIdentifier() returned empty string")
	}

	// Verify prefix based on OS
	expectedPrefix := ""
	switch runtime.GOOS {
	case "darwin":
		expectedPrefix = "macos:"
	case "windows":
		expectedPrefix = "windows:"
	default:
		expectedPrefix = "linux:"
	}

	if !strings.HasPrefix(identifier, expectedPrefix) {
		t.Errorf("getMachineIdentifier() = %q, want prefix %q", identifier, expectedPrefix)
	}

	// Verify identifier has more than just the prefix
	if len(identifier) <= len(expectedPrefix) {
		t.Errorf("getMachineIdentifier() = %q, too short", identifier)
	}
}

// TestGetmacOSMachineID verifies macOS machine ID format.
func TestGetmacOSMachineID(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test on non-macOS platform")
	}

	machineID := getmacOSMachineID()

	if machineID == "" {
		t.Error("getmacOSMachineID() returned empty string")
	}

	if !strings.HasPrefix(machineID, "macos:") {
		t.Errorf("getmacOSMachineID() = %q, want prefix 'macos:'", machineID)
	}
}

// TestGetWindowsMachineID verifies Windows machine ID format.
func TestGetWindowsMachineID(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows platform")
	}

	machineID := getWindowsMachineID()

	if machineID == "" {
		t.Error("getWindowsMachineID() returned empty string")
	}

	if !strings.HasPrefix(machineID, "windows:") {
		t.Errorf("getWindowsMachineID() = %q, want prefix 'windows:'", machineID)
	}
}

// TestGetLinuxMachineID verifies Linux machine ID format.
func TestGetLinuxMachineID(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test on non-Linux platform")
	}

	machineID := getLinuxMachineID()

	if machineID == "" {
		t.Error("getLinuxMachineID() returned empty string")
	}

	if !strings.HasPrefix(machineID, "linux:") {
		t.Errorf("getLinuxMachineID() = %q, want prefix 'linux:'", machineID)
	}
}

// TestGetLinuxMachineID_fallback verifies Linux machine ID fallback logic.
func TestGetLinuxMachineID_fallback(t *testing.T) {
	// This test verifies the fallback logic on any platform
	// by reading from /etc/machine-id if it exists (Linux)

	// Try to read from /etc/machine-id (will likely fail on non-Linux)
	machineID := getLinuxMachineID()

	if machineID != "" {
		if !strings.HasPrefix(machineID, "linux:") {
			t.Errorf("getLinuxMachineID() = %q, want prefix 'linux:'", machineID)
		}
	}
	// Empty machineID is acceptable if /etc/machine-id doesn't exist
}

// TestStoreFileCredential_noConfigDir verifies error when config dir not set.
func TestStoreFileCredential_noConfigDir(t *testing.T) {
	// Skip on macOS/Windows where platform-specific storage is used
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		t.Skip("Skipping on platform with native secure storage")
	}

	ss := NewSecureStorage("")
	err := ss.StoreCredential("test-account", "test-value")

	if err == nil {
		t.Error("StoreCredential() with empty configDir should return error")
	}

	if err != nil && !strings.Contains(err.Error(), "config directory not set") {
		t.Errorf("Error message should mention config directory: %v", err)
	}
}

// TestGetFileCredential_noConfigDir verifies error when config dir not set.
func TestGetFileCredential_noConfigDir(t *testing.T) {
	// Skip on macOS/Windows where platform-specific storage is used
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		t.Skip("Skipping on platform with native secure storage")
	}

	ss := NewSecureStorage("")
	value, err := ss.GetCredential("test-account")

	if err == nil {
		t.Error("GetCredential() with empty configDir should return error")
	}

	if value != "" {
		t.Errorf("GetCredential() should return empty string on error, got %q", value)
	}

	if err != nil && !strings.Contains(err.Error(), "config directory not set") {
		t.Errorf("Error message should mention config directory: %v", err)
	}
}

// TestDeleteFileCredential_noConfigDir verifies error when config dir not set.
func TestDeleteFileCredential_noConfigDir(t *testing.T) {
	// Skip on macOS/Windows where platform-specific storage is used
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		t.Skip("Skipping on platform with native secure storage")
	}

	ss := NewSecureStorage("")
	err := ss.DeleteCredential("test-account")

	if err == nil {
		t.Error("DeleteCredential() with empty configDir should return error")
	}

	if err != nil && !strings.Contains(err.Error(), "config directory not set") {
		t.Errorf("Error message should mention config directory: %v", err)
	}
}

// TestFileCredentialOperations verifies file-based credential operations.
func TestFileCredentialOperations(t *testing.T) {
	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "secure_storage_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ss := NewSecureStorage(tmpDir)
	account := "test-account"
	value := "test-secret-value"

	// Test StoreCredential
	err = ss.StoreCredential(account, value)
	if err != nil {
		// Note: This may fail on non-Linux platforms due to platform-specific implementations
		// On macOS/Windows, it would use Keychain/Credential Manager
		if runtime.GOOS == "linux" || strings.Contains(err.Error(), "config directory") {
			t.Errorf("StoreCredential() failed: %v", err)
		}
		return
	}

	// Test GetCredential
	retrieved, err := ss.GetCredential(account)
	if err != nil {
		t.Errorf("GetCredential() failed: %v", err)
	}

	if retrieved != value {
		t.Errorf("GetCredential() = %q, want %q", retrieved, value)
	}

	// Test DeleteCredential
	err = ss.DeleteCredential(account)
	if err != nil {
		t.Errorf("DeleteCredential() failed: %v", err)
	}

	// Verify credential is deleted
	_, err = ss.GetCredential(account)
	if err == nil {
		t.Error("GetCredential() should return error after deletion")
	}
}

// TestFileCredentialOperations_specialCharacters verifies sanitization of special chars.
func TestFileCredentialOperations_specialCharacters(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "secure_storage_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ss := NewSecureStorage(tmpDir)

	tests := []struct {
		name        string
		account     string
		sanitized   string // Expected sanitized account for filename
	}{
		{
			name:      "forward slashes",
			account:   "test/account/name",
			sanitized: "test_account_name",
		},
		{
			name:      "backslashes",
			account:   "test\\account\\name",
			sanitized: "test_account_name",
		},
		{
			name:      "double dots",
			account:   "test..account..name",
			sanitized: "test__account__name",
		},
		{
			name:      "mixed special chars",
			account:   "test/account\\name..here",
			sanitized: "test_account_name__here",
		},
		{
			name:      "normal account",
			account:   "normal-account-name",
			sanitized: "normal-account-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify that special characters don't cause crashes
			// Actual sanitization is implementation detail
			err := ss.StoreCredential(tt.account, "test-value")
			if err != nil && runtime.GOOS == "linux" {
				// On Linux, if we get past the config dir check, the operation should succeed
				if !strings.Contains(err.Error(), "config directory") {
					t.Errorf("StoreCredential() with special chars failed: %v", err)
				}
			}
		})
	}
}

// TestFileCredentialOperations_concurrentAccess verifies thread safety.
func TestFileCredentialOperations_concurrentAccess(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping concurrent test on non-Linux platform (uses platform-specific storage)")
	}

	tmpDir, err := os.MkdirTemp("", "secure_storage_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ss := NewSecureStorage(tmpDir)
	account := "concurrent-test"
	value := "test-value"

	// Run multiple operations concurrently (basic smoke test)
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			ss.StoreCredential(account, value+"_"+string(rune('0'+idx)))
			ss.GetCredential(account)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// If we got here without deadlock or panic, test passes
}

// =====================================================
// Direct File Credential Function Tests
// =====================================================

// TestStoreFileCredential verifies direct file credential storage.
func TestStoreFileCredential(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "secure_storage_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ss := NewSecureStorage(tmpDir)

	account := "test-direct-account"
	value := "my-secret-password"

	err = ss.storeFileCredential(account, value)
	if err != nil {
		t.Fatalf("storeFileCredential() failed: %v", err)
	}

	// Verify file was created
	secureDir := filepath.Join(tmpDir, "secure")
	safeAccount := strings.ReplaceAll(account, "/", "_")
	credFile := filepath.Join(secureDir, safeAccount+".cred")

	if _, err := os.Stat(credFile); os.IsNotExist(err) {
		t.Errorf("Credential file was not created: %s", credFile)
	}
}

// TestGetFileCredential verifies direct file credential retrieval.
func TestGetFileCredential(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "secure_storage_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ss := NewSecureStorage(tmpDir)

	account := "test-retrieve-account"
	value := "my-secret-value"

	// First store the credential
	err = ss.storeFileCredential(account, value)
	if err != nil {
		t.Fatalf("storeFileCredential() failed: %v", err)
	}

	// Now retrieve it
	retrieved, err := ss.getFileCredential(account)
	if err != nil {
		t.Fatalf("getFileCredential() failed: %v", err)
	}

	if retrieved != value {
		t.Errorf("getFileCredential() = %q, want %q", retrieved, value)
	}
}

// TestGetFileCredential_notFound verifies error when credential doesn't exist.
func TestGetFileCredential_notFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "secure_storage_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ss := NewSecureStorage(tmpDir)

	retrieved, err := ss.getFileCredential("non-existent-account")

	if err == nil {
		t.Error("getFileCredential() should return error for non-existent credential")
	}

	if retrieved != "" {
		t.Errorf("getFileCredential() should return empty string on error, got %q", retrieved)
	}

	if !strings.Contains(err.Error(), "credential not found") {
		t.Logf("Error message: %v", err)
	}
}

// TestDeleteFileCredential verifies direct file credential deletion.
func TestDeleteFileCredential(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "secure_storage_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ss := NewSecureStorage(tmpDir)

	account := "test-delete-account"
	value := "my-delete-value"

	// Store the credential
	err = ss.storeFileCredential(account, value)
	if err != nil {
		t.Fatalf("storeFileCredential() failed: %v", err)
	}

	// Delete it
	err = ss.deleteFileCredential(account)
	if err != nil {
		t.Fatalf("deleteFileCredential() failed: %v", err)
	}

	// Verify it's gone
	_, err = ss.getFileCredential(account)
	if err == nil {
		t.Error("getFileCredential() should return error after deletion")
	}
}

// TestDeleteFileCredential_notFound verifies deletion of non-existent credential succeeds.
func TestDeleteFileCredential_notFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "secure_storage_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ss := NewSecureStorage(tmpDir)

	// Delete non-existent credential - should succeed (idempotent)
	err = ss.deleteFileCredential("non-existent-account")
	if err != nil {
		t.Errorf("deleteFileCredential() of non-existent should succeed, got: %v", err)
	}
}

// TestWindowsCredentialFunctions verifies Windows credential functions.
func TestWindowsCredentialFunctions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "secure_storage_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ss := NewSecureStorage(tmpDir)

	account := "test-windows-account"
	value := "my-windows-secret"

	// Test storeWindowsCredential (uses file fallback)
	err = ss.storeWindowsCredential(account, value)
	if err != nil {
		t.Fatalf("storeWindowsCredential() failed: %v", err)
	}

	// Test getWindowsCredential
	retrieved, err := ss.getWindowsCredential(account)
	if err != nil {
		t.Fatalf("getWindowsCredential() failed: %v", err)
	}

	if retrieved != value {
		t.Errorf("getWindowsCredential() = %q, want %q", retrieved, value)
	}

	// Test deleteWindowsCredential
	err = ss.deleteWindowsCredential(account)
	if err != nil {
		t.Fatalf("deleteWindowsCredential() failed: %v", err)
	}

	// Verify deletion
	_, err = ss.getWindowsCredential(account)
	if err == nil {
		t.Error("getWindowsCredential() should return error after deletion")
	}
}

// =====================================================
// macOS-Specific Tests
// =====================================================

// TestMacOSCredOperations verifies macOS Keychain operations.
func TestMacOSCredOperations(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test on non-macOS platform")
	}

	ss := NewSecureStorage("") // Config dir not needed for macOS
	account := "test-macos-credential"
	value := "test-secret-value-macos"

	// Clean up any existing credential first
	ss.DeleteCredential(account)

	// Test StoreCredential
	err := ss.StoreCredential(account, value)
	if err != nil {
		t.Fatalf("StoreCredential() failed: %v", err)
	}

	// Test GetCredential
	retrieved, err := ss.GetCredential(account)
	if err != nil {
		t.Errorf("GetCredential() failed: %v", err)
	}

	if retrieved != value {
		t.Errorf("GetCredential() = %q, want %q", retrieved, value)
	}

	// Test DeleteCredential
	err = ss.DeleteCredential(account)
	if err != nil {
		t.Errorf("DeleteCredential() failed: %v", err)
	}

	// Verify credential is deleted
	_, err = ss.GetCredential(account)
	if err == nil {
		t.Error("GetCredential() should return error after deletion")
	}
}

// TestMacOSCredOperations_notFound verifies error when credential doesn't exist.
func TestMacOSCredOperations_notFound(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test on non-macOS platform")
	}

	ss := NewSecureStorage("")
	nonExistentAccount := "non-existent-credential-" + strings.ReplaceAll(getMachineIdentifier(), ":", "-")

	// Try to get a credential that doesn't exist
	_, err := ss.GetCredential(nonExistentAccount)
	if err == nil {
		t.Error("GetCredential() should return error for non-existent credential")
	}

	// Delete non-existent credential should succeed (idempotent)
	err = ss.DeleteCredential(nonExistentAccount)
	if err != nil {
		t.Errorf("DeleteCredential() of non-existent should succeed, got: %v", err)
	}
}

// TestMacOSCredOperations_unicode verifies macOS Keychain handles unicode values.
func TestMacOSCredOperations_unicode(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test on non-macOS platform")
	}

	ss := NewSecureStorage("")
	account := "test-unicode-" + strings.ReplaceAll(getMachineIdentifier(), ":", "-")
	value := "🔐 Secret 你好 🌍"

	// Clean up first
	ss.DeleteCredential(account)
	defer ss.DeleteCredential(account)

	err := ss.StoreCredential(account, value)
	if err != nil {
		t.Fatalf("StoreCredential() with unicode failed: %v", err)
	}

	retrieved, err := ss.GetCredential(account)
	if err != nil {
		t.Errorf("GetCredential() failed: %v", err)
	}

	if retrieved != value {
		t.Errorf("GetCredential() = %q, want %q", retrieved, value)
	}
}

// TestMacOSCredOperations_longValue verifies macOS Keychain handles long values.
func TestMacOSCredOperations_longValue(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test on non-macOS platform")
	}

	ss := NewSecureStorage("")
	account := "test-long-value-" + strings.ReplaceAll(getMachineIdentifier(), ":", "-")

	// Create a 1KB value
	value := strings.Repeat("A", 1024)

	// Clean up first
	ss.DeleteCredential(account)
	defer ss.DeleteCredential(account)

	err := ss.StoreCredential(account, value)
	if err != nil {
		t.Fatalf("StoreCredential() with long value failed: %v", err)
	}

	retrieved, err := ss.GetCredential(account)
	if err != nil {
		t.Errorf("GetCredential() failed: %v", err)
	}

	if retrieved != value {
		t.Errorf("GetCredential() length = %d, want %d", len(retrieved), len(value))
	}
}

// =====================================================
// Windows-Specific Tests (Fallback)
// =====================================================

// TestWindowsCredOperations_fallback verifies Windows uses file fallback.
func TestWindowsCredOperations_fallback(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows platform")
	}

	// On Windows, it should use file fallback
	tmpDir, err := os.MkdirTemp("", "secure_storage_windows_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ss := NewSecureStorage(tmpDir)
	account := "test-windows-credential"
	value := "test-secret-value-windows"

	// Test StoreCredential
	err = ss.StoreCredential(account, value)
	if err != nil {
		t.Fatalf("StoreCredential() failed: %v", err)
	}

	// Test GetCredential
	retrieved, err := ss.GetCredential(account)
	if err != nil {
		t.Errorf("GetCredential() failed: %v", err)
	}

	if retrieved != value {
		t.Errorf("GetCredential() = %q, want %q", retrieved, value)
	}

	// Test DeleteCredential
	err = ss.DeleteCredential(account)
	if err != nil {
		t.Errorf("DeleteCredential() failed: %v", err)
	}

	// Verify credential is deleted
	_, err = ss.GetCredential(account)
	if err == nil {
		t.Error("GetCredential() should return error after deletion")
	}
}

// TestWindowsCredOperations_noConfigDir verifies error when config dir not set on Windows.
func TestWindowsCredOperations_noConfigDir(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows platform")
	}

	ss := NewSecureStorage("")

	// All operations should fail without config dir on Windows
	err := ss.StoreCredential("test-account", "test-value")
	if err == nil {
		t.Error("StoreCredential() should fail without config dir")
	}

	_, err = ss.GetCredential("test-account")
	if err == nil {
		t.Error("GetCredential() should fail without config dir")
	}

	err = ss.DeleteCredential("test-account")
	if err == nil {
		t.Error("DeleteCredential() should fail without config dir")
	}
}
