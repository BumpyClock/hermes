// ABOUTME: ExtractFromMeta extracts content from HTML meta tags by matching names against cached selectors
// ABOUTME: This is a faithful port of the JavaScript extract-from-meta.js utility function

package dom

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

// removeComments recursively removes all HTML comment nodes from the tree
// Traverses the entire node tree starting from the given node and removes comment nodes at all levels.
func removeComments(node *html.Node) {
	// Traverse children and collect comments to remove
	// We can't remove during iteration as it modifies the linked list
	var toRemove []*html.Node
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.CommentNode {
			toRemove = append(toRemove, child)
		} else {
			// Recursively process non-comment nodes
			removeComments(child)
		}
	}

	// Remove all comment nodes found
	for _, comment := range toRemove {
		node.RemoveChild(comment)
	}
}

// StripTags removes all HTML tags from a string of text
// Returns plain text content with all HTML tags removed
// Removes non-content elements (script, style, noscript, head, meta, link) and HTML comments
// If the result is empty, returns the original text (JavaScript behavior).
func StripTags(text string) string {
	if text == "" {
		return text
	}

	// Fast-path: if no HTML tags present, return as-is
	if strings.IndexByte(text, '<') == -1 {
		return text
	}

	// Parse the HTML content directly to extract text
	// Previously, content was wrapped in a <span> tag to prevent parsing errors,
	// but this is unnecessary as goquery handles text fragments correctly.
	// If parsing fails, we return the original text as a fallback.
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(text))
	if err != nil {
		// If parsing fails, return original text
		return text
	}

	// Remove non-content elements before extracting text
	// These elements don't contribute to visible content
	doc.Find("script, style, noscript, head, meta, link").Remove()

	// Remove HTML comments at all levels using recursive traversal
	// Start from the document root to catch top-level comments
	if len(doc.Nodes) > 0 {
		removeComments(doc.Nodes[0])
	}

	cleanText := doc.Text()
	if cleanText == "" {
		// If extraction results in empty string, return original (JavaScript behavior)
		return text
	}

	return cleanText
}

// ExtractFromMeta extracts content from HTML meta tags
// Given a list of meta tag names to search for, find a meta tag associated.
// This function provides 100% JavaScript compatibility.
func ExtractFromMeta(doc *goquery.Document, metaNames []string, cachedNames []string, cleanTags bool) *string {
	// Filter metaNames to only include names that exist in cachedNames
	// JavaScript uses: metaNames.filter(name => cachedNames.indexOf(name) !== -1)
	// This maintains the order of metaNames, not cachedNames
	var foundNames []string
	for _, name := range metaNames {
		for _, cached := range cachedNames {
			if name == cached {
				foundNames = append(foundNames, name)
				break
			}
		}
	}

	// Process each found name in order
	for _, name := range foundNames {
		// JavaScript hardcodes type="name" and checks "value" attribute
		// However, standard HTML meta tags use "content", so we check both
		metaType := "name"

		// Find meta tags with the specified name
		selector := fmt.Sprintf("meta[%s=\"%s\"]", metaType, name)
		nodes := doc.Find(selector)

		// Get all non-empty values from both 'value' and 'content' attributes
		var values []string
		nodes.Each(func(index int, node *goquery.Selection) {
			// Check 'value' attribute first (matches JavaScript behavior)
			if val, exists := node.Attr("value"); exists && val != "" {
				values = append(values, val)
			} else if content, exists := node.Attr("content"); exists && content != "" {
				// Fallback to standard 'content' attribute
				values = append(values, content)
			}
		})

		// If we have exactly one value, return it
		// If we have more than one value, we have a conflict and can't trust any
		// If we have zero values, the meta tags had no values
		if len(values) == 1 {
			metaValue := values[0]

			// Meta values that contain HTML should be stripped, as they
			// weren't subject to cleaning previously
			if cleanTags {
				metaValue = StripTags(metaValue)
			}

			return &metaValue
		}
	}

	// If nothing is found, return nil
	return nil
}
