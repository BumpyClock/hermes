package hermes

import (
	"context"
	"strings"
	"testing"
)

func TestContentTypeMarkdownBehavior(t *testing.T) {
	// This test verifies that ContentType: "markdown" only affects the Content field,
	// while all other fields remain as properly typed values.
	client := New(WithContentType("markdown"))

	// Test HTML with various content types
	html := `
	<html>
		<head>
			<title>Test Article Title</title>
			<meta name="author" content="John Doe">
			<meta name="description" content="This is a test article">
		</head>
		<body>
			<article>
				<header>
					<h1>Main Heading</h1>
					<p class="author">By John Doe</p>
				</header>
				<div class="content">
					<p>This is the first paragraph with <strong>bold text</strong>.</p>
					<p>This is the second paragraph with <em>italic text</em>.</p>
					<ul>
						<li>List item 1</li>
						<li>List item 2</li>
					</ul>
				</div>
			</article>
		</body>
	</html>
	`

	result, err := client.ParseHTML(context.Background(), html, "https://example.com/test")
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	// Test that metadata fields are properly typed (not markdown)
	if result.Title == "" {
		t.Error("Title should not be empty")
	}
	if strings.Contains(result.Title, "#") || strings.Contains(result.Title, "**") {
		t.Errorf("Title should not contain markdown formatting: %s", result.Title)
	}

	// Author might be empty with generic extraction - that's OK
	if result.Author != "" && (strings.Contains(result.Author, "#") || strings.Contains(result.Author, "**")) {
		t.Errorf("Author should not contain markdown formatting: %s", result.Author)
	}

	if result.Description == "" {
		t.Error("Description should not be empty")
	}
	if strings.Contains(result.Description, "#") || strings.Contains(result.Description, "**") {
		t.Errorf("Description should not contain markdown formatting: %s", result.Description)
	}

	// Test that Content field IS in markdown format
	if result.Content == "" {
		t.Error("Content should not be empty")
	}
	
	// Content should contain markdown formatting
	hasMarkdown := strings.Contains(result.Content, "**") || strings.Contains(result.Content, "*") || strings.Contains(result.Content, "#")
	if !hasMarkdown {
		t.Errorf("Content should contain markdown formatting but got: %s", result.Content)
	}

	// Test that other fields are properly typed
	if result.WordCount <= 0 {
		t.Error("WordCount should be positive")
	}

	// Print results for manual verification
	t.Logf("Title: %q (type: %T)", result.Title, result.Title)
	t.Logf("Author: %q (type: %T)", result.Author, result.Author)
	t.Logf("Description: %q (type: %T)", result.Description, result.Description)
	t.Logf("Content: %q (type: %T)", result.Content, result.Content)
	t.Logf("WordCount: %d (type: %T)", result.WordCount, result.WordCount)

	// Test FormatMarkdown separately
	formatted := result.FormatMarkdown()
	if formatted == "" {
		t.Error("FormatMarkdown should not return empty string")
	}
	
	// FormatMarkdown should combine everything into one markdown document
	if !strings.Contains(formatted, "# "+result.Title) {
		t.Error("FormatMarkdown should contain title as H1")
	}
	
	t.Logf("FormatMarkdown output length: %d", len(formatted))
}