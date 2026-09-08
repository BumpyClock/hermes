// ABOUTME: Main parser implementation integrating all extractors and cleaners into complete extraction pipeline
// ABOUTME: Wires together resource layer, generic extractors, and content cleaners to create working end-to-end parser
// ABOUTME: Originally inspired by the Postlight Mercury parser, now rebranded as Hermes for consistency

package parser

import (
	"context"
	"fmt"
	"net/url"

	"github.com/BumpyClock/hermes/internal/resource"
	"github.com/BumpyClock/hermes/internal/validation"
)

// Hermes (formerly Mercury) is the main parser implementation.
type Hermes struct {
	options ParserOptions
}

// New creates a new Hermes parser instance.
func New(opts ...*ParserOptions) *Hermes {
	var options ParserOptions
	if len(opts) > 0 && opts[0] != nil {
		options = *opts[0]
	} else {
		options = *DefaultParserOptions()
	}

	return &Hermes{options: options}
}

// Parse extracts content from a URL.
func (h *Hermes) Parse(targetURL string, opts *ParserOptions) (*Result, error) {
	return h.ParseWithContext(context.Background(), targetURL, opts)
}

// ParseHTML extracts content from supplied HTML.
func (h *Hermes) ParseHTML(html, targetURL string, opts *ParserOptions) (*Result, error) {
	return h.ParseHTMLWithContext(context.Background(), html, targetURL, opts)
}

// ParseWithContext extracts content from a URL with context support.
func (h *Hermes) ParseWithContext(ctx context.Context, targetURL string, opts *ParserOptions) (*Result, error) {
	if opts == nil {
		opts = &h.options
	}
	// Validate URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	// Use unified URL validation
	validationOpts := validation.DefaultValidationOptions()
	validationOpts.AllowPrivateNetworks = opts.AllowPrivateNetworks
	validationOpts.AllowLocalhost = opts.AllowPrivateNetworks // Localhost should be allowed when private networks are allowed

	validationErr := validation.ValidateParsedURL(ctx, parsedURL, targetURL, validationOpts)
	if validationErr != nil {
		return nil, fmt.Errorf("URL validation failed: %w", validationErr)
	}

	// Use centralized HTTP client creation
	httpClient := ensureHTTPClient(opts)

	doc, err := resource.CreateDocument(ctx, targetURL, "", parsedURL, opts.Headers, httpClient)
	if err != nil {
		return nil, err
	}

	// Use the real extraction logic with context
	return h.extractAllFieldsWithContext(ctx, doc, targetURL, parsedURL, *opts)
}

// ParseHTMLWithContext extracts content from supplied HTML with context support.
func (h *Hermes) ParseHTMLWithContext(ctx context.Context, html, targetURL string, opts *ParserOptions) (*Result, error) {
	if opts == nil {
		opts = &h.options
	}
	// Validate URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	validationOpts := validation.DefaultValidationOptions()
	validationOpts.AllowPrivateNetworks = opts.AllowPrivateNetworks
	validationOpts.AllowLocalhost = opts.AllowPrivateNetworks
	validationErr := validation.ValidateParsedURL(ctx, parsedURL, targetURL, validationOpts)
	if validationErr != nil {
		return nil, fmt.Errorf("URL validation failed: %w", validationErr)
	}

	// HTML input is already prepared, so no HTTP client is needed for this path.
	doc, err := resource.CreateDocument(ctx, targetURL, html, parsedURL, opts.Headers, nil)
	if err != nil {
		return nil, err
	}

	// Use the real extraction logic with context
	return h.extractAllFieldsWithContext(ctx, doc, targetURL, parsedURL, *opts)
}
