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
		// Meta tags are normalized to name/value by NormalizeMetaTags; prefer value then content.
		if content := selection.Find("meta[name=\""+tagName+"\"]").AttrOr("value", ""); content != "" {
			return strings.TrimSpace(content)
		}
		if content := selection.Find("meta[name=\""+tagName+"\"]").AttrOr("content", ""); content != "" {
			return strings.TrimSpace(content)
		}
		// Backward compatibility in case normalization changes
		if content := selection.Find("meta[property=\""+tagName+"\"]").AttrOr("content", ""); content != "" {
			return strings.TrimSpace(content)
		}
		if content := selection.Find("meta[property=\""+tagName+"\"]").AttrOr("value", ""); content != "" {
			return strings.TrimSpace(content)
		}
	}

	// Fallback to domain name from URL if available
	return ""
}
