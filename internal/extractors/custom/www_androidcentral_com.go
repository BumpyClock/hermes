// ABOUTME: Android Central custom extractor with simple meta selector patterns
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.androidcentral.com/index.js

package custom

// WwwAndroidcentralComExtractor provides the custom extraction rules for www.androidcentral.com
// JavaScript equivalent: export const WwwAndroidcentralComExtractor = { ... }.
var WwwAndroidcentralComExtractor = &CustomExtractor{
	Domain: "www.androidcentral.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1"},
			{Selector: "h1.main-title"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"parsely-author\"]", Attribute: "value"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"article:published_time\"]", Attribute: "value"},
		},
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
			{"#article-body"},
		},

		// Clean selectors
		Clean: []string{
			".intro",
			"blockquote",
		},
	},
}

// GetWwwAndroidcentralComExtractor returns the Android Central custom extractor.
func GetWwwAndroidcentralComExtractor() *CustomExtractor {
	return WwwAndroidcentralComExtractor
}
