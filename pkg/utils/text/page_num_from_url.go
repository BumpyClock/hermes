// ABOUTME: PageNumFromURL extracts page numbers from URLs for pagination handling
// ABOUTME: Returns nil if no page number found or if page number >= 100

package text

import (
	"strconv"
)

// PageNumFromURL extracts a page number from a URL string.
// This is a faithful port of the JavaScript pageNumFromUrl function.
//
// The function looks for page number patterns in URLs like:
//   - page=1, pg=1, p=1
//   - paging=12, pag=7
//   - pagination/1, paging/88, pa/83, p/11
//
// Returns:
//   - *int: the page number if found and < 100
//   - nil: if no page number found or page number >= 100
//
// JavaScript equivalent:
//   export default function pageNumFromUrl(url) {
//     const matches = url.match(PAGE_IN_HREF_RE);
//     if (!matches) return null;
//     const pageNum = parseInt(matches[6], 10);
//     return pageNum < 100 ? pageNum : null;
//   }
func PageNumFromURL(url string) *int {
	// Find matches using the PAGE_IN_HREF_RE pattern
	matches := PAGE_IN_HREF_RE.FindStringSubmatch(url)
	if matches == nil {
		return nil
	}

	// The JavaScript regex captures 6 groups, with the page number in matches[6]
	// In Go, we need to check that we have enough submatch groups
	if len(matches) < 7 {
		return nil
	}

	// Extract the page number from the 7th capture group (index 6)
	pageNumStr := matches[6]
	if pageNumStr == "" {
		return nil
	}

	// Parse the page number as integer
	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil {
		return nil
	}

	// Return pageNum < 100, otherwise return nil
	// This matches the JavaScript logic exactly
	if pageNum < 100 {
		return &pageNum
	}
	return nil
}