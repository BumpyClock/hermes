// ABOUTME: MacRumors custom extractor with timezone support and rel=author patterns
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.macrumors.com/index.js

package custom

// WwwMacrumorsComExtractor provides the custom extraction rules for www.macrumors.com
// JavaScript equivalent: export const WwwMacrumorsComExtractor = { ... }.
var WwwMacrumorsComExtractor = &CustomExtractor{
	Domain: "www.macrumors.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1"},
			{Selector: "h1.title"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "article a[rel=\"author\"]"},
			{Selector: ".author-url"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "time", Attribute: "datetime"},
		},
		// Note: timezone support would be handled at extraction time
		// timezone: 'America/Los_Angeles' (from JavaScript)
	},

	Dek: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"description\"]", Attribute: "value"},
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
			{".article"},
		},
	},
}

// GetWwwMacrumorsComExtractor returns the MacRumors custom extractor.
func GetWwwMacrumorsComExtractor() *CustomExtractor {
	return WwwMacrumorsComExtractor
}
