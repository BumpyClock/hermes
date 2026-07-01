// ABOUTME: Word count extractor with JavaScript-compatible behavior and fallback methods
// ABOUTME: Provides accurate word counting from HTML content using primary and alternative algorithms

package generic

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/BumpyClock/hermes/internal/utils/text"
)

// HTML tag removal regex for alternative word counting method.
var htmlTagRE = regexp.MustCompile(`<[^>]*>`)

// Multiple whitespace regex for alternative method normalization.
var multipleSpacesRE = regexp.MustCompile(`\s+`)

// getWordCount implements the primary word counting method using goquery
// This matches the JavaScript cheerio.load($).first().text() approach.
func getWordCount(content string) int {
	// Load content with goquery (matches cheerio.load behavior)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		// Fallback to alternative method on parsing error
		return getWordCountAlt(content)
	}

	// Get the first div element (matches $('div').first() in JavaScript)
	contentDiv := doc.Find("div").First()

	// Extract text content
	contentText := contentDiv.Text()

	// Normalize spaces using our text utility (matches JavaScript normalizeSpaces)
	normalizedText := text.NormalizeSpaces(contentText)

	// Split on whitespace and count words (matches JavaScript .split(/\s+/).length)
	if normalizedText == "" {
		// Empty string should return 0 words in JavaScript
		return 0
	}

	// Count non-empty fields without allocating a filtered slice.
	words := multipleSpacesRE.Split(strings.TrimSpace(normalizedText), -1)
	count := 0
	for _, word := range words {
		if word != "" {
			count++
		}
	}

	return count
}

// getWordCountAlt implements the alternative/fallback word counting method
// This matches the JavaScript regex-based HTML stripping approach.
func getWordCountAlt(content string) int {
	// Remove HTML tags using regex (matches content.replace(/<[^>]*>/g, ' '))
	cleanContent := htmlTagRE.ReplaceAllString(content, " ")

	// Replace multiple whitespace with single space (matches content.replace(/\s+/g, ' '))
	cleanContent = multipleSpacesRE.ReplaceAllString(cleanContent, " ")

	// Trim leading and trailing whitespace (matches content.trim())
	cleanContent = strings.TrimSpace(cleanContent)

	// Split on single space and count words (matches content.split(' ').length)
	if cleanContent == "" {
		// Empty string should return 0 words
		return 0
	}

	words := strings.Split(cleanContent, " ")

	// Count non-empty fields without allocating a filtered slice.
	count := 0
	for _, word := range words {
		if word != "" {
			count++
		}
	}

	return count
}

// GenericWordCountExtractor extracts word count from content using JavaScript-compatible logic.
var GenericWordCountExtractor = struct {
	Extract func(options map[string]interface{}) int
}{
	Extract: func(options map[string]interface{}) int {
		// Handle nil or missing options gracefully
		if options == nil {
			return 0
		}

		// Extract content from options
		contentInterface, exists := options["content"]
		if !exists {
			return 0
		}

		// Ensure content is a string
		content, ok := contentInterface.(string)
		if !ok {
			return 0
		}

		// Use primary word counting method
		count := getWordCount(content)

		// Fall back when DOM extraction has no useful container text.
		if count <= 1 {
			count = getWordCountAlt(content)
		}

		return count
	},
}
