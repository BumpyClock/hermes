package dom

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// LinkDensity calculates the density of links in an element
// Returns the ratio of link text length to total text length.
func LinkDensity(element *goquery.Selection) float64 {
	totalText := strings.TrimSpace(element.Text())
	if totalText == "" {
		return 0
	}

	linkText := strings.TrimSpace(element.Find("a").Text())
	return float64(len(linkText)) / float64(len(totalText))
}

// NodeIsSufficient determines if a node has enough content to be considered sufficient
// This exactly matches the JavaScript nodeIsSufficient implementation
// JavaScript: export default function nodeIsSufficient($node) { return $node.text().trim().length >= 100; }.
func NodeIsSufficient(element *goquery.Selection) bool {
	// JavaScript: return $node.text().trim().length >= 100;
	return len(strings.TrimSpace(element.Text())) >= 100
}

// WithinComment checks if an element is within a comment section.
func WithinComment(element *goquery.Selection) bool {
	// Check if element or any parent has comment-related classes/IDs
	current := element
	for current.Length() > 0 {
		var classAndIdBuilder strings.Builder
		if class, exists := current.Attr("class"); exists {
			classAndIdBuilder.WriteString(class)
			classAndIdBuilder.WriteString(" ")
		}
		if id, exists := current.Attr("id"); exists {
			classAndIdBuilder.WriteString(id)
		}
		classAndId := strings.ToLower(classAndIdBuilder.String())

		// Check for comment indicators
		if strings.Contains(classAndId, "comment") ||
			strings.Contains(classAndId, "disqus") ||
			strings.Contains(classAndId, "respond") {
			return true
		}

		current = current.Parent()
	}

	return false
}
