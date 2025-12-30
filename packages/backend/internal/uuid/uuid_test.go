// Package uuid provides unit tests for UUID generation and validation.
package uuid

import (
	"regexp"
	"testing"
)

// TestNew tests that New() generates valid UUID v4 strings.
func TestNew(t *testing.T) {
	id := New()

	if id == "" {
		t.Fatal("Expected non-empty UUID string")
	}

	// UUID v4 format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	if !uuidRegex.MatchString(id) {
		t.Errorf("Generated UUID does not match v4 format: %s", id)
	}
}

// TestNewUniqueness tests that New() generates unique IDs.
func TestNewUniqueness(t *testing.T) {
	ids := make(map[string]bool)

	// Generate 1000 UUIDs and verify uniqueness
	for i := 0; i < 1000; i++ {
		id := New()
		if ids[id] {
			t.Errorf("Duplicate UUID generated: %s", id)
		}
		ids[id] = true
	}

	if len(ids) != 1000 {
		t.Errorf("Expected 1000 unique UUIDs, got %d", len(ids))
	}
}

// TestIsValid tests valid UUID v4 strings.
func TestIsValid(t *testing.T) {
	tests := []struct {
		name string
		uuid string
		want bool
	}{
		{
			name: "valid UUID v4",
			uuid: "f47ac10b-58cc-4372-a567-0e02b2c3d479",
			want: true,
		},
		{
			name: "valid UUID v4 with zeros",
			uuid: "00000000-0000-4000-8000-000000000000",
			want: true,
		},
		{
			name: "valid UUID v4 with all fs",
			uuid: "ffffffff-ffff-4fff-bfff-ffffffffffff",
			want: true,
		},
		{
			name: "valid UUID v4 lowercase",
			uuid: "6ba7b810-9dad-41d1-80b4-00c04fd430c8",
			want: true,
		},
		{
			name: "valid UUID v4 uppercase",
			uuid: "6BA7B810-9DAD-41D1-80B4-00C04FD430C8",
			want: true,
		},
		{
			name: "empty string",
			uuid: "",
			want: false,
		},
		{
			name: "too short",
			uuid: "f47ac10b-58cc-4372-a567",
			want: false,
		},
		{
			name: "too long",
			uuid: "f47ac10b-58cc-4372-a567-0e02b2c3d479-extra",
			want: false,
		},
		{
			name: "invalid format - missing dashes",
			uuid: "f47ac10b58cc4372a5670e02b2c3d479",
			want: false,
		},
		{
			name: "invalid version - v1 instead of v4",
			uuid: "f47ac10b-58cc-1372-a567-0e02b2c3d479",
			want: false,
		},
		{
			name: "invalid characters",
			uuid: "g47ac10b-58cc-4372-a567-0e02b2c3d479",
			want: false,
		},
		{
			name: "invalid variant",
			uuid: "f47ac10b-58cc-4372-c567-0e02b2c3d479",
			want: false,
		},
		{
			name: "random string",
			uuid: "not-a-uuid",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValid(tt.uuid)
			if got != tt.want {
				t.Errorf("IsValid(%q) = %v, want %v", tt.uuid, got, tt.want)
			}
		})
	}
}

// TestValidate tests Validate() function.
func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		uuid    string
		wantErr bool
	}{
		{
			name:    "valid UUID v4",
			uuid:    "f47ac10b-58cc-4372-a567-0e02b2c3d479",
			wantErr: false,
		},
		{
			name:    "invalid UUID",
			uuid:    "not-a-uuid",
			wantErr: true,
		},
		{
			name:    "empty string",
			uuid:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.uuid)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate(%q) error = %v, wantErr %v", tt.uuid, err, tt.wantErr)
			}
		})
	}
}

// TestNewFromString tests NewFromString() function.
func TestNewFromString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid UUID v4",
			input:   "f47ac10b-58cc-4372-a567-0e02b2c3d479",
			wantErr: false,
		},
		{
			name:    "invalid UUID format",
			input:   "not-a-uuid",
			wantErr: true,
		},
		{
			name:    "invalid version - v1",
			input:   "f47ac10b-58cc-1372-a567-0e02b2c3d479",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFromString(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

// TestNewFromStringRoundTrip tests that NewFromString can parse UUIDs from New().
func TestNewFromStringRoundTrip(t *testing.T) {
	original := New()

	parsed, err := NewFromString(original)
	if err != nil {
		t.Fatalf("NewFromString(%q) failed: %v", original, err)
	}

	if parsed.String() != original {
		t.Errorf("Round trip failed: got %q, want %q", parsed.String(), original)
	}
}

// BenchmarkNew benchmarks the New() function.
func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New()
	}
}

// BenchmarkIsValid benchmarks the IsValid() function.
func BenchmarkIsValid(b *testing.B) {
	validUUID := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	for i := 0; i < b.N; i++ {
		IsValid(validUUID)
	}
}
