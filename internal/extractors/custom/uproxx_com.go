// ABOUTME: Uproxx custom extractor for music and entertainment content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/uproxx.com/index.js

package custom

// UproxxCustomExtractor provides the custom extraction rules for uproxx.com
// JavaScript equivalent: export const UproxxComExtractor = { ... }.
var UproxxCustomExtractor = &CustomExtractor{
	Domain: "uproxx.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "div.entry-header h1"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"qc:author\"]", Attribute: "value"},
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
			{".entry-content"},
		},

		// Transform functions for Uproxx-specific content
		Transforms: map[string]TransformFunction{
			"div.image":                  &StringTransform{TargetTag: "figure"},
			"div.image .wp-media-credit": &StringTransform{TargetTag: "figcaption"},
		},
	},
}

// GetUproxxExtractor returns the Uproxx custom extractor.
func GetUproxxExtractor() *CustomExtractor {
	return UproxxCustomExtractor
}
