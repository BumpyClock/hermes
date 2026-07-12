// ABOUTME: ScienceFly custom extractor for science education content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/sciencefly.com/index.js

package custom

// ScienceflyComExtractor provides the custom extraction rules for sciencefly.com
// JavaScript equivalent: export const ScienceflyComExtractor = { ... }.
var ScienceflyComExtractor = &CustomExtractor{
	Domain: "sciencefly.com",

	Title: &FieldExtractor{
		Selectors: []interface{}{
			".entry-title",
			".cb-entry-title",
			".cb-single-title",
		},
	},

	Author: &FieldExtractor{
		Selectors: []interface{}{
			"div.cb-author",
			"div.cb-author-title",
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:published_time\"]", "value"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"div.theiaPostSlider_slides img", "src"},
		},
	},

	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				"div.theiaPostSlider_slides",
			},
		},
	},
}

// GetScienceflyComExtractor returns the ScienceFly custom extractor.
func GetScienceflyComExtractor() *CustomExtractor {
	return ScienceflyComExtractor
}
