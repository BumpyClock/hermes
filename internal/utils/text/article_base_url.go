// ABOUTME: Implements article base URL extraction by removing pagination parameters
// ABOUTME: Faithful port of JavaScript article-base-url.js with identical logic and behavior
package text

import (
	"net/url"
	"strings"
)

// IsGoodSegment determines if a URL segment should be kept or removed.
// This is a faithful port of the JavaScript isGoodSegment function.
//
// JavaScript logic:
// - If segment is purely a number and it's in first/second position with < 3 chars: keep it
// - If segment is "index" in first position: remove it
// - If segment is < 3 chars in first/second position and first segment has no letters: remove it
// - Otherwise: keep it
func IsGoodSegment(segment string, index int, firstSegmentHasLetters bool) bool {
	goodSegment := true

	// If this is purely a number, and it's the first or second
	// url_segment, it's probably a page number. Remove it.
	if index < 2 && IS_DIGIT_RE.MatchString(segment) && len(segment) < 3 {
		goodSegment = true
	}

	// If this is the first url_segment and it's just "index", remove it
	if index == 0 && strings.ToLower(segment) == "index" {
		goodSegment = false
	}

	// If our first or second url_segment is smaller than 3 characters,
	// and the first url_segment had no alphas, remove it.
	if index < 2 && len(segment) < 3 && !firstSegmentHasLetters {
		goodSegment = false
	}

	return goodSegment
}

// ArticleBaseURL takes a URL and returns the article base of said URL.
// That is, no pagination data exists in it. Useful for comparing to other links
// that might have pagination data within them.
//
// This is a faithful port of the JavaScript articleBaseUrl function.
//
// Parameters:
//   - urlStr: The URL string to process
//   - parsedURL: Optional pre-parsed URL (can be nil)
//
// Returns:
//   - string: The base URL with pagination data removed
//
// JavaScript equivalent:
//   export default function articleBaseUrl(url, parsed) {
//     const parsedUrl = parsed || URL.parse(url);
//     const { protocol, host, path } = parsedUrl;
//     ...
//   }
func ArticleBaseURL(urlStr string, parsedURL *url.URL) string {
	var parsedUrl *url.URL
	var err error

	if parsedURL != nil {
		parsedUrl = parsedURL
	} else {
		parsedUrl, err = url.Parse(urlStr)
		if err != nil {
			// If URL parsing fails, return the original URL
			return urlStr
		}
	}

	protocol := parsedUrl.Scheme
	host := parsedUrl.Host
	path := parsedUrl.Path

	// Handle empty path
	if path == "" || path == "/" {
		return protocol + "://" + host
	}

	var firstSegmentHasLetters bool
	// Split path by '/' exactly like JavaScript (keeping empty segments)
	segments := strings.Split(path, "/")
	
	// Reverse the segments to process from last to first (JavaScript does this)
	var reversedSegments []string
	for i := len(segments) - 1; i >= 0; i-- {
		reversedSegments = append(reversedSegments, segments[i])
	}

	var processedSegments []string
	for index, rawSegment := range reversedSegments {
		segment := rawSegment

		// Split off and save anything that looks like a file type.
		if strings.Contains(segment, ".") {
			parts := strings.Split(segment, ".")
			if len(parts) == 2 && IS_ALPHA_RE.MatchString(parts[1]) {
				segment = parts[0]
			}
		}

		// If our first or second segment has anything looking like a page
		// number, remove it.
		if PAGE_IN_HREF_RE.MatchString(segment) && index < 2 {
			segment = PAGE_IN_HREF_RE.ReplaceAllString(segment, "")
		}

		// If we're on the first segment, check to see if we have any
		// characters in it. The first segment is actually the last bit of
		// the URL, and this will be helpful to determine if we're on a URL
		// segment that looks like "/2/" for example.
		if index == 0 {
			firstSegmentHasLetters = HAS_ALPHA_RE.MatchString(segment)
		}

		// If it's not marked for deletion, push it to processed_segments.
		if IsGoodSegment(segment, index, firstSegmentHasLetters) {
			processedSegments = append(processedSegments, segment)
		}
	}

	// Reverse back to original order
	var finalSegments []string
	for i := len(processedSegments) - 1; i >= 0; i-- {
		finalSegments = append(finalSegments, processedSegments[i])
	}

	// Build the final URL - JavaScript joins all segments including empty ones
	// But we need to be careful about leading slashes
	finalPath := strings.Join(finalSegments, "/")
	
	// Remove double slashes that might be created by empty segments
	finalPath = strings.ReplaceAll(finalPath, "//", "/")
	
	// Ensure we start with a single slash if we have any path
	if finalPath != "" && !strings.HasPrefix(finalPath, "/") {
		finalPath = "/" + finalPath
	}
	
	return protocol + "://" + host + finalPath
}