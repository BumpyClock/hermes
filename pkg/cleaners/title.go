// ABOUTME: Title cleaner with 100% JavaScript compatibility for site name removal and normalization
// ABOUTME: Cleans extracted titles by removing site names, normalizing separators, and handling split titles

package cleaners

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/postlight/parser-go/pkg/utils/dom"
	"github.com/postlight/parser-go/pkg/utils/text"
)

// Title cleaning constants are now provided in constants.go

// CleanTitle cleans and normalizes title text by removing site names, HTML tags, and extra whitespace
// This is a faithful port of the JavaScript cleanTitle function
func CleanTitle(title string, url string, doc *goquery.Document) string {
	// First strip HTML tags to clean the title for processing
	cleaned := dom.StripTags(title, doc)
	cleaned = strings.TrimSpace(cleaned)

	// If title has |, :, or - in it, see if we can clean it up.
	if TITLE_SPLITTERS_RE.MatchString(cleaned) {
		cleaned = ResolveSplitTitle(cleaned, url)
	}

	// Final sanity check that we didn't get a crazy title.
	// if (title.length > 150 || title.length < 15) {
	if len(cleaned) > 150 {
		// If we did, return h1 from the document if it exists
		h1s := doc.Find("h1")
		if h1s.Length() == 1 {
			cleaned = h1s.Text()
		}
	}

	// Normalize spaces and return
	return text.NormalizeSpaces(cleaned)
}

// ResolveSplitTitle resolves whether any of the segments should be removed from a title with separators
// Given a title with separators in it (colons, dashes, etc), resolve whether any of the segments should be removed.
func ResolveSplitTitle(title, url string) string {
	// Splits while preserving splitters, like:
	// ['The New New York', ' - ', 'The Washington Post']
	splitTitle := SplitTitleWithSeparators(title)
	if len(splitTitle) <= 1 {
		return title
	}

	// Try extracting breadcrumb title
	if newTitle := ExtractBreadcrumbTitle(splitTitle, title); newTitle != "" {
		return newTitle
	}

	// Try cleaning domain from title
	if newTitle := CleanDomainFromTitle(splitTitle, url); newTitle != "" {
		return newTitle
	}

	// Fuzzy ratio didn't find anything, so this title is probably legit.
	// Just return it all.
	return title
}

// SplitTitleWithSeparators splits title while preserving separators
// This mimics JavaScript's split behavior with capturing groups
func SplitTitleWithSeparators(title string) []string {
	var result []string
	lastIndex := 0
	
	// Find all separator matches
	matches := TITLE_SPLITTERS_RE.FindAllStringIndex(title, -1)
	
	for _, match := range matches {
		start, end := match[0], match[1]
		
		// Add the text before the separator
		if start > lastIndex {
			result = append(result, title[lastIndex:start])
		}
		
		// Add the separator itself
		result = append(result, title[start:end])
		lastIndex = end
	}
	
	// Add any remaining text after the last separator
	if lastIndex < len(title) {
		result = append(result, title[lastIndex:])
	}
	
	return result
}

// ExtractBreadcrumbTitle extracts the most relevant title from breadcrumb-style titles
// This must be a very breadcrumbed title, like:
// The Best Gadgets on Earth : Bits : Blogs : NYTimes.com
// NYTimes - Blogs - Bits - The Best Gadgets on Earth
func ExtractBreadcrumbTitle(splitTitle []string, text string) string {
	if len(splitTitle) >= 6 {
		// Look to see if we can find a breadcrumb splitter that happens
		// more than once. If we can, we'll be able to better pull out
		// the title.
		termCounts := make(map[string]int)
		for _, titleText := range splitTitle {
			termCounts[titleText]++
		}

		maxTerm := ""
		termCount := 0
		for term, count := range termCounts {
			if count > termCount {
				maxTerm = term
				termCount = count
			}
		}

		// We found a splitter that was used more than once, so it
		// is probably the breadcrumber. Split our title on that instead.
		// Note: max_term should be <= 4 characters, so that " >> "
		// will match, but nothing longer than that.
		if termCount >= 2 && len(maxTerm) <= 4 {
			splitTitle = strings.Split(text, maxTerm)
		}

		// JavaScript: const splitEnds = [splitTitle[0], splitTitle.slice(-1)];
		// splitTitle.slice(-1) returns the last element as an array, then we get [0]
		splitEnds := []string{}
		if len(splitTitle) > 0 {
			splitEnds = append(splitEnds, splitTitle[0])
			splitEnds = append(splitEnds, splitTitle[len(splitTitle)-1])
		}

		// Get the longest end segment
		longestEnd := ""
		for _, end := range splitEnds {
			if len(end) > len(longestEnd) {
				longestEnd = end
			}
		}

		if len(longestEnd) > 10 {
			return strings.TrimSpace(longestEnd)
		}

		return text
	}

	return ""
}

// CleanDomainFromTitle removes domain name matches from title segments
// Search the ends of the title, looking for bits that fuzzy match
// the URL too closely. If one is found, discard it and return the rest.
func CleanDomainFromTitle(splitTitle []string, urlStr string) string {
	if urlStr == "" || len(splitTitle) < 2 {
		return ""
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	// Strip out the big TLDs - it just makes the matching a bit more accurate.
	// Not the end of the world if it doesn't strip right.
	nakedDomain := DOMAIN_ENDINGS_RE.ReplaceAllString(parsedURL.Host, "")

	// Check start of title
	if len(splitTitle) >= 2 {
		startSlug := strings.ToLower(strings.Replace(splitTitle[0], " ", "", 1))
		startSlugRatio := LevenshteinRatio(startSlug, nakedDomain)

		if startSlugRatio > 0.4 && len(startSlug) > 5 {
			// Join remaining segments (skip separator at index 1)
			if len(splitTitle) >= 3 {
				return strings.Join(splitTitle[2:], "")
			}
		}
	}

	// Check end of title
	if len(splitTitle) >= 2 {
		endSlug := strings.ToLower(strings.Replace(splitTitle[len(splitTitle)-1], " ", "", 1))
		endSlugRatio := LevenshteinRatio(endSlug, nakedDomain)

		if endSlugRatio > 0.4 && len(endSlug) >= 5 {
			// Join all segments except last two (content and separator)
			if len(splitTitle) >= 3 {
				return strings.Join(splitTitle[:len(splitTitle)-2], "")
			}
		}
	}

	return ""
}

// LevenshteinRatio calculates the Levenshtein similarity ratio between two strings
// This is compatible with the JavaScript wuzzy.levenshtein function
func LevenshteinRatio(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}
	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	// Compute actual Levenshtein distance
	distance := levenshteinDistance(s1, s2)
	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}

	// Return similarity ratio (1.0 - distance/maxLen)
	ratio := 1.0 - (float64(distance) / float64(maxLen))
	if ratio < 0 {
		ratio = 0
	}
	return ratio
}

// levenshteinDistance computes the Levenshtein distance between two strings
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	// Initialize first row and column
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = minInt(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

// minInt returns the minimum of three integers
func minInt(a, b, c int) int {
	if a < b && a < c {
		return a
	} else if b < c {
		return b
	}
	return c
}