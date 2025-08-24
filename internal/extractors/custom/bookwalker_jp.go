// ABOUTME: BookWalker Japan e-book platform custom extractor with timezone support
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/bookwalker.jp/index.js

package custom

// BookwalkerJpExtractor provides the custom extraction rules for bookwalker.jp
// JavaScript equivalent: export const BookwalkerJpExtractor = { ... }
var BookwalkerJpExtractor = &CustomExtractor{
	Domain: "bookwalker.jp",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1.p-main__title",
			"h1.main-heading",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			"div.p-author__list",
			"div.authors",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			"dl.p-information__data dd:nth-of-type(7)",
			".work-info .work-detail:first-of-type .work-detail-contents:last-of-type",
		},
		// timezone: 'Asia/Tokyo' in JavaScript - note: Go implementation handles timezone in date cleaner
		Timezone: "Asia/Tokyo",
	},
	
	Dek: nil,
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				"div.p-main__information",
				[]interface{}{"div.main-info", "div.main-cover-inner"},
			},
		},
		
		// defaultCleaner: false in JavaScript
		DefaultCleaner: false,
		
		// transforms: {} (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean trial labels and promotional content
		Clean: []string{
			"span.label.label--trial",
			"dt.info-head.info-head--coin",
			"dd.info-contents.info-contents--coin",
			"div.info-notice.fn-toggleClass",
		},
	},
}

// GetBookwalkerJpExtractor returns the BookWalker Japan custom extractor
func GetBookwalkerJpExtractor() *CustomExtractor {
	return BookwalkerJpExtractor
}