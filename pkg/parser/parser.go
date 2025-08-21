// ABOUTME: Main parser implementation integrating all extractors and cleaners into complete extraction pipeline
// ABOUTME: Wires together resource layer, generic extractors, and content cleaners to create working end-to-end parser

package parser

import (
	"fmt"
	"net/url"

	"github.com/BumpyClock/parser-go/pkg/resource"
	"github.com/BumpyClock/parser-go/pkg/utils/security"
)

// Mercury is the main parser implementation with built-in optimizations
type Mercury struct {
	options  ParserOptions
	htParser *HighThroughputParser
}

// New creates a new optimized Mercury parser instance
func New(opts ...*ParserOptions) *Mercury {
	var options ParserOptions
	if len(opts) > 0 && opts[0] != nil {
		options = *opts[0]
	} else {
		options = *DefaultParserOptions()
	}

	return &Mercury{
		options:  options,
		htParser: NewHighThroughputParser(&options),
	}
}

// NewParser creates a new parser instance (convenience function)
func NewParser() *Mercury {
	return New()
}

// Parse extracts content from a URL using optimized pooling
func (m *Mercury) Parse(targetURL string, opts *ParserOptions) (*Result, error) {
	// Use provided options or defaults
	if opts == nil {
		opts = &m.options
	}

	// Use the high-throughput parser for optimized performance
	return m.htParser.Parse(targetURL, opts)
}

// ParseHTML extracts content from provided HTML using optimized pooling
func (m *Mercury) ParseHTML(html string, targetURL string, opts *ParserOptions) (*Result, error) {
	// Use provided options or defaults
	if opts == nil {
		opts = &m.options
	}

	// Use the high-throughput parser for optimized performance
	return m.htParser.ParseHTML(html, targetURL, opts)
}

// ReturnResult returns a result to the object pool for reuse
// Call this when you're done with a Result to enable memory reuse
func (m *Mercury) ReturnResult(result *Result) {
	m.htParser.ReturnResult(result)
}

// GetStats returns performance statistics for this parser instance
func (m *Mercury) GetStats() *PoolStats {
	return m.htParser.GetStats()
}

// ResetStats resets performance statistics
func (m *Mercury) ResetStats() {
	m.htParser.ResetStats()
}

// parseWithoutOptimization performs basic parsing without optimization layers
// Used internally by the optimization framework to avoid circular dependencies
func (m *Mercury) parseWithoutOptimization(targetURL string, opts *ParserOptions) (*Result, error) {
	// Validate URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}
	
	if !validateURL(parsedURL) {
		return nil, fmt.Errorf("URL not allowed: %s", targetURL)
	}
	
	// Create resource instance and fetch content
	r := resource.NewResource()
	doc, err := r.Create(targetURL, "", parsedURL, opts.Headers)
	if err != nil {
		return nil, err
	}
	
	// Use the real extraction logic
	return m.extractAllFields(doc, targetURL, parsedURL, *opts)
}

// parseHTMLWithoutOptimization performs basic HTML parsing without optimization layers
func (m *Mercury) parseHTMLWithoutOptimization(html, targetURL string, opts *ParserOptions) (*Result, error) {
	// Validate URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}
	
	// Create resource instance and parse HTML
	r := resource.NewResource()
	doc, err := r.Create(targetURL, html, parsedURL, opts.Headers)
	if err != nil {
		return nil, err
	}
	
	// Use the real extraction logic
	return m.extractAllFields(doc, targetURL, parsedURL, *opts)
}

func validateURL(u *url.URL) bool {
	// Use the enhanced security validation
	if err := security.ValidateURL(u.String()); err != nil {
		return false
	}

	// Additional basic checks
	return security.IsValidWebURL(u)
}

// func (m *Mercury) collectAllPages(result *Result, extractor Extractor, opts ParserOptions) (*Result, error) {
// 	// Multi-page collection not implemented - would require:
// 	// 1. Next page URL detection from content
// 	// 2. Recursive fetching and content aggregation
// 	// 3. Deduplication and proper content merging
// 	return result, nil
// }
