// ABOUTME: Generic title extractor with JavaScript-compatible fallback logic and cleaning
// ABOUTME: Extracts article titles using meta tags and CSS selectors with domain name removal

package generic

import (
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/BumpyClock/hermes/internal/cleaners"
	"github.com/BumpyClock/hermes/internal/utils/dom"
	"github.com/BumpyClock/hermes/internal/utils/text"
)

// Title extraction constants matching JavaScript behavior exactly.
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
)

// GenericTitleExtractor extracts article titles using multiple fallback strategies.
var GenericTitleExtractor = struct {
	Extract func(doc *goquery.Selection, url string, metaCache []string) string
}{
	Extract: func(doc *goquery.Selection, url string, metaCache []string) string {
		if doc.Length() == 0 {
			return ""
		}
		document := &goquery.Document{Selection: doc}

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

// cleanTitle cleans and normalizes the title text.
func cleanTitle(title string, url string, doc *goquery.Selection) string {
	// If title has |, :, or - in it, see if we can clean it up.
	if cleaners.TITLE_SPLITTERS_RE.MatchString(title) {
		title = cleaners.ResolveExtractedTitle(title, url)
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
	cleaned := dom.StripTags(title)
	return text.NormalizeSpaces(strings.TrimSpace(cleaned))
}
