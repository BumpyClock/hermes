// ABOUTME: URL text extraction utilities for pattern matching
// ABOUTME: Provides extractFromURL function to extract text from URLs using regex patterns

package text

import (
	"regexp"
)

// ExtractFromURL searches for patterns in a URL and returns the first capture group from the first matching regex.
// Given a URL and a list of regular expressions, this function tests each regex against the URL
// and returns the first capture group (group 1) from the first matching pattern.
// This is primarily used for extracting date information from URLs in date published extraction.
//
// Parameters:
//   - url: The URL string to search within
//   - regexList: A slice of compiled regular expressions to test against the URL
//
// Returns:
//   - string: The first capture group from the first matching regex, or empty string if no match
//   - bool: true if a match was found, false otherwise
//
// The function expects each regex to have at least one capture group, and will return
// the content of the first capture group from the first regex that matches the URL.
func ExtractFromURL(url string, regexList []*regexp.Regexp) (string, bool) {
	for _, re := range regexList {
		if matches := re.FindStringSubmatch(url); matches != nil && len(matches) > 1 {
			return matches[1], true
		}
	}
	return "", false
}