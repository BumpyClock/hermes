// ABOUTME: PopSugar custom extractor for lifestyle with shopping content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.popsugar.com/index.js

package custom

// PopSugarCustomExtractor provides the custom extraction rules for www.popsugar.com
// JavaScript equivalent: export const WwwPopsugarComExtractor = { ... }.
var PopSugarCustomExtractor = &CustomExtractor{
	Domain: "www.popsugar.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h2.post-title"},
			{Selector: "title-text"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"article:author\"]", Attribute: "value"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"article:published_time\"]", Attribute: "value"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"#content"},
		},

		// Clean selectors - remove unwanted elements
		Clean: []string{
			".share-copy-title",
			".post-tags",
			".reactions",
		},
	},
}

// GetPopSugarExtractor returns the PopSugar custom extractor.
func GetPopSugarExtractor() *CustomExtractor {
	return PopSugarCustomExtractor
}
