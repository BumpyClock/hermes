// ABOUTME: Mashable custom extractor with string transforms (.image-credit to figcaption)
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/mashable.com/index.js

package custom

// MashableComExtractor provides the custom extraction rules for mashable.com
// JavaScript equivalent: export const MashableComExtractor = { ... }.
var MashableComExtractor = &CustomExtractor{
	Domain: "mashable.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "header h1"},
			{Selector: "h1.title"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"article:author\"]", Attribute: "value"},
			{Selector: "span.author_name a"},
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
			{"#article"},
			{"section.article-content.blueprint"},
		},

		// Transform functions for Mashable-specific content
		Transforms: map[string]TransformFunction{
			".image-credit": &StringTransform{
				TargetTag: "figcaption",
			},
		},
	},
}

// GetMashableComExtractor returns the Mashable custom extractor.
func GetMashableComExtractor() *CustomExtractor {
	return MashableComExtractor
}
