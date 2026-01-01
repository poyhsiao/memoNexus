// Package crypto provides platform-secure storage for sensitive credentials.
// T225: Platform-specific secure storage using Keychain (macOS), Credential Manager (Windows),
// or encrypted file (Linux fallback).
package crypto

/*
#cgo darwin CFLAGS: -x objective-c
#cgo darwin LDFLAGS: -framework Foundation -framework Security

#import <Foundation/Foundation.h>
#import <Security/Security.h>

// macOS Keychain functions
int macOSWriteKeychain(const char *service, const char *account, const char *password) {
	NSString *nsService = [NSString stringWithUTF8String:service];
	NSString *nsAccount = [NSString stringWithUTF8String:account];
	NSString *nsPassword = [NSString stringWithUTF8String:password];

	NSDictionary *query = @{
		(__bridge id)kSecClass: (__bridge id)kSecClassGenericPassword,
		(__bridge id)kSecAttrService: nsService,
		(__bridge id)kSecAttrAccount: nsAccount,
		(__bridge id)kSecValueData: [nsPassword dataUsingEncoding:NSUTF8StringEncoding],
	};

	OSStatus status = SecItemAdd((__bridge CFDictionaryRef)query, NULL);
	return (int)status;
}

char* macOSReadKeychain(const char *service, const char *account) {
	NSString *nsService = [NSString stringWithUTF8String:service];
	NSString *nsAccount = [NSString stringWithUTF8String:account];

	NSDictionary *query = @{
		(__bridge id)kSecClass: (__bridge id)kSecClassGenericPassword,
		(__bridge id)kSecAttrService: nsService,
		(__bridge id)kSecAttrAccount: nsAccount,
		(__bridge id)kSecReturnData: @YES,
		(__bridge id)kSecMatchLimit: (__bridge id)kSecMatchLimitOne,
	};

	CFTypeRef result = NULL;
	OSStatus status = SecItemCopyMatching((__bridge CFDictionaryRef)query, &result);

	if (status == errSecSuccess) {
		NSData *data = (__bridge_transfer NSData *)result;
		NSString *password = [[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding];
		return strdup([password UTF8String]);
	}
	return NULL;
}

int macOSDeleteKeychain(const char *service, const char *account) {
	NSString *nsService = [NSString stringWithUTF8String:service];
	NSString *nsAccount = [NSString stringWithUTF8String:account];

	NSDictionary *query = @{
		(__bridge id)kSecClass: (__bridge id)kSecClassGenericPassword,
		(__bridge id)kSecAttrService: nsService,
		(__bridge id)kSecAttrAccount: nsAccount,
	};

	OSStatus status = SecItemDelete((__bridge CFDictionaryRef)query);
	return (int)status;
}
*/
import "C"
import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unsafe"
)

const (
	// ServiceName is the name used for storing credentials in platform key stores.
	ServiceName = "com.kimhsiao.memonexus"
)

// SecureStorage provides platform-specific secure credential storage.
type SecureStorage struct {
	serviceName string
	configDir   string // For Linux fallback
}

// NewSecureStorage creates a new SecureStorage instance.
// T225: Platform-specific secure storage initialization.
func NewSecureStorage(configDir string) *SecureStorage {
	return &SecureStorage{
		serviceName: ServiceName,
		configDir:   configDir,
	}
}

// StoreCredential securely stores a credential (username/password or key/value pair).
// T225: Uses platform-specific secure storage:
// - macOS: Keychain Services
// - Windows: Credential Manager (via fallback for now)
// - Linux: Encrypted file fallback
func (s *SecureStorage) StoreCredential(account, value string) error {
	switch runtime.GOOS {
	case "darwin":
		return s.storemacOSCredential(account, value)
	case "windows":
		return s.storeWindowsCredential(account, value)
	default:
		return s.storeFileCredential(account, value)
	}
}

// GetCredential retrieves a securely stored credential.
func (s *SecureStorage) GetCredential(account string) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return s.getmacOSCredential(account)
	case "windows":
		return s.getWindowsCredential(account)
	default:
		return s.getFileCredential(account)
	}
}

// DeleteCredential removes a securely stored credential.
func (s *SecureStorage) DeleteCredential(account string) error {
	switch runtime.GOOS {
	case "darwin":
		return s.deletemacOSCredential(account)
	case "windows":
		return s.deleteWindowsCredential(account)
	default:
		return s.deleteFileCredential(account)
	}
}

// =====================================================
// macOS Implementation using Keychain
// =====================================================

func (s *SecureStorage) storemacOSCredential(account, value string) error {
	cService := C.CString(s.serviceName)
	defer C.free(unsafe.Pointer(cService))

	cAccount := C.CString(account)
	defer C.free(unsafe.Pointer(cAccount))

	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	status := C.macOSWriteKeychain(cService, cAccount, cValue)
	if status != 0 {
		return fmt.Errorf("macOS Keychain write failed with status: %d", status)
	}
	return nil
}

func (s *SecureStorage) getmacOSCredential(account string) (string, error) {
	cService := C.CString(s.serviceName)
	defer C.free(unsafe.Pointer(cService))

	cAccount := C.CString(account)
	defer C.free(unsafe.Pointer(cAccount))

	cValue := C.macOSReadKeychain(cService, cAccount)
	if cValue == nil {
		return "", fmt.Errorf("credential not found in Keychain")
	}
	defer C.free(unsafe.Pointer(cValue))

	value := C.GoString(cValue)
	return value, nil
}

func (s *SecureStorage) deletemacOSCredential(account string) error {
	cService := C.CString(s.serviceName)
	defer C.free(unsafe.Pointer(cService))

	cAccount := C.CString(account)
	defer C.free(unsafe.Pointer(cAccount))

	status := C.macOSDeleteKeychain(cService, cAccount)
	if status != 0 && status != -25300 { // -25300 = errSecItemNotFound
		return fmt.Errorf("macOS Keychain delete failed with status: %d", status)
	}
	return nil
}

// =====================================================
// Windows Implementation (using encrypted file fallback)
// TODO: Implement using Win32 CredRead/CredWrite via cgo
// =====================================================

func (s *SecureStorage) storeWindowsCredential(account, value string) error {
	// Fallback to file storage for now
	// In production, use Windows Credential Manager via Win32 API
	return s.storeFileCredential(account, value)
}

func (s *SecureStorage) getWindowsCredential(account string) (string, error) {
	return s.getFileCredential(account)
}

func (s *SecureStorage) deleteWindowsCredential(account string) error {
	return s.deleteFileCredential(account)
}

// =====================================================
// Linux/Other Implementation (encrypted file fallback)
// =====================================================

func (s *SecureStorage) storeFileCredential(account, value string) error {
	if s.configDir == "" {
		return fmt.Errorf("config directory not set for secure storage")
	}

	// Create secure directory
	secureDir := filepath.Join(s.configDir, "secure")
	if err := os.MkdirAll(secureDir, 0700); err != nil {
		return fmt.Errorf("failed to create secure directory: %w", err)
	}

	// Sanitize account name for filename
	safeAccount := strings.ReplaceAll(account, "/", "_")
	safeAccount = strings.ReplaceAll(safeAccount, "\\", "_")
	safeAccount = strings.ReplaceAll(safeAccount, "..", "_")

	credFile := filepath.Join(secureDir, safeAccount+".cred")

	// Encrypt the value before storing
	machineKey := GetMachineKey(getMachineIdentifier())
	encrypted, err := EncryptString(value, string(machineKey))
	if err != nil {
		return fmt.Errorf("failed to encrypt credential: %w", err)
	}

	// Write with restrictive permissions
	if err := os.WriteFile(credFile, []byte(encrypted), 0600); err != nil {
		return fmt.Errorf("failed to write credential file: %w", err)
	}

	return nil
}

func (s *SecureStorage) getFileCredential(account string) (string, error) {
	if s.configDir == "" {
		return "", fmt.Errorf("config directory not set for secure storage")
	}

	safeAccount := strings.ReplaceAll(account, "/", "_")
	safeAccount = strings.ReplaceAll(safeAccount, "\\", "_")
	safeAccount = strings.ReplaceAll(safeAccount, "..", "_")

	credFile := filepath.Join(s.configDir, "secure", safeAccount+".cred")

	data, err := os.ReadFile(credFile)
	if err != nil {
		return "", fmt.Errorf("credential not found")
	}

	encrypted := string(data)
	machineKey := GetMachineKey(getMachineIdentifier())
	value, err := DecryptString(encrypted, string(machineKey))
	if err != nil {
		return "", fmt.Errorf("failed to decrypt credential: %w", err)
	}

	return value, nil
}

func (s *SecureStorage) deleteFileCredential(account string) error {
	if s.configDir == "" {
		return fmt.Errorf("config directory not set for secure storage")
	}

	safeAccount := strings.ReplaceAll(account, "/", "_")
	safeAccount = strings.ReplaceAll(safeAccount, "\\", "_")
	safeAccount = strings.ReplaceAll(safeAccount, "..", "_")

	credFile := filepath.Join(s.configDir, "secure", safeAccount+".cred")

	if err := os.Remove(credFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete credential file: %w", err)
	}

	return nil
}

// =====================================================
// Machine Identifier Helper
// =====================================================

// getMachineIdentifier returns a platform-specific machine identifier.
// Used as part of the encryption key for file-based credential storage.
func getMachineIdentifier() string {
	switch runtime.GOOS {
	case "darwin":
		return getmacOSMachineID()
	case "windows":
		return getWindowsMachineID()
	default:
		return getLinuxMachineID()
	}
}

// getmacOSMachineID returns the macOS hardware UUID.
func getmacOSMachineID() string {
	// Use platform-specific hardware UUID as machine identifier
	// This is a unique identifier for the Mac hardware
	// In production, read from IOKit or use system_profiler
	hostname, _ := os.Hostname()
	return "macos:" + hostname
}

// getWindowsMachineID returns the Windows MachineGUID.
func getWindowsMachineID() string {
	// In production, read from Registry:
	// HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Cryptography\MachineGuid
	hostname, _ := os.Hostname()
	return "windows:" + hostname
}

// getLinuxMachineID returns a Linux machine identifier.
func getLinuxMachineID() string {
	// Try to read machine-id from /etc/machine-id (systemd)
	// or fallback to /var/lib/dbus/machine-id
	// or fallback to hostname
	if data, err := os.ReadFile("/etc/machine-id"); err == nil {
		return "linux:" + strings.TrimSpace(string(data))
	}
	if data, err := os.ReadFile("/var/lib/dbus/machine-id"); err == nil {
		return "linux:" + strings.TrimSpace(string(data))
	}
	hostname, _ := os.Hostname()
	return "linux:" + hostname
}
