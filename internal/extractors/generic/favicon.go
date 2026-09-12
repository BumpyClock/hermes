package generic

import "github.com/PuerkitoBio/goquery"

// GenericFaviconExtractor extracts the favicon URL.
type GenericFaviconExtractor struct{}

// Extract extracts the favicon URL from the page.
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
		href := selection.Find("link[rel=\""+rel+"\"]").AttrOr("href", "")
		if href != "" {
			return normalizeResourceURL(href, pageURL)
		}
	}

	// Default favicon.ico - resolve it properly against the base URL
	return normalizeResourceURL("/favicon.ico", pageURL)
}
