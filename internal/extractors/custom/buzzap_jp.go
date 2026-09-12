// ABOUTME: BuzzAP Japan news site custom extractor with entry content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/buzzap.jp/index.js

package custom

// BuzzapJpExtractor provides the custom extraction rules for buzzap.jp
// JavaScript equivalent: export const BuzzapJpExtractor = { ... }.
var BuzzapJpExtractor = &CustomExtractor{
	Domain: "buzzap.jp",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1.entry-title"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "time.entry-date", Attribute: "datetime"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"div.ctiframe"},
		},

		// defaultCleaner: false in JavaScript
		DisableDefaultCleaner: true,
	},
}

// GetBuzzapJpExtractor returns the BuzzAP Japan custom extractor.
func GetBuzzapJpExtractor() *CustomExtractor {
	return BuzzapJpExtractor
}
