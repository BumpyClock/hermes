// ABOUTME: Centralized HTTP client creation and management utilities to eliminate duplication
// ABOUTME: Provides consistent HTTP client wrapping and header handling across the parser layer

package parser

import "github.com/BumpyClock/hermes/internal/resource"

// ensureHTTPClient ensures we have a proper HTTPClient wrapper, creating a default if needed.
func ensureHTTPClient(opts *ParserOptions) *resource.HTTPClient {
	if opts.HTTPClient != nil {
		return &resource.HTTPClient{
			Client:  opts.HTTPClient,
			Headers: opts.Headers,
		}
	}

	defaultClient := resource.CreateDefaultHTTPClient()
	defaultClient.Headers = opts.Headers
	return defaultClient
}
