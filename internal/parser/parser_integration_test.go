// ABOUTME: Integration tests for the complete parser pipeline with real HTML content and end-to-end validation
// ABOUTME: Tests full extraction flow from URL/HTML through resource layer to final Result with all fields populated

package parser

import (
	"strings"
	"testing"
	"time"
)

const sampleNewsHTML = `
<!DOCTYPE html>
<html>
<head>
	<title>Sample News Article | News Site</title>
	<meta name="author" content="John Smith">
	<meta name="description" content="This is a test article for parser integration">
	<meta property="article:published_time" content="2023-01-15T10:30:00Z">
	<meta property="og:image" content="https://example.com/image.jpg">
</head>
<body>
	<article>
		<h1>Sample News Article</h1>
		<p class="byline">By John Smith</p>
		<p class="dek">This is a test article for parser integration</p>
		<div class="content">
			<p>This is the main content of the article. It contains multiple paragraphs with meaningful content that should be extracted by the parser.</p>
			<p>This is another paragraph with more content. The parser should be able to extract this along with the previous paragraph.</p>
			<p>And here's a final paragraph to ensure the parser captures all the content properly.</p>
		</div>
	</article>
</body>
</html>`

func TestParserIntegration_ParseHTML_BasicExtraction(t *testing.T) {
	parser := New()
	opts := ParserOptions{
		Fallback:    true,
		ContentType: "html",
	}

	result, err := parser.ParseHTML(sampleNewsHTML, "https://example.com/article", &opts)
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	// Verify all fields are extracted
	if result.Title == "" {
		t.Error("Expected title to be extracted")
	}
	
	if result.Author == "" {
		t.Error("Expected author to be extracted")
	}

	if result.Content == "" {
		t.Error("Expected content to be extracted")
	}

	if result.DatePublished == nil {
		t.Error("Expected date published to be extracted")
	}

	if result.LeadImageURL == "" {
		t.Error("Expected lead image URL to be extracted")
	}

	if result.URL != "https://example.com/article" {
		t.Errorf("Expected URL to be %s, got %s", "https://example.com/article", result.URL)
	}

	if result.Domain != "example.com" {
		t.Errorf("Expected domain to be example.com, got %s", result.Domain)
	}

	// Test specific values
	if !contains(result.Title, "Sample News Article") {
		t.Errorf("Expected title to contain 'Sample News Article', got: %s", result.Title)
	}

	if !contains(result.Author, "John Smith") {
		t.Errorf("Expected author to contain 'John Smith', got: %s", result.Author)
	}

	if !contains(result.Content, "main content of the article") {
		t.Errorf("Expected content to contain article text, got: %s", result.Content)
	}

	// Verify date parsing
	expectedTime, _ := time.Parse(time.RFC3339, "2023-01-15T10:30:00Z")
	if !result.DatePublished.Equal(expectedTime) {
		t.Errorf("Expected date to be %v, got %v", expectedTime, result.DatePublished)
	}

	if result.LeadImageURL != "https://example.com/image.jpg" {
		t.Errorf("Expected lead image URL to be https://example.com/image.jpg, got: %s", result.LeadImageURL)
	}
}

func TestParserIntegration_ParseHTML_ContentTypes(t *testing.T) {
	parser := New()
	
	tests := []struct {
		name        string
		contentType string
		expectHTML  bool
	}{
		{
			name:        "HTML content type",
			contentType: "html",
			expectHTML:  true,
		},
		{
			name:        "Text content type",
			contentType: "text",
			expectHTML:  false,
		},
		{
			name:        "Markdown content type", 
			contentType: "markdown",
			expectHTML:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := ParserOptions{
				Fallback:    true,
				ContentType: tt.contentType,
			}

			result, err := parser.ParseHTML(sampleNewsHTML, "https://example.com/article", &opts)
			if err != nil {
				t.Fatalf("ParseHTML failed: %v", err)
			}

			if tt.expectHTML {
				// HTML should contain tags
				if !contains(result.Content, "<") {
					t.Error("Expected HTML content to contain tags")
				}
			} else {
				// Text/Markdown should not contain HTML tags
				if contains(result.Content, "<p>") || contains(result.Content, "</p>") {
					t.Error("Expected non-HTML content to not contain HTML tags")
				}
			}
		})
	}
}

func TestParserIntegration_ParseHTML_FallbackBehavior(t *testing.T) {
	parser := New()

	// Test with fallback enabled
	opts := ParserOptions{
		Fallback:    true,
		ContentType: "html",
	}

	result, err := parser.ParseHTML(sampleNewsHTML, "https://example.com/article", &opts)
	if err != nil {
		t.Fatalf("ParseHTML with fallback failed: %v", err)
	}

	if result.Content == "" {
		t.Error("Expected content extraction with fallback enabled")
	}

	// Test with fallback disabled (should still work with generic extractor)
	opts.Fallback = false
	result2, err := parser.ParseHTML(sampleNewsHTML, "https://example.com/article", &opts)
	if err != nil {
		t.Fatalf("ParseHTML without fallback failed: %v", err)
	}

	if result2.Content == "" {
		t.Error("Expected content extraction even without fallback")
	}
}

func TestParserIntegration_ParseHTML_ErrorHandling(t *testing.T) {
	parser := New()
	opts := ParserOptions{
		Fallback:    true,
		ContentType: "html",
	}

	// Test with invalid URL
	_, err := parser.ParseHTML(sampleNewsHTML, "not-a-url", &opts)
	if err == nil {
		t.Error("Expected error for invalid URL")
	}

	// Test with malformed HTML (should still work)
	malformedHTML := `<html><body><p>Unclosed paragraph`
	result, err := parser.ParseHTML(malformedHTML, "https://example.com/malformed", &opts)
	if err != nil {
		t.Fatalf("ParseHTML should handle malformed HTML: %v", err)
	}

	if result.URL != "https://example.com/malformed" {
		t.Error("Expected URL to be set even with malformed HTML")
	}
}

func TestParserIntegration_ParseHTML_EmptyContent(t *testing.T) {
	parser := New()
	opts := ParserOptions{
		Fallback:    true,
		ContentType: "html",
	}

	emptyHTML := `<html><head><title>Empty</title></head><body></body></html>`
	
	result, err := parser.ParseHTML(emptyHTML, "https://example.com/empty", &opts)
	if err != nil {
		t.Fatalf("ParseHTML should handle empty content: %v", err)
	}

	// Should still extract basic info
	if result.URL == "" {
		t.Error("Expected URL to be set")
	}
	
	if result.Domain == "" {
		t.Error("Expected domain to be set")
	}

	if result.Title != "Empty" {
		t.Errorf("Expected title to be 'Empty', got: %s", result.Title)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}