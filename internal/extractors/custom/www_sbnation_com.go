// ABOUTME: SB Nation custom extractor with Vox Media content patterns
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.sbnation.com/index.js

package custom

// WwwSbnationComExtractor provides the custom extraction rules for www.sbnation.com
// JavaScript equivalent: export const WwwSbnationComExtractor = { ... }.
var WwwSbnationComExtractor = &CustomExtractor{
	Domain: "www.sbnation.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1.c-page-title"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"author\"]", Attribute: "value"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"article:published_time\"]", Attribute: "value"},
		},
	},

	Dek: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "p.c-entry-summary.p-dek"},
			{Selector: "h2.c-entry-summary.p-dek"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"div.c-entry-content"},
		},
	},
}

// GetWwwSbnationComExtractor returns the SB Nation custom extractor.
func GetWwwSbnationComExtractor() *CustomExtractor {
	return WwwSbnationComExtractor
}
