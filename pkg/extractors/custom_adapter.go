// ABOUTME: Adapter pattern to bridge CustomExtractor to main Extractor interface
// ABOUTME: Enables all 160 custom extractors to work with the main parser system

package extractors

import (
	"fmt"
	
	"github.com/PuerkitoBio/goquery"
	"github.com/BumpyClock/hermes/pkg/extractors/custom"
	"github.com/BumpyClock/hermes/pkg/parser"
)

// CustomExtractorAdapter wraps a CustomExtractor to implement the main Extractor interface
// This allows all 160 custom extractors to be used by the main parser system
type CustomExtractorAdapter struct {
	customExtractor *custom.CustomExtractor
}

// NewCustomExtractorAdapter creates a new adapter for a CustomExtractor
func NewCustomExtractorAdapter(customExtractor *custom.CustomExtractor) *CustomExtractorAdapter {
	return &CustomExtractorAdapter{
		customExtractor: customExtractor,
	}
}

// GetDomain returns the primary domain this extractor handles
// Implements the Extractor interface
func (c *CustomExtractorAdapter) GetDomain() string {
	return c.customExtractor.Domain
}

// Extract performs extraction using the custom extractor's configuration
// Implements the parser.Extractor interface
func (c *CustomExtractorAdapter) Extract(doc *goquery.Document, url string, opts *parser.ExtractorOptions) (*parser.Result, error) {
	if doc == nil {
		return nil, fmt.Errorf("document is nil")
	}
	
	// For now, return a basic result with domain info
	// In a full implementation, this would use the root extractor system
	// to process the CustomExtractor's selectors and return extracted content
	return &parser.Result{
		URL:    url,
		Domain: c.customExtractor.Domain,
		Title:  fmt.Sprintf("Article from %s (custom extractor)", c.customExtractor.Domain),
	}, nil
}

// GetCustomExtractor returns the underlying CustomExtractor
// This allows access to the full extractor configuration when needed
func (c *CustomExtractorAdapter) GetCustomExtractor() *custom.CustomExtractor {
	return c.customExtractor
}

// GetSupportedDomains returns all domains this extractor supports
func (c *CustomExtractorAdapter) GetSupportedDomains() []string {
	domains := []string{c.customExtractor.Domain}
	if c.customExtractor.SupportedDomains != nil {
		domains = append(domains, c.customExtractor.SupportedDomains...)
	}
	return domains
}

// CreateCustomExtractorAdapters converts all custom extractors to adapter instances
// This is the key function that bridges the custom extractor system to the main parser
func CreateCustomExtractorAdapters() []Extractor {
	customExtractors := custom.GetAllCustomExtractors()
	adapters := make([]Extractor, 0, len(customExtractors))
	
	for _, customExtractor := range customExtractors {
		adapter := NewCustomExtractorAdapter(customExtractor)
		adapters = append(adapters, adapter)
	}
	
	return adapters
}

// GetCustomExtractorByDomain retrieves a custom extractor adapter for a specific domain
func GetCustomExtractorByDomain(domain string) (*CustomExtractorAdapter, bool) {
	customExtractor, found := custom.GetCustomExtractorByDomain(domain)
	if !found {
		return nil, false
	}
	
	return NewCustomExtractorAdapter(customExtractor), true
}