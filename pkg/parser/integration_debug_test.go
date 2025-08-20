// ABOUTME: Debug test for integration to see exactly what content extractor returns in full pipeline
// ABOUTME: Traces the content extraction step by step to identify where content is lost

package parser

import (
	"testing"
)

func TestIntegrationContentDebug(t *testing.T) {
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Test Article</title>
		<meta name="author" content="Test Author">
	</head>
	<body>
		<article>
			<h1>Test Article</h1>
			<p>This is test content.</p>
		</article>
	</body>
	</html>
	`

	parser := New()
	opts := ParserOptions{
		Fallback:    true,
		ContentType: "html", // Explicitly test HTML content type
	}

	// Let me trace through the extraction
	t.Logf("Testing with ContentType: %s", opts.ContentType)
	
	result, err := parser.ParseHTML(html, "https://example.com/article", opts)
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	t.Logf("Final Result:")
	t.Logf("  Title: '%s'", result.Title)
	t.Logf("  Author: '%s'", result.Author)  
	t.Logf("  Content: '%s'", result.Content)
	t.Logf("  Content Length: %d", len(result.Content))
	t.Logf("  Excerpt: '%s'", result.Excerpt)
	t.Logf("  WordCount: %d", result.WordCount)

	// Test with text content type
	opts.ContentType = "text"
	result2, err := parser.ParseHTML(html, "https://example.com/article", opts)
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	t.Logf("Text Content Result:")
	t.Logf("  Content: '%s'", result2.Content)
	t.Logf("  Content Length: %d", len(result2.Content))
}