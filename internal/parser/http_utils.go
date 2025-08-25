// ABOUTME: Centralized HTTP client creation and management utilities to eliminate duplication
// ABOUTME: Provides consistent HTTP client wrapping and header handling across the parser layer

package parser

import (
	"net/http"

	"github.com/BumpyClock/hermes/internal/resource"
)

// createHTTPClientWrapper wraps an http.Client with headers in a consistent way
// This function eliminates the duplication of HTTP client wrapping logic
func createHTTPClientWrapper(httpClient *http.Client, headers map[string]string) *resource.HTTPClient {
	if httpClient == nil {
		// Should not happen, but defensive programming
		httpClient = http.DefaultClient
	}
	
	return &resource.HTTPClient{
		Client:  httpClient,
		Headers: headers,
	}
}

// ensureHTTPClient ensures we have a proper HTTPClient wrapper, creating a default if needed
// This centralizes the logic for HTTP client creation and header management
func ensureHTTPClient(opts *ParserOptions) *resource.HTTPClient {
	if opts.HTTPClient != nil {
		// Create HTTPClient wrapper for the provided client
		return createHTTPClientWrapper(opts.HTTPClient, opts.Headers)
	}
	
	// Create a default HTTP client when none is provided
	defaultClient := resource.CreateDefaultHTTPClient()
	defaultClient.Headers = opts.Headers
	return defaultClient
}

// ensureHTTPClientForHTML ensures we have a proper HTTPClient wrapper for HTML parsing
// Even though HTML parsing doesn't need HTTP, we keep this for API consistency
func ensureHTTPClientForHTML(opts *ParserOptions) *resource.HTTPClient {
	// Use the same logic as regular parsing for consistency
	return ensureHTTPClient(opts)
}