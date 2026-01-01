// Package s3 provides S3-compatible storage provider implementations.
// T153: Cloudflare R2 provider with S3-compatible API.
package s3

import (
	"fmt"
	"strings"

	"github.com/kimhsiao/memonexus/backend/internal/sync"
)

// R2Config holds Cloudflare R2-specific configuration.
type R2Config struct {
	AccountID  string // Cloudflare Account ID
	BucketName string
	AccessKey  string // R2 API Token (Access Key ID)
	SecretKey  string // R2 API Token (Secret Access Key)
}

// NewR2Client creates an S3 client configured for Cloudflare R2.
// T153: R2 uses S3-compatible API with account-specific endpoints.
//
// The R2 endpoint format is: https://<accountid>.r2.cloudflarestorage.com
//
// Example:
//
//	client := NewR2Client(&R2Config{
//	    AccountID:  "abc123...",
//	    BucketName: "my-bucket",
//	    AccessKey:  "access_key_id",
//	    SecretKey:  "secret_access_key",
//	})
func NewR2Client(config *R2Config) *sync.S3Client {
	// Validate account ID
	if config.AccountID == "" {
		panic("R2 AccountID is required")
	}

	// Build R2 endpoint: <accountid>.r2.cloudflarestorage.com
	endpoint := fmt.Sprintf("%s.r2.cloudflarestorage.com", config.AccountID)

	return sync.NewS3Client(&sync.S3Config{
		Endpoint:       endpoint,
		BucketName:     config.BucketName,
		AccessKey:      config.AccessKey,
		SecretKey:      config.SecretKey,
		Region:         "auto", // R2 doesn't use regions like AWS
		ForcePathStyle: false,  // Virtual-host style for R2
	})
}

// R2EndpointForAccount returns the R2 endpoint for a given account ID.
func R2EndpointForAccount(accountID string) string {
	return fmt.Sprintf("%s.r2.cloudflarestorage.com", accountID)
}

// IsValidAccountID performs basic validation of a Cloudflare Account ID.
// R2 Account IDs are typically 32-character hex strings.
func IsValidR2AccountID(accountID string) bool {
	// R2 account IDs are typically 32 hex characters
	if len(accountID) != 32 {
		return false
	}

	// Check if all hex characters
	for _, c := range accountID {
		if !strings.ContainsAny(string(c), "0123456789abcdefABCDEF") {
			return false
		}
	}

	return true
}

// R2PublicURL returns the public URL for an object in R2.
// This requires a custom domain to be configured for the bucket.
//
// Example:
//
//	// If you have custom domain "cdn.example.com" bound to bucket
//	publicURL := R2PublicURL("cdn.example.com", "items/file.json")
//	// Returns: https://cdn.example.com/items/file.json
func R2PublicURL(customDomain, key string) string {
	return fmt.Sprintf("https://%s/%s", customDomain, key)
}

// R2S3URL returns the S3 API URL for an object (for internal use, not public access).
func R2S3URL(accountID, bucket, key string) string {
	endpoint := R2EndpointForAccount(accountID)
	return fmt.Sprintf("https://%s.%s/%s", bucket, endpoint, key)
}
