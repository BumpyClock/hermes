package generic

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// GenericFaviconExtractor extracts the favicon URL
type GenericFaviconExtractor struct{}

// Extract extracts the favicon URL from the page
func (extractor *GenericFaviconExtractor) Extract(selection *goquery.Selection, pageURL string, metaCache []string) string {
	// Priority order for favicon extraction
	linkRels := []string{
		"apple-touch-icon",
		"apple-touch-icon-precomposed",
		"icon",
		"shortcut icon",
	}

	// Check each link rel in priority order
	for _, rel := range linkRels {
		href := selection.Find("link[rel=\"" + rel + "\"]").AttrOr("href", "")
		if href != "" {
			return extractor.normalizeURL(href, pageURL)
		}
	}

	// Default favicon.ico
	return "/favicon.ico"
}

// normalizeURL ensures the favicon URL is absolute
func (extractor *GenericFaviconExtractor) normalizeURL(href, pageURL string) string {
	href = strings.TrimSpace(href)
	
	// Already absolute
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	
	// Protocol-relative
	if strings.HasPrefix(href, "//") {
		return "https:" + href
	}
	
	// Relative URL - for now just return as-is
	// TODO: Properly resolve relative URLs against the page URL
	return href
}