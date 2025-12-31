// Package s3 provides S3-compatible storage provider implementations.
// T152: AWS S3 provider with virtual-host style URLs.
package s3

import (
	"fmt"

	"github.com/kimhsiao/memonexus/backend/internal/sync"
)

// Default AWS S3 endpoints by region.
// T152: Standard AWS S3 regional endpoints.
var awsEndpoints = map[string]string{
	"us-east-1":      "s3.amazonaws.com",
	"us-east-2":      "s3.us-east-2.amazonaws.com",
	"us-west-1":      "s3.us-west-1.amazonaws.com",
	"us-west-2":      "s3.us-west-2.amazonaws.com",
	"eu-west-1":      "s3.eu-west-1.amazonaws.com",
	"eu-west-2":      "s3.eu-west-2.amazonaws.com",
	"eu-west-3":      "s3.eu-west-3.amazonaws.com",
	"eu-central-1":   "s3.eu-central-1.amazonaws.com",
	"eu-north-1":     "s3.eu-north-1.amazonaws.com",
	"eu-south-1":     "s3.eu-south-1.amazonaws.com",
	"ap-northeast-1": "s3.ap-northeast-1.amazonaws.com",
	"ap-northeast-2": "s3.ap-northeast-2.amazonaws.com",
	"ap-northeast-3": "s3.ap-northeast-3.amazonaws.com",
	"ap-southeast-1": "s3.ap-southeast-1.amazonaws.com",
	"ap-southeast-2": "s3.ap-southeast-2.amazonaws.com",
	"ap-south-1":     "s3.ap-south-1.amazonaws.com",
	"ca-central-1":   "s3.ca-central-1.amazonaws.com",
	"sa-east-1":      "s3.sa-east-1.amazonaws.com",
	"me-south-1":     "s3.me-south-1.amazonaws.com",
	"af-south-1":     "s3.af-south-1.amazonaws.com",
}

// AWSConfig holds AWS S3-specific configuration.
type AWSConfig struct {
	BucketName string
	AccessKey  string
	SecretKey  string
	Region     string // Default: us-east-1
}

// NewAWSClient creates an S3 client configured for AWS S3.
// T152: AWS S3 uses virtual-host style URLs (bucket.s3.amazonaws.com).
//
// Example:
//
//	client := NewAWSClient(&AWSConfig{
//	    BucketName: "my-bucket",
//	    AccessKey:  "AKIA...",
//	    SecretKey:  "...",
//	    Region:     "us-west-2",
//	})
func NewAWSClient(config *AWSConfig) *sync.S3Client {
	// Default to us-east-1 if not specified
	region := config.Region
	if region == "" {
		region = "us-east-1"
	}

	// Get regional endpoint
	endpoint, ok := awsEndpoints[region]
	if !ok {
		// Fallback to global endpoint for unknown regions
		endpoint = "s3.amazonaws.com"
	}

	return sync.NewS3Client(&sync.S3Config{
		Endpoint:       endpoint,
		BucketName:     config.BucketName,
		AccessKey:      config.AccessKey,
		SecretKey:      config.SecretKey,
		Region:         region,
		ForcePathStyle: false, // Virtual-host style for AWS S3
	})
}

// AWSEndpointForRegion returns the S3 endpoint for a given region.
// Returns an error if the region is not recognized.
func AWSEndpointForRegion(region string) (string, error) {
	endpoint, ok := awsEndpoints[region]
	if !ok {
		return "", fmt.Errorf("unknown AWS region: %s", region)
	}
	return endpoint, nil
}

// IsSupportedRegion checks if a region is supported.
func IsSupportedAWSRegion(region string) bool {
	_, ok := awsEndpoints[region]
	return ok
}

// SupportedAWSRegions returns a list of all supported AWS regions.
func SupportedAWSRegions() []string {
	regions := make([]string, 0, len(awsEndpoints))
	for region := range awsEndpoints {
		regions = append(regions, region)
	}
	return regions
}
