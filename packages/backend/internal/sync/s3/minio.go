// Package s3 provides S3-compatible storage provider implementations.
// T154: MinIO provider for self-hosted S3-compatible storage.
package s3

import (
	"fmt"
	"strings"

	"github.com/kimhsiao/memonexus/backend/internal/sync"
)

// MinIOConfig holds MinIO-specific configuration.
type MinIOConfig struct {
	Endpoint  string // MinIO server endpoint (e.g., "localhost:9000" or "https://minio.example.com")
	BucketName string
	AccessKey  string // MinIO Root Username or access key
	SecretKey  string // MinIO Root Password or secret key
	UseSSL    bool   // Use HTTPS connection (default: false for development)
}

// NewMinIOClient creates an S3 client configured for MinIO.
// T154: MinIO requires path-style URLs (endpoint/bucket/key).
//
// Example:
//
//	client := NewMinIOClient(&MinIOConfig{
//	    Endpoint:  "localhost:9000",
//	    BucketName: "my-bucket",
//	    AccessKey:  "minioadmin",
//	    SecretKey:  "minioadmin",
//	    UseSSL:    false, // For local development
//	})
//
// For production with TLS:
//
//	client := NewMinIOClient(&MinIOConfig{
//	    Endpoint:  "minio.example.com",
//	    BucketName: "my-bucket",
//	    AccessKey:  "access_key",
//	    SecretKey:  "secret_key",
//	    UseSSL:    true,
//	})
func NewMinIOClient(config *MinIOConfig) *sync.S3Client {
	// Build endpoint with proper scheme
	endpoint := config.Endpoint

	// Add scheme if not present
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		if config.UseSSL {
			endpoint = "https://" + endpoint
		} else {
			endpoint = "http://" + endpoint
		}
	}

	// Remove trailing slash for consistency
	endpoint = strings.TrimSuffix(endpoint, "/")

	return sync.NewS3Client(&sync.S3Config{
		Endpoint:       endpoint,
		BucketName:     config.BucketName,
		AccessKey:      config.AccessKey,
		SecretKey:      config.SecretKey,
		Region:         "us-east-1", // MinIO doesn't use regions, default required
		ForcePathStyle: true, // Path-style URLs required for MinIO
	})
}

// MinIODefaultCredentials returns the default MinIO credentials.
// WARNING: Only use for local development, never in production.
func MinIODefaultCredentials() (accessKey, secretKey string) {
	return "minioadmin", "minioadmin"
}

// MinIOLocalEndpoint returns the default local MinIO endpoint.
func MinIOLocalEndpoint() string {
	return "localhost:9000"
}

// MinIOHealthCheckURL returns the health check URL for a MinIO server.
func MinIOHealthCheckURL(endpoint string, useSSL bool) string {
	baseEndpoint := endpoint
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		if useSSL {
			baseEndpoint = "https://" + endpoint
		} else {
			baseEndpoint = "http://" + endpoint
		}
	}
	return fmt.Sprintf("%s/minio/health/live", baseEndpoint)
}

// IsMinIOEndpoint checks if the endpoint appears to be a MinIO server.
// This is a heuristic check based on common MinIO deployment patterns.
func IsMinIOEndpoint(endpoint string) bool {
	lowerEndpoint := strings.ToLower(endpoint)

	// Common MinIO patterns
	minioIndicators := []string{
		"minio",
		":9000",
		":9001",
		"s3.amazonaws.com", // Not MinIO (exclude)
	}

	hasMinioIndicator := false
	for _, indicator := range minioIndicators {
		if strings.Contains(lowerEndpoint, indicator) {
			if indicator == "s3.amazonaws.com" {
				return false // Definitely not MinIO
			}
			hasMinioIndicator = true
		}
	}

	return hasMinioIndicator
}

// ParseMinIOEndpoint parses and validates a MinIO endpoint.
// Returns the endpoint with proper scheme.
func ParseMinIOEndpoint(endpoint string, useSSL bool) (string, error) {
	if endpoint == "" {
		return "", fmt.Errorf("endpoint cannot be empty")
	}

	// Add scheme if not present
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		if useSSL {
			endpoint = "https://" + endpoint
		} else {
			endpoint = "http://" + endpoint
		}
	}

	// Remove trailing slash
	endpoint = strings.TrimSuffix(endpoint, "/")

	return endpoint, nil
}
