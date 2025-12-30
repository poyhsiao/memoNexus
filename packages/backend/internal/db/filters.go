// Package db provides search filter building functionality.
package db

import (
	"fmt"
	"strings"
	"time"
)

// Filter represents a single search filter condition.
type Filter interface {
	// SQL returns the SQL fragment for this filter
	SQL() string

	// Args returns the arguments for this filter
	Args() []interface{}

	// Valid checks if the filter is valid
	Valid() bool
}

// MediaTypeFilter filters by content media type.
type MediaTypeFilter struct {
	MediaType string
}

// Valid checks if the media type is valid.
func (f *MediaTypeFilter) Valid() bool {
	validTypes := map[string]bool{
		"web":      true,
		"image":    true,
		"video":    true,
		"pdf":      true,
		"markdown": true,
	}
	return validTypes[f.MediaType]
}

// SQL returns the SQL fragment for media type filtering.
func (f *MediaTypeFilter) SQL() string {
	return "ci.media_type = ?"
}

// Args returns the arguments for media type filtering.
func (f *MediaTypeFilter) Args() []interface{} {
	return []interface{}{f.MediaType}
}

// DateRangeFilter filters by creation date range.
type DateRangeFilter struct {
	From int64 // Unix timestamp
	To   int64 // Unix timestamp
}

// Valid checks if the date range is valid.
func (f *DateRangeFilter) Valid() bool {
	// At least one boundary should be set
	if f.From == 0 && f.To == 0 {
		return false
	}
	// From should be before To (if both are set)
	if f.From > 0 && f.To > 0 && f.From > f.To {
		return false
	}
	// To should not be in the future
	if f.To > 0 && f.To > time.Now().Unix()+86400 {
		return false // Allow 1 day of clock skew
	}
	return true
}

// SQL returns the SQL fragment for date range filtering.
func (f *DateRangeFilter) SQL() string {
	var parts []string
	if f.From > 0 {
		parts = append(parts, "ci.created_at >= ?")
	}
	if f.To > 0 {
		parts = append(parts, "ci.created_at <= ?")
	}
	return strings.Join(parts, " AND ")
}

// Args returns the arguments for date range filtering.
func (f *DateRangeFilter) Args() []interface{} {
	var args []interface{}
	if f.From > 0 {
		args = append(args, f.From)
	}
	if f.To > 0 {
		args = append(args, f.To)
	}
	return args
}

// TagsFilter filters by tag names.
type TagsFilter struct {
	Tags []string // Tag names to match
}

// Valid checks if the tag filter is valid.
func (f *TagsFilter) Valid() bool {
	if len(f.Tags) == 0 {
		return false
	}
	for _, tag := range f.Tags {
		if strings.TrimSpace(tag) == "" {
			return false
		}
	}
	return true
}

// SQL returns the SQL fragment for tag filtering.
// Uses OR logic to match any of the specified tags.
func (f *TagsFilter) SQL() string {
	var conditions []string
	for _, tag := range f.Tags {
		if tag != "" {
			conditions = append(conditions, "ci.tags LIKE ?")
		}
	}
	if len(conditions) == 0 {
		return "1=0" // No valid tags, never match
	}
	return "(" + strings.Join(conditions, " OR ") + ")"
}

// Args returns the arguments for tag filtering.
func (f *TagsFilter) Args() []interface{} {
	var args []interface{}
	for _, tag := range f.Tags {
		if tag != "" {
			args = append(args, "%"+tag+"%")
		}
	}
	return args
}

// FilterBuilder builds SQL filter conditions from multiple filters.
type FilterBuilder struct {
	filters []Filter
}

// NewFilterBuilder creates a new FilterBuilder.
func NewFilterBuilder() *FilterBuilder {
	return &FilterBuilder{
		filters: make([]Filter, 0),
	}
}

// MediaType adds a media type filter.
func (fb *FilterBuilder) MediaType(mediaType string) *FilterBuilder {
	filter := &MediaTypeFilter{MediaType: mediaType}
	if filter.Valid() {
		fb.filters = append(fb.filters, filter)
	}
	return fb
}

// DateRange adds a date range filter.
func (fb *FilterBuilder) DateRange(from, to int64) *FilterBuilder {
	filter := &DateRangeFilter{From: from, To: to}
	if filter.Valid() {
		fb.filters = append(fb.filters, filter)
	}
	return fb
}

// DateFrom adds a "from date" filter.
func (fb *FilterBuilder) DateFrom(from int64) *FilterBuilder {
	return fb.DateRange(from, 0)
}

// DateTo adds a "to date" filter.
func (fb *FilterBuilder) DateTo(to int64) *FilterBuilder {
	return fb.DateRange(0, to)
}

// Tags adds a tag filter.
func (fb *FilterBuilder) Tags(tags ...string) *FilterBuilder {
	filter := &TagsFilter{Tags: tags}
	if filter.Valid() {
		fb.filters = append(fb.filters, filter)
	}
	return fb
}

// TagsFromCommaString adds tags from a comma-separated string.
func (fb *FilterBuilder) TagsFromCommaString(tagsStr string) *FilterBuilder {
	tags := strings.Split(tagsStr, ",")
	cleanTags := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			cleanTags = append(cleanTags, tag)
		}
	}
	return fb.Tags(cleanTags...)
}

// HasFilters returns true if any filters have been added.
func (fb *FilterBuilder) HasFilters() bool {
	return len(fb.filters) > 0
}

// Count returns the number of filters.
func (fb *FilterBuilder) Count() int {
	return len(fb.filters)
}

// Build builds the SQL WHERE clause and returns the arguments.
// Returns the SQL fragment and the arguments slice.
func (fb *FilterBuilder) Build() (string, []interface{}) {
	if !fb.HasFilters() {
		return "", nil
	}

	var sqlParts []string
	var args []interface{}

	for _, filter := range fb.filters {
		sqlParts = append(sqlParts, filter.SQL())
		args = append(args, filter.Args()...)
	}

	sql := strings.Join(sqlParts, " AND ")
	return sql, args
}

// BuildForSearch builds filters specifically for search queries.
// Ensures proper integration with FTS5 search.
func (fb *FilterBuilder) BuildForSearch() (string, []interface{}) {
	sql, args := fb.Build()
	return sql, args
}

// Reset clears all filters.
func (fb *FilterBuilder) Reset() *FilterBuilder {
	fb.filters = make([]Filter, 0)
	return fb
}

// Clone creates a copy of the FilterBuilder.
func (fb *FilterBuilder) Clone() *FilterBuilder {
	clone := NewFilterBuilder()
	clone.filters = append(clone.filters, fb.filters...)
	return clone
}

// String returns a string representation of the filters (for debugging).
func (fb *FilterBuilder) String() string {
	if !fb.HasFilters() {
		return "(no filters)"
	}

	var parts []string
	for _, filter := range fb.filters {
		parts = append(parts, fmt.Sprintf("%T", filter))
	}
	return strings.Join(parts, ", ")
}

// ParseSearchOptionsToFilters converts SearchOptions to a FilterBuilder.
// Useful for external code that wants to work with FilterBuilder API.
func ParseSearchOptionsToFilters(opts *SearchOptions) *FilterBuilder {
	fb := NewFilterBuilder()

	if opts.MediaType != "" {
		fb.MediaType(opts.MediaType)
	}

	if opts.DateFrom > 0 || opts.DateTo > 0 {
		fb.DateRange(opts.DateFrom, opts.DateTo)
	}

	if opts.Tags != "" {
		fb.TagsFromCommaString(opts.Tags)
	}

	return fb
}

// MediaTypeFromString validates and returns a media type string.
func MediaTypeFromString(mt string) (string, error) {
	filter := &MediaTypeFilter{MediaType: mt}
	if !filter.Valid() {
		return "", fmt.Errorf("invalid media type: %s", mt)
	}
	return mt, nil
}

// TagsFromCommaString parses tags from a comma-separated string.
func TagsFromCommaString(tagsStr string) []string {
	tags := strings.Split(tagsStr, ",")
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			result = append(result, tag)
		}
	}
	return result
}

// ValidateDateRange validates a date range.
func ValidateDateRange(from, to int64) error {
	filter := &DateRangeFilter{From: from, To: to}
	if !filter.Valid() {
		return fmt.Errorf("invalid date range: from=%d, to=%d", from, to)
	}
	return nil
}

// NormalizeTags converts tags to a comma-separated string.
func NormalizeTags(tags []string) string {
	cleanTags := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			cleanTags = append(cleanTags, tag)
		}
	}
	return strings.Join(cleanTags, ",")
}
