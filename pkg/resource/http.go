package resource

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient provides a configured HTTP client for fetching resources
type HTTPClient struct {
	client  *http.Client
	headers map[string]string
}

// NewHTTPClient creates a new HTTP client with sensible defaults
func NewHTTPClient(headers map[string]string) *HTTPClient {
	// Use enhanced timeout from constants
	timeout := FETCH_TIMEOUT
	if timeout == 0 {
		timeout = 30 * time.Second // fallback
	}
	
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:       10,
				IdleConnTimeout:    90 * time.Second,
				DisableCompression: false,
				ForceAttemptHTTP2:  false, // Disable HTTP/2 to avoid stuck connection bug
				TLSNextProto:       make(map[string]func(authority string, c *tls.Conn) http.RoundTripper), // Prevent HTTP/2 negotiation
			},
		},
		headers: headers,
	}
}

// Get performs a GET request with optional retries
func (c *HTTPClient) Get(url string) (*Response, error) {
	return c.GetWithRetry(url, 3)
}

// GetWithRetry performs a GET request with specified number of retries
func (c *HTTPClient) GetWithRetry(url string, maxRetries int) (*Response, error) {
	var lastErr error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			delay := time.Duration(1<<uint(attempt-1)) * time.Second
			time.Sleep(delay)
		}
		
		resp, err := c.doRequest(url)
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

// doRequest performs the actual HTTP request
func (c *HTTPClient) doRequest(url string) (*Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.client.Timeout)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	
	// Set all headers using centralized configuration
	allHeaders := MergeHeaders(c.headers)
	for key, value := range allHeaders {
		req.Header.Set(key, value)
	}
	// Note: Accept-Encoding is handled automatically by Go's HTTP client when DisableCompression=false
	
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("performing request: %w", err)
	}
	
	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		// Read error response body before closing for better error reporting
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("HTTP %d: %s (failed to read error response)", resp.StatusCode, resp.Status)
		}
		return &Response{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Headers:    resp.Header,
			Body:       body,
		}, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
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

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Status     string
	Headers    http.Header
	Body       []byte
}

// GetHeader returns a header value
func (r *Response) GetHeader(key string) string {
	return r.Headers.Get(key)
}

// GetContentType returns the content type header
func (r *Response) GetContentType() string {
	return r.GetHeader("Content-Type")
}