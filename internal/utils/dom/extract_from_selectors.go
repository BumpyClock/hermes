// ABOUTME: CSS selector-based content extraction utility function
// ABOUTME: Extracts content from DOM elements using CSS selectors with JavaScript compatibility

package dom

import "github.com/PuerkitoBio/goquery"

// isGoodNode checks if a node is suitable for content extraction.
func isGoodNode(node *goquery.Selection, maxChildren int) bool {
	// If it has a number of children, it's more likely a container
	// element. Skip it.
	if node.Children().Length() > maxChildren {
		return false
	}

	// If it looks to be within a comment, skip it.
	if WithinComment(node) {
		return false
	}

	return true
}

// ExtractFromSelectors finds content that may be extractable from the document
// using CSS selectors. This is for flat meta-information, like author, title,
// date published, etc.
//
// Parameters:
// - doc: The goquery document/selection to search within
// - selectors: List of CSS selectors to try in order
// - maxChildren: Maximum number of child elements allowed (default 1)
// - textOnly: If true, extract text content; if false, extract HTML (default true)
//
// Returns:
// - *string: The extracted content, or nil if nothing suitable found.
func ExtractFromSelectors(doc *goquery.Selection, selectors []string, maxChildren int, textOnly bool) *string {
	return ExtractFromSelectorsWithFilter(doc, selectors, maxChildren, textOnly, nil)
}

// ExtractFromSelectorsWithFilter additionally rejects candidates outside the caller's metadata scope.
func ExtractFromSelectorsWithFilter(doc *goquery.Selection, selectors []string, maxChildren int, textOnly bool, accept func(*goquery.Selection) bool) *string {
	// eslint-disable-next-line no-restricted-syntax
	for _, selector := range selectors {
		nodes := doc.Find(selector)

		// If we didn't get exactly one of this selector, this may be
		// a list of articles or comments. Skip it.
		if nodes.Length() != 1 {
			continue
		}
		node := nodes.First()
		if !isGoodNode(node, maxChildren) {
			continue
		}
		if accept != nil && !accept(node) {
			continue
		}

		var content string
		if textOnly {
			content = node.Text()
		} else {
			content, _ = node.Html()
		}

		// Normalize whitespace to match JavaScript's text normalization
		// Replace all whitespace sequences with single spaces
		content = normalizeSpaces(content)
		if content != "" {
			return &content
		}
	}

	return nil
}
