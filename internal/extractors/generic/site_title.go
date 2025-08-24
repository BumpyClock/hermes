package generic

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// GenericSiteTitleExtractor extracts the site title
type GenericSiteTitleExtractor struct{}

// Extract extracts the site title from the page
func (extractor *GenericSiteTitleExtractor) Extract(selection *goquery.Selection, pageURL string, metaCache []string) string {
	// First try Open Graph title
	ogTitle := selection.Find("meta[property=\"og:title\"]").AttrOr("content", "")
	if ogTitle != "" {
		return strings.TrimSpace(ogTitle)
	}

	// Try Twitter title
	twitterTitle := selection.Find("meta[name=\"twitter:title\"]").AttrOr("content", "")
	if twitterTitle != "" {
		return strings.TrimSpace(twitterTitle)
	}

	// Fallback to page title
	pageTitle := selection.Find("title").Text()
	if pageTitle != "" {
		return strings.TrimSpace(pageTitle)
	}

	return ""
}