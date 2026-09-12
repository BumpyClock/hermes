// ABOUTME: Fandom Wikia.com custom extractor with wiki content and community-driven articles
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/fandom.wikia.com/index.js

package custom

// FandomWikiaCustomExtractor provides the custom extraction rules for fandom.wikia.com
// JavaScript equivalent: export const WikiaExtractor = { ... }.
var FandomWikiaCustomExtractor = &CustomExtractor{
	Domain: "fandom.wikia.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1.entry-title"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".author vcard"},
			{Selector: ".fn"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{".grid-content"},
			{".entry-content"},
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
}

// GetFandomWikiaExtractor returns the Fandom Wikia custom extractor.
func GetFandomWikiaExtractor() *CustomExtractor {
	return FandomWikiaCustomExtractor
}
