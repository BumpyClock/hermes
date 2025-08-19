package dom

import (
	"github.com/PuerkitoBio/goquery"
)

// GetAttrs returns all attributes of a goquery node as a map
// This mimics the JavaScript getAttrs function that works with both
// cheerio's attribs and browser DOM attributes
func GetAttrs(selection *goquery.Selection) map[string]string {
	attrs := make(map[string]string)
	
	// Check if there are any nodes in the selection
	if len(selection.Nodes) == 0 {
		return attrs
	}
	
	// Get the first node (consistent with JS behavior)
	node := selection.Nodes[0]
	
	// Iterate through all attributes
	for _, attr := range node.Attr {
		attrs[attr.Key] = attr.Val
	}
	
	return attrs
}

// GetAttr is a convenience function to get a single attribute
// It's equivalent to selection.Attr() but provides consistent behavior
func GetAttr(selection *goquery.Selection, attrName string) (string, bool) {
	return selection.Attr(attrName)
}


// RemoveAttr is a convenience function to remove an attribute
func RemoveAttr(selection *goquery.Selection, attrName string) *goquery.Selection {
	return selection.RemoveAttr(attrName)
}

// HasAttr checks if an element has a specific attribute
func HasAttr(selection *goquery.Selection, attrName string) bool {
	_, exists := selection.Attr(attrName)
	return exists
}