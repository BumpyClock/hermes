// ABOUTME: Debug test in external package to match failing test conditions exactly
// ABOUTME: Tests from parser_test perspective to identify why content extraction differs

package parser_test

import (
	"testing"

	"github.com/BumpyClock/parser-go/pkg/parser"
)

func TestExternalDebug(t *testing.T) {
	p := parser.New()

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

	result, err := p.ParseHTML(html, "https://example.com/article", parser.ParserOptions{})
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	t.Logf("External Debug Result:")
	t.Logf("  Title: '%s'", result.Title)
	t.Logf("  Author: '%s'", result.Author)
	t.Logf("  Content: '%s'", result.Content)
	t.Logf("  Content Length: %d", len(result.Content))
	t.Logf("  Error: %v", result.Error)
	t.Logf("  Message: '%s'", result.Message)

	// Check what the actual assertion is testing
	if result.Content == "" {
		t.Logf("Content is empty - this explains the test failure")
	} else {
		t.Logf("Content is not empty - assertion should pass")
	}
}