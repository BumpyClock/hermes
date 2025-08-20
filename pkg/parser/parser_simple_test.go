// ABOUTME: Debug test for simple HTML parsing to diagnose content extraction issues
// ABOUTME: Tests basic HTML extraction to verify compatibility with existing test expectations

package parser

import (
	"strings"
	"testing"
)

func TestSimpleHTML_Debug(t *testing.T) {
	p := New()

	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Test Article</title>
		<meta property="og:title" content="Test Article">
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

	result, err := p.ParseHTML(html, "https://example.com/article", ParserOptions{})
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	t.Logf("Result Debug:")
	t.Logf("  Title: '%s'", result.Title)
	t.Logf("  Author: '%s'", result.Author)
	t.Logf("  Content: '%s'", result.Content)
	t.Logf("  Content Length: %d", len(result.Content))
	t.Logf("  Contains 'This is test content': %v", strings.Contains(result.Content, "This is test content"))

	html2 := `
	<!DOCTYPE html>
	<html>
	<body>
		<h1>Header Title</h1>
		<p>Some content here.</p>
	</body>
	</html>
	`

	result2, err := p.ParseHTML(html2, "https://example.com/article", ParserOptions{})
	if err != nil {
		t.Fatalf("ParseHTML2 failed: %v", err)
	}

	t.Logf("Result2 Debug:")
	t.Logf("  Title: '%s'", result2.Title)
	t.Logf("  Content: '%s'", result2.Content)
	t.Logf("  Content Length: %d", len(result2.Content))
	t.Logf("  Contains 'Some content here': %v", strings.Contains(result2.Content, "Some content here"))
}