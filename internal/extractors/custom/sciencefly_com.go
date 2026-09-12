// ABOUTME: ScienceFly custom extractor for science education content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/sciencefly.com/index.js

package custom

// ScienceflyComExtractor provides the custom extraction rules for sciencefly.com
// JavaScript equivalent: export const ScienceflyComExtractor = { ... }.
var ScienceflyComExtractor = &CustomExtractor{
	Domain: "sciencefly.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".entry-title"},
			{Selector: ".cb-entry-title"},
			{Selector: ".cb-single-title"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "div.cb-author"},
			{Selector: "div.cb-author-title"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"article:published_time\"]", Attribute: "value"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "div.theiaPostSlider_slides img", Attribute: "src"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"div.theiaPostSlider_slides"},
		},
	},
}

// GetScienceflyComExtractor returns the ScienceFly custom extractor.
func GetScienceflyComExtractor() *CustomExtractor {
	return ScienceflyComExtractor
}
