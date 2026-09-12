package resource

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"time"
)

// CreateDefaultHTTPClient creates a new HTTP client with default settings
// This is used when no custom client is provided.
func CreateDefaultHTTPClient() *HTTPClient {
	// Create cookie jar
	jar, err := cookiejar.New(nil)
	if err != nil {
		// If cookie jar creation fails, create client without it
		jar = nil
	}

	// Create client with optimized connection pooling
	client := &http.Client{
		Timeout: FETCH_TIMEOUT,
		Jar:     jar,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			DisableCompression:  false,
			ForceAttemptHTTP2:   false, // Keep HTTP/2 disabled for stability
			TLSNextProto:        make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("stopped after 5 redirects")
			}
			return nil
		},
	}

	return &HTTPClient{
		Client:  client,
		Headers: make(map[string]string),
	}
}

// Fetch retrieves and validates a response with the supplied HTTP client.
func Fetch(ctx context.Context, rawURL string, parsedURL *url.URL, headers map[string]string, httpClient *HTTPClient) (*Response, error) {
	// Parse URL if not provided
	if parsedURL == nil {
		var err error
		parsedURL, err = url.Parse(rawURL)
		if err != nil {
			//nolint:staticcheck // ST1005: Public error text must remain identical.
			return nil, fmt.Errorf("Invalid URL: %v", err)
		}
	}

	// Require HTTP client to be provided
	if httpClient == nil {
		return nil, fmt.Errorf("HTTP client is required")
	}

	// Create a temporary client wrapper with request-specific headers.
	// HTTPClient.doRequest performs the default-header merge once.
	clientWithHeaders := &HTTPClient{
		Client:  httpClient.Client,
		Headers: mergeHeaders(httpClient.Headers, headers),
	}

	// Perform request with retry using the pooled client
	response, err := clientWithHeaders.Get(ctx, parsedURL.String())
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}

	// Validate response
	if err := ValidateResponse(response, false); err != nil {
		return nil, err
	}

	return response, nil
}

// ValidateResponse validates that the response is suitable for parsing.
func ValidateResponse(response *Response, parseNon200 bool) error {
	// Check status code
	if response.StatusCode != 200 {
		if !parseNon200 {
			//nolint:staticcheck // ST1005: Public error text must remain identical.
			return fmt.Errorf("Resource returned a response status code of %d and resource was instructed to reject non-200 status codes", response.StatusCode)
		}
	}

	contentType := response.GetContentType()
	contentLengthStr := response.GetHeader("Content-Length")

	// Check content type
	if BAD_CONTENT_TYPES_RE.MatchString(contentType) {
		return fmt.Errorf("content-type for this resource was %s and is not allowed", contentType)
	}

	// Check content length
	if contentLengthStr != "" {
		contentLength, err := strconv.ParseInt(contentLengthStr, 10, 64)
		if err == nil && contentLength > MAX_CONTENT_LENGTH {
			return fmt.Errorf("content for this resource was too large. Maximum content length is %d", MAX_CONTENT_LENGTH)
		}
	}

	return nil
}
