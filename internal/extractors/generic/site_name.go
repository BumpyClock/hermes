package generic

import "github.com/PuerkitoBio/goquery"

// GenericSiteNameExtractor extracts the site name from meta tags.
type GenericSiteNameExtractor struct{}

// Extract extracts the site name from various meta tags.
func (extractor *GenericSiteNameExtractor) Extract(selection *goquery.Selection, pageURL string, metaCache []string) string {
	// Priority order for site name extraction
	metaTags := []string{
		"og:site_name",
		"twitter:site",
		"application-name",
		"al:ios:app_name",
		"al:android:app_name",
	}

	if value := firstMetaValue(selection, metaTags, nil); value != "" {
		return value
	}

	// Fallback to domain name from URL if available
	return ""
}
