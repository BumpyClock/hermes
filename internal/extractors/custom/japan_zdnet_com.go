// ABOUTME: ZDNet Japan custom extractor with cXenseParse:author meta pattern
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/japan.zdnet.com/index.js

package custom

// JapanZdnetComExtractor provides the custom extraction rules for japan.zdnet.com
// JavaScript equivalent: export const JapanZdnetComExtractor = { ... }.
var JapanZdnetComExtractor = &CustomExtractor{
	Domain: "japan.zdnet.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"cXenseParse:author\"]", Attribute: "value"},
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
			{"div.article_body"},
		},
	},
}

// GetJapanZdnetComExtractor returns the ZDNet Japan custom extractor.
func GetJapanZdnetComExtractor() *CustomExtractor {
	return JapanZdnetComExtractor
}
