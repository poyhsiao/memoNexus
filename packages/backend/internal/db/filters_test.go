// Package db tests for search filter building functionality.
package db

import (
	"strings"
	"testing"
	"time"
)

// TestMediaTypeFilter_Valid verifies media type validation.
func TestMediaTypeFilter_Valid(t *testing.T) {
	tests := []struct {
		name      string
		mediaType string
		expected  bool
	}{
		{"valid web", "web", true},
		{"valid image", "image", true},
		{"valid video", "video", true},
		{"valid pdf", "pdf", true},
		{"valid markdown", "markdown", true},
		{"invalid empty", "", false},
		{"invalid unknown", "unknown", false},
		{"invalid uppercase", "Web", false},
		{"invalid mixed case", "PDF", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &MediaTypeFilter{MediaType: tt.mediaType}
			result := filter.Valid()
			if result != tt.expected {
				t.Errorf("MediaTypeFilter.Valid(%q) = %v, want %v", tt.mediaType, result, tt.expected)
			}
		})
	}
}

// TestMediaTypeFilter_SQL verifies SQL generation.
func TestMediaTypeFilter_SQL(t *testing.T) {
	filter := &MediaTypeFilter{MediaType: "web"}
	expected := "ci.media_type = ?"
	result := filter.SQL()
	if result != expected {
		t.Errorf("MediaTypeFilter.SQL() = %q, want %q", result, expected)
	}
}

// TestMediaTypeFilter_Args verifies argument generation.
func TestMediaTypeFilter_Args(t *testing.T) {
	filter := &MediaTypeFilter{MediaType: "pdf"}
	args := filter.Args()
	if len(args) != 1 {
		t.Fatalf("Args() returned %d args, want 1", len(args))
	}
	if args[0] != "pdf" {
		t.Errorf("Args()[0] = %q, want 'pdf'", args[0])
	}
}

// TestDateRangeFilter_Valid verifies date range validation.
func TestDateRangeFilter_Valid(t *testing.T) {
	now := time.Now().Unix()

	tests := []struct {
		name     string
		from     int64
		to       int64
		expected bool
	}{
		{"valid from only", now - 86400, 0, true},
		{"valid to only", 0, now - 86400, true},
		{"valid range", now - 86400, now - 3600, true},
		{"invalid none", 0, 0, false},
		{"invalid from after to", now - 3600, now - 86400, false},
		{"invalid to future", 0, now + 100000, false},
		{"valid boundary", now - 86400, now + 86400, true}, // exactly 1 day ahead
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &DateRangeFilter{From: tt.from, To: tt.to}
			result := filter.Valid()
			if result != tt.expected {
				t.Errorf("DateRangeFilter.Valid(from=%d, to=%d) = %v, want %v", tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

// TestDateRangeFilter_SQL verifies SQL generation for date ranges.
func TestDateRangeFilter_SQL(t *testing.T) {
	tests := []struct {
		name     string
		from     int64
		to       int64
		expected string
	}{
		{"from only", 1000, 0, "ci.created_at >= ?"},
		{"to only", 0, 2000, "ci.created_at <= ?"},
		{"both", 1000, 2000, "ci.created_at >= ? AND ci.created_at <= ?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &DateRangeFilter{From: tt.from, To: tt.to}
			result := filter.SQL()
			if result != tt.expected {
				t.Errorf("DateRangeFilter.SQL(from=%d, to=%d) = %q, want %q", tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

// TestDateRangeFilter_Args verifies argument generation.
func TestDateRangeFilter_Args(t *testing.T) {
	tests := []struct {
		name     string
		from     int64
		to       int64
		expected []interface{}
	}{
		{"from only", 1000, 0, []interface{}{int64(1000)}},
		{"to only", 0, 2000, []interface{}{int64(2000)}},
		{"both", 1000, 2000, []interface{}{int64(1000), int64(2000)}},
		{"none", 0, 0, []interface{}{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &DateRangeFilter{From: tt.from, To: tt.to}
			args := filter.Args()
			if len(args) != len(tt.expected) {
				t.Fatalf("Args() returned %d args, want %d", len(args), len(tt.expected))
			}
			for i, arg := range args {
				if arg != tt.expected[i] {
					t.Errorf("Args()[%d] = %v, want %v", i, arg, tt.expected[i])
				}
			}
		})
	}
}

// TestTagsFilter_Valid verifies tag validation.
func TestTagsFilter_Valid(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		expected bool
	}{
		{"valid single", []string{"python"}, true},
		{"valid multiple", []string{"python", "golang"}, true},
		{"invalid empty", []string{}, false},
		{"invalid whitespace", []string{"python", " "}, false},
		{"valid leading space", []string{" python"}, true},     // Valid() trims whitespace
		{"valid trailing space", []string{"python "}, true},    // Valid() trims whitespace
		{"valid after trim", []string{" python ", "golang"}, true}, // Valid() trims whitespace
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &TagsFilter{Tags: tt.tags}
			result := filter.Valid()
			if result != tt.expected {
				t.Errorf("TagsFilter.Valid(%v) = %v, want %v", tt.tags, result, tt.expected)
			}
		})
	}
}

// TestTagsFilter_SQL verifies SQL generation for tags.
func TestTagsFilter_SQL(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		expected string
	}{
		{"single tag", []string{"python"}, "(ci.tags LIKE ?)"},
		{"multiple tags", []string{"python", "golang"}, "(ci.tags LIKE ? OR ci.tags LIKE ?)"},
		{"empty tags", []string{}, "1=0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &TagsFilter{Tags: tt.tags}
			result := filter.SQL()
			if result != tt.expected {
				t.Errorf("TagsFilter.SQL(%v) = %q, want %q", tt.tags, result, tt.expected)
			}
		})
	}
}

// TestTagsFilter_Args verifies argument generation for tags.
func TestTagsFilter_Args(t *testing.T) {
	filter := &TagsFilter{Tags: []string{"python", "golang"}}
	args := filter.Args()
	if len(args) != 2 {
		t.Fatalf("Args() returned %d args, want 2", len(args))
	}
	if args[0] != "%python%" {
		t.Errorf("Args()[0] = %q, want '%%python%%'", args[0])
	}
	if args[1] != "%golang%" {
		t.Errorf("Args()[1] = %q, want '%%golang%%'", args[1])
	}
}

// TestNewFilterBuilder verifies FilterBuilder creation.
func TestNewFilterBuilder(t *testing.T) {
	fb := NewFilterBuilder()
	if fb == nil {
		t.Fatal("NewFilterBuilder() returned nil")
	}
	if fb.filters == nil {
		t.Error("filters slice is nil")
	}
	if len(fb.filters) != 0 {
		t.Errorf("filters length = %d, want 0", len(fb.filters))
	}
}

// TestFilterBuilder_HasFilters verifies HasFilters method.
func TestFilterBuilder_HasFilters(t *testing.T) {
	fb := NewFilterBuilder()

	if fb.HasFilters() {
		t.Error("HasFilters() on empty builder should return false")
	}

	fb.MediaType("web")
	if !fb.HasFilters() {
		t.Error("HasFilters() after adding filter should return true")
	}
}

// TestFilterBuilder_Count verifies Count method.
func TestFilterBuilder_Count(t *testing.T) {
	fb := NewFilterBuilder()

	if fb.Count() != 0 {
		t.Errorf("Count() on empty builder = %d, want 0", fb.Count())
	}

	fb.MediaType("web")
	if fb.Count() != 1 {
		t.Errorf("Count() after adding 1 filter = %d, want 1", fb.Count())
	}

	fb.DateFrom(1000)
	if fb.Count() != 2 {
		t.Errorf("Count() after adding 2 filters = %d, want 2", fb.Count())
	}
}

// TestFilterBuilder_MediaType verifies MediaType filter addition.
func TestFilterBuilder_MediaType(t *testing.T) {
	fb := NewFilterBuilder()

	result := fb.MediaType("web")
	if result != fb {
		t.Error("MediaType() should return the same FilterBuilder instance")
	}

	if fb.Count() != 1 {
		t.Errorf("Count() after MediaType() = %d, want 1", fb.Count())
	}

	// Invalid media type should not be added
	fb.MediaType("invalid")
	if fb.Count() != 1 {
		t.Errorf("Count() after invalid MediaType() = %d, want 1", fb.Count())
	}
}

// TestFilterBuilder_DateRange verifies DateRange filter addition.
func TestFilterBuilder_DateRange(t *testing.T) {
	fb := NewFilterBuilder()

	result := fb.DateRange(1000, 2000)
	if result != fb {
		t.Error("DateRange() should return the same FilterBuilder instance")
	}

	if fb.Count() != 1 {
		t.Errorf("Count() after DateRange() = %d, want 1", fb.Count())
	}

	// Invalid range should not be added
	fb.DateRange(2000, 1000)
	if fb.Count() != 1 {
		t.Errorf("Count() after invalid DateRange() = %d, want 1", fb.Count())
	}
}

// TestFilterBuilder_DateFrom verifies DateFrom filter addition.
func TestFilterBuilder_DateFrom(t *testing.T) {
	fb := NewFilterBuilder()
	fb.DateFrom(1000)

	if fb.Count() != 1 {
		t.Errorf("Count() after DateFrom() = %d, want 1", fb.Count())
	}

	sql, args := fb.Build()
	if sql != "ci.created_at >= ?" {
		t.Errorf("SQL = %q, want 'ci.created_at >= ?'", sql)
	}
	if len(args) != 1 || args[0] != int64(1000) {
		t.Errorf("Args = %v, want [1000]", args)
	}
}

// TestFilterBuilder_DateTo verifies DateTo filter addition.
func TestFilterBuilder_DateTo(t *testing.T) {
	fb := NewFilterBuilder()
	fb.DateTo(2000)

	if fb.Count() != 1 {
		t.Errorf("Count() after DateTo() = %d, want 1", fb.Count())
	}

	sql, args := fb.Build()
	if sql != "ci.created_at <= ?" {
		t.Errorf("SQL = %q, want 'ci.created_at <= ?'", sql)
	}
	if len(args) != 1 || args[0] != int64(2000) {
		t.Errorf("Args = %v, want [2000]", args)
	}
}

// TestFilterBuilder_Tags verifies Tags filter addition.
func TestFilterBuilder_Tags(t *testing.T) {
	fb := NewFilterBuilder()

	result := fb.Tags("python", "golang")
	if result != fb {
		t.Error("Tags() should return the same FilterBuilder instance")
	}

	if fb.Count() != 1 {
		t.Errorf("Count() after Tags() = %d, want 1", fb.Count())
	}

	// Empty tags should not be added
	fb.Tags()
	if fb.Count() != 1 {
		t.Errorf("Count() after empty Tags() = %d, want 1", fb.Count())
	}
}

// TestFilterBuilder_TagsFromCommaString verifies tag parsing from string.
func TestFilterBuilder_TagsFromCommaString(t *testing.T) {
	fb := NewFilterBuilder()

	fb.TagsFromCommaString("python, golang, rust")

	if fb.Count() != 1 {
		t.Errorf("Count() after TagsFromCommaString() = %d, want 1", fb.Count())
	}

	sql, args := fb.Build()
	if !strings.Contains(sql, "ci.tags LIKE ?") {
		t.Error("SQL should contain ci.tags LIKE ?")
	}
	if len(args) != 3 {
		t.Errorf("Args length = %d, want 3", len(args))
	}
}

// TestFilterBuilder_TagsFromCommaString_whitespace verifies whitespace handling.
func TestFilterBuilder_TagsFromCommaString_whitespace(t *testing.T) {
	fb := NewFilterBuilder()

	fb.TagsFromCommaString("python , golang , rust")

	if fb.Count() != 1 {
		t.Errorf("Count() = %d, want 1", fb.Count())
	}

	_, args := fb.Build()
	if len(args) != 3 {
		t.Errorf("Args length = %d, want 3 (whitespace trimmed)", len(args))
	}
}

// TestFilterBuilder_Build verifies SQL building.
func TestFilterBuilder_Build(t *testing.T) {
	tests := []struct {
		name           string
		mediaType      string
		from           int64
		to             int64
		tags           []string
		expectedSQL    string
		expectedArgLen  int
	}{
		{
			name:          "no filters",
			mediaType:     "",
			from:          0,
			to:            0,
			tags:          nil,
			expectedSQL:   "",
			expectedArgLen: 0,
		},
		{
			name:          "media type only",
			mediaType:     "web",
			from:          0,
			to:            0,
			tags:          nil,
			expectedSQL:   "ci.media_type = ?",
			expectedArgLen: 1,
		},
		{
			name:          "date range only",
			mediaType:     "",
			from:          1000,
			to:            2000,
			tags:          nil,
			expectedSQL:   "ci.created_at >= ? AND ci.created_at <= ?",
			expectedArgLen: 2,
		},
		{
			name:          "all filters",
			mediaType:     "pdf",
			from:          1000,
			to:            2000,
			tags:          []string{"python"},
			expectedSQL:   "ci.media_type = ? AND ci.created_at >= ? AND ci.created_at <= ? AND (ci.tags LIKE ?)",
			expectedArgLen: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fb := NewFilterBuilder()
			if tt.mediaType != "" {
				fb.MediaType(tt.mediaType)
			}
			if tt.from > 0 || tt.to > 0 {
				fb.DateRange(tt.from, tt.to)
			}
			if tt.tags != nil {
				fb.Tags(tt.tags...)
			}

			sql, args := fb.Build()
			if sql != tt.expectedSQL {
				t.Errorf("SQL = %q, want %q", sql, tt.expectedSQL)
			}
			if len(args) != tt.expectedArgLen {
				t.Errorf("Args length = %d, want %d", len(args), tt.expectedArgLen)
			}
		})
	}
}

// TestFilterBuilder_BuildForSearch verifies search-specific building.
func TestFilterBuilder_BuildForSearch(t *testing.T) {
	fb := NewFilterBuilder()
	fb.MediaType("web").DateFrom(1000)

	sql, args := fb.BuildForSearch()
	if sql == "" {
		t.Error("BuildForSearch() should return SQL")
	}
	if len(args) == 0 {
		t.Error("BuildForSearch() should return args")
	}
}

// TestFilterBuilder_Reset verifies Reset method.
func TestFilterBuilder_Reset(t *testing.T) {
	fb := NewFilterBuilder()
	fb.MediaType("web").DateFrom(1000)

	if fb.Count() != 2 {
		t.Errorf("Count() before Reset() = %d, want 2", fb.Count())
	}

	result := fb.Reset()
	if result != fb {
		t.Error("Reset() should return the same FilterBuilder instance")
	}

	if fb.Count() != 0 {
		t.Errorf("Count() after Reset() = %d, want 0", fb.Count())
	}
}

// TestFilterBuilder_Reset_keeps_same_instance verifies Reset doesn't create new instance.
func TestFilterBuilder_Reset_keeps_same_instance(t *testing.T) {
	fb := NewFilterBuilder()
	originalAddr := &fb.filters
	fb.MediaType("web")

	fb.Reset()
	if &fb.filters != originalAddr {
		t.Error("Reset() should keep the same slice")
	}
}

// TestFilterBuilder_Clone verifies Clone method.
func TestFilterBuilder_Clone(t *testing.T) {
	fb := NewFilterBuilder()
	fb.MediaType("web").DateFrom(1000)

	clone := fb.Clone()

	if clone.Count() != fb.Count() {
		t.Errorf("Clone Count() = %d, want %d", clone.Count(), fb.Count())
	}

	// Verify it's a different instance - modifying original shouldn't affect clone
	originalCount := fb.Count() // Should be 2
	fb.MediaType("pdf")
	newCount := fb.Count()     // Should be 3

	if newCount != originalCount + 1 {
		t.Errorf("Original Count() after adding filter = %d, want %d", newCount, originalCount+1)
	}

	// Clone should still have the original 2 filters
	if clone.Count() != 2 {
		t.Errorf("Clone Count() after modifying original = %d, want 2 (unchanged)", clone.Count())
	}
}

// TestFilterBuilder_String verifies String representation.
func TestFilterBuilder_String(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*FilterBuilder)
		expected string
	}{
		{
			name:     "empty",
			setup:    func(fb *FilterBuilder) {},
			expected: "(no filters)",
		},
		{
			name: "single filter",
			setup: func(fb *FilterBuilder) {
				fb.MediaType("web")
			},
			expected: "*db.MediaTypeFilter",
		},
		{
			name: "multiple filters",
			setup: func(fb *FilterBuilder) {
				fb.MediaType("web").DateFrom(1000)
			},
			expected: "*db.MediaTypeFilter, *db.DateRangeFilter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fb := NewFilterBuilder()
			tt.setup(fb)
			result := fb.String()
			if result != tt.expected {
				t.Errorf("String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestMediaTypeFromString verifies media type validation function.
func TestMediaTypeFromString(t *testing.T) {
	tests := []struct {
		name      string
		mediaType string
		expected  string
		hasError  bool
	}{
		{"valid web", "web", "web", false},
		{"valid image", "image", "image", false},
		{"invalid", "invalid", "", true},
		{"invalid empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MediaTypeFromString(tt.mediaType)
			if tt.hasError {
				if err == nil {
					t.Error("MediaTypeFromString() should return error")
				}
			} else {
				if err != nil {
					t.Errorf("MediaTypeFromString() unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("MediaTypeFromString() = %q, want %q", result, tt.expected)
				}
			}
		})
	}
}

// TestTagsFromCommaString verifies tag parsing function.
func TestTagsFromCommaString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"single", "python", []string{"python"}},
		{"multiple", "python,golang,rust", []string{"python", "golang", "rust"}},
		{"with spaces", "python, golang, rust", []string{"python", "golang", "rust"}},
		{"empty tags", "", []string{}},
		{"trailing comma", "python,", []string{"python"}},
		{"leading comma", ",python", []string{"python"}},
		{"multiple commas", "python,,golang", []string{"python", "golang"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TagsFromCommaString(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("TagsFromCommaString() length = %d, want %d", len(result), len(tt.expected))
			}
			for i, tag := range result {
				if tag != tt.expected[i] {
					t.Errorf("TagsFromCommaString()[%d] = %q, want %q", i, tag, tt.expected[i])
				}
			}
		})
	}
}

// TestValidateDateRange verifies date range validation function.
func TestValidateDateRange(t *testing.T) {
	now := time.Now().Unix()

	tests := []struct {
		name     string
		from     int64
		to       int64
		hasError bool
	}{
		{"valid from", now - 86400, 0, false},
		{"valid to", 0, now - 86400, false},
		{"valid range", now - 86400, now - 3600, false},
		{"invalid none", 0, 0, true},
		{"invalid inverted", now - 3600, now - 86400, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDateRange(tt.from, tt.to)
			if tt.hasError {
				if err == nil {
					t.Error("ValidateDateRange() should return error")
				}
			} else {
				if err != nil {
					t.Errorf("ValidateDateRange() unexpected error: %v", err)
				}
			}
		})
	}
}

// TestNormalizeTags verifies tag normalization function.
func TestNormalizeTags(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		expected string
	}{
		{"single", []string{"python"}, "python"},
		{"multiple", []string{"python", "golang", "rust"}, "python,golang,rust"},
		{"with spaces", []string{" python ", "golang"}, "python,golang"},
		{"empty tags", []string{}, ""},
		{"empty strings", []string{"", "python", ""}, "python"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeTags(tt.tags)
			if result != tt.expected {
				t.Errorf("NormalizeTags() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestParseSearchOptionsToFilters verifies SearchOptions parsing.
func TestParseSearchOptionsToFilters(t *testing.T) {
	opts := &SearchOptions{
		MediaType: "web",
		DateFrom:   1000,
		DateTo:     2000,
		Tags:       "python,golang",
	}

	fb := ParseSearchOptionsToFilters(opts)

	if fb.Count() != 3 {
		t.Errorf("Count() = %d, want 3", fb.Count())
	}

	sql, args := fb.Build()
	if !strings.Contains(sql, "ci.media_type = ?") {
		t.Error("SQL should contain media_type filter")
	}
	if !strings.Contains(sql, "ci.created_at") {
		t.Error("SQL should contain date filter")
	}
	if !strings.Contains(sql, "ci.tags LIKE") {
		t.Error("SQL should contain tags filter")
	}
	if len(args) != 5 { // 1 mediaType + 2 dates + 2 tags (2 LIKE conditions)
		t.Errorf("Args length = %d, want 5", len(args))
	}
}

// TestFilterBuilder_chain_methods verifies method chaining.
func TestFilterBuilder_chain_methods(t *testing.T) {
	fb := NewFilterBuilder().
		MediaType("pdf").
		DateRange(1000, 2000).
		Tags("python", "golang")

	if fb.Count() != 3 {
		t.Errorf("Count() = %d, want 3", fb.Count())
	}

	sql, args := fb.Build()
	if !strings.Contains(sql, "ci.media_type = ?") {
		t.Error("SQL should contain media_type filter")
	}
	if len(args) != 5 { // 1 mediaType + 2 dates + 2 tags
		t.Errorf("Args length = %d, want 5", len(args))
	}
}

// TestFilterBuilder_chain_filters_can_be_modified verifies filters can be added incrementally.
func TestFilterBuilder_chain_filters_can_be_modified(t *testing.T) {
	fb := NewFilterBuilder()

	fb.MediaType("web")
	sql1, _ := fb.Build()

	fb.DateFrom(1000)
	sql2, _ := fb.Build()

	if !strings.Contains(sql2, sql1) {
		t.Error("Second SQL should contain first SQL")
	}
	if !strings.Contains(sql2, "ci.created_at >= ?") {
		t.Error("Second SQL should contain date filter")
	}
}
