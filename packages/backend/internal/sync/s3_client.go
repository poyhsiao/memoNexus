// Package sync provides S3-compatible storage client.
package sync

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// S3Config holds S3 connection configuration.
type S3Config struct {
	Endpoint        string
	BucketName      string
	AccessKey       string
	SecretKey       string
	Region          string
	ForcePathStyle  bool // Use path-style URLs (minio, localstack)
}

// S3Client implements ObjectStore for S3-compatible storage.
type S3Client struct {
	config     *S3Config
	httpClient *http.Client
}

// ListBucketResult represents the S3 ListObjectsV2 response.
type ListBucketResult struct {
	XMLName xml.Name `xml:"ListBucketResult"`
	Name    string   `xml:"Name"`
	Prefix  string   `xml:"Prefix"`
	Contents []struct {
		Key          string `xml:"Key"`
		LastModified string `xml:"LastModified"`
		Size         int64  `xml:"Size"`
	} `xml:"Contents"`
}

// LocationConstraint represents the S3 GetBucketLocation response.
type LocationConstraint struct {
	XMLName          xml.Name `xml:"LocationConstraint"`
	LocationConstraint string   `xml:",chardata"`
}

// NewS3Client creates a new S3Client.
func NewS3Client(config *S3Config) *S3Client {
	return &S3Client{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:       10,
				IdleConnTimeout:    30 * time.Second,
				DisableCompression: false,
			},
		},
	}
}

// Upload uploads data to S3.
func (c *S3Client) Upload(ctx context.Context, key string, data []byte) error {
	// Create PUT request
	req, err := c.createRequest(ctx, http.MethodPut, key, bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(data)))

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Download downloads data from S3.
func (c *S3Client) Download(ctx context.Context, key string) ([]byte, error) {
	// Create GET request
	req, err := c.createRequest(ctx, http.MethodGet, key, nil)
	if err != nil {
		return nil, err
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("object not found: %s", key)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}

// Delete deletes data from S3.
func (c *S3Client) Delete(ctx context.Context, key string) error {
	// Create DELETE request
	req, err := c.createRequest(ctx, http.MethodDelete, key, nil)
	if err != nil {
		return err
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// List lists all objects with a prefix.
func (c *S3Client) List(ctx context.Context, prefix string) ([]string, error) {
	// Create ListObjectsV2 request
	// Note: For bucket-level operations like ListObjectsV2, the key is empty
	// and query parameters are handled differently
	// Build query params without '?' prefix (buildSignedRequest handles that)
	queryParams := "list-type=2"
	if prefix != "" {
		queryParams += "&prefix=" + url.QueryEscape(prefix)
	}
	req, err := c.createRequestForBucket(ctx, http.MethodGet, queryParams)
	if err != nil {
		return nil, err
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse XML response
	var result ListBucketResult
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract keys
	var keys []string
	for _, content := range result.Contents {
		keys = append(keys, content.Key)
	}

	return keys, nil
}

// buildSignedRequest builds and signs an S3 request (unified for object and bucket operations).
func (c *S3Client) buildSignedRequest(
	ctx context.Context,
	method, canonicalURI, rawQuery string,
	body io.Reader,
) (*http.Request, error) {
	// Build base URL
	base := c.config.Endpoint
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "https://" + base
	}

	u, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("failed to parse endpoint: %w", err)
	}

	// Set path and query for virtual-host style
	if !c.config.ForcePathStyle {
		u.Host = fmt.Sprintf("%s.%s", c.config.BucketName, u.Host)
		u.Path = canonicalURI
		u.RawQuery = rawQuery
	} else {
		// Path-style: include bucket in path
		u.Path = canonicalURI
		u.RawQuery = rawQuery
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}

	// Set host header for virtual-host style
	if !c.config.ForcePathStyle {
		req.Host = fmt.Sprintf("%s.%s", c.config.BucketName, u.Host)
	}

	// Add AWS V4 signature headers
	timestamp := time.Now().UTC()
	amzDate := timestamp.Format("20060102T150405Z")

	req.Header.Set("Host", req.Host)
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Content-Sha256", "UNSIGNED-PAYLOAD")

	// Extract key from canonicalURI for signature calculation
	// canonicalURI is either "/{bucket}/key" for objects or "/{bucket}" for bucket operations
	key := ""
	bucketPrefix := "/" + c.config.BucketName + "/"
	if strings.HasPrefix(canonicalURI, bucketPrefix) {
		key = canonicalURI[len(bucketPrefix):]
	} else if canonicalURI == "/"+c.config.BucketName {
		// Bucket-level operation, key remains empty
		key = ""
	}

	// Calculate signature using unified method
	auth := c.calculateAuthorization(method, key, amzDate, rawQuery)
	req.Header.Set("Authorization", auth)

	return req, nil
}

// createRequest creates an S3 request with authentication (object-level).
func (c *S3Client) createRequest(ctx context.Context, method, key string, body io.Reader) (*http.Request, error) {
	canonicalURI := "/" + c.config.BucketName + "/" + key
	return c.buildSignedRequest(ctx, method, canonicalURI, "", body)
}

// createRequestForBucket creates a bucket-level S3 request with authentication.
// Used for operations like ListObjectsV2 that operate on the bucket itself.
func (c *S3Client) createRequestForBucket(ctx context.Context, method, rawQuery string) (*http.Request, error) {
	canonicalURI := "/" + c.config.BucketName
	return c.buildSignedRequest(ctx, method, canonicalURI, rawQuery, nil)
}


// calculateAuthorization calculates AWS V4 signature authorization header.
func (c *S3Client) calculateAuthorization(method, key, amzDate, rawQuery string) string {
	// This is a simplified version of AWS Signature V4
	// In production, use the full AWS V4 signing process

	// Scope
	dateStamp := amzDate[:8]
	scope := fmt.Sprintf("%s/%s/s3/aws4_request", dateStamp, c.config.Region)

	// Canonical request - build URI based on whether key is present
	// For bucket operations (empty key): /{bucket}
	// For object operations: /{bucket}/{key}
	var canonicalURI string
	if key == "" {
		canonicalURI = "/" + c.config.BucketName
	} else {
		canonicalURI = "/" + c.config.BucketName + "/" + key
	}
	canonicalQuery := rawQuery

	// Build host header based on URL style
	var hostHeader string
	if c.config.ForcePathStyle {
		hostHeader = c.config.Endpoint
	} else {
		hostHeader = c.config.BucketName + "." + c.config.Endpoint
	}

	canonicalHeaders := fmt.Sprintf("host:%s\nx-amz-date:%s\n", hostHeader, amzDate)
	signedHeaders := "host;x-amz-date"

	payloadHash := "UNSIGNED-PAYLOAD"

	// Canonical request format per AWS V4 spec:
	// HTTPMethod\nCanonicalURI\nCanonicalQueryString\nCanonicalHeaders\nSignedHeaders\nHash(Payload)
	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s%s\n%s",
		method, canonicalURI, canonicalQuery, canonicalHeaders, signedHeaders, payloadHash)

	// Create string to sign
	algorithm := "AWS4-HMAC-SHA256"
	credentialScope := scope
	stringToSign := fmt.Sprintf("%s\n%s\n%s\n%s",
		algorithm, amzDate, credentialScope, hex.EncodeToString(hashSHA256([]byte(canonicalRequest))))

	// Calculate signature
	kSecret := []byte("AWS4" + c.config.SecretKey)
	kDate := hmacSHA256(kSecret, dateStamp)
	kRegion := hmacSHA256(kDate, c.config.Region)
	kService := hmacSHA256(kRegion, "s3")
	kSigning := hmacSHA256(kService, "aws4_request")
	signature := hex.EncodeToString(hmacSHA256(kSigning, stringToSign))

	// Build authorization header
	accessKey := c.config.AccessKey
	return fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		algorithm, accessKey, scope, signedHeaders, signature)
}

// hmacSHA256 calculates HMAC-SHA256.
func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

// hashSHA256 calculates SHA256 hash.
func hashSHA256(data []byte) []byte {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
}

// TestConnection tests the S3 connection by listing the bucket.
func (c *S3Client) TestConnection(ctx context.Context) error {
	// Try to list with empty prefix
	_, err := c.List(ctx, "")
	return err
}

// GetBucketLocation gets the bucket location (region).
func (c *S3Client) GetBucketLocation(ctx context.Context) (string, error) {
	// Create GET bucket location request (bucket-level operation)
	// Use "location" query param without '?' prefix
	req, err := c.createRequestForBucket(ctx, http.MethodGet, "location")
	if err != nil {
		return "", err
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("location request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("location request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse XML response
	var locationConstraint LocationConstraint
	if err := xml.NewDecoder(resp.Body).Decode(&locationConstraint); err != nil {
		return "", fmt.Errorf("failed to parse location response: %w", err)
	}

	// Empty LocationConstraint means us-east-1
	if locationConstraint.LocationConstraint == "" {
		return "us-east-1", nil
	}

	return locationConstraint.LocationConstraint, nil
}
