// ABOUTME: GenericDekExtractor extracts article subtitles/descriptions from meta tags and selectors
// ABOUTME: Implements dek validation, cleaning, and fallback extraction strategy with JavaScript compatibility

package generic

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/postlight/parser-go/pkg/utils/text"
)

var (
	// TEXT_LINK_RE matches plain text URLs that should disqualify a dek
	textLinkRE = regexp.MustCompile(`https?://`)
	
	// Meta tag names for dek extraction, ordered by priority
	dekMetaTags = []string{
		"description",
		"og:description", 
		"twitter:description",
		"dc.description",
	}
	
	// CSS selectors for dek extraction, ordered by priority
	dekSelectors = []string{
		".entry-summary",
		"h2[itemprop=\"description\"]",
		".subtitle",
		".sub-title", 
		".deck",
		".dek",
		".standfirst",
		".summary",
		".description",
	}
)

// GenericDekExtractor extracts article subtitles/descriptions (deks)
type GenericDekExtractor struct{}

// Extract extracts dek from meta tags and selectors with validation and cleaning
func (e *GenericDekExtractor) Extract(doc *goquery.Document, opts map[string]interface{}) string {
	selection := doc.Selection
	if s, ok := opts["$"].(*goquery.Selection); ok {
		selection = s
	}
	
	var excerpt string
	if ex, ok := opts["excerpt"].(string); ok {
		excerpt = ex
	}
	
	// Try meta tags first (higher priority)
	if dek := e.extractFromMeta(selection); dek != "" {
		if cleaned := e.cleanDek(dek, excerpt); cleaned != "" {
			return cleaned
		}
	}
	
	// Fall back to CSS selectors
	if dek := e.extractFromSelectors(selection); dek != "" {
		if cleaned := e.cleanDek(dek, excerpt); cleaned != "" {
			return cleaned
		}
	}
	
	return ""
}

// extractFromMeta extracts dek from meta tags
func (e *GenericDekExtractor) extractFromMeta(selection *goquery.Selection) string {
	for _, metaName := range dekMetaTags {
		var content string
		var found bool
		
		// Try name attribute first
		selection.Find("meta[name=\"" + metaName + "\"]").Each(func(_ int, s *goquery.Selection) {
			if !found {
				if attr, exists := s.Attr("content"); exists {
					found = true
					if strings.TrimSpace(attr) != "" {
						content = attr
					}
				}
			}
		})
		
		// Try property attribute for OpenGraph if not found with name
		if !found {
			selection.Find("meta[property=\"" + metaName + "\"]").Each(func(_ int, s *goquery.Selection) {
				if !found {
					if attr, exists := s.Attr("content"); exists {
						found = true
						if strings.TrimSpace(attr) != "" {
							content = attr
						}
					}
				}
			})
		}
		
		// If we found the meta tag but it's empty, reject entirely (don't fall back)
		if found {
			return content
		}
	}
	
	return ""
}

// extractFromSelectors extracts dek from CSS selectors
func (e *GenericDekExtractor) extractFromSelectors(selection *goquery.Selection) string {
	for _, selector := range dekSelectors {
		if element := selection.Find(selector).First(); element.Length() > 0 {
			return element.Text()
		}
	}
	
	return ""
}

// cleanDek validates and cleans extracted dek text
func (e *GenericDekExtractor) cleanDek(dek, excerpt string) string {
	if dek == "" {
		return ""
	}
	
	// Strip HTML tags if present
	dekText := e.stripTags(dek)
	
	// Sanity check length (5-1000 characters)
	if len(dekText) > 1000 || len(dekText) < 5 {
		return ""
	}
	
	// Check that dek isn't the same as excerpt (first 10 words)
	if excerpt != "" {
		dekExcerpt := text.ExcerptContent(dekText, 10)
		excerptContent := text.ExcerptContent(excerpt, 10)
		// Debug: fmt.Printf("DekExcerpt: %q, ExcerptContent: %q\n", dekExcerpt, excerptContent)
		if dekExcerpt == excerptContent {
			return ""
		}
	}
	
	// Plain text links shouldn't exist in the dek
	if textLinkRE.MatchString(dekText) {
		return ""
	}
	
	// Normalize whitespace and trim
	return text.NormalizeSpaces(strings.TrimSpace(dekText))
}

// stripTags removes HTML tags from text while preserving content
func (e *GenericDekExtractor) stripTags(html string) string {
	if html == "" {
		return ""
	}
	
	// Wrap in span to ensure valid HTML parsing (avoid nesting issues)
	wrapped := "<span>" + html + "</span>"
	
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(wrapped))
	if err != nil {
		// If parsing fails, return original text
		return html
	}
	
	text := doc.Find("span").First().Text()
	if text == "" {
		// If extraction results in empty string, return original
		return html
	}
	
	return text
}