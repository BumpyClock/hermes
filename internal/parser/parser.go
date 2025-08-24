// ABOUTME: Main parser implementation integrating all extractors and cleaners into complete extraction pipeline
// ABOUTME: Wires together resource layer, generic extractors, and content cleaners to create working end-to-end parser

package parser

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/BumpyClock/hermes/internal/resource"
	"github.com/BumpyClock/hermes/internal/utils/security"
)

// Mercury is the main parser implementation
type Mercury struct {
	options    ParserOptions
	httpClient *http.Client // Store HTTP client
}

// New creates a new Mercury parser instance
func New(opts ...*ParserOptions) *Mercury {
	var options ParserOptions
	if len(opts) > 0 && opts[0] != nil {
		options = *opts[0]
	} else {
		options = *DefaultParserOptions()
	}

	m := &Mercury{
		options: options,
	}
	
	// Store HTTP client if provided
	if options.HTTPClient != nil {
		m.httpClient = options.HTTPClient
	}
	
	return m
}

// NewParser creates a new parser instance (convenience function)
func NewParser() *Mercury {
	return New()
}

// Parse extracts content from a URL
func (m *Mercury) Parse(targetURL string, opts *ParserOptions) (*Result, error) {
	// Use provided options or defaults
	if opts == nil {
		opts = &m.options
	}

	// Use the simple parsing approach
	return m.parseWithoutOptimization(targetURL, opts)
}

// ParseWithContext extracts content from a URL with context support
func (m *Mercury) ParseWithContext(ctx context.Context, targetURL string, opts *ParserOptions) (*Result, error) {
	// Use provided options or defaults
	if opts == nil {
		opts = &m.options
	}

	// Use the context-aware parsing path
	return m.parseWithoutOptimizationContext(ctx, targetURL, opts)
}

// ParseHTML extracts content from provided HTML
func (m *Mercury) ParseHTML(html string, targetURL string, opts *ParserOptions) (*Result, error) {
	// Use provided options or defaults
	if opts == nil {
		opts = &m.options
	}

	// Use the simple parsing approach
	return m.parseHTMLWithoutOptimization(html, targetURL, opts)
}

// ParseHTMLWithContext extracts content from provided HTML with context support
func (m *Mercury) ParseHTMLWithContext(ctx context.Context, html string, targetURL string, opts *ParserOptions) (*Result, error) {
	// Use provided options or defaults
	if opts == nil {
		opts = &m.options
	}

	// Use the context-aware parsing path
	return m.parseHTMLWithoutOptimizationContext(ctx, html, targetURL, opts)
}

// ReturnResult is deprecated - no longer needed without object pooling
func (m *Mercury) ReturnResult(result *Result) {
	// No-op - object pooling has been removed
}

// GetStats is deprecated - no longer tracks statistics
func (m *Mercury) GetStats() *PoolStats {
	// Return empty stats for backward compatibility
	return &PoolStats{}
}

// ResetStats is deprecated - no longer tracks statistics
func (m *Mercury) ResetStats() {
	// No-op - statistics tracking has been removed
}

// parseWithoutOptimization performs basic parsing without optimization layers
// Used internally by the optimization framework to avoid circular dependencies
func (m *Mercury) parseWithoutOptimization(targetURL string, opts *ParserOptions) (*Result, error) {
	// Use background context for backward compatibility
	// Callers should use ParseWithContext for proper context handling
	return m.parseWithoutOptimizationContext(context.Background(), targetURL, opts)
}

// parseWithoutOptimizationContext performs basic parsing with context support
func (m *Mercury) parseWithoutOptimizationContext(ctx context.Context, targetURL string, opts *ParserOptions) (*Result, error) {
	// Validate URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}
	
	if !validateURLWithOptions(parsedURL, opts.AllowPrivateNetworks) {
		return nil, fmt.Errorf("URL not allowed: %s", targetURL)
	}
	
	// Create resource instance and fetch content with context
	r := resource.NewResource()
	
	// Use HTTP client from options if available, or create a default one
	var httpClient *resource.HTTPClient
	if opts.HTTPClient != nil {
		// Create HTTPClient wrapper for the provided client
		httpClient = &resource.HTTPClient{
			Client: opts.HTTPClient,
			Headers: opts.Headers,
		}
	} else {
		// Create a default HTTP client when none is provided
		httpClient = resource.CreateDefaultHTTPClient()
		httpClient.Headers = opts.Headers
	}
	
	doc, err := r.CreateWithClient(ctx, targetURL, "", parsedURL, opts.Headers, httpClient)
	if err != nil {
		return nil, err
	}
	
	// Use the real extraction logic with context
	return m.extractAllFieldsWithContext(ctx, doc, targetURL, parsedURL, *opts)
}

// parseHTMLWithoutOptimization performs basic HTML parsing without optimization layers
func (m *Mercury) parseHTMLWithoutOptimization(html, targetURL string, opts *ParserOptions) (*Result, error) {
	// Use background context for backward compatibility
	// Callers should use ParseHTMLWithContext for proper context handling
	return m.parseHTMLWithoutOptimizationContext(context.Background(), html, targetURL, opts)
}

// parseHTMLWithoutOptimizationContext performs HTML parsing with context support
func (m *Mercury) parseHTMLWithoutOptimizationContext(ctx context.Context, html, targetURL string, opts *ParserOptions) (*Result, error) {
	// Validate URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}
	
	// Create resource instance and parse HTML with context
	r := resource.NewResource()
	
	// Use HTTP client from options if available (though not needed for HTML parsing)
	var httpClient *resource.HTTPClient
	if opts.HTTPClient != nil {
		// Create HTTPClient wrapper for the provided client
		httpClient = &resource.HTTPClient{
			Client: opts.HTTPClient,
			Headers: opts.Headers,
		}
	} else {
		// Create a default HTTP client when none is provided (for consistency)
		httpClient = resource.CreateDefaultHTTPClient()
		httpClient.Headers = opts.Headers
	}
	
	doc, err := r.CreateWithClient(ctx, targetURL, html, parsedURL, opts.Headers, httpClient)
	if err != nil {
		return nil, err
	}
	
	// Use the real extraction logic with context
	return m.extractAllFieldsWithContext(ctx, doc, targetURL, parsedURL, *opts)
}

func validateURL(u *url.URL) bool {
	return validateURLWithOptions(u, false)
}

func validateURLWithOptions(u *url.URL, allowPrivateNetworks bool) bool {
	// Use the enhanced security validation with options
	if err := security.ValidateURLWithOptions(context.Background(), u.String(), allowPrivateNetworks); err != nil {
		return false
	}

	// Additional basic checks
	return security.IsValidWebURL(u)
}

// TODO: Implement multi-page article collection and merging
// The FetchAllPages configuration option exists but doesn't trigger actual merging.
// Infrastructure exists in pkg/extractors/collect_all_pages.go but needs integration.
// func (m *Mercury) collectAllPages(result *Result, extractor Extractor, opts ParserOptions) (*Result, error) {
// 	// Multi-page collection not implemented - would require:
// 	// 1. Next page URL detection from content (✓ partially implemented)
// 	// 2. Recursive fetching and content aggregation (needs implementation)
// 	// 3. Deduplication and proper content merging (needs implementation)
// 	// 4. Integration with main extraction pipeline (needs implementation)
// 	return result, nil
// }
