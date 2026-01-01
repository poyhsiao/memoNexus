// Package models tests for data model definitions.
package models

import (
	"database/sql/driver"
	"encoding/json"
	"testing"
	"time"
)

// =====================================================
// UUID Type Tests
// =====================================================

// TestUUID_Value verifies the Value() method returns correct string.
func TestUUID_Value(t *testing.T) {
	uuid := UUID("123e4567-e89b-12d3-a456-426614174000")

	val, err := uuid.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}

	if val != "123e4567-e89b-12d3-a456-426614174000" {
		t.Errorf("Value() = %v, want '123e4567-e89b-12d3-a456-426614174000'", val)
	}
}

// TestUUID_Scan_nil verifies nil value handling.
func TestUUID_Scan_nil(t *testing.T) {
	var uuid UUID
	err := uuid.Scan(nil)

	if err != nil {
		t.Fatalf("Scan(nil) error = %v", err)
	}

	if uuid != "" {
		t.Errorf("Scan(nil) = %q, want empty string", uuid)
	}
}

// TestUUID_Scan_bytes verifies []byte handling.
func TestUUID_Scan_bytes(t *testing.T) {
	var uuid UUID
	input := []byte("123e4567-e89b-12d3-a456-426614174000")

	err := uuid.Scan(input)
	if err != nil {
		t.Fatalf("Scan([]byte) error = %v", err)
	}

	if uuid != "123e4567-e89b-12d3-a456-426614174000" {
		t.Errorf("Scan([]byte) = %q, want '123e4567-e89b-12d3-a456-426614174000'", uuid)
	}
}

// TestUUID_Scan_string verifies string handling.
func TestUUID_Scan_string(t *testing.T) {
	var uuid UUID
	input := "123e4567-e89b-12d3-a456-426614174000"

	err := uuid.Scan(input)
	if err != nil {
		t.Fatalf("Scan(string) error = %v", err)
	}

	if uuid != "123e4567-e89b-12d3-a456-426614174000" {
		t.Errorf("Scan(string) = %q, want '123e4567-e89b-12d3-a456-426614174000'", uuid)
	}
}

// TestUUID_Scan_invalidType verifies error for invalid types.
func TestUUID_Scan_invalidType(t *testing.T) {
	var uuid UUID
	err := uuid.Scan(12345) // int is invalid

	if err == nil {
		t.Error("Scan(int) should return error")
	}
}

// TestUUID_Scan_invalidLength verifies error for invalid UUID length.
func TestUUID_Scan_invalidLength(t *testing.T) {
	var uuid UUID
	err := uuid.Scan("too-short")

	if err == nil {
		t.Error("Scan(too-short) should return error for invalid length")
	}
}

// TestUUID_String verifies String() method.
func TestUUID_String(t *testing.T) {
	uuid := UUID("test-uuid-string")
	if uuid.String() != "test-uuid-string" {
		t.Errorf("String() = %q, want 'test-uuid-string'", uuid.String())
	}
}

// =====================================================
// ContentItem Tests
// =====================================================

// TestContentItem_TableName verifies table name.
func TestContentItem_TableName(t *testing.T) {
	c := ContentItem{}
	if c.TableName() != "content_items" {
		t.Errorf("TableName() = %q, want 'content_items'", c.TableName())
	}
}

// TestContentItem_CreatedAtTime verifies timestamp conversion.
func TestContentItem_CreatedAtTime(t *testing.T) {
	expected := time.Unix(1609459200, 0) // 2021-01-01 00:00:00 UTC
	c := ContentItem{CreatedAt: 1609459200}

	result := c.CreatedAtTime()
	if !result.Equal(expected) {
		t.Errorf("CreatedAtTime() = %v, want %v", result, expected)
	}
}

// TestContentItem_UpdatedAtTime verifies timestamp conversion.
func TestContentItem_UpdatedAtTime(t *testing.T) {
	expected := time.Unix(1609459200, 0)
	c := ContentItem{UpdatedAt: 1609459200}

	result := c.UpdatedAtTime()
	if !result.Equal(expected) {
		t.Errorf("UpdatedAtTime() = %v, want %v", result, expected)
	}
}

// TestContentItem_Touch verifies Touch() updates timestamp and version.
func TestContentItem_Touch(t *testing.T) {
	c := ContentItem{
		UpdatedAt: 1609459200,
		Version:   1,
	}

	before := time.Now().Unix()
	c.Touch()
	after := time.Now().Unix()

	// UpdatedAt should be updated to current time
	if c.UpdatedAt < before || c.UpdatedAt > after {
		t.Errorf("Touch() UpdatedAt = %d, want between %d and %d", c.UpdatedAt, before, after)
	}

	// Version should be incremented
	if c.Version != 2 {
		t.Errorf("Touch() Version = %d, want 2", c.Version)
	}
}

// =====================================================
// Tag Tests
// =====================================================

// TestTag_TableName verifies table name.
func TestTag_TableName(t *testing.T) {
	tag := Tag{}
	if tag.TableName() != "tags" {
		t.Errorf("TableName() = %q, want 'tags'", tag.TableName())
	}
}

// TestTag_CreatedAtTime verifies timestamp conversion.
func TestTag_CreatedAtTime(t *testing.T) {
	expected := time.Unix(1609459200, 0)
	tag := Tag{CreatedAt: 1609459200}

	result := tag.CreatedAtTime()
	if !result.Equal(expected) {
		t.Errorf("CreatedAtTime() = %v, want %v", result, expected)
	}
}

// TestTag_UpdatedAtTime verifies timestamp conversion.
func TestTag_UpdatedAtTime(t *testing.T) {
	expected := time.Unix(1609459200, 0)
	tag := Tag{UpdatedAt: 1609459200}

	result := tag.UpdatedAtTime()
	if !result.Equal(expected) {
		t.Errorf("UpdatedAtTime() = %v, want %v", result, expected)
	}
}

// TestTag_Touch verifies Touch() updates timestamp.
func TestTag_Touch(t *testing.T) {
	tag := Tag{UpdatedAt: 1609459200}

	before := time.Now().Unix()
	tag.Touch()
	after := time.Now().Unix()

	if tag.UpdatedAt < before || tag.UpdatedAt > after {
		t.Errorf("Touch() UpdatedAt = %d, want between %d and %d", tag.UpdatedAt, before, after)
	}
}

// =====================================================
// AIConfig Tests
// =====================================================

// TestAIConfig_TableName verifies table name.
func TestAIConfig_TableName(t *testing.T) {
	config := AIConfig{}
	if config.TableName() != "ai_config" {
		t.Errorf("TableName() = %q, want 'ai_config'", config.TableName())
	}
}

// TestAIConfig_CreatedAtTime verifies timestamp conversion.
func TestAIConfig_CreatedAtTime(t *testing.T) {
	expected := time.Unix(1609459200, 0)
	config := AIConfig{CreatedAt: 1609459200}

	result := config.CreatedAtTime()
	if !result.Equal(expected) {
		t.Errorf("CreatedAtTime() = %v, want %v", result, expected)
	}
}

// TestAIConfig_UpdatedAtTime verifies timestamp conversion.
func TestAIConfig_UpdatedAtTime(t *testing.T) {
	expected := time.Unix(1609459200, 0)
	config := AIConfig{UpdatedAt: 1609459200}

	result := config.UpdatedAtTime()
	if !result.Equal(expected) {
		t.Errorf("UpdatedAtTime() = %v, want %v", result, expected)
	}
}

// TestAIConfig_SetAPIKey verifies API key encryption.
func TestAIConfig_SetAPIKey(t *testing.T) {
	config := AIConfig{}
	apiKey := "sk-test-key-12345"
	machineID := "test-machine"

	err := config.SetAPIKey(apiKey, machineID)
	if err != nil {
		t.Fatalf("SetAPIKey() error = %v", err)
	}

	if config.APIKeyEncrypted == "" {
		t.Error("SetAPIKey() should set APIKeyEncrypted")
	}

	if config.APIKeyEncrypted == apiKey {
		t.Error("SetAPIKey() should encrypt the key, not store plaintext")
	}
}

// TestAIConfig_GetAPIKey verifies API key decryption.
func TestAIConfig_GetAPIKey(t *testing.T) {
	config := AIConfig{}
	apiKey := "sk-test-key-12345"
	machineID := "test-machine"

	// Set the key first
	err := config.SetAPIKey(apiKey, machineID)
	if err != nil {
		t.Fatalf("SetAPIKey() error = %v", err)
	}

	// Get it back
	retrieved, err := config.GetAPIKey(machineID)
	if err != nil {
		t.Fatalf("GetAPIKey() error = %v", err)
	}

	if retrieved != apiKey {
		t.Errorf("GetAPIKey() = %q, want %q", retrieved, apiKey)
	}
}

// TestAIConfig_GetAPIKey_empty verifies empty encrypted key returns empty.
func TestAIConfig_GetAPIKey_empty(t *testing.T) {
	config := AIConfig{}

	retrieved, err := config.GetAPIKey("test-machine")
	if err != nil {
		t.Fatalf("GetAPIKey() error = %v", err)
	}

	if retrieved != "" {
		t.Errorf("GetAPIKey() with empty encrypted = %q, want empty", retrieved)
	}
}

// TestAIConfig_HasAPIKey verifies HasAPIKey() method.
func TestAIConfig_HasAPIKey(t *testing.T) {
	config := AIConfig{}

	if config.HasAPIKey() {
		t.Error("HasAPIKey() on empty config should return false")
	}

	config.APIKeyEncrypted = "encrypted-value"
	if !config.HasAPIKey() {
		t.Error("HasAPIKey() with encrypted value should return true")
	}
}

// =====================================================
// SyncCredential Tests
// =====================================================

// TestSyncCredential_TableName verifies table name.
func TestSyncCredential_TableName(t *testing.T) {
	cred := SyncCredential{}
	if cred.TableName() != "sync_credentials" {
		t.Errorf("TableName() = %q, want 'sync_credentials'", cred.TableName())
	}
}

// TestSyncCredential_CreatedAtTime verifies timestamp conversion.
func TestSyncCredential_CreatedAtTime(t *testing.T) {
	expected := time.Unix(1609459200, 0)
	cred := SyncCredential{CreatedAt: 1609459200}

	result := cred.CreatedAtTime()
	if !result.Equal(expected) {
		t.Errorf("CreatedAtTime() = %v, want %v", result, expected)
	}
}

// TestSyncCredential_UpdatedAtTime verifies timestamp conversion.
func TestSyncCredential_UpdatedAtTime(t *testing.T) {
	expected := time.Unix(1609459200, 0)
	cred := SyncCredential{UpdatedAt: 1609459200}

	result := cred.UpdatedAtTime()
	if !result.Equal(expected) {
		t.Errorf("UpdatedAtTime() = %v, want %v", result, expected)
	}
}

// TestSyncCredential_SetAccessKey verifies access key encryption.
func TestSyncCredential_SetAccessKey(t *testing.T) {
	cred := SyncCredential{}
	accessKey := "AKIAIOSFODNN7EXAMPLE"
	machineID := "test-machine"

	err := cred.SetAccessKey(accessKey, machineID)
	if err != nil {
		t.Fatalf("SetAccessKey() error = %v", err)
	}

	if cred.AccessKeyEncrypted == "" {
		t.Error("SetAccessKey() should set AccessKeyEncrypted")
	}

	if cred.AccessKeyEncrypted == accessKey {
		t.Error("SetAccessKey() should encrypt the key")
	}
}

// TestSyncCredential_GetAccessKey verifies access key decryption.
func TestSyncCredential_GetAccessKey(t *testing.T) {
	cred := SyncCredential{}
	accessKey := "AKIAIOSFODNN7EXAMPLE"
	machineID := "test-machine"

	err := cred.SetAccessKey(accessKey, machineID)
	if err != nil {
		t.Fatalf("SetAccessKey() error = %v", err)
	}

	retrieved, err := cred.GetAccessKey(machineID)
	if err != nil {
		t.Fatalf("GetAccessKey() error = %v", err)
	}

	if retrieved != accessKey {
		t.Errorf("GetAccessKey() = %q, want %q", retrieved, accessKey)
	}
}

// TestSyncCredential_GetAccessKey_empty verifies empty encrypted key returns empty.
func TestSyncCredential_GetAccessKey_empty(t *testing.T) {
	cred := SyncCredential{}

	retrieved, err := cred.GetAccessKey("test-machine")
	if err != nil {
		t.Fatalf("GetAccessKey() error = %v", err)
	}

	if retrieved != "" {
		t.Errorf("GetAccessKey() with empty encrypted = %q, want empty", retrieved)
	}
}

// TestSyncCredential_SetSecretKey verifies secret key encryption.
func TestSyncCredential_SetSecretKey(t *testing.T) {
	cred := SyncCredential{}
	secretKey := "secret-key-12345"
	machineID := "test-machine"

	err := cred.SetSecretKey(secretKey, machineID)
	if err != nil {
		t.Fatalf("SetSecretKey() error = %v", err)
	}

	if cred.SecretKeyEncrypted == "" {
		t.Error("SetSecretKey() should set SecretKeyEncrypted")
	}

	if cred.SecretKeyEncrypted == secretKey {
		t.Error("SetSecretKey() should encrypt the key")
	}
}

// TestSyncCredential_GetSecretKey verifies secret key decryption.
func TestSyncCredential_GetSecretKey(t *testing.T) {
	cred := SyncCredential{}
	secretKey := "secret-key-12345"
	machineID := "test-machine"

	err := cred.SetSecretKey(secretKey, machineID)
	if err != nil {
		t.Fatalf("SetSecretKey() error = %v", err)
	}

	retrieved, err := cred.GetSecretKey(machineID)
	if err != nil {
		t.Fatalf("GetSecretKey() error = %v", err)
	}

	if retrieved != secretKey {
		t.Errorf("GetSecretKey() = %q, want %q", retrieved, secretKey)
	}
}

// TestSyncCredential_GetSecretKey_empty verifies empty encrypted key returns empty.
func TestSyncCredential_GetSecretKey_empty(t *testing.T) {
	cred := SyncCredential{}

	retrieved, err := cred.GetSecretKey("test-machine")
	if err != nil {
		t.Fatalf("GetSecretKey() error = %v", err)
	}

	if retrieved != "" {
		t.Errorf("GetSecretKey() with empty encrypted = %q, want empty", retrieved)
	}
}

// TestSyncCredential_HasCredentials verifies HasCredentials() method.
func TestSyncCredential_HasCredentials(t *testing.T) {
	cred := SyncCredential{}

	if cred.HasCredentials() {
		t.Error("HasCredentials() on empty config should return false")
	}

	// Only access key set
	cred.AccessKeyEncrypted = "encrypted-access"
	if cred.HasCredentials() {
		t.Error("HasCredentials() with only access key should return false")
	}

	// Both keys set
	cred.SecretKeyEncrypted = "encrypted-secret"
	if !cred.HasCredentials() {
		t.Error("HasCredentials() with both keys should return true")
	}
}

// =====================================================
// ChangeLog Tests
// =====================================================

// TestChangeLog_TableName verifies table name.
func TestChangeLog_TableName(t *testing.T) {
	log := ChangeLog{}
	if log.TableName() != "change_log" {
		t.Errorf("TableName() = %q, want 'change_log'", log.TableName())
	}
}

// TestChangeLog_Time verifies timestamp conversion.
func TestChangeLog_Time(t *testing.T) {
	expected := time.Unix(1609459200, 0)
	log := ChangeLog{Timestamp: 1609459200}

	result := log.Time()
	if !result.Equal(expected) {
		t.Errorf("Time() = %v, want %v", result, expected)
	}
}

// =====================================================
// ConflictLog Tests
// =====================================================

// TestConflictLog_TableName verifies table name.
func TestConflictLog_TableName(t *testing.T) {
	log := ConflictLog{}
	if log.TableName() != "conflict_log" {
		t.Errorf("TableName() = %q, want 'conflict_log'", log.TableName())
	}
}

// TestConflictLog_DetectedAtTime verifies timestamp conversion.
func TestConflictLog_DetectedAtTime(t *testing.T) {
	expected := time.Unix(1609459200, 0)
	log := ConflictLog{DetectedAt: 1609459200}

	result := log.DetectedAtTime()
	if !result.Equal(expected) {
		t.Errorf("DetectedAtTime() = %v, want %v", result, expected)
	}
}

// =====================================================
// SyncQueue Tests
// =====================================================

// TestSyncQueue_TableName verifies table name.
func TestSyncQueue_TableName(t *testing.T) {
	queue := SyncQueue{}
	if queue.TableName() != "sync_queue" {
		t.Errorf("TableName() = %q, want 'sync_queue'", queue.TableName())
	}
}

// TestSyncQueue_Payload verifies JSON payload handling.
func TestSyncQueue_Payload(t *testing.T) {
	payloadData := map[string]interface{}{
		"item_id": "123",
		"action":  "upload",
	}
	payloadBytes, _ := json.Marshal(payloadData)

	queue := SyncQueue{
		Payload: json.RawMessage(payloadBytes),
	}

	if queue.Payload == nil {
		t.Error("Payload should be set")
	}

	// Verify payload can be unmarshaled
	var result map[string]interface{}
	err := json.Unmarshal(queue.Payload, &result)
	if err != nil {
		t.Errorf("Failed to unmarshal payload: %v", err)
	}
}

// =====================================================
// ExportArchive Tests
// =====================================================

// TestExportArchive_TableName verifies table name.
func TestExportArchive_TableName(t *testing.T) {
	archive := ExportArchive{}
	if archive.TableName() != "export_archives" {
		t.Errorf("TableName() = %q, want 'export_archives'", archive.TableName())
	}
}

// TestExportArchive_CreatedAtTime verifies timestamp conversion.
func TestExportArchive_CreatedAtTime(t *testing.T) {
	expected := time.Unix(1609459200, 0)
	archive := ExportArchive{CreatedAt: 1609459200}

	result := archive.CreatedAtTime()
	if !result.Equal(expected) {
		t.Errorf("CreatedAtTime() = %v, want %v", result, expected)
	}
}

// =====================================================
// Driver.Valuer and sql.Scanner Tests
// =====================================================

// TestUUID_Valuer verifies UUID implements driver.Valuer.
func TestUUID_Valuer(t *testing.T) {
	uuid := UUID("test-uuid")
	var _ driver.Valuer = uuid // Should compile

	val, err := uuid.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}

	if val != "test-uuid" {
		t.Errorf("Value() = %v, want 'test-uuid'", val)
	}
}

// TestUUID_Scanner verifies UUID implements sql.Scanner.
func TestUUID_Scanner_interface(t *testing.T) {
	var uuid UUID
	var scanner interface{ Scan(interface{}) error } = &uuid // Should compile

	err := scanner.Scan("test-uuid")
	if err != nil {
		// Invalid UUID format is expected to error
		t.Logf("Scan() correctly rejected invalid UUID: %v", err)
	}
}
