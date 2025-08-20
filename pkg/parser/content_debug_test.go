// ABOUTME: Debug test for content extraction to diagnose why simple HTML returns empty content
// ABOUTME: Tests generic content extractor directly to understand extraction behavior

package parser

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/BumpyClock/parser-go/pkg/extractors/generic"
)

func TestContentExtraction_Debug(t *testing.T) {
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Test Article</title>
	</head>
	<body>
		<article>
			<h1>Test Article</h1>
			<p>This is test content.</p>
		</article>
	</body>
	</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	// Test content extractor directly
	contentExtractor := generic.NewGenericContentExtractor()
	contentParams := generic.ExtractorParams{
		Doc:   doc,
		HTML:  html,
		Title: "Test Article",
		URL:   "https://example.com/test",
	}
	contentOpts := generic.ExtractorOptions{
		StripUnlikelyCandidates: true,
		WeightNodes:             true,
		CleanConditionally:      true,
	}

	content := contentExtractor.Extract(contentParams, contentOpts)
	t.Logf("Content extractor returned: '%s'", content)
	t.Logf("Content length: %d", len(content))

	// Test with more relaxed options
	relaxedOpts := generic.ExtractorOptions{
		StripUnlikelyCandidates: false,
		WeightNodes:             false,
		CleanConditionally:      false,
	}

	content2 := contentExtractor.Extract(contentParams, relaxedOpts)
	t.Logf("Content with relaxed options: '%s'", content2)
	t.Logf("Content2 length: %d", len(content2))

	// Test with just the fallback extraction
	fallbackContent := doc.Find("article, .article, #article, .content, #content, .entry-content").First().Text()
	t.Logf("Fallback content: '%s'", strings.TrimSpace(fallbackContent))
}