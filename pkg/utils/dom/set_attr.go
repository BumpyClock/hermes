// ABOUTME: Single attribute setter utility that mirrors JavaScript setAttr behavior
// ABOUTME: Handles goquery Selection objects for DOM manipulation
package dom

import (
	"github.com/PuerkitoBio/goquery"
)

// SetAttr sets a single attribute on a DOM node
// This function mirrors the JavaScript setAttr behavior, handling goquery selections
// which are equivalent to cheerio nodes in the original JavaScript implementation.
//
// In JavaScript, this function handled two cases:
// 1. Cheerio nodes (with attribs property) - equivalent to our goquery selections
// 2. Browser DOM nodes (with setAttribute method) - not applicable in Go/server environment
//
// Parameters:
//   - selection: The goquery selection to modify
//   - attr: The attribute name to set
//   - val: The attribute value to set
//
// Returns:
//   - The modified goquery selection for method chaining
func SetAttr(selection *goquery.Selection, attr, val string) *goquery.Selection {
	// In Go with goquery, we only need to handle the cheerio-style case
	// since we're always working in a server-side environment
	return selection.SetAttr(attr, val)
}