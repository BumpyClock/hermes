// ABOUTME: Qdaily.com custom extractor with Chinese content support and lazy-load cleanup
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.qdaily.com/index.js

package custom

// QdailyCustomExtractor provides the custom extraction rules for www.qdaily.com
// JavaScript equivalent: export const WwwQdailyComExtractor = { ... }.
var QdailyCustomExtractor = &CustomExtractor{
	Domain: "www.qdaily.com",

	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h2",
			"h2.title",
		},
	},

	Author: &FieldExtractor{
		Selectors: []interface{}{
			".name",
		},
	},

	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				".detail",
			},
		},

		// Clean lazy-load elements
		Clean: []string{
			".lazyload",
			".lazylad",
			".lazylood",
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{".date.smart-date", "data-origindate"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{".article-detail-hd img", "src"},
		},
	},

	Dek: &FieldExtractor{
		Selectors: []interface{}{
			".excerpt",
		},
	},
}

// GetQdailyExtractor returns the Qdaily custom extractor.
func GetQdailyExtractor() *CustomExtractor {
	return QdailyCustomExtractor
}
