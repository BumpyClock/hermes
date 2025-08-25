package hermes

import (
	"context"
)

// Parser is the interface for content extraction.
// Implement this interface to create mock parsers for testing.
type Parser interface {
	// Parse extracts content from the given URL.
	// The context can be used to cancel the request or set a deadline.
	Parse(ctx context.Context, url string) (*Result, error)
	
	// ParseHTML extracts content from pre-fetched HTML.
	// This is useful when you already have the HTML content.
	ParseHTML(ctx context.Context, html, url string) (*Result, error)
}

// Ensure Client implements the Parser interface
var _ Parser = (*Client)(nil)