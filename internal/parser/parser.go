// ABOUTME: Main parser implementation integrating all extractors and cleaners into complete extraction pipeline
// ABOUTME: Wires together resource layer, generic extractors, and content cleaners to create working end-to-end parser
// ABOUTME: Originally inspired by the Postlight Mercury parser, now rebranded as Hermes for consistency

package parser

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/BumpyClock/hermes/internal/resource"
	"github.com/BumpyClock/hermes/internal/validation"
)

// Hermes (formerly Mercury) is the main parser implementation
type Hermes struct {
	options    ParserOptions
	httpClient *http.Client // Store HTTP client
}

// New creates a new Hermes parser instance
func New(opts ...*ParserOptions) *Hermes {
	var options ParserOptions
	if len(opts) > 0 && opts[0] != nil {
		options = *opts[0]
	} else {
		options = *DefaultParserOptions()
	}

	h := &Hermes{
		options: options,
	}
	
	// Store HTTP client if provided
	if options.HTTPClient != nil {
		h.httpClient = options.HTTPClient
	}
	
	return h
}

// NewParser creates a new parser instance (convenience function)
func NewParser() *Hermes {
	return New()
}

// Parse extracts content from a URL
func (h *Hermes) Parse(targetURL string, opts *ParserOptions) (*Result, error) {
	// Use provided options or defaults
	if opts == nil {
		opts = &h.options
	}

	// Use the simple parsing approach
	return h.parseWithoutOptimization(targetURL, opts)
}

// ParseWithContext extracts content from a URL with context support
func (h *Hermes) ParseWithContext(ctx context.Context, targetURL string, opts *ParserOptions) (*Result, error) {
	// Use provided options or defaults
	if opts == nil {
		opts = &h.options
	}

	// Use the context-aware parsing path
	return h.parseWithoutOptimizationContext(ctx, targetURL, opts)
}

// ParseHTML extracts content from provided HTML
func (h *Hermes) ParseHTML(html string, targetURL string, opts *ParserOptions) (*Result, error) {
	// Use provided options or defaults
	if opts == nil {
		opts = &h.options
	}

	// Use the simple parsing approach
	return h.parseHTMLWithoutOptimization(html, targetURL, opts)
}

// ParseHTMLWithContext extracts content from provided HTML with context support
func (h *Hermes) ParseHTMLWithContext(ctx context.Context, html string, targetURL string, opts *ParserOptions) (*Result, error) {
	// Use provided options or defaults
	if opts == nil {
		opts = &h.options
	}

	// Use the context-aware parsing path
	return h.parseHTMLWithoutOptimizationContext(ctx, html, targetURL, opts)
}

// ReturnResult is deprecated - no longer needed without object pooling
func (h *Hermes) ReturnResult(result *Result) {
	// No-op - object pooling has been removed
}

// GetStats is deprecated - no longer tracks statistics
func (h *Hermes) GetStats() *PoolStats {
	// Return empty stats for backward compatibility
	return &PoolStats{}
}

// ResetStats is deprecated - no longer tracks statistics
func (h *Hermes) ResetStats() {
	// No-op - statistics tracking has been removed
}

// parseWithoutOptimization performs basic parsing without optimization layers
// Used internally by the optimization framework to avoid circular dependencies
// DEPRECATED: This method uses context.Background() which prevents proper cancellation.
// Use parseWithoutOptimizationContext instead.
func (h *Hermes) parseWithoutOptimization(targetURL string, opts *ParserOptions) (*Result, error) {
	// Use background context for backward compatibility - DEPRECATED
	// Callers should use ParseWithContext for proper context handling
	return h.parseWithoutOptimizationContext(context.Background(), targetURL, opts)
}

// parseWithoutOptimizationContext performs basic parsing with context support
func (h *Hermes) parseWithoutOptimizationContext(ctx context.Context, targetURL string, opts *ParserOptions) (*Result, error) {
	// Validate URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}
	
	// Use unified URL validation
	validationOpts := validation.DefaultValidationOptions()
	validationOpts.AllowPrivateNetworks = opts.AllowPrivateNetworks
	validationOpts.AllowLocalhost = opts.AllowPrivateNetworks // Localhost should be allowed when private networks are allowed
	
	if err := validation.ValidateURL(ctx, targetURL, validationOpts); err != nil {
		return nil, fmt.Errorf("URL validation failed: %w", err)
	}
	
	// Create resource instance and fetch content with context
	r := resource.NewResource()
	
	// Use centralized HTTP client creation
	httpClient := ensureHTTPClient(opts)
	
	doc, err := r.CreateWithClient(ctx, targetURL, "", parsedURL, opts.Headers, httpClient)
	if err != nil {
		return nil, err
	}
	
	// Use the real extraction logic with context
	return h.extractAllFieldsWithContext(ctx, doc, targetURL, parsedURL, *opts)
}

// parseHTMLWithoutOptimization performs basic HTML parsing without optimization layers
// DEPRECATED: This method uses context.Background() which prevents proper cancellation.
// Use parseHTMLWithoutOptimizationContext instead.
func (h *Hermes) parseHTMLWithoutOptimization(html, targetURL string, opts *ParserOptions) (*Result, error) {
	// Use background context for backward compatibility - DEPRECATED
	// Callers should use ParseHTMLWithContext for proper context handling
	return h.parseHTMLWithoutOptimizationContext(context.Background(), html, targetURL, opts)
}

// parseHTMLWithoutOptimizationContext performs HTML parsing with context support
func (h *Hermes) parseHTMLWithoutOptimizationContext(ctx context.Context, html, targetURL string, opts *ParserOptions) (*Result, error) {
	// Validate URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}
	
	// Create resource instance and parse HTML with context
	r := resource.NewResource()
	
	// Use centralized HTTP client creation (for consistency, even though HTML parsing doesn't need HTTP)
	httpClient := ensureHTTPClientForHTML(opts)
	
	doc, err := r.CreateWithClient(ctx, targetURL, html, parsedURL, opts.Headers, httpClient)
	if err != nil {
		return nil, err
	}
	
	// Use the real extraction logic with context
	return h.extractAllFieldsWithContext(ctx, doc, targetURL, parsedURL, *opts)
}


// TODO: Implement multi-page article collection and merging
// The FetchAllPages configuration option exists but doesn't trigger actual merging.
// Infrastructure exists in pkg/extractors/collect_all_pages.go but needs integration.
// func (h *Hermes) collectAllPages(result *Result, extractor Extractor, opts ParserOptions) (*Result, error) {
// 	// Multi-page collection not implemented - would require:
// 	// 1. Next page URL detection from content (✓ partially implemented)
// 	// 2. Recursive fetching and content aggregation (needs implementation)
// 	// 3. Deduplication and proper content merging (needs implementation)
// 	// 4. Integration with main extraction pipeline (needs implementation)
// 	return result, nil
// }
