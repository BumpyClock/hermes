// ABOUTME: Takagi Hiromitsu academic researcher personal site extractor
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/takagi-hiromitsu.jp/index.js

package custom

// TakagihiromitsuJpExtractor provides the custom extraction rules for takagi-hiromitsu.jp
// JavaScript equivalent: export const TakagihiromitsuJpExtractor = { ... }.
var TakagihiromitsuJpExtractor = &CustomExtractor{
	Domain: "takagi-hiromitsu.jp",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h3"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"author\"]", Attribute: "value"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[http-equiv=\"Last-Modified\"]", Attribute: "value"},
		},
	},

	// Dek is null in JavaScript

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"div.body"},
		},
		DisableDefaultCleaner: true,
	},
}

// GetTakagihiromitsuJpExtractor returns the Takagi Hiromitsu custom extractor.
func GetTakagihiromitsuJpExtractor() *CustomExtractor {
	return TakagihiromitsuJpExtractor
}
