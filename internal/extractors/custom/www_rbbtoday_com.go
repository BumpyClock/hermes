// ABOUTME: RBB TODAY Japan tech news site custom extractor with article content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.rbbtoday.com/index.js

package custom

// WwwRbbtodayComExtractor provides the custom extraction rules for www.rbbtoday.com
// JavaScript equivalent: export const WwwRbbtodayComExtractor = { ... }.
var WwwRbbtodayComExtractor = &CustomExtractor{
	Domain: "www.rbbtoday.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".writer.writer-name"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "header time", Attribute: "datetime"},
		},
	},

	Dek: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"description\"]", Attribute: "value"},
			{Selector: ".arti-summary"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{".arti-content"},
		},

		// Clean promotional content
		Clean: []string{
			".arti-giga",
		},
	},
}

// GetWwwRbbtodayComExtractor returns the RBB TODAY Japan custom extractor.
func GetWwwRbbtodayComExtractor() *CustomExtractor {
	return WwwRbbtodayComExtractor
}
