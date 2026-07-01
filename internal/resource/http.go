package resource

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/BumpyClock/hermes/internal/pools"
)

// HTTPClient provides a configured HTTP client for fetching resources.
type HTTPClient struct {
	Client  *http.Client      // Exported for external use
	Headers map[string]string // Exported for external use
}

// NewHTTPClient creates a new HTTP client with sensible defaults.
func NewHTTPClient(headers map[string]string) *HTTPClient {
	client := CreateDefaultHTTPClient()
	client.Headers = headers
	return client
}

// Get performs a GET request with optional retries.
func (c *HTTPClient) Get(ctx context.Context, url string) (*Response, error) {
	return c.GetWithRetry(ctx, url, 3)
}

// GetWithRetry performs a GET request with specified number of retries.
func (c *HTTPClient) GetWithRetry(ctx context.Context, url string, maxRetries int) (*Response, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check if context is cancelled before each attempt
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("context cancelled: %w", err)
		}

		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			delay := time.Duration(1<<uint(attempt-1)) * time.Second

			// Use context-aware sleep
			select {
			case <-time.After(delay):
				// Continue with retry
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during retry backoff: %w", ctx.Err())
			}
		}

		resp, err := c.doRequest(ctx, url)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Don't retry on client errors (4xx)
		if resp != nil && resp.StatusCode >= 400 && resp.StatusCode < 500 {
			break
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries+1, lastErr)
}

// doRequest performs the actual HTTP request.
func (c *HTTPClient) doRequest(ctx context.Context, url string) (*Response, error) {
	// Use the provided context directly - no longer create our own timeout
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Set all headers using centralized configuration
	allHeaders := MergeHeaders(c.Headers)
	for key, value := range allHeaders {
		req.Header.Set(key, value)
	}

	// Note: Accept-Encoding is handled automatically by Go's HTTP client when DisableCompression=false
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("performing request: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		// Read error response body using pooled buffer for better error reporting
		errorBody, readErr := pools.GlobalResponseBodyPool.ReadResponseBody(resp)
		if readErr != nil {
			return nil, fmt.Errorf("HTTP %d: %s (failed to read error response)", resp.StatusCode, resp.Status)
		}
		return &Response{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Headers:    resp.Header,
			Body:       errorBody,
		}, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Read response body using pooled buffer for efficiency
	body, err := pools.GlobalResponseBodyPool.ReadResponseBody(resp)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    resp.Header,
		Body:       body,
	}, nil
}

// Response represents an HTTP response.
type Response struct {
	StatusCode int
	Status     string
	Headers    http.Header
	Body       []byte
}

// GetHeader returns a header value.
func (r *Response) GetHeader(key string) string {
	return r.Headers.Get(key)
}

// GetContentType returns the content type header.
func (r *Response) GetContentType() string {
	return r.GetHeader("Content-Type")
}
