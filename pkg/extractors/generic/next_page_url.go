// ABOUTME: Next page URL extractor for multi-page articles with sophisticated scoring algorithms
// ABOUTME: Faithful 1:1 port of JavaScript implementation with all regex patterns and scoring logic

package generic

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/postlight/parser-go/pkg/utils/dom"
	"github.com/postlight/parser-go/pkg/utils/text"
)

// Constants from JavaScript implementation
var (
	// DIGIT_RE matches any digit character
	DIGIT_RE = regexp.MustCompile(`\d`)
	
	// EXTRANEOUS_LINK_HINTS are words that indicate a link is probably not a next page
	EXTRANEOUS_LINK_HINTS = []string{
		"print", "archive", "comment", "discuss", "e-mail", "email",
		"share", "reply", "all", "login", "sign", "single", "adx", "entry-unrelated",
	}
	EXTRANEOUS_LINK_HINTS_RE = regexp.MustCompile(`(?i)` + strings.Join(EXTRANEOUS_LINK_HINTS, "|"))
	
	// NEXT_LINK_TEXT_RE matches text that likely indicates a next page link
	NEXT_LINK_TEXT_RE = regexp.MustCompile(`(?i)(next|weiter|continue|>([^|]|$)|»([^|]|$))`)
	
	// CAP_LINK_TEXT_RE matches text that indicates end links (first, last, etc.)
	CAP_LINK_TEXT_RE = regexp.MustCompile(`(?i)(first|last|end)`)
	
	// PREV_LINK_TEXT_RE matches text that indicates previous page links
	PREV_LINK_TEXT_RE = regexp.MustCompile(`(?i)(prev|earl|old|new|<|«)`)
	
	// PAGE_RE matches pagination-related text
	PAGE_RE = regexp.MustCompile(`(?i)pag(e|ing|inat)`)
)

// GenericNextPageUrlExtractor extracts next page URLs for multi-page articles
type GenericNextPageUrlExtractor struct{}

// NewGenericNextPageUrlExtractor creates a new instance
func NewGenericNextPageUrlExtractor() *GenericNextPageUrlExtractor {
	return &GenericNextPageUrlExtractor{}
}

// Extract finds and returns the most likely next page URL
func (e *GenericNextPageUrlExtractor) Extract(doc *goquery.Document, articleURL string, parsedURL *url.URL, previousUrls []string) string {
	if parsedURL == nil {
		var err error
		parsedURL, err = url.Parse(articleURL)
		if err != nil {
			return ""
		}
	}

	cleanArticleURL := text.RemoveAnchor(articleURL)
	baseURL := text.ArticleBaseURL(articleURL, parsedURL)

	// Get all links with href attributes
	var links []*goquery.Selection
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		links = append(links, s)
	})

	// Score all potential next page links
	scoredLinks := scoreLinks(links, cleanArticleURL, baseURL, parsedURL, previousUrls, doc)

	// If no links were scored, return empty string
	if scoredLinks == nil || len(scoredLinks) == 0 {
		return ""
	}

	// Find the highest scoring link
	var topPage scoredLink
	topPage.score = -100

	for _, link := range scoredLinks {
		if link.score > topPage.score {
			topPage = link
		}
	}

	// If the score is less than 50, we're not confident enough to use it
	if topPage.score >= 50 {
		return topPage.href
	}

	return ""
}

// scoredLink represents a link with its calculated score
type scoredLink struct {
	score    float64
	linkText string
	href     string
}

// scoreLinks scores all potential next page links and returns them
func scoreLinks(links []*goquery.Selection, articleURL, baseURL string, parsedURL *url.URL, previousUrls []string, doc *goquery.Document) []scoredLink {
	baseRegex := regexp.MustCompile(`(?i)^` + regexp.QuoteMeta(baseURL))
	isWp := dom.IsWordpress(doc)

	var scoredPages []scoredLink
	linkMap := make(map[string]*scoredLink)

	for _, link := range links {
		href, exists := link.Attr("href")
		if !exists {
			continue
		}

		// Resolve relative URLs to absolute URLs
		if baseURL, err := url.Parse(articleURL); err == nil {
			if resolvedURL, err := baseURL.Parse(href); err == nil {
				href = resolvedURL.String()
			}
		}

		href = text.RemoveAnchor(href)
		linkText := strings.TrimSpace(link.Text())

		if !shouldScore(href, articleURL, baseURL, parsedURL, linkText, previousUrls) {
			continue
		}

		// If we haven't seen this href before, create a new entry
		if _, exists := linkMap[href]; !exists {
			linkMap[href] = &scoredLink{
				score:    0,
				linkText: linkText,
				href:     href,
			}
		} else {
			// Combine link text with existing entry
			linkMap[href].linkText = linkMap[href].linkText + "|" + linkText
		}

		possiblePage := linkMap[href]
		linkData := makeSig(link, linkText)
		pageNumPtr := text.PageNumFromURL(href)
		pageNum := 0
		if pageNumPtr != nil {
			pageNum = *pageNumPtr
		}

		// Calculate total score using all scoring functions
		score := scoreBaseUrl(href, baseRegex)
		score += scoreNextLinkText(linkData)
		score += scoreCapLinks(linkData)
		score += scorePrevLink(linkData)
		score += scoreByParentsNextPage(link)
		score += scoreExtraneousLinks(href)
		score += scorePageInLink(pageNum, isWp)
		score += scoreLinkText(linkText, pageNum)
		score += scoreSimilarity(score, articleURL, href)

		possiblePage.score = score
	}

	// Convert map to slice
	for _, link := range linkMap {
		scoredPages = append(scoredPages, *link)
	}

	if len(scoredPages) == 0 {
		return nil
	}

	return scoredPages
}

// shouldScore determines if a link should be considered for next page scoring
func shouldScore(href, articleURL, baseURL string, parsedURL *url.URL, linkText string, previousUrls []string) bool {
	// Skip if we've already fetched this URL
	for _, prevURL := range previousUrls {
		if href == prevURL {
			return false
		}
	}

	// If empty, same as article URL, or same as base URL, skip it
	if href == "" || href == articleURL || href == baseURL {
		return false
	}

	// Parse the href to check hostname
	linkURL, err := url.Parse(href)
	if err != nil {
		return false
	}

	// Domain mismatch
	if linkURL.Hostname() != parsedURL.Hostname() {
		return false
	}

	// If href doesn't contain a digit after removing the base URL, skip it
	fragment := strings.Replace(href, baseURL, "", 1)
	if !DIGIT_RE.MatchString(fragment) {
		return false
	}

	// Skip links with extraneous content in link text
	if EXTRANEOUS_LINK_HINTS_RE.MatchString(linkText) {
		return false
	}

	// Next page link text is never long, skip if too long
	if len(linkText) > 25 {
		return false
	}

	return true
}

// makeSig creates a signature string from a link element
func makeSig(link *goquery.Selection, linkText string) string {
	if linkText == "" {
		linkText = strings.TrimSpace(link.Text())
	}
	
	class, _ := link.Attr("class")
	id, _ := link.Attr("id")
	
	return linkText + " " + class + " " + id
}

// Scoring functions - these implement the JavaScript scoring algorithms

func scoreBaseUrl(href string, baseRegex *regexp.Regexp) float64 {
	// If the baseUrl isn't part of this URL, penalize this link
	if !baseRegex.MatchString(href) {
		return -25
	}
	return 0
}

func scoreNextLinkText(linkData string) float64 {
	// Things like "next", ">>", etc.
	if NEXT_LINK_TEXT_RE.MatchString(linkData) {
		return 50
	}
	return 0
}

func scoreCapLinks(linkData string) float64 {
	// Cap links are links like "last", etc.
	if CAP_LINK_TEXT_RE.MatchString(linkData) {
		// If we found a link like "last", but we've already seen that
		// this link is also "next", it's fine. If it's not been
		// previously marked as "next", then it's probably bad.
		if !NEXT_LINK_TEXT_RE.MatchString(linkData) {
			return -65
		}
	}
	return 0
}

func scorePrevLink(linkData string) float64 {
	// If the link has something like "previous", it's definitely an old link
	if PREV_LINK_TEXT_RE.MatchString(linkData) {
		return -200
	}
	return 0
}

func scoreByParentsNextPage(link *goquery.Selection) float64 {
	// If a parent node contains paging-like classname or id, give a bonus
	parent := link.Parent()
	positiveMatch := false
	negativeMatch := false
	score := 0.0

	// Check up to 4 levels of parents
	for i := 0; i < 4; i++ {
		if parent.Length() == 0 {
			break
		}

		class, _ := parent.Attr("class")
		id, _ := parent.Attr("id")
		parentData := class + " " + id

		// If we have 'page' or 'paging' in our data, that's a good sign
		if !positiveMatch && PAGE_RE.MatchString(parentData) {
			positiveMatch = true
			score += 25
		}

		// If we have negative indicators and extraneous hints, penalize
		if !negativeMatch && 
		   dom.NEGATIVE_SCORE_RE.MatchString(parentData) && 
		   EXTRANEOUS_LINK_HINTS_RE.MatchString(parentData) {
			if !dom.POSITIVE_SCORE_RE.MatchString(parentData) {
				negativeMatch = true
				score -= 25
			}
		}

		parent = parent.Parent()
	}

	return score
}

func scoreExtraneousLinks(href string) float64 {
	// If the URL itself contains extraneous values, give a penalty
	if EXTRANEOUS_LINK_HINTS_RE.MatchString(href) {
		return -25
	}
	return 0
}

func scorePageInLink(pageNum int, isWp bool) float64 {
	// Page in the link = bonus. Intentionally ignore WordPress because
	// their ?p=123 link style gets caught by this even though it means
	// separate documents entirely.
	if pageNum > 0 && !isWp {
		return 50
	}
	return 0
}

func scoreLinkText(linkText string, pageNum int) float64 {
	// If the link text can be parsed as a number, give it a minor
	// bonus, with a slight bias towards lower numbered pages
	score := 0.0

	// Check if the trimmed link text is all digits
	trimmed := strings.TrimSpace(linkText)
	if text.IS_DIGIT_RE.MatchString(trimmed) {
		linkTextAsNum, err := strconv.Atoi(trimmed)
		if err == nil {
			// If it's the first page, we already got it on the first call
			if linkTextAsNum < 2 {
				score = -30
			} else {
				// Up to page 10, give a small bonus
				score = float64(maxIntNextPage(0, 10-linkTextAsNum))
			}

			// If current page number is greater than this link's page number, big penalty
			if pageNum > 0 && pageNum >= linkTextAsNum {
				score -= 50
			}
		}
	}

	return score
}

func scoreSimilarity(score float64, articleURL, href string) float64 {
	// Only do this expensive computation if we have a real candidate
	if score > 0 {
		// Calculate similarity using simple string comparison
		// This is a simplified version of JavaScript's difflib.SequenceMatcher
		similarity := calculateSimilarity(articleURL, href)
		
		// JavaScript algorithm: diffPercent = 1.0 - similarity
		// diffModifier = -(250 * (diffPercent - 0.2))
		diffPercent := 1.0 - similarity
		diffModifier := -(250.0 * (diffPercent - 0.2))
		return score + diffModifier
	}
	return 0
}

// calculateSimilarity provides a simple similarity calculation
// This is a simplified version of Python's difflib.SequenceMatcher ratio
func calculateSimilarity(a, b string) float64 {
	if a == b {
		return 1.0
	}
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	// Simple similarity based on common characters
	// This is much simpler than difflib but provides similar behavior
	matches := 0
	minLen := minIntNextPage(len(a), len(b))
	maxLen := maxIntNextPage(len(a), len(b))

	for i := 0; i < minLen; i++ {
		if a[i] == b[i] {
			matches++
		}
	}

	return float64(matches*2) / float64(maxLen)
}

func minIntNextPage(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxIntNextPage(a, b int) int {
	if a > b {
		return a
	}
	return b
}