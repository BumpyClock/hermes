// ABOUTME: Field transformation utilities for converting extracted data to standard formats
// ABOUTME: Provides transformers for string normalization, URL resolution, date parsing, and format conversion

package fields

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// FieldTransformer converts extracted data to standard formats
type FieldTransformer interface {
	Transform(value interface{}) interface{}
	TargetType() string
}

// StringTransformer normalizes string fields
type StringTransformer struct{}

// NewStringTransformer creates a new string transformer
func NewStringTransformer() *StringTransformer {
	return &StringTransformer{}
}

// Transform normalizes a string value
func (st *StringTransformer) Transform(value interface{}) interface{} {
	str, ok := value.(string)
	if !ok {
		return value
	}
	
	// Trim whitespace and normalize spaces
	str = strings.TrimSpace(str)
	str = normalizeSpaces(str)
	
	return str
}

// TargetType returns the target type
func (st *StringTransformer) TargetType() string {
	return "string"
}

// URLTransformer resolves and normalizes URL fields
type URLTransformer struct {
	baseURL string
}

// NewURLTransformer creates a new URL transformer with base URL for relative resolution
func NewURLTransformer(baseURL string) *URLTransformer {
	return &URLTransformer{
		baseURL: baseURL,
	}
}

// Transform resolves and normalizes a URL
func (ut *URLTransformer) Transform(value interface{}) interface{} {
	str, ok := value.(string)
	if !ok {
		return value
	}
	
	str = strings.TrimSpace(str)
	if str == "" {
		return ""
	}
	
	// Parse the URL
	parsedURL, err := url.Parse(str)
	if err != nil {
		return str // Return original if parsing fails
	}
	
	// Resolve relative URLs
	if ut.baseURL != "" && !parsedURL.IsAbs() {
		if baseURL, err := url.Parse(ut.baseURL); err == nil {
			parsedURL = baseURL.ResolveReference(parsedURL)
		}
	}
	
	// Normalize the URL
	return normalizeURL(parsedURL)
}

// TargetType returns the target type
func (ut *URLTransformer) TargetType() string {
	return "string"
}

// DateTransformer parses and normalizes date fields
type DateTransformer struct {
	outputFormat string
}

// NewDateTransformer creates a new date transformer
func NewDateTransformer() *DateTransformer {
	return &DateTransformer{
		outputFormat: time.RFC3339,
	}
}

// NewDateTransformerWithFormat creates a date transformer with custom output format
func NewDateTransformerWithFormat(format string) *DateTransformer {
	return &DateTransformer{
		outputFormat: format,
	}
}

// Transform parses and formats a date
func (dt *DateTransformer) Transform(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		if parsedTime, err := parseDate(v); err == nil {
			return parsedTime
		}
		return value
	case time.Time:
		return v
	default:
		return value
	}
}

// TargetType returns the target type
func (dt *DateTransformer) TargetType() string {
	return "time.Time"
}

// ArrayTransformer processes array fields
type ArrayTransformer struct {
	elementTransformer FieldTransformer
	deduplicateItems   bool
	maxItems           int
}

// NewArrayTransformer creates a new array transformer
func NewArrayTransformer(elementTransformer FieldTransformer) *ArrayTransformer {
	return &ArrayTransformer{
		elementTransformer: elementTransformer,
		deduplicateItems:   true,
		maxItems:           0, // No limit
	}
}

// SetDeduplication enables or disables item deduplication
func (at *ArrayTransformer) SetDeduplication(enabled bool) {
	at.deduplicateItems = enabled
}

// SetMaxItems sets the maximum number of items
func (at *ArrayTransformer) SetMaxItems(max int) {
	at.maxItems = max
}

// Transform processes an array of values
func (at *ArrayTransformer) Transform(value interface{}) interface{} {
	var items []interface{}
	
	switch v := value.(type) {
	case []interface{}:
		items = v
	case []string:
		for _, str := range v {
			items = append(items, str)
		}
	case string:
		// Split string into array if it contains delimiters
		if strings.Contains(v, ",") {
			parts := strings.Split(v, ",")
			for _, part := range parts {
				items = append(items, strings.TrimSpace(part))
			}
		} else {
			items = []interface{}{v}
		}
	default:
		return value
	}
	
	// Transform each element
	var transformedItems []interface{}
	seen := make(map[string]bool)
	
	for _, item := range items {
		transformed := item
		if at.elementTransformer != nil {
			transformed = at.elementTransformer.Transform(item)
		}
		
		// Handle deduplication
		if at.deduplicateItems {
			key := getStringRepresentation(transformed)
			if seen[key] {
				continue
			}
			seen[key] = true
		}
		
		transformedItems = append(transformedItems, transformed)
		
		// Check item limit
		if at.maxItems > 0 && len(transformedItems) >= at.maxItems {
			break
		}
	}
	
	return transformedItems
}

// TargetType returns the target type
func (at *ArrayTransformer) TargetType() string {
	if at.elementTransformer != nil {
		return "[]" + at.elementTransformer.TargetType()
	}
	return "[]interface{}"
}

// JSONTransformer handles structured JSON data
type JSONTransformer struct {
	fieldMappings map[string]FieldTransformer
}

// NewJSONTransformer creates a new JSON transformer
func NewJSONTransformer() *JSONTransformer {
	return &JSONTransformer{
		fieldMappings: make(map[string]FieldTransformer),
	}
}

// AddFieldMapping adds a transformer for a specific field
func (jt *JSONTransformer) AddFieldMapping(fieldName string, transformer FieldTransformer) {
	jt.fieldMappings[fieldName] = transformer
}

// Transform processes structured data
func (jt *JSONTransformer) Transform(value interface{}) interface{} {
	data, ok := value.(map[string]interface{})
	if !ok {
		return value
	}
	
	transformed := make(map[string]interface{})
	
	for key, val := range data {
		if transformer, exists := jt.fieldMappings[key]; exists {
			transformed[key] = transformer.Transform(val)
		} else {
			transformed[key] = val
		}
	}
	
	return transformed
}

// TargetType returns the target type
func (jt *JSONTransformer) TargetType() string {
	return "map[string]interface{}"
}

// Helper functions

// normalizeSpaces normalizes whitespace in a string
func normalizeSpaces(s string) string {
	// Replace multiple consecutive whitespace characters with single space
	var result strings.Builder
	var lastWasSpace bool
	
	for _, char := range s {
		isSpace := char == ' ' || char == '\t' || char == '\n' || char == '\r'
		
		if isSpace {
			if !lastWasSpace {
				result.WriteRune(' ')
				lastWasSpace = true
			}
		} else {
			result.WriteRune(char)
			lastWasSpace = false
		}
	}
	
	return result.String()
}

// normalizeURL normalizes a URL by removing unnecessary components
func normalizeURL(parsedURL *url.URL) string {
	// Remove fragment unless it's meaningful
	if parsedURL.Fragment != "" && !isMeaningfulFragment(parsedURL.Fragment) {
		parsedURL.Fragment = ""
	}
	
	// Normalize query parameters (could be enhanced)
	if parsedURL.RawQuery != "" {
		values := parsedURL.Query()
		// Remove tracking parameters
		trackingParams := []string{"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content", "fbclid", "gclid"}
		for _, param := range trackingParams {
			values.Del(param)
		}
		parsedURL.RawQuery = values.Encode()
	}
	
	// Remove default ports
	if (parsedURL.Scheme == "http" && parsedURL.Port() == "80") ||
		(parsedURL.Scheme == "https" && parsedURL.Port() == "443") {
		parsedURL.Host = parsedURL.Hostname()
	}
	
	return parsedURL.String()
}

// isMeaningfulFragment checks if a URL fragment is meaningful (not just tracking)
func isMeaningfulFragment(fragment string) bool {
	// Consider fragments meaningful if they look like section references
	meaningfulPrefixes := []string{"section", "chapter", "page", "anchor", "content"}
	
	lower := strings.ToLower(fragment)
	for _, prefix := range meaningfulPrefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	
	// Also consider fragments with letters and numbers meaningful
	hasLetters := false
	hasNumbers := false
	for _, char := range fragment {
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' {
			hasLetters = true
		}
		if char >= '0' && char <= '9' {
			hasNumbers = true
		}
	}
	
	return hasLetters && hasNumbers
}

// parseDate attempts to parse a date string using common formats
func parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"January 2, 2006",
		"Jan 2, 2006",
		"2006/01/02",
		"01/02/2006",
		"02-01-2006",
		"2006-01-02T15:04:05.000Z",
		"Mon, 02 Jan 2006 15:04:05 MST",
		"Mon, 02 Jan 2006 15:04:05 -0700",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse date with any known format")
}

// getStringRepresentation returns a string representation of any value for deduplication
func getStringRepresentation(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return string(rune(v))
	case float64:
		return string(rune(int(v)))
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}

// ChainTransformer chains multiple transformers together
type ChainTransformer struct {
	transformers []FieldTransformer
}

// NewChainTransformer creates a new chain transformer
func NewChainTransformer(transformers ...FieldTransformer) *ChainTransformer {
	return &ChainTransformer{
		transformers: transformers,
	}
}

// Transform applies all transformers in sequence
func (ct *ChainTransformer) Transform(value interface{}) interface{} {
	result := value
	for _, transformer := range ct.transformers {
		result = transformer.Transform(result)
	}
	return result
}

// TargetType returns the target type of the last transformer
func (ct *ChainTransformer) TargetType() string {
	if len(ct.transformers) > 0 {
		return ct.transformers[len(ct.transformers)-1].TargetType()
	}
	return "interface{}"
}