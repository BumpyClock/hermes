package hermes

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestHTTPClientInjection verifies that custom HTTP client is actually used.
func TestHTTPClientInjection(t *testing.T) {
	// Create a test server that tracks if it was called
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		// Check that custom header is present
		if r.Header.Get("X-Custom-Header") != "test-value" {
			t.Error("Custom header not found")
		}
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>Test Page</title></head>
<body>
  <article>
    <h1>Test Article</h1>
    <p>This is test content that should be extracted. It contains enough text to be considered valid article content by the parser.</p>
    <p>Another paragraph with more content to ensure extraction works properly.</p>
  </article>
</body>
</html>`))
	}))
	defer ts.Close()

	// Create a custom HTTP client with a custom transport
	customTransport := &customRoundTripper{
		base: http.DefaultTransport,
	}
	customHTTPClient := &http.Client{
		Transport: customTransport,
		Timeout:   10 * time.Second,
	}

	// Create Hermes client with custom HTTP client
	// Note: We need to allow private networks since httptest uses localhost
	client := New(
		WithHTTPClient(customHTTPClient),
		WithUserAgent("TestAgent/1.0"),
		WithAllowPrivateNetworks(true),
	)

	// Parse a URL
	ctx := context.Background()
	result, err := client.Parse(ctx, ts.URL)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify the server was called
	if !called {
		t.Error("Test server was not called")
	}

	if !customTransport.used {
		t.Error("Custom transport was not used - HTTP client injection failed")
	}

	// Verify result - title extraction can vary
	if result.Title == "" {
		t.Error("No title extracted")
	}

	// The important test is that our custom client was used and content was extracted
	if result.Content == "" {
		t.Error("No content extracted")
	}

	// Verify the content contains our test text
	if !strings.Contains(strings.ToLower(result.Content), "test content") {
		t.Errorf("Content does not contain expected text, got: %s", result.Content)
	}
}

// customRoundTripper adds a custom header to all requests.
type customRoundTripper struct {
	base http.RoundTripper
	used bool
}

func (c *customRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	c.used = true
	// Add custom header
	req.Header.Set("X-Custom-Header", "test-value")
	return c.base.RoundTrip(req)
}

// TestSSRFProtection verifies that SSRF protection works.
func TestSSRFProtection(t *testing.T) {

	tests := []struct {
		name    string
		url     string
		allowed bool
	}{
		{"localhost", "http://localhost/test", false},
		{"127.0.0.1", "http://127.0.0.1/test", false},
		{"private IP", "http://192.168.1.1/test", false},
		{"public IP", "http://8.8.8.8/test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			client := New(WithTransport(testTransport(func(req *http.Request) (*http.Response, error) {
				called = true
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": {"text/html"}},
					Body:       io.NopCloser(strings.NewReader("<html><head><title>Allowed</title></head><body>Public content</body></html>")),
					Request:    req,
				}, nil
			})))
			result, err := client.Parse(context.Background(), tt.url)
			if tt.allowed {
				if err != nil {
					t.Fatal(err)
				}
				if !called || result.Title != "Allowed" {
					t.Fatalf("transport called = %v, title = %q", called, result.Title)
				}
			} else {
				var parseErr *ParseError
				if !errors.As(err, &parseErr) || parseErr.Code != ErrSSRF {
					t.Fatalf("expected ErrSSRF, got %v", err)
				}
				if called {
					t.Fatal("blocked URL reached transport")
				}
			}
		})
	}
}

// TestAllowPrivateNetworks verifies that the SSRF protection can be disabled.
func TestAllowPrivateNetworks(t *testing.T) {
	// Create a test server on localhost
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><head><title>Private Network</title></head><body>Content</body></html>`))
	}))
	defer ts.Close()

	// Client with SSRF protection disabled
	client := New(WithAllowPrivateNetworks(true))

	ctx := context.Background()
	result, err := client.Parse(ctx, ts.URL)
	if err != nil {
		t.Fatalf("Parse failed with private networks allowed: %v", err)
	}

	if result.Title != "Private Network" {
		t.Errorf("Expected title 'Private Network', got '%s'", result.Title)
	}
}
