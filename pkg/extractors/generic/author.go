// ABOUTME: Port of extractors/generic/author/extractor.js to Go
// This file provides 100% JavaScript-compatible author extraction with the
// same three-tier strategy: meta tags, CSS selectors, and byline regex patterns.
//
// JavaScript Compatibility: Maintains exact extraction order and logic:
// 1. extractFromMeta() with AUTHOR_META_TAGS priority
// 2. extractFromSelectors() with AUTHOR_SELECTORS priority  
// 3. BYLINE_SELECTORS_RE with /^[\n\s]*By/i pattern matching
// 4. cleanAuthor() with CLEAN_AUTHOR_RE for 'By' prefix removal
//
// Implementation: Uses existing DOM utilities (extractFromMeta, extractFromSelectors)
// and text utilities (normalizeSpaces) to maintain consistency with other extractors.
// All constants and patterns match JavaScript exactly for compatibility.
//
// Performance: Optimized Go implementation with efficient regex compilation
// and string manipulation while preserving JavaScript behavior.

package generic

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/BumpyClock/parser-go/pkg/utils/dom"
	"github.com/BumpyClock/parser-go/pkg/utils/text"
)

// Author extraction constants - exact ports from JavaScript

// AUTHOR_META_TAGS - ordered list of meta tag names that denote likely article authors
// From most distinct to least distinct. Note: "author" is too often the developer
// of the page, so it is not included here.
var AUTHOR_META_TAGS = []string{
	"byl",
	"clmst",
	"dc.author",
	"dcsext.author",
	"dc.creator",
	"rbauthors",
	"authors",
}

// AUTHOR_MAX_LENGTH - maximum length for valid author names
const AUTHOR_MAX_LENGTH = 300

// AUTHOR_SELECTORS - ordered list of CSS selectors to find likely article authors
// From most explicit to least explicit. Uses class substring matching like JavaScript.
var AUTHOR_SELECTORS = []string{
	".entry .entry-author",
	".author.vcard .fn",
	".author .vcard .fn",
	".byline.vcard .fn",
	".byline .vcard .fn",
	".byline .by .author",
	".byline .by",
	".byline .author",
	".post-author.vcard",
	".post-author .vcard",
	"a[rel=author]",
	"#by_author",
	".by_author",
	"#entryAuthor",
	".entryAuthor",
	".byline a[href*=author]",
	"#author .authorname",
	".author .authorname",
	"#author",
	".author",
	".articleauthor",
	".ArticleAuthor",
	".byline",
}

// BYLINE_SELECTORS_RE - selectors with regex patterns for byline content
// Matches /^[\n\s]*By/i pattern from JavaScript
var bylineRe = regexp.MustCompile(`(?i)^[\n\s]*By`)
var BYLINE_SELECTORS_RE = [][2]interface{}{
	{"#byline", bylineRe},
	{".byline", bylineRe},
}

// CLEAN_AUTHOR_RE - regex for cleaning author prefixes
// Matches /^\s*(posted |written )?by\s*:?\s*(.*)/i from JavaScript
var CLEAN_AUTHOR_RE = regexp.MustCompile(`(?i)^\s*(posted |written )?by\s*:?\s*(.*)`)

// GenericAuthorExtractor provides author extraction functionality
type GenericAuthorExtractor struct{}

// Extract extracts author information from HTML using the three-tier strategy
// Returns *string to allow nil for no author found (matching JavaScript behavior)
func (e *GenericAuthorExtractor) Extract(doc *goquery.Selection, metaCache []string) *string {
	var author string

	// First, check to see if we have a matching meta tag that we can make use of.
	// Need to get the document from the selection for meta tag extraction
	var document *goquery.Document
	if doc.Is("html") {
		// Already a document root
		if docRoot := doc.Get(0); docRoot != nil && docRoot.Type == 9 { // NodeDocument
			document = goquery.NewDocumentFromNode(docRoot)
		}
	}
	if document == nil {
		// Create a document from the current HTML
		if html, err := doc.Html(); err == nil {
			if strings.Contains(html, "<html") {
				document, _ = goquery.NewDocumentFromReader(strings.NewReader(html))
			} else {
				document, _ = goquery.NewDocumentFromReader(strings.NewReader("<html>" + html + "</html>"))
			}
		}
	}
	
	if document != nil {
		authorPtr := dom.ExtractFromMeta(document, AUTHOR_META_TAGS, metaCache, true)
		if authorPtr != nil {
			author = *authorPtr
			if len(author) < AUTHOR_MAX_LENGTH {
				cleaned := cleanAuthor(author)
				return &cleaned
			}
		}
	}

	// Second, look through our selectors looking for potential authors.
	authorPtr := dom.ExtractFromSelectors(doc, AUTHOR_SELECTORS, 2, true)
	if authorPtr != nil {
		author = *authorPtr
		if len(author) < AUTHOR_MAX_LENGTH {
			cleaned := cleanAuthor(author)
			return &cleaned
		}
	}

	// Last, use our looser regular-expression based selectors for potential authors.
	for _, selectorRegex := range BYLINE_SELECTORS_RE {
		selector := selectorRegex[0].(string)
		regex := selectorRegex[1].(*regexp.Regexp)

		node := doc.Find(selector)
		if node.Length() == 1 {
			text := strings.TrimSpace(node.Text())
			if regex.MatchString(text) {
				cleaned := cleanAuthor(text)
				return &cleaned
			}
		}
	}

	return nil
}

// cleanAuthor cleans author strings by removing prefixes like "By", "posted by", etc.
// Matches the JavaScript cleanAuthor function exactly
func cleanAuthor(author string) string {
	// Apply the CLEAN_AUTHOR_RE regex to remove prefixes
	matches := CLEAN_AUTHOR_RE.FindStringSubmatch(author)
	if len(matches) >= 3 {
		// Use the third capture group (the actual author name)
		author = matches[2]
	}

	// Normalize spaces and trim
	return text.NormalizeSpaces(strings.TrimSpace(author))
}