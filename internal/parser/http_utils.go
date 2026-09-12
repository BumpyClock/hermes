// ABOUTME: Centralized HTTP client creation and management utilities to eliminate duplication
// ABOUTME: Provides consistent HTTP client wrapping and header handling across the parser layer

package parser

import "github.com/BumpyClock/hermes/internal/resource"

// ensureHTTPClient ensures we have a proper HTTPClient wrapper, creating a default if needed.
func (h *Hermes) ensureHTTPClient(opts *ParserOptions) *resource.HTTPClient {
	client := opts.HTTPClient
	if client == nil {
		h.defaultHTTPClientOnce.Do(func() {
			h.defaultHTTPClient = resource.CreateDefaultHTTPClient().Client
		})
		client = h.defaultHTTPClient
	}

	return &resource.HTTPClient{
		Client:  client,
		Headers: opts.Headers,
	}
}
