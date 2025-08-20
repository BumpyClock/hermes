// ABOUTME: Generic title extractor with JavaScript-compatible fallback logic and cleaning
// ABOUTME: Extracts article titles using meta tags and CSS selectors with domain name removal

package generic

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/postlight/parser-go/pkg/utils/dom"
	"github.com/postlight/parser-go/pkg/utils/text"
)

// Title extraction constants matching JavaScript behavior exactly
var (
	// An ordered list of meta tag names that denote likely article titles.
	// All attributes should be lowercase for faster case-insensitive matching.
	// From most distinct to least distinct.
	STRONG_TITLE_META_TAGS = []string{
		"tweetmeme-title",
		"dc.title",
		"rbtitle",
		"headline",
		"title",
	}

	// og:title is weak because it typically contains context that we don't like,
	// for example the source site's name. Gotta get that brand into facebook!
	WEAK_TITLE_META_TAGS = []string{
		"og:title",
	}

	// An ordered list of CSS Selectors to find likely article titles.
	// From most explicit to least explicit.
	//
	// Note - this does not use classes like CSS. This checks to see if the string
	// exists in the className, which is not as accurate as .className (which
	// splits on spaces/endlines), but for our purposes it's close enough.
	STRONG_TITLE_SELECTORS = []string{
		".hentry .entry-title",
		"h1#articleHeader",
		"h1.articleHeader",
		"h1.article",
		".instapaper_title",
		"#meebo-title",
	}

	WEAK_TITLE_SELECTORS = []string{
		"article h1",
		"#entry-title",
		".entry-title",
		"#entryTitle",
		"#entrytitle",
		".entryTitle",
		".entrytitle",
		"#articleTitle",
		".articleTitle",
		"post post-title",
		"h1.title",
		"h2.article",
		"h1",
		"html head title",
		"title",
	}

	// Regular expression for title separators
	TITLE_SPLITTERS_RE = regexp.MustCompile(`(: | - | \| )`)

	// Domain endings regex for cleaning
	DOMAIN_ENDINGS_RE = regexp.MustCompile(`\.com$|\.net$|\.org$|\.co\.uk$`)
)

// GenericTitleExtractor extracts article titles using multiple fallback strategies
var GenericTitleExtractor = struct {
	Extract func(doc *goquery.Selection, url string, metaCache []string) string
}{
	Extract: func(doc *goquery.Selection, url string, metaCache []string) string {
		// Convert selection to document for meta tag extraction
		// Get the full HTML from the selection to create a proper document
		html := "<html></html>" // Default fallback
		if doc.Length() > 0 {
			if fullHtml, err := doc.Html(); err == nil && fullHtml != "" {
				html = "<html>" + fullHtml + "</html>"
			} else {
				// Try to get the parent document HTML
				if doc.Parent().Length() > 0 {
					if parentHtml, err := doc.Parent().Html(); err == nil {
						html = "<html>" + parentHtml + "</html>"
					}
				}
			}
		} else {
			return ""
		}

		document, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			return ""
		}

		// First, check to see if we have a matching meta tag that we can make
		// use of that is strongly associated with the headline.
		title := dom.ExtractFromMeta(document, STRONG_TITLE_META_TAGS, metaCache, true)
		if title != nil && *title != "" {
			return cleanTitle(*title, url, doc)
		}

		// Second, look through our content selectors for the most likely
		// article title that is strongly associated with the headline.
		title = dom.ExtractFromSelectors(doc, STRONG_TITLE_SELECTORS, 1, true)
		if title != nil && *title != "" {
			return cleanTitle(*title, url, doc)
		}

		// Third, check for weaker meta tags that may match.
		title = dom.ExtractFromMeta(document, WEAK_TITLE_META_TAGS, metaCache, true)
		if title != nil && *title != "" {
			return cleanTitle(*title, url, doc)
		}

		// Last, look for weaker selector tags that may match.
		title = dom.ExtractFromSelectors(doc, WEAK_TITLE_SELECTORS, 1, true)
		if title != nil && *title != "" {
			return cleanTitle(*title, url, doc)
		}

		// If no matches, return an empty string
		return ""
	},
}

// cleanTitle cleans and normalizes the title text
func cleanTitle(title string, url string, doc *goquery.Selection) string {
	// If title has |, :, or - in it, see if we can clean it up.
	if TITLE_SPLITTERS_RE.MatchString(title) {
		title = resolveSplitTitle(title, url)
	}

	// Final sanity check that we didn't get a crazy title.
	if len(title) > 150 {
		// If we did, return h1 from the document if it exists
		h1s := doc.Find("h1")
		if h1s.Length() == 1 {
			title = h1s.Text()
		}
	}

	// strip any html tags in the title text and normalize spaces
	// Create a document for StripTags function
	tempDoc, _ := goquery.NewDocumentFromReader(strings.NewReader("<html></html>"))
	cleaned := dom.StripTags(title, tempDoc)
	return text.NormalizeSpaces(strings.TrimSpace(cleaned))
}

// splitTitleWithSeparators splits title while preserving separators
// This mimics JavaScript's split behavior with capturing groups
func splitTitleWithSeparators(title string) []string {
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

// resolveSplitTitle resolves whether any of the segments should be removed
func resolveSplitTitle(title, url string) string {
	// Splits while preserving splitters - use FindAllString to get separators too
	// This mimics JavaScript's behavior with capturing groups in regex
	splitTitle := splitTitleWithSeparators(title)
	if len(splitTitle) <= 1 {
		return title
	}

	// Try extracting breadcrumb title
	if newTitle := extractBreadcrumbTitle(splitTitle, title); newTitle != "" {
		return newTitle
	}

	// Try cleaning domain from title
	if newTitle := cleanDomainFromTitle(splitTitle, url); newTitle != "" {
		return newTitle
	}

	// Fuzzy ratio didn't find anything, so this title is probably legit.
	// Just return it all.
	return title
}

// extractBreadcrumbTitle extracts the most relevant title from breadcrumb-style titles
func extractBreadcrumbTitle(splitTitle []string, text string) string {
	// This must be a very breadcrumbed title, like:
	// The Best Gadgets on Earth : Bits : Blogs : NYTimes.com
	// NYTimes - Blogs - Bits - The Best Gadgets on Earth
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
			return longestEnd
		}

		return text
	}

	return ""
}

// cleanDomainFromTitle removes domain name matches from title segments
func cleanDomainFromTitle(splitTitle []string, urlStr string) string {
	if urlStr == "" || len(splitTitle) < 2 {
		return ""
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	// Strip out the big TLDs - it just makes the matching a bit more accurate.
	nakedDomain := DOMAIN_ENDINGS_RE.ReplaceAllString(parsedURL.Host, "")

	// Check start of title
	if len(splitTitle) >= 2 {
		startSlug := strings.ToLower(strings.ReplaceAll(splitTitle[0], " ", ""))
		if levenshteinRatio(startSlug, nakedDomain) > 0.4 && len(startSlug) > 5 {
			// Join remaining segments (skip separator at index 1)
			if len(splitTitle) >= 3 {
				return strings.Join(splitTitle[2:], "")
			}
		}
	}

	// Check end of title
	if len(splitTitle) >= 2 {
		endSlug := strings.ToLower(strings.ReplaceAll(splitTitle[len(splitTitle)-1], " ", ""))
		if levenshteinRatio(endSlug, nakedDomain) > 0.4 && len(endSlug) >= 5 {
			// Join all segments except last two (content and separator)
			if len(splitTitle) >= 3 {
				return strings.Join(splitTitle[:len(splitTitle)-2], "")
			}
		}
	}

	return ""
}

// levenshteinRatio calculates the Levenshtein similarity ratio between two strings
// This is compatible with the JavaScript wuzzy.levenshtein function
func levenshteinRatio(s1, s2 string) float64 {
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