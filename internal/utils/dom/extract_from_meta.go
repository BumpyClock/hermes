// ABOUTME: ExtractFromMeta extracts content from HTML meta tags by matching names against cached selectors
// ABOUTME: This is a faithful port of the JavaScript extract-from-meta.js utility function

package dom

import (
	"fmt"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
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
	return stripTags(text, false)
}

// StripTagsWithBlockBoundaries preserves visible separation between article blocks.
func StripTagsWithBlockBoundaries(text string) string {
	return stripTags(text, true)
}

func stripTags(text string, blockBoundaries bool) string {
	if text == "" {
		return text
	}

	// Fast-path: if no HTML tags present, return as-is. Article HTML can still
	// hold entities without tags, for example after sanitizing removed them.
	if strings.IndexByte(text, '<') == -1 && (!blockBoundaries || strings.IndexByte(text, '&') == -1) {
		return text
	}

	// Parse the HTML content directly to extract text
	// Previously, content was wrapped in a <span> tag to prevent parsing errors,
	// but this is unnecessary as goquery handles text fragments correctly.
	// If parsing fails, we return the original text as a fallback.
	root, err := html.Parse(strings.NewReader(text))
	if err != nil {
		// If parsing fails, return original text
		return text
	}
	if blockBoundaries {
		var output strings.Builder
		appendArticleText(&output, root)
		return output.String()
	}
	doc := goquery.NewDocumentFromNode(root)

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

func appendArticleText(output *strings.Builder, node *html.Node) {
	if node.Type == html.CommentNode {
		return
	}
	block := false
	if node.Type == html.ElementNode {
		switch node.Data {
		case "script", "style", "noscript", "head", "meta", "link", "template":
			return
		case "address", "article", "aside", "blockquote", "br", "caption", "dd", "div",
			"dl", "dt", "fieldset", "figcaption", "figure", "footer", "form",
			"h1", "h2", "h3", "h4", "h5", "h6", "header", "hr", "li",
			"main", "nav", "ol", "p", "pre", "section", "table", "tbody", "td",
			"tfoot", "th", "thead", "tr", "ul":
			block = true
		}
	}
	if block {
		output.WriteByte(' ')
	}
	if node.Type == html.TextNode {
		output.WriteString(node.Data)
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		appendArticleText(output, child)
	}
	if block {
		output.WriteByte(' ')
	}
}

// MetaNames is an ordered list of meta tag names, each paired with its
// precompiled meta[name="..."] selector.
type MetaNames struct {
	names    []string
	matchers []goquery.Matcher
}

// MustCompileMetaNames compiles the meta[name="..."] selector for each name in order.
func MustCompileMetaNames(names ...string) MetaNames {
	metaNames := MetaNames{names: names, matchers: make([]goquery.Matcher, len(names))}
	for i, name := range names {
		// JavaScript hardcodes type="name"; ExtractFromMeta checks "value" and "content".
		metaNames.matchers[i] = cascadia.MustCompile(fmt.Sprintf("meta[name=\"%s\"]", name))
	}
	return metaNames
}

// ExtractFromMeta extracts content from HTML meta tags
// Given a list of meta tag names to search for, find a meta tag associated.
// This function provides 100% JavaScript compatibility.
func ExtractFromMeta(doc *goquery.Document, metaNames MetaNames, cachedNames []string, cleanTags bool) *string {
	// Process only names that exist in cachedNames
	// JavaScript uses: metaNames.filter(name => cachedNames.indexOf(name) !== -1)
	// This maintains the order of metaNames, not cachedNames
	for i, name := range metaNames.names {
		if !slices.Contains(cachedNames, name) {
			continue
		}

		// Find meta tags with the specified name
		nodes := doc.FindMatcher(metaNames.matchers[i])

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
