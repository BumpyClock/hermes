// ABOUTME: Debug test for meta tag extraction to verify ExtractFromMeta functionality
// ABOUTME: Tests the DOM utils meta extraction directly to isolate the issue

package parser

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/BumpyClock/parser-go/pkg/utils/dom"
)

func TestMetaExtraction_Debug(t *testing.T) {
	html := `
<!DOCTYPE html>
<html>
<head>
	<meta property="article:published_time" content="2023-01-15T10:30:00Z">
	<meta name="pubdate" content="2023-01-15">
	<meta name="date" content="January 15, 2023">
</head>
<body>
	<p>Test content</p>
</body>
</html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	// Test ExtractFromMeta directly
	dateMetaTags := []string{
		"article:published_time",
		"displaydate",
		"dc.date",
		"dc.date.issued",
		"rbpubdate",
		"publish_date",
		"pub_date",
		"pagedate",
		"pubdate",
		"revision_date",
		"doc_date",
		"date_created",
		"content_create_date",
		"lastmodified",
		"created",
		"date",
	}

	// Build proper meta cache
	metaCache := buildMetaCache(doc)
	t.Logf("Meta cache: %v", metaCache)
	
	meta := dom.ExtractFromMeta(doc, dateMetaTags, metaCache, false)
	t.Logf("ExtractFromMeta returned: %v", meta)
	if meta != nil {
		t.Logf("Meta value: %s", *meta)
	} else {
		t.Log("ExtractFromMeta returned nil")
	}

	// Test each meta tag individually
	for _, tagName := range dateMetaTags {
		singleTag := []string{tagName}
		result := dom.ExtractFromMeta(doc, singleTag, metaCache, false)
		if result != nil {
			t.Logf("Tag '%s' found: %s", tagName, *result)
		}
	}
}