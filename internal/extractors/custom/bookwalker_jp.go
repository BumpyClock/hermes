// ABOUTME: BookWalker Japan e-book platform custom extractor with timezone support
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/bookwalker.jp/index.js

package custom

// BookwalkerJpExtractor provides the custom extraction rules for bookwalker.jp
// JavaScript equivalent: export const BookwalkerJpExtractor = { ... }.
var BookwalkerJpExtractor = &CustomExtractor{
	Domain: "bookwalker.jp",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1.p-main__title"},
			{Selector: "h1.main-heading"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "div.p-author__list"},
			{Selector: "div.authors"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "dl.p-information__data dd:nth-of-type(7)"},
			{Selector: ".work-info .work-detail:first-of-type .work-detail-contents:last-of-type"},
		},
		// timezone: 'Asia/Tokyo' in JavaScript - note: Go implementation handles timezone in date cleaner
		Timezone: "Asia/Tokyo",
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"div.p-main__information"},
			{"div.main-info", "div.main-cover-inner"},
		},

		// defaultCleaner: false in JavaScript
		DisableDefaultCleaner: true,

		// Clean trial labels and promotional content
		Clean: []string{
			"span.label.label--trial",
			"dt.info-head.info-head--coin",
			"dd.info-contents.info-contents--coin",
			"div.info-notice.fn-toggleClass",
		},
	},
}

// GetBookwalkerJpExtractor returns the BookWalker Japan custom extractor.
func GetBookwalkerJpExtractor() *CustomExtractor {
	return BookwalkerJpExtractor
}
