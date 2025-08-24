package hermes_test

import (
	"context"
	"testing"
	"time"

	"github.com/BumpyClock/hermes"
)

func TestNewClient(t *testing.T) {
	// Test creating a new client with default options
	client := hermes.New()
	if client == nil {
		t.Fatal("Expected client to be created")
	}

	// Test creating a client with options
	client = hermes.New(
		hermes.WithTimeout(10*time.Second),
		hermes.WithUserAgent("TestClient/1.0"),
	)
	if client == nil {
		t.Fatal("Expected client to be created with options")
	}
}

func TestParseInvalidURL(t *testing.T) {
	client := hermes.New()
	ctx := context.Background()

	// Test empty URL
	_, err := client.Parse(ctx, "")
	if err == nil {
		t.Fatal("Expected error for empty URL")
	}

	// Check it's a ParseError
	parseErr, ok := err.(*hermes.ParseError)
	if !ok {
		t.Fatalf("Expected ParseError, got %T", err)
	}

	if parseErr.Code != hermes.ErrInvalidURL {
		t.Fatalf("Expected ErrInvalidURL, got %v", parseErr.Code)
	}
}

func TestParseHTMLInvalidInputs(t *testing.T) {
	client := hermes.New()
	ctx := context.Background()

	// Test empty URL
	_, err := client.ParseHTML(ctx, "<html></html>", "")
	if err == nil {
		t.Fatal("Expected error for empty URL")
	}

	// Test empty HTML
	_, err = client.ParseHTML(ctx, "", "https://example.com")
	if err == nil {
		t.Fatal("Expected error for empty HTML")
	}
}

func TestParserInterface(t *testing.T) {
	// Verify that Client implements Parser interface
	var _ hermes.Parser = (*hermes.Client)(nil)
}

func TestErrorTypes(t *testing.T) {
	// Test error code string representations
	codes := []hermes.ErrorCode{
		hermes.ErrInvalidURL,
		hermes.ErrFetch,
		hermes.ErrTimeout,
		hermes.ErrSSRF,
		hermes.ErrExtract,
		hermes.ErrContext,
	}

	for _, code := range codes {
		str := code.String()
		if str == "unknown error" {
			t.Errorf("Unexpected string for error code %d", code)
		}
	}
}

func TestParseError(t *testing.T) {
	// Create a parse error
	err := &hermes.ParseError{
		Code: hermes.ErrFetch,
		URL:  "https://example.com",
		Op:   "Parse",
		Err:  nil,
	}

	// Test error string
	errStr := err.Error()
	if errStr == "" {
		t.Fatal("Expected non-empty error string")
	}

	// Test type checking methods
	if !err.IsFetch() {
		t.Fatal("Expected IsFetch to return true")
	}
	if err.IsTimeout() {
		t.Fatal("Expected IsTimeout to return false")
	}
}

func TestResultHelpers(t *testing.T) {
	// Test empty result
	result := &hermes.Result{}
	if !result.IsEmpty() {
		t.Fatal("Expected IsEmpty to return true for empty result")
	}

	// Test result with content
	result = &hermes.Result{
		Title:   "Test Article",
		Content: "Some content",
	}
	if result.IsEmpty() {
		t.Fatal("Expected IsEmpty to return false for non-empty result")
	}

	// Test author helpers
	if result.HasAuthor() {
		t.Fatal("Expected HasAuthor to return false")
	}
	result.Author = "John Doe"
	if !result.HasAuthor() {
		t.Fatal("Expected HasAuthor to return true")
	}

	// Test date helpers
	if result.HasDate() {
		t.Fatal("Expected HasDate to return false")
	}
	now := time.Now()
	result.DatePublished = &now
	if !result.HasDate() {
		t.Fatal("Expected HasDate to return true")
	}

	// Test image helpers
	if result.HasImage() {
		t.Fatal("Expected HasImage to return false")
	}
	result.LeadImageURL = "https://example.com/image.jpg"
	if !result.HasImage() {
		t.Fatal("Expected HasImage to return true")
	}
}

func TestResultFormatMarkdown(t *testing.T) {
	now := time.Now()
	result := &hermes.Result{
		Title:         "Test Article",
		Content:       "This is the article content.",
		Author:        "John Doe",
		DatePublished: &now,
		URL:           "https://example.com/article",
		SiteName:      "Example Site",
		WordCount:     100,
		Description:   "Article description",
	}

	markdown := result.FormatMarkdown()
	if markdown == "" {
		t.Fatal("Expected non-empty markdown")
	}

	// Check that markdown contains expected elements
	expectedStrings := []string{
		"# Test Article",
		"**Author:** John Doe",
		"**URL:** https://example.com/article",
		"**Word Count:** 100",
		"## Content",
		"This is the article content.",
	}

	for _, expected := range expectedStrings {
		if !contains(markdown, expected) {
			t.Errorf("Expected markdown to contain '%s'", expected)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && len(substr) == 0 || (len(substr) > 0 && findSubstring(s, substr) >= 0))
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}