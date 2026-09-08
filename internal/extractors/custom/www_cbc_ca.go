// ABOUTME: CBC (Canadian broadcaster) custom extractor with simple story-based extraction
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.cbc.ca/index.js

package custom

// WwwCbcCaExtractor provides the custom extraction rules for www.cbc.ca
// JavaScript equivalent: export const WwwCbcCaExtractor = { ... }.
var WwwCbcCaExtractor = &CustomExtractor{
	Domain: "www.cbc.ca",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".authorText"},
			{Selector: ".bylineDetails"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{".story"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".timeStamp[datetime]", Attribute: "datetime"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Dek: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".deck"},
		},
	},
}

// GetWwwCbcCaExtractor returns the CBC custom extractor.
func GetWwwCbcCaExtractor() *CustomExtractor {
	return WwwCbcCaExtractor
}
