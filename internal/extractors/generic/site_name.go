package generic

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// GenericSiteNameExtractor extracts the site name from meta tags
type GenericSiteNameExtractor struct{}

// Extract extracts the site name from various meta tags
func (extractor *GenericSiteNameExtractor) Extract(selection *goquery.Selection, pageURL string, metaCache []string) string {
	// Priority order for site name extraction
	metaTags := []string{
		"og:site_name",
		"twitter:site",
		"application-name",
		"al:ios:app_name",
		"al:android:app_name",
	}

	// Check each meta tag in priority order
	for _, tagName := range metaTags {
		// Try meta[property="..."]
		content := selection.Find("meta[property=\"" + tagName + "\"]").AttrOr("content", "")
		if content != "" {
			return strings.TrimSpace(content)
		}

		// Try meta[name="..."]
		content = selection.Find("meta[name=\"" + tagName + "\"]").AttrOr("content", "")
		if content != "" {
			return strings.TrimSpace(content)
		}
	}

	// Fallback to domain name from URL if available
	return ""
}