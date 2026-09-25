package generic

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"

	"github.com/BumpyClock/hermes/internal/utils/dom"
	"github.com/BumpyClock/hermes/internal/utils/text"
)

// Author extraction constants - exact ports from JavaScript

// AUTHOR_META_TAGS - ordered list of meta tag names that denote likely article authors
// From most distinct to least distinct. Note: "author" is too often the developer
// of the page, so it is not included here.
var AUTHOR_META_TAGS = dom.MustCompileMetaNames(
	"byl",
	"clmst",
	"dc.author",
	"dcsext.author",
	"dc.creator",
	"rbauthors",
	"authors",
)

// AUTHOR_MAX_LENGTH - maximum length for valid author names.
const AUTHOR_MAX_LENGTH = 300

// AUTHOR_SELECTORS - ordered list of CSS selectors to find likely article authors
// From most explicit to least explicit. Uses class substring matching like JavaScript.
var AUTHOR_SELECTORS = []goquery.Matcher{
	cascadia.MustCompile(".entry .entry-author"),
	cascadia.MustCompile(".author.vcard .fn"),
	cascadia.MustCompile(".author .vcard .fn"),
	cascadia.MustCompile(".byline.vcard .fn"),
	cascadia.MustCompile(".byline .vcard .fn"),
	cascadia.MustCompile(".byline .by .author"),
	cascadia.MustCompile(".byline .by"),
	cascadia.MustCompile(".byline .author"),
	cascadia.MustCompile(".post-author.vcard"),
	cascadia.MustCompile(".post-author .vcard"),
	cascadia.MustCompile("a[rel=author]"),
	cascadia.MustCompile("#by_author"),
	cascadia.MustCompile(".by_author"),
	cascadia.MustCompile("#entryAuthor"),
	cascadia.MustCompile(".entryAuthor"),
	cascadia.MustCompile(".byline a[href*=author]"),
	cascadia.MustCompile("#author .authorname"),
	cascadia.MustCompile(".author .authorname"),
	cascadia.MustCompile("#author"),
	cascadia.MustCompile(".author"),
	cascadia.MustCompile(".articleauthor"),
	cascadia.MustCompile(".ArticleAuthor"),
	cascadia.MustCompile(".byline"),
}

// BYLINE_SELECTORS_RE - compiled selectors with regex patterns for byline content
// Matches /^[\n\s]*By/i pattern from JavaScript.
var bylineRe = regexp.MustCompile(`(?i)^[\n\s]*By`)
var BYLINE_SELECTORS_RE = []struct {
	Selector goquery.Matcher
	Pattern  *regexp.Regexp
}{
	{cascadia.MustCompile("#byline"), bylineRe},
	{cascadia.MustCompile(".byline"), bylineRe},
}

// CLEAN_AUTHOR_RE - regex for cleaning author prefixes
// Matches /^\s*(posted |written )?by\s*:?\s*(.*)/i from JavaScript.
var CLEAN_AUTHOR_RE = regexp.MustCompile(`(?i)^\s*(posted |written )?by\s*:?\s*(.*)`)

var authorNavigationRE = regexp.MustCompile(`(?i)(^|[\s_-])(sidebar|navigation|nav|menu)($|[\s_-])`)

var (
	authorLinkMatcher       = cascadia.MustCompile("a[rel=author]")
	authorContextMatcher    = cascadia.MustCompile("article, .byline, .author, .post-author, [itemprop=author]")
	authorNavigationMatcher = cascadia.MustCompile("aside, nav, [role=navigation], [role=complementary]")
)

// GenericAuthorExtractor provides author extraction functionality.
type GenericAuthorExtractor struct{}

// Extract extracts author information from HTML using the three-tier strategy
// Returns *string to allow nil for no author found (matching JavaScript behavior).
func (e *GenericAuthorExtractor) Extract(doc *goquery.Selection, metaCache []string) *string {
	var author string

	// First, check to see if we have a matching meta tag that we can make use of.
	document := &goquery.Document{Selection: doc}
	if authorPtr := dom.ExtractFromMeta(document, AUTHOR_META_TAGS, metaCache, true); authorPtr != nil {
		author = *authorPtr
		if len(author) < AUTHOR_MAX_LENGTH {
			cleaned := cleanAuthor(author)
			return &cleaned
		}
	}

	// Second, look through our selectors looking for potential authors.
	authorPtr := dom.ExtractFromMatchers(doc, AUTHOR_SELECTORS, 2, true, isArticleAuthorCandidate)
	if authorPtr != nil {
		author = *authorPtr
		if len(author) < AUTHOR_MAX_LENGTH {
			cleaned := cleanAuthor(author)
			return &cleaned
		}
	}

	// Last, use our looser regular-expression based selectors for potential authors.
	for _, byline := range BYLINE_SELECTORS_RE {
		node := doc.FindMatcher(byline.Selector)
		if node.Length() == 1 {
			text := strings.TrimSpace(node.Text())
			if byline.Pattern.MatchString(text) {
				cleaned := cleanAuthor(text)
				return &cleaned
			}
		}
	}

	return nil
}

func isArticleAuthorCandidate(node *goquery.Selection) bool {
	if !node.IsMatcher(authorLinkMatcher) || node.ClosestMatcher(authorContextMatcher).Length() != 0 {
		return true
	}
	for parent := node.Parent(); parent.Length() != 0; parent = parent.Parent() {
		if parent.IsMatcher(authorNavigationMatcher) ||
			authorNavigationRE.MatchString(parent.AttrOr("id", "")+" "+parent.AttrOr("class", "")) {
			return false
		}
	}
	return true
}

// cleanAuthor cleans author strings by removing prefixes like "By", "posted by", etc.
// Matches the JavaScript cleanAuthor function exactly.
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
