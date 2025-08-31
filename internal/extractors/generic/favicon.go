package generic

import (
	"net/url"
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

	// Default favicon.ico - resolve it properly against the base URL
	return extractor.normalizeURL("/favicon.ico", pageURL)
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
	
	// Parse the base URL
	baseURL, err := url.Parse(pageURL)
	if err != nil {
		// If we can't parse the page URL, return the href as-is
		return href
	}
	
	// Parse the relative URL
	relativeURL, err := url.Parse(href)
	if err != nil {
		// If we can't parse the href, return it as-is
		return href
	}
	
	// Resolve the relative URL against the base URL
	resolved := baseURL.ResolveReference(relativeURL)
	return resolved.String()
}