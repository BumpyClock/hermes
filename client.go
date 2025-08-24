package hermes

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/BumpyClock/hermes/internal/parser"
	"github.com/BumpyClock/hermes/internal/validation"
)

// Client is a thread-safe, reusable parser client for extracting content from web pages.
// It manages its own HTTP client for connection pooling and can be shared across goroutines.
type Client struct {
	httpClient           *http.Client
	userAgent            string
	timeout              time.Duration
	allowPrivateNetworks bool
	contentType          string
	
	// Internal parser instance
	parser *parser.Mercury
}

// New creates a new Hermes client with the provided options.
// The client is thread-safe and should be reused across requests.
//
// Example:
//
//	client := hermes.New(
//	    hermes.WithTimeout(30*time.Second),
//	    hermes.WithUserAgent("MyApp/1.0"),
//	)
func New(opts ...Option) *Client {
	// Default configuration
	c := &Client{
		userAgent: "Hermes/1.0",
		timeout:   30 * time.Second,
		allowPrivateNetworks: false,
		contentType: "html",
	}
	
	// Apply options
	for _, opt := range opts {
		opt(c)
	}
	
	// Create HTTP client if not provided
	if c.httpClient == nil {
		c.httpClient = &http.Client{
			Timeout: c.timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
				DisableCompression:  false,
				// Re-enable HTTP/2 by default (remove old workaround)
			},
		}
	}
	
	// Create internal parser
	// Note: HTTP client will be passed through headers/options
	// until we can refactor the parser to accept it directly
	c.parser = parser.New()
	
	return c
}

// Parse extracts content from the given URL.
// The context can be used to cancel the request or set a deadline.
//
// Example:
//
//	ctx := context.Background()
//	result, err := client.Parse(ctx, "https://example.com/article")
//	if err != nil {
//	    // Handle error
//	}
//	fmt.Println(result.Title)
func (c *Client) Parse(ctx context.Context, url string) (*Result, error) {
	// Validate URL
	if url == "" {
		return nil, &ParseError{
			Code: ErrInvalidURL,
			URL:  url,
			Op:   "Parse",
			Err:  fmt.Errorf("empty URL"),
		}
	}
	
	// Create parser options with client configuration
	opts := c.buildParserOptions()
	
	// Parse the URL with context support
	internalResult, err := c.parser.ParseWithContext(ctx, url, opts)
	if err != nil {
		// Use proper error classification instead of string matching
		code := ErrorCode(parser.ClassifyErrorCode(err, ctx, "Parse"))
		// Wrap error with type information
		return nil, &ParseError{
			Code: code,
			URL:  url,
			Op:   "Parse",
			Err:  err,
		}
	}
	
	// Map internal result to public result
	result := mapInternalResult(internalResult)
	return result, nil
}

// ParseHTML extracts content from pre-fetched HTML.
// This is useful when you already have the HTML content and want to avoid an additional HTTP request.
//
// Example:
//
//	html := "<html>...</html>"
//	result, err := client.ParseHTML(ctx, html, "https://example.com/article")
func (c *Client) ParseHTML(ctx context.Context, html, url string) (*Result, error) {
	// Validate inputs
	if url == "" {
		return nil, &ParseError{
			Code: ErrInvalidURL,
			URL:  url,
			Op:   "ParseHTML",
			Err:  fmt.Errorf("empty URL"),
		}
	}
	
	if html == "" {
		return nil, &ParseError{
			Code: ErrInvalidURL,
			URL:  url,
			Op:   "ParseHTML",
			Err:  fmt.Errorf("empty HTML content"),
		}
	}
	
	// Validate URL format
	validationOpts := validation.DefaultValidationOptions()
	validationOpts.AllowPrivateNetworks = c.allowPrivateNetworks
	validationOpts.AllowLocalhost = c.allowPrivateNetworks // Localhost should be allowed when private networks are allowed
	
	if err := validation.ValidateURL(ctx, url, validationOpts); err != nil {
		return nil, &ParseError{
			Code: ErrInvalidURL,
			URL:  url,
			Op:   "ParseHTML",
			Err:  err,
		}
	}
	
	// Create parser options with client configuration
	opts := c.buildParserOptions()
	
	// Parse the HTML with context support
	internalResult, err := c.parser.ParseHTMLWithContext(ctx, html, url, opts)
	if err != nil {
		// Use proper error classification instead of hardcoded ErrExtract
		code := ErrorCode(parser.ClassifyErrorCode(err, ctx, "ParseHTML"))
		// Wrap error with type information
		return nil, &ParseError{
			Code: code,
			URL:  url,
			Op:   "ParseHTML",
			Err:  err,
		}
	}
	
	// Map internal result to public result
	result := mapInternalResult(internalResult)
	return result, nil
}

// buildParserOptions creates parser options with client configuration
// This centralizes the option building logic to avoid duplication
func (c *Client) buildParserOptions() *parser.ParserOptions {
	return &parser.ParserOptions{
		FetchAllPages:        false,
		ContentType:          c.contentType,
		Headers:              map[string]string{"User-Agent": c.userAgent},
		HTTPClient:           c.httpClient,
		AllowPrivateNetworks: c.allowPrivateNetworks,
	}
}

// mapInternalResult converts the internal parser.Result to our public Result type
func mapInternalResult(internal *parser.Result) *Result {
	if internal == nil {
		return nil
	}
	
	return &Result{
		URL:           internal.URL,
		Title:         internal.Title,
		Content:       internal.Content,
		Author:        internal.Author,
		DatePublished: internal.DatePublished,
		LeadImageURL:  internal.LeadImageURL,
		Dek:           internal.Dek,
		Domain:        internal.Domain,
		Excerpt:       internal.Excerpt,
		WordCount:     internal.WordCount,
		Direction:     internal.Direction,
		TotalPages:    internal.TotalPages,
		RenderedPages: internal.RenderedPages,
		SiteName:      internal.SiteName,
		Description:   internal.Description,
		Language:      internal.Language,
	}
}