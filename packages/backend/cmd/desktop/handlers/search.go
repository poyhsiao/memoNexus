// Package handlers provides REST API handlers for search.
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/kimhsiao/memonexus/backend/internal/db"
	"github.com/kimhsiao/memonexus/backend/internal/models"
)

// SearchHandler handles search operations using FTS5.
type SearchHandler struct {
	repo *db.Repository
}

// NewSearchHandler creates a new SearchHandler.
func NewSearchHandler(repo *db.Repository) *SearchHandler {
	return &SearchHandler{repo: repo}
}

// Search handles GET /search
// Implements T118: FTS5 search endpoint with T119: query validation
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")

	// T119: Query validation
	if err := validateSearchQuery(query); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse and validate limit parameter
	limitStr := r.URL.Query().Get("limit")
	var limit int
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			http.Error(w, "limit must be between 1 and 100", http.StatusBadRequest)
			return
		}
	} else {
		limit = 20 // default
	}

	// Parse optional filters
	var dateFrom, dateTo int64
	if df := r.URL.Query().Get("date_from"); df != "" {
		d, err := strconv.ParseInt(df, 10, 64)
		if err != nil {
			http.Error(w, "invalid date_from format", http.StatusBadRequest)
			return
		}
		dateFrom = d
	}
	if dt := r.URL.Query().Get("date_to"); dt != "" {
		d, err := strconv.ParseInt(dt, 10, 64)
		if err != nil {
			http.Error(w, "invalid date_to format", http.StatusBadRequest)
			return
		}
		dateTo = d
	}

	// Build search options
	opts := &db.SearchOptions{
		Query:     query,
		Limit:     limit,
		MediaType: r.URL.Query().Get("media_type"),
		Tags:      r.URL.Query().Get("tags"),
		DateFrom:  dateFrom,
		DateTo:    dateTo,
	}

	// Validate media type if provided
	if opts.MediaType != "" {
		if _, err := validateMediaType(opts.MediaType); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Execute search using Repository.Search (T115)
	response, err := h.repo.Search(opts)
	if err != nil {
		http.Error(w, fmt.Sprintf("search failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Build response with proper total count
	result := map[string]interface{}{
		"results": toSearchResults(response.Results),
		"total":   response.Total,
		"query":   response.Query,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// validateSearchQuery validates the search query (T119)
// Returns an error if the query is invalid.
func validateSearchQuery(query string) error {
	// Check if query is empty
	if strings.TrimSpace(query) == "" {
		return fmt.Errorf("search query 'q' is required")
	}

	// Check query length (max 500 characters to prevent abuse)
	if utf8.RuneCountInString(query) > 500 {
		return fmt.Errorf("search query too long (max 500 characters)")
	}

	// Check for potentially dangerous FTS5 syntax
	// Allow basic FTS5 operators but limit complex queries
	dangerousPatterns := []string{
		"NEAR/", // NEAR with large distance could be slow
		"^",    // Prefix searches on very short terms
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(strings.ToUpper(query), pattern) {
			return fmt.Errorf("search query contains unsupported operator: %s", pattern)
		}
	}

	// Check for very short search terms that could cause performance issues
	terms := strings.Fields(query)
	for _, term := range terms {
		// Skip operators
		upperTerm := strings.ToUpper(term)
		if upperTerm == "AND" || upperTerm == "OR" || upperTerm == "NOT" {
			continue
		}
		// Remove trailing wildcard
		term = strings.TrimSuffix(term, "*")
		term = strings.Trim(term, "\"()")
		if utf8.RuneCountInString(term) < 2 && !strings.ContainsAny(term, "()*") {
			// Single character searches (without wildcards) are too broad
			return fmt.Errorf("search terms must be at least 2 characters")
		}
	}

	return nil
}

// validateMediaType validates the media type parameter.
func validateMediaType(mediaType string) (bool, error) {
	validTypes := map[string]bool{
		"web":      true,
		"image":    true,
		"video":    true,
		"pdf":      true,
		"markdown": true,
	}

	if !validTypes[mediaType] {
		return false, fmt.Errorf("invalid media_type: %s (must be one of: web, image, video, pdf, markdown)", mediaType)
	}
	return true, nil
}

// toSearchResults converts db.SearchResult to API response format.
func toSearchResults(results []*db.SearchResult) []map[string]interface{} {
	apiResults := make([]map[string]interface{}, len(results))

	for i, result := range results {
		item := result.Item

		apiItem := map[string]interface{}{
			"id":          item.ID,
			"title":       item.Title,
			"content_text": item.ContentText,
			"media_type":  item.MediaType,
			"tags":        item.Tags,
			"created_at":  item.CreatedAt,
			"updated_at":  item.UpdatedAt,
			"version":     item.Version,
		}

		if item.SourceURL != "" {
			apiItem["source_url"] = item.SourceURL
		}
		if item.Summary != "" {
			apiItem["summary"] = item.Summary
		}
		if item.ContentHash != "" {
			apiItem["content_hash"] = item.ContentHash
		}

		// Extract matched terms for highlighting
		matchedTerms := extractMatchedTermsFromResult(item, result.MatchedTerms)
		if len(matchedTerms) > 0 {
			apiItem["matched_terms"] = matchedTerms
		}

		apiResults[i] = map[string]interface{}{
			"item":         apiItem,
			"relevance":    result.Relevance,
			"matched_terms": result.MatchedTerms,
		}
	}

	return apiResults
}

// extractMatchedTermsFromResult processes matched terms from search result.
func extractMatchedTermsFromResult(item *models.ContentItem, matchedTerms []string) []string {
	// If matched terms are already populated, use them
	if len(matchedTerms) > 0 {
		return matchedTerms
	}

	// Otherwise, extract from the item's text content
	// This is a simplified implementation - production would use FTS5 matchinfo
	terms := make([]string, 0)
	seen := make(map[string]bool)

	// Search in title and content for common patterns
	searchText := item.Title + " " + item.ContentText + " " + item.Tags
	words := regexp.MustCompile(`\w+`).FindAllString(searchText, -1)

	for _, word := range words {
		word = strings.ToLower(word)
		if len(word) >= 3 && !seen[word] {
			terms = append(terms, word)
			seen[word] = true
			if len(terms) >= 5 { // Limit to 5 terms
				break
			}
		}
	}

	return terms
}
