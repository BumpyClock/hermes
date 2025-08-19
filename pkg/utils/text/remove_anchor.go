// ABOUTME: Removes URL anchors/fragments for clean URL comparison
// ABOUTME: Faithfully ports the JavaScript removeAnchor function behavior
package text

import (
	"strings"
)

// RemoveAnchor removes the anchor/fragment portion from a URL and trailing slashes
// This function provides 100% compatibility with the JavaScript removeAnchor function:
// - Splits URL on '#' and takes the first part (removes fragment)
// - Removes trailing slashes from the result
func RemoveAnchor(url string) string {
	// Split on '#' and take the first part (removes fragment/anchor)
	parts := strings.Split(url, "#")
	urlWithoutAnchor := parts[0]
	
	// Remove trailing slash if present (matches JavaScript regex /\/$/)
	urlWithoutAnchor = strings.TrimSuffix(urlWithoutAnchor, "/")
	
	return urlWithoutAnchor
}