// Package s3 provides unit tests for MinIO provider.
// T154: Unit test for MinIO provider configuration.
package s3

import (
	"testing"
)

// TestNewMinIOClient tests creating a MinIO client.
func TestNewMinIOClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *MinIOConfig
		wantNil bool
	}{
		{
			name: "HTTP endpoint",
			config: &MinIOConfig{
				Endpoint:   "localhost:9000",
				BucketName: "test-bucket",
				AccessKey:  "minioadmin",
				SecretKey:  "minioadmin",
				UseSSL:     false,
			},
			wantNil: false,
		},
		{
			name: "HTTPS endpoint",
			config: &MinIOConfig{
				Endpoint:   "minio.example.com",
				BucketName: "test-bucket",
				AccessKey:  "access_key",
				SecretKey:  "secret_key",
				UseSSL:     true,
			},
			wantNil: false,
		},
		{
			name: "endpoint with scheme",
			config: &MinIOConfig{
				Endpoint:   "https://minio.example.com",
				BucketName: "test-bucket",
				AccessKey:  "minioadmin",
				SecretKey:  "minioadmin",
				UseSSL:     true,
			},
			wantNil: false,
		},
		{
			name: "endpoint with trailing slash",
			config: &MinIOConfig{
				Endpoint:   "localhost:9000/",
				BucketName: "test-bucket",
				AccessKey:  "minioadmin",
				SecretKey:  "minioadmin",
				UseSSL:     false,
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewMinIOClient(tt.config)
			if (client == nil) != tt.wantNil {
				t.Errorf("NewMinIOClient() = %v, wantNil %v", client, tt.wantNil)
			}
		})
	}
}

// TestMinIODefaultCredentials tests getting default MinIO credentials.
func TestMinIODefaultCredentials(t *testing.T) {
	accessKey, secretKey := MinIODefaultCredentials()

	if accessKey != "minioadmin" {
		t.Errorf("Expected access key 'minioadmin', got %s", accessKey)
	}

	if secretKey != "minioadmin" {
		t.Errorf("Expected secret key 'minioadmin', got %s", secretKey)
	}
}

// TestMinIOLocalEndpoint tests getting default local endpoint.
func TestMinIOLocalEndpoint(t *testing.T) {
	endpoint := MinIOLocalEndpoint()

	if endpoint != "localhost:9000" {
		t.Errorf("Expected endpoint 'localhost:9000', got %s", endpoint)
	}
}

// TestMinIOHealthCheckURL tests generating health check URLs.
func TestMinIOHealthCheckURL(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		useSSL   bool
		expected string
	}{
		{
			name:     "HTTP localhost",
			endpoint: "localhost:9000",
			useSSL:   false,
			expected: "http://localhost:9000/minio/health/live",
		},
		{
			name:     "HTTPS remote",
			endpoint: "minio.example.com",
			useSSL:   true,
			expected: "https://minio.example.com/minio/health/live",
		},
		{
			name:     "with scheme",
			endpoint: "https://minio.example.com",
			useSSL:   true,
			expected: "https://minio.example.com/minio/health/live",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := MinIOHealthCheckURL(tt.endpoint, tt.useSSL)
			if url != tt.expected {
				t.Errorf("Expected URL %s, got %s", tt.expected, url)
			}
		})
	}
}

// TestIsMinIOEndpoint tests MinIO endpoint detection.
func TestIsMinIOEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		expected bool
	}{
		{"contains minio", "minio.example.com", true},
		{"port 9000", "localhost:9000", true},
		{"port 9001", "storage.example.com:9001", true},
		{"AWS S3", "s3.amazonaws.com", false},
		{"AWS regional", "s3.us-west-2.amazonaws.com", false},
		{"Cloudflare R2", "abc123.r2.cloudflarestorage.com", false},
		{"generic endpoint", "storage.example.com", false},
		{"localhost other port", "localhost:8080", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMinIOEndpoint(tt.endpoint)
			if result != tt.expected {
				t.Errorf("IsMinIOEndpoint(%s) = %v, want %v", tt.endpoint, result, tt.expected)
			}
		})
	}
}

// TestParseMinIOEndpoint tests parsing and validating MinIO endpoints.
func TestParseMinIOEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		useSSL   bool
		expected string
		wantErr  bool
	}{
		{
			name:     "HTTP without scheme",
			endpoint: "localhost:9000",
			useSSL:   false,
			expected: "http://localhost:9000",
			wantErr:  false,
		},
		{
			name:     "HTTPS without scheme",
			endpoint: "minio.example.com",
			useSSL:   true,
			expected: "https://minio.example.com",
			wantErr:  false,
		},
		{
			name:     "with HTTP scheme",
			endpoint: "http://localhost:9000",
			useSSL:   false,
			expected: "http://localhost:9000",
			wantErr:  false,
		},
		{
			name:     "with HTTPS scheme",
			endpoint: "https://minio.example.com",
			useSSL:   true,
			expected: "https://minio.example.com",
			wantErr:  false,
		},
		{
			name:     "with trailing slash",
			endpoint: "localhost:9000/",
			useSSL:   false,
			expected: "http://localhost:9000",
			wantErr:  false,
		},
		{
			name:     "empty endpoint",
			endpoint: "",
			useSSL:   false,
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMinIOEndpoint(tt.endpoint, tt.useSSL)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error for invalid endpoint")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected endpoint %s, got %s", tt.expected, result)
			}
		})
	}
}
