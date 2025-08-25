package generic

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// GenericSiteImageExtractor extracts the main site image
type GenericSiteImageExtractor struct{}

// Extract extracts the site's main image from meta tags
func (extractor *GenericSiteImageExtractor) Extract(selection *goquery.Selection, pageURL string, metaCache []string) string {
	// Priority order for image extraction
	metaTags := []string{
		"og:image",
		"twitter:image",
		"twitter:image:src",
		"thumbnail",
		"image",
	}

	// Check each meta tag in priority order
	for _, tagName := range metaTags {
		// Try meta[property="..."]
		content := selection.Find("meta[property=\"" + tagName + "\"]").AttrOr("content", "")
		if content != "" && extractor.isValidImageURL(content) {
			return strings.TrimSpace(content)
		}

		// Try meta[name="..."]
		content = selection.Find("meta[name=\"" + tagName + "\"]").AttrOr("content", "")
		if content != "" && extractor.isValidImageURL(content) {
			return strings.TrimSpace(content)
		}
	}

	// Try link[rel="image_src"]
	imageSrc := selection.Find("link[rel=\"image_src\"]").AttrOr("href", "")
	if imageSrc != "" && extractor.isValidImageURL(imageSrc) {
		return strings.TrimSpace(imageSrc)
	}

	return ""
}

// isValidImageURL checks if the URL looks like a valid image URL
func (extractor *GenericSiteImageExtractor) isValidImageURL(url string) bool {
	if url == "" {
		return false
	}
	
	// Basic validation - has protocol or starts with /
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "//") || strings.HasPrefix(url, "/") {
		return true
	}
	
	return false
}