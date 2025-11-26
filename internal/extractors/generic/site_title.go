package generic

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// GenericSiteTitleExtractor extracts the site title
type GenericSiteTitleExtractor struct{}

// Extract extracts the site title from the page
func (extractor *GenericSiteTitleExtractor) Extract(selection *goquery.Selection, pageURL string, metaCache []string) string {
	// First try Open Graph title (meta tags normalized to name/value)
	if ogTitle := selection.Find("meta[name=\"og:title\"]").AttrOr("value", ""); ogTitle != "" {
		return strings.TrimSpace(ogTitle)
	}
	if ogTitle := selection.Find("meta[name=\"og:title\"]").AttrOr("content", ""); ogTitle != "" {
		return strings.TrimSpace(ogTitle)
	}
	// Backward compat for unnormalized property
	if ogTitle := selection.Find("meta[property=\"og:title\"]").AttrOr("content", ""); ogTitle != "" {
		return strings.TrimSpace(ogTitle)
	}
	if ogTitle := selection.Find("meta[property=\"og:title\"]").AttrOr("value", ""); ogTitle != "" {
		return strings.TrimSpace(ogTitle)
	}

	// Try Twitter title
	if twitterTitle := selection.Find("meta[name=\"twitter:title\"]").AttrOr("value", ""); twitterTitle != "" {
		return strings.TrimSpace(twitterTitle)
	}
	if twitterTitle := selection.Find("meta[name=\"twitter:title\"]").AttrOr("content", ""); twitterTitle != "" {
		return strings.TrimSpace(twitterTitle)
	}

	// Fallback to page title
	pageTitle := selection.Find("title").Text()
	if pageTitle != "" {
		return strings.TrimSpace(pageTitle)
	}

	return ""
}
