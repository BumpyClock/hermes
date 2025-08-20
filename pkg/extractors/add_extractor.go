// ABOUTME: Runtime extractor registration system with extended types support and JavaScript-compatible API
// ABOUTME: Direct port of JavaScript add-extractor.js with identical validation, error handling, and registry management

package extractors

import (
	"sync"
	"fmt"

	"github.com/PuerkitoBio/goquery"
)

// FullExtractor represents a complete custom extractor with all field definitions
type FullExtractor struct {
	Domain           string                                `json:"domain"`
	SupportedDomains []string                             `json:"supportedDomains,omitempty"`
	
	// Field extractors for standard fields
	Title         *FieldExtractor   `json:"title,omitempty"`
	Author        *FieldExtractor   `json:"author,omitempty"`
	Content       *ContentExtractor `json:"content,omitempty"`
	DatePublished *FieldExtractor   `json:"date_published,omitempty"`
	LeadImageURL  *FieldExtractor   `json:"lead_image_url,omitempty"`
	Dek           *FieldExtractor   `json:"dek,omitempty"`
	NextPageURL   *FieldExtractor   `json:"next_page_url,omitempty"`
	Excerpt       *FieldExtractor   `json:"excerpt,omitempty"`
	WordCount     *FieldExtractor   `json:"word_count,omitempty"`
	Direction     *FieldExtractor   `json:"direction,omitempty"`
	URL           *FieldExtractor   `json:"url,omitempty"`
	
	// Extended types for custom fields
	Extend map[string]*FieldExtractor `json:"extend,omitempty"`
}

// FieldExtractor configuration for extracting a field
type FieldExtractor struct {
	Selectors      []interface{} `json:"selectors,omitempty"`      // string or [string, string] for [selector, attr]
	AllowMultiple  bool          `json:"allowMultiple,omitempty"`  
	DefaultCleaner bool          `json:"defaultCleaner"`           // defaults to true in JavaScript
}

// ContentExtractor configuration for content extraction with transforms and cleaning
type ContentExtractor struct {
	Selectors      []interface{}             `json:"selectors,omitempty"`
	AllowMultiple  bool                      `json:"allowMultiple,omitempty"`
	DefaultCleaner bool                      `json:"defaultCleaner"`
	Clean          []string                  `json:"clean,omitempty"`      // Selectors to remove
	Transforms     map[string]interface{}    `json:"transforms,omitempty"` // Transform functions
}

// GetDomain implements the Extractor interface for FullExtractor
func (f *FullExtractor) GetDomain() string {
	return f.Domain
}

// Extract implements the Extractor interface for FullExtractor
func (f *FullExtractor) Extract(doc *goquery.Document) (interface{}, error) {
	// This will be implemented with the root extractor logic
	return nil, fmt.Errorf("not implemented yet")
}

// ExtractorError represents error response from addExtractor
type ExtractorError struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

// API extractor registry - thread-safe storage for runtime-registered extractors
var (
	apiExtractorsMutex sync.RWMutex
	apiExtractors      = make(map[string]*FullExtractor)
)

// mergeSupportedDomains creates a map of domain -> extractor mappings
// Direct port of JavaScript mergeSupportedDomains function
func mergeSupportedDomains(extractor *FullExtractor) map[string]*FullExtractor {
	var domains []string
	
	// Add the primary domain
	domains = append(domains, extractor.Domain)
	
	// Add supported domains if they exist
	if extractor.SupportedDomains != nil {
		domains = append(domains, extractor.SupportedDomains...)
	}
	
	// Create the map - each domain points to the same extractor
	result := make(map[string]*FullExtractor)
	for _, domain := range domains {
		result[domain] = extractor
	}
	
	return result
}

// AddExtractor adds a custom extractor to the runtime registry
// Direct port of JavaScript addExtractor function with 100% behavioral compatibility
func AddExtractor(extractor *FullExtractor) interface{} {
	// JavaScript validation: if (!extractor || !extractor.domain)
	if extractor == nil || extractor.Domain == "" {
		return ExtractorError{
			Error:   true,
			Message: "Unable to add custom extractor. Invalid parameters.",
		}
	}
	
	// Generate domain mappings using mergeSupportedDomains
	domainMappings := mergeSupportedDomains(extractor)
	
	// Thread-safe registry update
	apiExtractorsMutex.Lock()
	defer apiExtractorsMutex.Unlock()
	
	// JavaScript equivalent: Object.assign(apiExtractors, mergeSupportedDomains(extractor))
	for domain, extractorCopy := range domainMappings {
		apiExtractors[domain] = extractorCopy
	}
	
	// Return a copy of all current extractors (JavaScript compatibility)
	result := make(map[string]*FullExtractor)
	for k, v := range apiExtractors {
		result[k] = v
	}
	return result
}

// GetAPIExtractors returns a copy of all runtime-registered extractors
func GetAPIExtractors() map[string]*FullExtractor {
	apiExtractorsMutex.RLock()
	defer apiExtractorsMutex.RUnlock()
	
	result := make(map[string]*FullExtractor)
	for k, v := range apiExtractors {
		result[k] = v
	}
	return result
}

// GetExtractorByDomain retrieves a specific extractor by domain from API registry
func GetExtractorByDomain(domain string) (*FullExtractor, bool) {
	apiExtractorsMutex.RLock()
	defer apiExtractorsMutex.RUnlock()
	
	extractor, exists := apiExtractors[domain]
	return extractor, exists
}

// HasExtractor checks if an extractor is registered for the given domain
func HasExtractor(domain string) bool {
	apiExtractorsMutex.RLock()
	defer apiExtractorsMutex.RUnlock()
	
	_, exists := apiExtractors[domain]
	return exists
}

// GetExtractorCount returns the number of registered extractors
func GetExtractorCount() int {
	apiExtractorsMutex.RLock()
	defer apiExtractorsMutex.RUnlock()
	
	return len(apiExtractors)
}

// ClearAPIExtractors clears all registered extractors (useful for testing)
func ClearAPIExtractors() {
	apiExtractorsMutex.Lock()
	defer apiExtractorsMutex.Unlock()
	
	apiExtractors = make(map[string]*FullExtractor)
}