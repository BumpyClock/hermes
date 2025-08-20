// ABOUTME: Debug test specifically for date extraction to identify why date parsing fails
// ABOUTME: Tests date extractor in isolation to verify meta tag detection and date parsing

package parser

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/BumpyClock/parser-go/pkg/extractors/generic"
)

func TestDateExtraction_Debug(t *testing.T) {
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

	// Test the generic date extractor directly
	metaCache := []string{}
	
	// Check what doc.Selection actually is
	t.Logf("doc.Selection element: %s", goquery.NodeName(doc.Selection))
	t.Logf("doc.Selection.Is('html'): %v", doc.Selection.Is("html"))
	
	// Try extracting with different selections
	dateStr := generic.GenericDateExtractor.Extract(doc.Selection, "https://example.com/test", metaCache)
	t.Logf("Date with doc.Selection: %v", dateStr)
	
	// Try with HTML element specifically
	htmlSelection := doc.Find("html")
	if htmlSelection.Length() > 0 {
		dateStr2 := generic.GenericDateExtractor.Extract(htmlSelection, "https://example.com/test", metaCache)
		t.Logf("Date with html selection: %v", dateStr2)
	}
	
	t.Logf("Date extractor returned: %v", dateStr)
	if dateStr != nil {
		t.Logf("Date string: %s", *dateStr)
		
		// Test our parseDate function
		if date, err := parseDate(*dateStr); err == nil {
			t.Logf("Parsed date: %v", date)
		} else {
			t.Logf("Failed to parse date: %v", err)
		}
	} else {
		t.Log("Date extractor returned nil")
	}

	// Also test meta tag extraction manually
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		if prop, exists := s.Attr("property"); exists {
			if content, exists := s.Attr("content"); exists {
				t.Logf("Found meta property='%s' content='%s'", prop, content)
			}
		}
		if name, exists := s.Attr("name"); exists {
			if content, exists := s.Attr("content"); exists {
				t.Logf("Found meta name='%s' content='%s'", name, content)
			}
		}
	})
}