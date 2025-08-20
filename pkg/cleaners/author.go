// ABOUTME: Author name cleaning and normalization functionality
// ABOUTME: Faithful port of JavaScript cleaners/author.js with 100% compatibility

package cleaners

import (
	"strings"
	
	"github.com/postlight/parser-go/pkg/utils/text"
)

// CleanAuthor takes an author string (like 'By David Smith ') and cleans it to
// just the name(s): 'David Smith'.
// 
// This is a faithful 1:1 port of the JavaScript cleanAuthor function:
// - Removes "By", "Posted by", "Written by" prefixes (case insensitive)
// - Handles optional colons after prefixes  
// - Normalizes all whitespace to single spaces
// - Trims leading and trailing whitespace
//
// JavaScript equivalent:
// export default function cleanAuthor(author) {
//   return normalizeSpaces(author.replace(CLEAN_AUTHOR_RE, '$2').trim());
// }
func CleanAuthor(author string) string {
	// Use the regex to match and capture the author part (group $2)
	// JavaScript: author.replace(CLEAN_AUTHOR_RE, '$2')
	matches := CLEAN_AUTHOR_RE.FindStringSubmatch(author)
	
	var authorPart string
	if len(matches) >= 3 {
		// Group $2 is at index 2 (group $1 is prefix, group $2 is author name)
		authorPart = matches[2]
	} else {
		// No match found, use original string
		authorPart = author
	}
	
	// Trim whitespace first, then normalize spaces
	// JavaScript: normalizeSpaces(result.trim())
	authorPart = strings.TrimSpace(authorPart)
	
	// Apply normalizeSpaces to handle multiple consecutive whitespace
	return text.NormalizeSpaces(authorPart)
}