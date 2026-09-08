// ABOUTME: Bustle custom extractor for fashion/lifestyle content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.bustle.com/index.js

package custom

// BustleCustomExtractor provides the custom extraction rules for www.bustle.com
// JavaScript equivalent: export const WwwBustleComExtractor = { ... }.
var BustleCustomExtractor = &CustomExtractor{
	Domain: "www.bustle.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1"},
			{Selector: "h1.post-page__title"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "a[href*=\"profile\"]"},
			{Selector: "div.content-meta__author"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "time", Attribute: "datetime"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"article"},
			{".post-page__body"},
		},
	},
}

// GetBustleExtractor returns the Bustle custom extractor.
func GetBustleExtractor() *CustomExtractor {
	return BustleCustomExtractor
}
