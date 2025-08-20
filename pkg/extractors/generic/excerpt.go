// ABOUTME: Excerpt extractor that generates article summaries from meta tags or content with JavaScript compatibility
// ABOUTME: Faithful 1:1 port of JavaScript GenericExcerptExtractor with ellipsize functionality and meta tag priority

package generic

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/postlight/parser-go/pkg/utils/dom"
)

// EXCERPT_META_SELECTORS defines the meta tag names to search for excerpt content
// This matches the JavaScript constants exactly: ['og:description', 'twitter:description']
var EXCERPT_META_SELECTORS = []string{"og:description", "twitter:description"}

// GenericExcerptExtractor implements excerpt extraction logic
type GenericExcerptExtractor struct{}

// NewGenericExcerptExtractor creates a new excerpt extractor
func NewGenericExcerptExtractor() *GenericExcerptExtractor {
	return &GenericExcerptExtractor{}
}

// Extract extracts excerpt from meta tags or falls back to content
// This is a faithful port of the JavaScript GenericExcerptExtractor.extract method
func (e *GenericExcerptExtractor) Extract(doc *goquery.Document, content string, metaCache []string) string {
	// Try to extract from meta tags first (matches JavaScript behavior)
	excerpt := dom.ExtractFromMeta(doc, EXCERPT_META_SELECTORS, metaCache, true)
	if excerpt != nil && *excerpt != "" {
		return clean(*excerpt, doc, 200)
	}

	// Fall back to excerpting from the extracted content (JavaScript behavior)
	maxLength := 200
	shortContent := content
	if len(content) > maxLength*5 {
		// JavaScript: content.slice(0, maxLength * 5)
		shortContent = content[:maxLength*5]
	}

	// JavaScript: clean($(shortContent).text(), $, maxLength)
	// We need to parse the content as HTML and extract text
	if shortContent != "" {
		// Create a temporary document from the content to extract text like JavaScript $(content).text()
		contentDoc, err := goquery.NewDocumentFromReader(strings.NewReader("<div>" + shortContent + "</div>"))
		if err != nil {
			// If HTML parsing fails, use the content directly
			return clean(shortContent, doc, maxLength)
		}
		textContent := contentDoc.Find("div").Text()
		return clean(textContent, doc, maxLength)
	}

	return ""
}

// clean normalizes whitespace and applies ellipsize with JavaScript compatibility
// This is a faithful port of the JavaScript clean function
func clean(content string, doc *goquery.Document, maxLength int) string {
	if content == "" {
		return ""
	}

	// JavaScript: content.replace(/[\s\n]+/g, ' ').trim()
	// Normalize all whitespace sequences to single spaces and trim
	whitespaceRegex := regexp.MustCompile(`[\s\n]+`)
	normalized := strings.TrimSpace(whitespaceRegex.ReplaceAllString(content, " "))

	if normalized == "" {
		return ""
	}

	// JavaScript: ellipsize(content, maxLength, { ellipse: '&hellip;' })
	return ellipsize(normalized, maxLength)
}

// ellipsize truncates content to maxLength and adds ellipsis if needed
// This matches the JavaScript ellipsize library behavior with { ellipse: '&hellip;' }
// The JavaScript library truncates and trims trailing spaces before adding ellipsis
func ellipsize(content string, maxLength int) string {
	if content == "" {
		return ""
	}

	if maxLength <= 0 {
		return "&hellip;"
	}

	// Convert to runes to handle UTF-8 properly
	runes := []rune(content)
	
	if len(runes) <= maxLength {
		return content
	}

	// JavaScript ellipsize library truncates at maxLength and trims trailing spaces
	truncated := strings.TrimRight(string(runes[:maxLength]), " ")
	return truncated + "&hellip;"
}