// ABOUTME: Debug test for parser integration to verify what fields are being extracted
// ABOUTME: Helps diagnose extraction issues with specific field outputs and error conditions

package parser

import (
	"fmt"
	"testing"
)

const debugHTML = `
<!DOCTYPE html>
<html>
<head>
	<title>Debug Article | Test Site</title>
	<meta name="author" content="Test Author">
	<meta name="description" content="Test description">
	<meta property="article:published_time" content="2023-01-15T10:30:00Z">
	<meta property="og:image" content="https://example.com/image.jpg">
</head>
<body>
	<article>
		<h1>Debug Article</h1>
		<p class="byline">By Test Author</p>
		<div class="content">
			<p>This is the main content of the article.</p>
		</div>
	</article>
</body>
</html>`

func TestParserDebug_Basic(t *testing.T) {
	parser := New()
	opts := ParserOptions{
		Fallback:    true,
		ContentType: "html",
	}

	result, err := parser.ParseHTML(debugHTML, "https://example.com/debug", opts)
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	// Debug output - print all fields
	fmt.Printf("Debug Results:\n")
	fmt.Printf("  URL: %s\n", result.URL)
	fmt.Printf("  Domain: %s\n", result.Domain)
	fmt.Printf("  Title: %s\n", result.Title)
	fmt.Printf("  Author: %s\n", result.Author)
	fmt.Printf("  Content: %s\n", result.Content)
	fmt.Printf("  DatePublished: %v\n", result.DatePublished)
	fmt.Printf("  LeadImageURL: %s\n", result.LeadImageURL)
	fmt.Printf("  Dek: %s\n", result.Dek)
	fmt.Printf("  Excerpt: %s\n", result.Excerpt)
	fmt.Printf("  WordCount: %d\n", result.WordCount)

	// Basic validations
	if result.URL != "https://example.com/debug" {
		t.Errorf("Expected URL to be https://example.com/debug, got: %s", result.URL)
	}

	if result.Domain != "example.com" {
		t.Errorf("Expected domain to be example.com, got: %s", result.Domain)
	}

	// Don't fail on missing fields yet, just check what we get
	t.Logf("Title extracted: %s", result.Title)
	t.Logf("Author extracted: %s", result.Author) 
	t.Logf("Content extracted: %s", result.Content)
}