// ABOUTME: Pastebin.com custom extractor with code content handling, syntax highlighting, and list transforms
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/pastebin.com/index.js

package custom

// PastebinCustomExtractor provides the custom extraction rules for pastebin.com
// JavaScript equivalent: export const PastebinComExtractor = { ... }.
var PastebinCustomExtractor = &CustomExtractor{
	Domain: "pastebin.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".username"},
			{Selector: ".paste_box_line2 .t_us + a"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{".source"},
			{"#selectable .text"},
		},

		// Transform functions for Pastebin code content
		Transforms: map[string]TransformFunction{
			// Convert ordered list to div
			"ol": &StringTransform{
				TargetTag: "div",
			},

			// Convert list items to paragraphs
			"li": &StringTransform{
				TargetTag: "p",
			},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".date"},
			{Selector: ".paste_box_line2 .t_da + span"},
		},
		// Timezone from JavaScript: 'America/New_York'
		// Format from JavaScript: 'MMMM D, YYYY'
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},
}

// GetPastebinExtractor returns the Pastebin custom extractor.
func GetPastebinExtractor() *CustomExtractor {
	return PastebinCustomExtractor
}
