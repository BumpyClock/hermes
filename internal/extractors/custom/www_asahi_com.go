// ABOUTME: Asahi Shimbun (major Japanese newspaper) custom extractor with main content selection
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.asahi.com/index.js

package custom

// WwwAsahiComExtractor provides the custom extraction rules for www.asahi.com
// JavaScript equivalent: export const WwwAsahiComExtractor = { ... }.
var WwwAsahiComExtractor = &CustomExtractor{
	Domain: "www.asahi.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "main h1"},
			{Selector: ".ArticleTitle h1"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"article:author\"]", Attribute: "value"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"pubdate\"]", Attribute: "value"},
		},
	},

	Excerpt: &FieldExtractor{
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
			{"main"},
		},

		// defaultCleaner: false in JavaScript
		DisableDefaultCleaner: true,

		// Clean selectors to remove ads and unwanted content
		Clean: []string{
			"div.AdMod",
			"div.LoginSelectArea",
			"time",
			"div.notPrint",
		},
	},
}

// GetWwwAsahiComExtractor returns the Asahi Shimbun custom extractor.
func GetWwwAsahiComExtractor() *CustomExtractor {
	return WwwAsahiComExtractor
}
