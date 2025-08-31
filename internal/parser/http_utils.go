// ABOUTME: Centralized HTTP client creation and management utilities to eliminate duplication
// ABOUTME: Provides consistent HTTP client wrapping and header handling across the parser layer

package parser

import (
	"fmt"
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
	fmt.Printf("[HERMES FETCH DEBUG] ensureHTTPClient called\n")
	fmt.Printf("[HERMES FETCH DEBUG] - HTTPClient provided: %t\n", opts.HTTPClient != nil)
	
	if opts.HTTPClient != nil {
		fmt.Printf("[HERMES FETCH DEBUG] Using provided HTTP client (from integration)\n")
		// Create HTTPClient wrapper for the provided client
		wrapper := createHTTPClientWrapper(opts.HTTPClient, opts.Headers)
		
		// Log the wrapped client details
		if transport, ok := wrapper.Client.Transport.(*http.Transport); ok {
			fmt.Printf("[HERMES FETCH DEBUG] Provided client transport config:\n")
			fmt.Printf("[HERMES FETCH DEBUG] - MaxIdleConns: %d\n", transport.MaxIdleConns)
			fmt.Printf("[HERMES FETCH DEBUG] - ForceAttemptHTTP2: %t\n", transport.ForceAttemptHTTP2)
			fmt.Printf("[HERMES FETCH DEBUG] - TLSNextProto disabled: %t\n", len(transport.TLSNextProto) == 0)
		}
		
		return wrapper
	}
	
	fmt.Printf("[HERMES FETCH DEBUG] Creating default HTTP client (CreateDefaultHTTPClient)\n")
	// Create a default HTTP client when none is provided
	defaultClient := resource.CreateDefaultHTTPClient()
	defaultClient.Headers = opts.Headers
	
	// Log the default client details
	if transport, ok := defaultClient.Client.Transport.(*http.Transport); ok {
		fmt.Printf("[HERMES FETCH DEBUG] Default client transport config:\n")
		fmt.Printf("[HERMES FETCH DEBUG] - MaxIdleConns: %d\n", transport.MaxIdleConns)
		fmt.Printf("[HERMES FETCH DEBUG] - ForceAttemptHTTP2: %t\n", transport.ForceAttemptHTTP2)
		fmt.Printf("[HERMES FETCH DEBUG] - TLSNextProto disabled: %t\n", len(transport.TLSNextProto) == 0)
	}
	
	return defaultClient
}

// ensureHTTPClientForHTML ensures we have a proper HTTPClient wrapper for HTML parsing
// Even though HTML parsing doesn't need HTTP, we keep this for API consistency
func ensureHTTPClientForHTML(opts *ParserOptions) *resource.HTTPClient {
	// Use the same logic as regular parsing for consistency
	return ensureHTTPClient(opts)
}