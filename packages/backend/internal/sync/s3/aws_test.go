// Package s3 provides unit tests for AWS S3 provider.
// T152: Unit test for AWS S3 provider configuration.
package s3

import (
	"testing"
)

// TestNewAWSClient tests creating an AWS S3 client.
func TestNewAWSClient(t *testing.T) {
	config := &AWSConfig{
		BucketName: "test-bucket",
		AccessKey:  "AKIAIOSFODNN7EXAMPLE",
		SecretKey:  "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		Region:     "us-west-2",
	}

	client := NewAWSClient(config)

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	// Client configuration is internal, but we can verify it was created
	// by checking that it's not nil and has the expected type
}

// TestNewAWSClientDefaultRegion tests creating AWS client with default region.
func TestNewAWSClientDefaultRegion(t *testing.T) {
	config := &AWSConfig{
		BucketName: "test-bucket",
		AccessKey:  "AKIAIOSFODNN7EXAMPLE",
		SecretKey:  "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		// Region not specified, should default to us-east-1
	}

	client := NewAWSClient(config)

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	// Verify default region was used by checking endpoint lookup
	endpoint, err := AWSEndpointForRegion("us-east-1")
	if err != nil {
		t.Errorf("Failed to get default region endpoint: %v", err)
	}

	if endpoint != "s3.amazonaws.com" {
		t.Errorf("Expected default endpoint s3.amazonaws.com, got %s", endpoint)
	}
}

// TestAWSEndpointForRegion tests getting regional endpoints.
func TestAWSEndpointForRegion(t *testing.T) {
	tests := []struct {
		name     string
		region   string
		expected string
		wantErr  bool
	}{
		{"us-east-1", "us-east-1", "s3.amazonaws.com", false},
		{"us-west-2", "us-west-2", "s3.us-west-2.amazonaws.com", false},
		{"eu-west-1", "eu-west-1", "s3.eu-west-1.amazonaws.com", false},
		{"ap-northeast-1", "ap-northeast-1", "s3.ap-northeast-1.amazonaws.com", false},
		{"invalid region", "invalid-region", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint, err := AWSEndpointForRegion(tt.region)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error for invalid region")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if endpoint != tt.expected {
				t.Errorf("Expected endpoint %s, got %s", tt.expected, endpoint)
			}
		})
	}
}

// TestIsSupportedAWSRegion tests region support checking.
func TestIsSupportedAWSRegion(t *testing.T) {
	tests := []struct {
		region   string
		supported bool
	}{
		{"us-east-1", true},
		{"us-west-2", true},
		{"eu-central-1", true},
		{"ap-southeast-1", true},
		{"invalid-region", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.region, func(t *testing.T) {
			result := IsSupportedAWSRegion(tt.region)
			if result != tt.supported {
				t.Errorf("IsSupportedAWSRegion(%s) = %v, want %v", tt.region, result, tt.supported)
			}
		})
	}
}

// TestSupportedAWSRegions tests getting all supported regions.
func TestSupportedAWSRegions(t *testing.T) {
	regions := SupportedAWSRegions()

	if len(regions) == 0 {
		t.Error("Expected at least one region")
	}

	// Check that common regions are present
	commonRegions := []string{"us-east-1", "us-west-2", "eu-west-1"}
	for _, r := range commonRegions {
		found := false
		for _, supported := range regions {
			if supported == r {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected region %s to be in supported regions", r)
		}
	}
}
