// Package sync provides unit tests for S3 client.
// T148: Unit test for S3 client upload/download.
package sync

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestS3ClientUpload tests the S3 Upload method.
func TestS3ClientUpload(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}

		// Verify required headers
		if r.Header.Get("X-Amz-Date") == "" {
			t.Error("Missing X-Amz-Date header")
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("Missing Authorization header")
		}

		// Return success
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create S3 client config
	config := &S3Config{
		Endpoint:       server.URL,
		BucketName:     "test-bucket",
		AccessKey:      "test-access-key",
		SecretKey:      "test-secret-key",
		Region:         "us-east-1",
		ForcePathStyle: true,
	}

	// Create client
	client := NewS3Client(config)

	// Test upload
	ctx := context.Background()
	data := []byte("test content")
	err := client.Upload(ctx, "test-key", data)

	if err != nil {
		t.Errorf("Upload failed: %v", err)
	}
}

// TestS3ClientDownload tests the S3 Download method.
func TestS3ClientDownload(t *testing.T) {
	// Test data
	testData := []byte("downloaded content")

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Return test data
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		w.Write(testData)
	}))
	defer server.Close()

	// Create S3 client config
	config := &S3Config{
		Endpoint:       server.URL,
		BucketName:     "test-bucket",
		AccessKey:      "test-access-key",
		SecretKey:      "test-secret-key",
		Region:         "us-east-1",
		ForcePathStyle: true,
	}

	// Create client
	client := NewS3Client(config)

	// Test download
	ctx := context.Background()
	data, err := client.Download(ctx, "test-key")

	if err != nil {
		t.Errorf("Download failed: %v", err)
	}

	if string(data) != string(testData) {
		t.Errorf("Expected %s, got %s", testData, data)
	}
}

// TestS3ClientDownloadNotFound tests the S3 Download method with 404 response.
func TestS3ClientDownloadNotFound(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Create S3 client config
	config := &S3Config{
		Endpoint:       server.URL,
		BucketName:     "test-bucket",
		AccessKey:      "test-access-key",
		SecretKey:      "test-secret-key",
		Region:         "us-east-1",
		ForcePathStyle: true,
	}

	// Create client
	client := NewS3Client(config)

	// Test download
	ctx := context.Background()
	_, err := client.Download(ctx, "nonexistent-key")

	if err == nil {
		t.Error("Expected error for nonexistent key")
	}
}

// TestS3ClientDelete tests the S3 Delete method.
func TestS3ClientDelete(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}

		// Return success (S3 returns 204 NoContent or 200 OK)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Create S3 client config
	config := &S3Config{
		Endpoint:       server.URL,
		BucketName:     "test-bucket",
		AccessKey:      "test-access-key",
		SecretKey:      "test-secret-key",
		Region:         "us-east-1",
		ForcePathStyle: true,
	}

	// Create client
	client := NewS3Client(config)

	// Test delete
	ctx := context.Background()
	err := client.Delete(ctx, "test-key")

	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}
}

// TestS3ClientList tests the S3 List method.
func TestS3ClientList(t *testing.T) {
	// Create test server with XML response
	xmlResponse := `<?xml version="1.0" encoding="UTF-8"?>
<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <Name>test-bucket</Name>
  <Prefix>items/</Prefix>
  <Contents>
    <Key>items/file1.json</Key>
    <LastModified>2024-01-01T00:00:00.000Z</LastModified>
    <Size>1024</Size>
  </Contents>
  <Contents>
    <Key>items/file2.json</Key>
    <LastModified>2024-01-02T00:00:00.000Z</LastModified>
    <Size>2048</Size>
  </Contents>
</ListBucketResult>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, xmlResponse)
	}))
	defer server.Close()

	// Create S3 client config
	config := &S3Config{
		Endpoint:       server.URL,
		BucketName:     "test-bucket",
		AccessKey:      "test-access-key",
		SecretKey:      "test-secret-key",
		Region:         "us-east-1",
		ForcePathStyle: true,
	}

	// Create client
	client := NewS3Client(config)

	// Test list
	ctx := context.Background()
	keys, err := client.List(ctx, "items/")

	if err != nil {
		t.Errorf("List failed: %v", err)
	}

	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(keys))
	}

	if keys[0] != "items/file1.json" {
		t.Errorf("Expected items/file1.json, got %s", keys[0])
	}
}

// TestS3ClientTimeout tests client timeout behavior.
func TestS3ClientTimeout(t *testing.T) {
	// Create test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep longer than client timeout
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create S3 client config with short timeout
	config := &S3Config{
		Endpoint:       server.URL,
		BucketName:     "test-bucket",
		AccessKey:      "test-access-key",
		SecretKey:      "test-secret-key",
		Region:         "us-east-1",
		ForcePathStyle: true,
	}

	// Create client
	client := NewS3Client(config)

	// Test with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := client.Upload(ctx, "test-key", []byte("test"))
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

// TestS3ClientConnectionError tests connection error handling.
func TestS3ClientConnectionError(t *testing.T) {
	// Create S3 client config with invalid endpoint
	config := &S3Config{
		Endpoint:       "http://invalid-endpoint-that-does-not-exist:12345",
		BucketName:     "test-bucket",
		AccessKey:      "test-access-key",
		SecretKey:      "test-secret-key",
		Region:         "us-east-1",
		ForcePathStyle: true,
	}

	// Create client
	client := NewS3Client(config)

	// Test list with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := client.List(ctx, "")
	if err == nil {
		t.Error("Expected connection error, got nil")
	}
}
