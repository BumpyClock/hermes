package hermes

import (
	"context"
	"net/http"
	"testing"
	"time"
)

// TestRealURL tests parsing a real URL from The Verge
func TestRealURL(t *testing.T) {
	// Create a custom HTTP client to verify it's being used
	customTransport := &testRoundTripper{
		base: http.DefaultTransport,
	}
	customHTTPClient := &http.Client{
		Transport: customTransport,
		Timeout:   30 * time.Second,
	}

	// Create Hermes client with custom HTTP client
	client := New(
		WithHTTPClient(customHTTPClient),
		WithUserAgent("Hermes-Test/1.0"),
	)

	// Parse The Verge article
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	url := "https://www.theverge.com/notepad-microsoft-newsletter/763357/microsoft-asus-xbox-ally-handheld-hands-on-notepad"
	result, err := client.Parse(ctx, url)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify custom transport was used
	if !customTransport.used {
		t.Error("Custom transport was not used - HTTP client injection failed")
	} else {
		t.Log("✓ Custom HTTP client was successfully injected and used")
	}

	// Verify we got content
	if result.Title == "" {
		t.Error("No title extracted")
	} else {
		t.Logf("✓ Title: %s", result.Title)
	}

	if result.Content == "" {
		t.Error("No content extracted")
	} else {
		t.Logf("✓ Content extracted: %d characters", len(result.Content))
	}

	if result.Author != "" {
		t.Logf("✓ Author: %s", result.Author)
	}

	if result.DatePublished != nil {
		t.Logf("✓ Date: %s", result.DatePublished.Format("2006-01-02"))
	}
	
	if result.LeadImageURL != "" {
		t.Logf("✓ Lead image: %s", result.LeadImageURL)
	}

	// Log site metadata
	if result.SiteName != "" {
		t.Logf("✓ Site name: %s", result.SiteName)
	}
	
	t.Logf("✓ Word count: %d", result.WordCount)
}

// testRoundTripper tracks if it was used
type testRoundTripper struct {
	base http.RoundTripper
	used bool
}

func (c *testRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	c.used = true
	// Add custom header to verify our client is being used
	req.Header.Set("X-Test-Client", "true")
	return c.base.RoundTrip(req)
}