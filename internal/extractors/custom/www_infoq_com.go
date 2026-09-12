// ABOUTME: InfoQ custom extractor with default cleaning disabled
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.infoq.com/index.js

package custom

// WwwInfoqComExtractor provides the custom extraction rules for www.infoq.com
// JavaScript equivalent: export const WwwInfoqComExtractor = { ... }.
var WwwInfoqComExtractor = &CustomExtractor{
	Domain: "www.infoq.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1.heading"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "div.widget.article__authors"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".article__readTime.date"},
		},
		// Note: format and timezone would be handled at extraction time
		// format: 'YYYY年MM月DD日' (from JavaScript)
		// timezone: 'Asia/Tokyo' (from JavaScript)
	},

	Dek: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:description\"]", Attribute: "value"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"div.article__data"},
		},
		DisableDefaultCleaner: true, // JavaScript disables default cleaning.
	},
}

// GetWwwInfoqComExtractor returns the InfoQ custom extractor.
func GetWwwInfoqComExtractor() *CustomExtractor {
	return WwwInfoqComExtractor
}
