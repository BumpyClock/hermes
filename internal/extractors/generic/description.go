// ABOUTME: GenericDescriptionExtractor extracts site-level descriptions from meta tags and JSON-LD
// ABOUTME: Prioritizes general site descriptions over article-specific descriptions with validation

package generic

import (
	"encoding/json"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// GenericDescriptionExtractor extracts site descriptions
type GenericDescriptionExtractor struct{}

// Meta tags for description extraction, ordered by priority
var descriptionMetaTags = []string{
	"description",       // Standard meta description
	"og:description",    // Open Graph description
	"twitter:description", // Twitter card description
	"dc.description",    // Dublin Core description
}

// Extract extracts site description using priority-based strategies
func (extractor *GenericDescriptionExtractor) Extract(selection *goquery.Selection, pageURL string, metaCache []string) string {
	// Strategy 1: Try meta tags first
	if description := extractor.extractFromMetaTags(selection); description != "" {
		return extractor.cleanDescription(description)
	}

	// Strategy 2: Try JSON-LD structured data
	if description := extractor.extractFromJSONLD(selection); description != "" {
		return extractor.cleanDescription(description)
	}
	return ""
}

// extractFromMetaTags extracts description from meta tags using priority order
func (extractor *GenericDescriptionExtractor) extractFromMetaTags(selection *goquery.Selection) string {
	// Check each meta tag in priority order
	for _, tagName := range descriptionMetaTags {
		// First try meta[name="..."] - after normalization, content becomes value
		content := selection.Find("meta[name=\"" + tagName + "\"]").AttrOr("value", "")
		if content != "" {
			if extractor.isValidDescription(content) {
				return strings.TrimSpace(content)
			}
		}

		// For og/twitter tags, after normalization property becomes name
		// So we already checked them above, no need for separate property check
	}

	return ""
}

// extractFromJSONLD extracts description from JSON-LD structured data
func (extractor *GenericDescriptionExtractor) extractFromJSONLD(selection *goquery.Selection) string {
	var foundDescription string

	// Find all JSON-LD script tags
	selection.Find("script[type=\"application/ld+json\"]").Each(func(i int, s *goquery.Selection) {
		if foundDescription != "" {
			return // Already found a description
		}

		jsonText := strings.TrimSpace(s.Text())
		if jsonText == "" {
			return
		}

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(jsonText), &data); err != nil {
			return // Skip invalid JSON
		}

		// Try to extract description based on @type
		if typeVal, ok := data["@type"].(string); ok {
			var description string

			switch typeVal {
			case "WebSite", "Organization", "NewsMediaOrganization":
				if desc, ok := data["description"].(string); ok {
					description = desc
				}
			case "Article", "NewsArticle":
				// For articles, check if there's a publisher with description
				if publisher, ok := data["publisher"].(map[string]interface{}); ok {
					if desc, ok := publisher["description"].(string); ok {
						description = desc
					}
				}
			}

			if description != "" && extractor.isValidDescription(description) {
				foundDescription = description
				return
			}
		}
	})

	return foundDescription
}

// isValidDescription validates that the description is suitable as site metadata
func (extractor *GenericDescriptionExtractor) isValidDescription(description string) bool {
	description = strings.TrimSpace(description)
	
	// Must not be empty
	if description == "" {
		return false
	}

	// Should be reasonable length (not too short, not too long)
	if len(description) < 10 || len(description) > 500 {
		return false
	}

	// Should not contain URLs (likely spam or article-specific)
	if strings.Contains(description, "http://") || strings.Contains(description, "https://") {
		return false
	}

	// Should not start with common article-specific prefixes
	lowerDesc := strings.ToLower(description)
	articlePrefixes := []string{
		"in this article",
		"this article",
		"read more about",
		"continue reading",
		"full story:",
	}

	for _, prefix := range articlePrefixes {
		if strings.HasPrefix(lowerDesc, prefix) {
			return false
		}
	}

	return true
}

// cleanDescription cleans and normalizes the description
func (extractor *GenericDescriptionExtractor) cleanDescription(description string) string {
	description = strings.TrimSpace(description)
	
	// Remove extra whitespace
	description = strings.Join(strings.Fields(description), " ")
	
	// Remove common suffixes that are site-specific but not descriptive
	suffixes := []string{
		" - Read more",
		" | Read more",
		" - Continue reading",
		" | Continue reading",
	}
	
	for _, suffix := range suffixes {
		if strings.HasSuffix(description, suffix) {
			description = strings.TrimSuffix(description, suffix)
			break
		}
	}
	
	return strings.TrimSpace(description)
}