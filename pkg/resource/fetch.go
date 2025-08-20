package resource

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// FetchResource fetches a resource from the given URL with retry logic
func FetchResource(rawURL string, parsedURL *url.URL, headers map[string]string) (*FetchResult, error) {
	// Parse URL if not provided
	if parsedURL == nil {
		var err error
		parsedURL, err = url.Parse(rawURL)
		if err != nil {
			return &FetchResult{
				Error:   true,
				Message: fmt.Sprintf("Invalid URL: %v", err),
			}, nil
		}
	}

	// Create HTTP client with cookie jar
	jar, err := cookiejar.New(nil)
	if err != nil {
		return &FetchResult{
			Error:   true,
			Message: fmt.Sprintf("Failed to create cookie jar: %v", err),
		}, nil
	}

	// Use centralized header merging
	allHeaders := MergeHeaders(headers)

	// Create enhanced HTTP client
	client := &HTTPClient{
		client: &http.Client{
			Timeout: FETCH_TIMEOUT,
			Jar:     jar,
			Transport: &http.Transport{
				MaxIdleConns:       10,
				IdleConnTimeout:    90 * time.Second,
				DisableCompression: false,
				ForceAttemptHTTP2:  false, // Disable HTTP/2 to avoid stuck connection bug
				TLSNextProto:       make(map[string]func(authority string, c *tls.Conn) http.RoundTripper), // Prevent HTTP/2 negotiation
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 5 {
					return fmt.Errorf("stopped after 5 redirects")
				}
				return nil
			},
		},
		headers: allHeaders,
	}

	// Perform request with retry
	response, err := client.Get(parsedURL.String())
	if err != nil {
		return &FetchResult{
			Error:   true,
			Message: fmt.Sprintf("HTTP request failed: %v", err),
		}, nil
	}

	// Validate response
	if err := ValidateResponse(response, false); err != nil {
		return &FetchResult{
			Error:   true,
			Message: err.Error(),
		}, nil
	}

	return &FetchResult{
		Response: response,
	}, nil
}

// ValidateResponse validates that the response is suitable for parsing
func ValidateResponse(response *Response, parseNon200 bool) error {
	// Check status code
	if response.StatusCode != 200 {
		if !parseNon200 {
			return fmt.Errorf("Resource returned a response status code of %d and resource was instructed to reject non-200 status codes", response.StatusCode)
		}
	}

	contentType := response.GetContentType()
	contentLengthStr := response.GetHeader("Content-Length")

	// Check content type
	if BAD_CONTENT_TYPES_RE.MatchString(contentType) {
		return fmt.Errorf("Content-type for this resource was %s and is not allowed", contentType)
	}

	// Check content length
	if contentLengthStr != "" {
		contentLength, err := strconv.ParseInt(contentLengthStr, 10, 64)
		if err == nil && contentLength > MAX_CONTENT_LENGTH {
			return fmt.Errorf("Content for this resource was too large. Maximum content length is %d", MAX_CONTENT_LENGTH)
		}
	}

	return nil
}

// BaseDomain extracts the base domain from a host
// Gets the last two pieces of the URL and joins them back together
// This is to get 'livejournal.com' from 'erotictrains.livejournal.com'
func BaseDomain(host string) string {
	parts := strings.Split(host, ".")
	if len(parts) < 2 {
		return host
	}
	
	return strings.Join(parts[len(parts)-2:], ".")
}

// FetchResult represents the result of fetching a resource
type FetchResult struct {
	Response      *Response
	Error         bool
	Message       string
	AlreadyDecoded bool
}

// IsError returns true if the fetch result contains an error
func (fr *FetchResult) IsError() bool {
	return fr.Error
}