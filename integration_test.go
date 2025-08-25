package hermes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Helper function to check if content contains substring
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// TestHTTPClientInjection verifies that custom HTTP client is actually used
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
		w.Write([]byte(`<!DOCTYPE html>
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
	
	// Debug output
	t.Logf("Result Title: %s", result.Title)
	t.Logf("Result Content length: %d", len(result.Content))
	t.Logf("Result Excerpt: %s", result.Excerpt)

	// Verify the server was called
	if !called {
		t.Error("Test server was not called")
	}

	// Debug: Check if custom transport was used
	if !customTransport.used {
		t.Error("Custom transport was not used - HTTP client injection failed")
	} else {
		t.Log("Custom transport was successfully used")
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
	if !contains(result.Content, "test content") {
		t.Errorf("Content does not contain expected text, got: %s", result.Content)
	}
}

// customRoundTripper adds a custom header to all requests
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

// TestContextCancellation verifies that context cancellation works
func TestContextCancellation(t *testing.T) {
	// Create a test server that delays response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond) // Delay longer than our context timeout
		w.Write([]byte(`<html><body>Too late</body></html>`))
	}))
	defer ts.Close()

	client := New(WithAllowPrivateNetworks(true)) // Allow localhost for testing

	// Create a context with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Try to parse - should fail with timeout
	_, err := client.Parse(ctx, ts.URL)
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}

	// Check that it's a timeout or fetch error (context cancelled)
	if perr, ok := err.(*ParseError); ok {
		if perr.Code != ErrTimeout && perr.Code != ErrFetch {
			t.Errorf("Expected ErrTimeout or ErrFetch, got %v", perr.Code)
		}
	} else {
		t.Errorf("Expected ParseError, got %T", err)
	}
}

// TestSSRFProtection verifies that SSRF protection works
func TestSSRFProtection(t *testing.T) {
	client := New()

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
			ctx := context.Background()
			_, err := client.Parse(ctx, tt.url)
			
			if tt.allowed {
				// Should work (though might fail for other reasons)
				// We're just checking it doesn't fail with SSRF error
				if err != nil {
					if perr, ok := err.(*ParseError); ok && perr.Code == ErrSSRF {
						t.Errorf("URL %s should be allowed but got SSRF error", tt.url)
					}
				}
			} else {
				// Should fail with SSRF or related error
				if err == nil {
					t.Errorf("URL %s should be blocked but parsing succeeded", tt.url)
				}
			}
		})
	}
}

// TestAllowPrivateNetworks verifies that the SSRF protection can be disabled
func TestAllowPrivateNetworks(t *testing.T) {
	// Create a test server on localhost
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><head><title>Private Network</title></head><body>Content</body></html>`))
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