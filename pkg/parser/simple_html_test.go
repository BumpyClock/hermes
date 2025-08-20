// ABOUTME: Debug test for simple HTML without article tags to fix content extraction
// ABOUTME: Tests extraction from minimal HTML structure with just body content

package parser_test

import (
	"testing"

	"github.com/BumpyClock/parser-go/pkg/parser"
)

func TestSimpleHTMLDebug(t *testing.T) {
	p := parser.New()

	html := `
	<!DOCTYPE html>
	<html>
	<body>
		<h1>Header Title</h1>
		<p>Some content here.</p>
	</body>
	</html>
	`

	result, err := p.ParseHTML(html, "https://example.com/article", parser.ParserOptions{})
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	t.Logf("Simple HTML Debug:")
	t.Logf("  Title: '%s'", result.Title)
	t.Logf("  Content: '%s'", result.Content)
	t.Logf("  Content Length: %d", len(result.Content))

	// Check what selectors would match
	// This would normally be done with goquery but I'll just log what we're checking
	t.Logf("Fallback selectors: article, .article, #article, .content, #content, .entry-content")
	t.Logf("HTML structure: body > h1 + p")
}