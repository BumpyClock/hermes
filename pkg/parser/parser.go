// ABOUTME: Main parser implementation integrating all extractors and cleaners into complete extraction pipeline
// ABOUTME: Wires together resource layer, generic extractors, and content cleaners to create working end-to-end parser

package parser

import (
	"fmt"
	"net/url"

	"github.com/BumpyClock/parser-go/pkg/resource"
	"github.com/BumpyClock/parser-go/pkg/utils/security"
)

// Mercury is the main parser implementation
type Mercury struct {
	options ParserOptions
}

// New creates a new Mercury parser instance
func New(opts ...ParserOptions) *Mercury {
	parser := &Mercury{}
	if len(opts) > 0 {
		parser.options = opts[0]
	} else {
		parser.options = ParserOptions{
			FetchAllPages: true,
			Fallback:      true,
			ContentType:   "html",
		}
	}
	return parser
}

// Parse extracts content from a URL
func (m *Mercury) Parse(targetURL string, opts ParserOptions) (*Result, error) {
	// Parse and validate URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if !validateURL(parsedURL) {
		return &Result{
			Error:   true,
			Message: "The url parameter passed does not look like a valid URL. Please check your URL and try again.",
		}, nil
	}

	// Create resource from URL (with fetching)
	r := resource.NewResource()
	doc, err := r.Create(targetURL, "", parsedURL, opts.Headers)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Extract all fields using generic extractors
	result, err := m.extractAllFields(doc, targetURL, parsedURL, opts)
	if err != nil {
		return nil, fmt.Errorf("extraction failed: %w", err)
	}

	// Multi-page collection is not implemented in this version
	// This feature requires implementing next page URL detection
	// and recursive content aggregation

	return result, nil
}

// ParseHTML extracts content from provided HTML
func (m *Mercury) ParseHTML(html string, targetURL string, opts ParserOptions) (*Result, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if !validateURL(parsedURL) {
		return nil, fmt.Errorf("the url parameter passed does not look like a valid URL: %s", targetURL)
	}

	// Create resource from provided HTML
	r := resource.NewResource()
	doc, err := r.Create(targetURL, html, parsedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource from HTML: %w", err)
	}

	// Extract all fields using generic extractors
	result, err := m.extractAllFields(doc, targetURL, parsedURL, opts)
	if err != nil {
		return nil, fmt.Errorf("extraction failed: %w", err)
	}

	return result, nil
}

func validateURL(u *url.URL) bool {
	// Use the enhanced security validation
	if err := security.ValidateURL(u.String()); err != nil {
		return false
	}
	
	// Additional basic checks
	return security.IsValidWebURL(u)
}

func (m *Mercury) collectAllPages(result *Result, extractor Extractor, opts ParserOptions) (*Result, error) {
	// Multi-page collection not implemented - would require:
	// 1. Next page URL detection from content
	// 2. Recursive fetching and content aggregation
	// 3. Deduplication and proper content merging
	return result, nil
}