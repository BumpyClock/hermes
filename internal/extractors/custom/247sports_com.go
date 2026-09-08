// ABOUTME: 247Sports custom extractor with college sports content patterns and data-published date handling
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/247sports.com/index.js

package custom

// TwofortysevensportsComExtractor provides the custom extraction rules for 247sports.com
// JavaScript equivalent: export const twofortysevensportsComExtractor = { ... }.
var TwofortysevensportsComExtractor = &CustomExtractor{
	Domain: "247sports.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "title"},
			{Selector: "article header h1"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".article-cnt__author"},
			{Selector: ".author"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "time[data-published]", Attribute: "data-published"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{".article-body"},
			{"section.body.article"},
		},
	},
}

// GetTwofortysevensportsComExtractor returns the 247Sports custom extractor.
func GetTwofortysevensportsComExtractor() *CustomExtractor {
	return TwofortysevensportsComExtractor
}
