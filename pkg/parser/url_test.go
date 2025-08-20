// ABOUTME: Test URL validation behavior for debugging error handling test failure
// ABOUTME: Verifies URL parsing and validation logic in parser integration

package parser

import (
	"net/url"
	"testing"
)

func TestURLParsing(t *testing.T) {
	// Test what happens with "not-a-url"
	parsed, err := url.Parse("not-a-url")
	t.Logf("url.Parse('not-a-url') = %v, err = %v", parsed, err)
	
	if err == nil {
		t.Logf("Scheme: '%s', Host: '%s'", parsed.Scheme, parsed.Host)
		t.Logf("validateURL result: %v", validateURL(parsed))
	}
	
	// Test with actual invalid formats
	invalidURLs := []string{
		"not-a-url",
		"://invalid",
		"http://",
		"",
	}
	
	for _, testURL := range invalidURLs {
		parsed, err := url.Parse(testURL)
		if err != nil {
			t.Logf("URL '%s' failed parsing: %v", testURL, err)
		} else {
			t.Logf("URL '%s' parsed as: scheme='%s', host='%s', valid=%v", 
				testURL, parsed.Scheme, parsed.Host, validateURL(parsed))
		}
	}
}

func TestParseHTMLWithInvalidURL(t *testing.T) {
	parser := New()
	opts := ParserOptions{
		Fallback:    true,
		ContentType: "html",
	}

	html := `<html><body><p>Test</p></body></html>`
	
	_, err := parser.ParseHTML(html, "not-a-url", opts)
	t.Logf("ParseHTML with 'not-a-url': err = %v", err)
	
	_, err = parser.ParseHTML(html, "", opts)
	t.Logf("ParseHTML with empty URL: err = %v", err)
}