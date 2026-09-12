// ABOUTME: Main parser implementation integrating all extractors and cleaners into complete extraction pipeline
// ABOUTME: Wires together resource layer, generic extractors, and content cleaners to create working end-to-end parser
// ABOUTME: Originally inspired by the Postlight Mercury parser, now rebranded as Hermes for consistency

package parser

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/BumpyClock/hermes/internal/resource"
	"github.com/BumpyClock/hermes/internal/validation"
)

// Hermes (formerly Mercury) is the main parser implementation.
type Hermes struct {
	options               ParserOptions
	defaultHTTPClientOnce sync.Once
	defaultHTTPClient     *http.Client
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

// ParseWithContext extracts content from a URL with context support.
func (h *Hermes) ParseWithContext(ctx context.Context, targetURL string, opts *ParserOptions) (*Result, error) {
	if opts == nil {
		opts = &h.options
	}
	// Validate URL
	parsedURL, err := parseAndValidateURL(ctx, targetURL, opts.AllowPrivateNetworks)
	if err != nil {
		return nil, err
	}

	// Use centralized HTTP client creation
	httpClient := h.ensureHTTPClient(opts)

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
	parsedURL, err := parseAndValidateURL(ctx, targetURL, opts.AllowPrivateNetworks)
	if err != nil {
		return nil, err
	}

	// HTML input is already prepared, so no HTTP client is needed for this path.
	doc, err := resource.CreateDocument(ctx, targetURL, html, parsedURL, opts.Headers, nil)
	if err != nil {
		return nil, err
	}

	// Use the real extraction logic with context
	return h.extractAllFieldsWithContext(ctx, doc, targetURL, parsedURL, *opts)
}

func parseAndValidateURL(ctx context.Context, targetURL string, allowPrivateNetworks bool) (*url.URL, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	validationOpts := validation.DefaultValidationOptions()
	validationOpts.AllowPrivateNetworks = allowPrivateNetworks
	validationOpts.AllowLocalhost = allowPrivateNetworks
	if err := validation.ValidateParsedURL(ctx, parsedURL, targetURL, validationOpts); err != nil {
		return nil, fmt.Errorf("URL validation failed: %w", err)
	}

	return parsedURL, nil
}
