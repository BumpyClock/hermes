// ABOUTME: Wired.com custom extractor with article content patterns
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.wired.com/index.js

package custom

// WwwWiredComExtractor provides the custom extraction rules for www.wired.com
// JavaScript equivalent: export const WiredExtractor = { ... }.
var WwwWiredComExtractor = &CustomExtractor{
	Domain: "www.wired.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1[data-testId=\"ContentHeaderHed\"]"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"article:author\"]", Attribute: "value"},
			{Selector: "a[rel=\"author\"]"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"article.article.main-content"},
			{"article.content"},
		},

		// Clean selectors - remove unwanted elements
		Clean: []string{
			".visually-hidden",
			"figcaption img.photo",
			".alert-message",
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

// GetWwwWiredComExtractor returns the Wired.com custom extractor.
func GetWwwWiredComExtractor() *CustomExtractor {
	return WwwWiredComExtractor
}
