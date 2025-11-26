// ABOUTME: Integration tests for site-level metadata extraction across parser pipeline
// ABOUTME: Ensures normalized meta tags still populate site_name/title/image in results

package parser

import (
	"context"
	"testing"
)

func TestSiteMetadataFromNormalizedMetaTags(t *testing.T) {
	h := New()

	html := `
	<!doctype html>
	<html>
	<head>
		<meta name="og:site_name" value="Example Site">
		<meta name="og:title" value="Example Site Title">
		<meta name="og:image" value="https://cdn.example.com/social.png">
	</head>
	<body><main><h1>Ignored</h1></main></body>
	</html>`

	opts := DefaultParserOptions()
	opts.Fallback = true

	res, err := h.ParseHTMLWithContext(context.Background(), html, "https://example.com/test", opts)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if res.SiteName != "Example Site" {
		t.Fatalf("expected SiteName 'Example Site', got '%s'", res.SiteName)
	}
	if res.SiteTitle != "Example Site Title" {
		t.Fatalf("expected SiteTitle 'Example Site Title', got '%s'", res.SiteTitle)
	}
	if res.SiteImage != "https://cdn.example.com/social.png" {
		t.Fatalf("expected SiteImage to be set, got '%s'", res.SiteImage)
	}
}
