package dom

import (
	"math"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// LinkDensity calculates the density of links in an element
// Returns the ratio of link text length to total text length
func LinkDensity(element *goquery.Selection) float64 {
	totalText := strings.TrimSpace(element.Text())
	if len(totalText) == 0 {
		return 0
	}

	linkText := ""
	element.Find("a").Each(func(index int, link *goquery.Selection) {
		linkText += link.Text()
	})

	linkLength := len(strings.TrimSpace(linkText))
	totalLength := len(totalText)

	if totalLength == 0 {
		return 0
	}

	return float64(linkLength) / float64(totalLength)
}

// NodeIsSufficient determines if a node has enough content to be considered sufficient
// This exactly matches the JavaScript nodeIsSufficient implementation
// JavaScript: export default function nodeIsSufficient($node) { return $node.text().trim().length >= 100; }
func NodeIsSufficient(element *goquery.Selection) bool {
	// JavaScript: return $node.text().trim().length >= 100;
	return len(strings.TrimSpace(element.Text())) >= 100
}

// WithinComment checks if an element is within a comment section
func WithinComment(element *goquery.Selection) bool {
	// Check if element or any parent has comment-related classes/IDs
	current := element
	for current.Length() > 0 {
		classAndId := ""
		if class, exists := current.Attr("class"); exists {
			classAndId += class + " "
		}
		if id, exists := current.Attr("id"); exists {
			classAndId += id
		}

		// Check for comment indicators
		if strings.Contains(strings.ToLower(classAndId), "comment") ||
			strings.Contains(strings.ToLower(classAndId), "disqus") ||
			strings.Contains(strings.ToLower(classAndId), "respond") {
			return true
		}

		current = current.Parent()
	}

	return false
}

// IsWordpress detects if a page is likely WordPress-based
func IsWordpress(doc *goquery.Document) bool {
	// Check for WordPress generator meta tag
	generator := doc.Find(`meta[name="generator"]`)
	if generator.Length() > 0 {
		content, exists := generator.Attr("content")
		if exists && strings.Contains(strings.ToLower(content), "wordpress") {
			return true
		}
	}

	// Check for common WordPress classes/IDs
	wpIndicators := []string{
		".wp-content",
		"#wp-content",
		".wordpress",
		".wp-",
		"[class*='wp-']",
	}

	for _, indicator := range wpIndicators {
		if doc.Find(indicator).Length() > 0 {
			return true
		}
	}

	// Check for WordPress-specific script sources
	scripts := doc.Find("script[src]")
	scripts.Each(func(index int, script *goquery.Selection) {
		src, exists := script.Attr("src")
		if exists && (strings.Contains(src, "wp-content") || strings.Contains(src, "wp-includes")) {
			return // Found WordPress indicator
		}
	})

	return false
}

// HasSentenceEnd checks if text ends with proper sentence punctuation
func HasSentenceEnd(text string) bool {
	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return false
	}

	lastChar := text[len(text)-1:]
	return lastChar == "." || lastChar == "!" || lastChar == "?" || lastChar == ":" || lastChar == ";"
}

// DetectTextDirection attempts to detect the text direction (LTR/RTL) of content
func DetectTextDirection(text string) string {
	if len(text) == 0 {
		return "ltr"
	}

	rtlChars := 0
	totalChars := 0

	// Count RTL characters (Arabic, Hebrew, etc.)
	for _, r := range text {
		totalChars++
		// Arabic: U+0600-U+06FF
		// Hebrew: U+0590-U+05FF
		if (r >= 0x0600 && r <= 0x06FF) || (r >= 0x0590 && r <= 0x05FF) {
			rtlChars++
		}
	}

	if totalChars == 0 {
		return "ltr"
	}

	rtlRatio := float64(rtlChars) / float64(totalChars)
	if rtlRatio > 0.3 {
		return "rtl"
	}

	return "ltr"
}

// GetContentScore calculates a basic content score for an element
func GetContentScore(element *goquery.Selection) float64 {
	text := strings.TrimSpace(element.Text())
	textLength := len(text)

	if textLength == 0 {
		return 0
	}

	score := float64(0)

	// Base score on text length (with diminishing returns)
	score += math.Log(float64(textLength)) * 2

	// Bonus for paragraph tags
	paragraphs := element.Find("p")
	score += float64(paragraphs.Length()) * 3

	// Penalty for high link density
	linkDensity := LinkDensity(element)
	score -= linkDensity * 10

	// Check class and ID for positive/negative indicators
	classAndId := ""
	if class, exists := element.Attr("class"); exists {
		classAndId += class + " "
	}
	if id, exists := element.Attr("id"); exists {
		classAndId += id
	}

	if POSITIVE_SCORE_RE.MatchString(classAndId) {
		score += 25
	}

	if NEGATIVE_SCORE_RE.MatchString(classAndId) {
		score -= 25
	}

	return score
}

// CountWords counts the number of words in text
func CountWords(text string) int {
	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return 0
	}

	words := strings.Fields(text)
	return len(words)
}

// CountSentences estimates the number of sentences in text
func CountSentences(text string) int {
	if len(text) == 0 {
		return 0
	}

	// Simple sentence counting based on punctuation
	sentences := 0
	for _, char := range text {
		if char == '.' || char == '!' || char == '?' {
			sentences++
		}
	}

	// Ensure at least 1 sentence if there's text
	if sentences == 0 && len(strings.TrimSpace(text)) > 0 {
		sentences = 1
	}

	return sentences
}

// IsLikelyArticleElement checks if an element is likely to contain article content
func IsLikelyArticleElement(element *goquery.Selection) bool {
	tagName := goquery.NodeName(element)
	
	// Check tag name
	if tagName == "article" || tagName == "main" {
		return true
	}

	// Check class and ID
	classAndId := ""
	if class, exists := element.Attr("class"); exists {
		classAndId += class + " "
	}
	if id, exists := element.Attr("id"); exists {
		classAndId += id
	}

	// Look for article-related keywords
	if POSITIVE_SCORE_RE.MatchString(classAndId) {
		return true
	}

	// Check if it has good content characteristics
	textLength := len(strings.TrimSpace(element.Text()))
	paragraphs := element.Find("p").Length()
	
	return textLength > 500 && paragraphs >= 2 && LinkDensity(element) < 0.3
}