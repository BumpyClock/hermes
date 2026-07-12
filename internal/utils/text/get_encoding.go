// ABOUTME: Character encoding detection from HTML content and headers
// ABOUTME: Faithful port of JavaScript getEncoding function with shared charset validation

package text

import "github.com/BumpyClock/hermes/internal/resource"

// GetEncoding extracts and validates character encoding from a string.
// It checks a string for encoding using ENCODING_RE pattern and validates
// the charset exists before returning it, otherwise returns DEFAULT_ENCODING.
func GetEncoding(str string) string {
	encoding := DEFAULT_ENCODING

	// Use ENCODING_RE to extract charset from string.
	matches := ENCODING_RE.FindStringSubmatch(str)
	if len(matches) > 1 {
		str = matches[1]
	}

	if resource.GetEncodingByCharset(str) != nil {
		encoding = str
	}

	return encoding
}
