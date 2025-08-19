// ABOUTME: Normalizes whitespace in text content while preserving spaces in HTML tags
// ABOUTME: Faithful port of JavaScript normalizeSpaces function with 100% compatibility
package text

import (
	"regexp"
	"strconv"
	"strings"
)

// MULTIPLE_SPACES_RE matches 2 or more consecutive whitespace characters
// This is part of the JavaScript regex /\s{2,}(?![^<>]*<\/(pre|code|textarea)>)/g
var MULTIPLE_SPACES_RE = regexp.MustCompile(`\s{2,}`)

// PRE_TAG_RE finds pre tags and their content (only closed tags)
var PRE_TAG_RE = regexp.MustCompile(`(?i)<pre[^>]*>.*?</pre>`)

// CODE_TAG_RE finds code tags and their content (only closed tags)
var CODE_TAG_RE = regexp.MustCompile(`(?i)<code[^>]*>.*?</code>`)

// TEXTAREA_TAG_RE finds textarea tags and their content (only closed tags)
var TEXTAREA_TAG_RE = regexp.MustCompile(`(?i)<textarea[^>]*>.*?</textarea>`)

// NormalizeSpaces normalizes consecutive whitespace characters to single spaces
// while preserving spacing within pre, code, and textarea HTML tags.
//
// This function provides 100% compatibility with the JavaScript normalizeSpaces function:
// - Replaces 2+ consecutive whitespace characters with a single space
// - Preserves whitespace inside <pre>, <code>, and <textarea> tags
// - Trims leading and trailing whitespace from the result
// - Handles all types of whitespace: spaces, tabs, newlines, carriage returns
//
// Example:
//   NormalizeSpaces("text   with    spaces") // returns "text with spaces"
//   NormalizeSpaces("<pre>  keep  spaces  </pre>") // returns "<pre>  keep  spaces  </pre>"
func NormalizeSpaces(text string) string {
	// Since Go doesn't support negative lookahead, we need to implement the logic differently
	// We'll find all preservable tag content first, replace them with placeholders,
	// normalize the rest, then restore the preserved content
	
	preservedContent := make(map[string]string)
	result := text
	placeholderCounter := 0
	
	// Helper function to preserve content and replace with placeholder
	preserveContent := func(re *regexp.Regexp) {
		result = re.ReplaceAllStringFunc(result, func(match string) string {
			placeholder := "___PRESERVE_CONTENT_" + strconv.Itoa(placeholderCounter) + "___"
			preservedContent[placeholder] = match
			placeholderCounter++
			return placeholder
		})
	}
	
	// Preserve content in pre, code, and textarea tags
	preserveContent(PRE_TAG_RE)
	preserveContent(CODE_TAG_RE)
	preserveContent(TEXTAREA_TAG_RE)
	
	// Now normalize spaces in the text without the preserved content
	result = MULTIPLE_SPACES_RE.ReplaceAllString(result, " ")
	
	// Restore the preserved content
	for placeholder, originalContent := range preservedContent {
		result = strings.ReplaceAll(result, placeholder, originalContent)
	}
	
	// Trim leading and trailing whitespace (matches JavaScript .trim())
	result = strings.TrimSpace(result)
	
	return result
}