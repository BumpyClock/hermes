// ABOUTME: HuffingtonPost.com custom extractor with news article support and content cleaning
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.huffingtonpost.com/index.js

package custom

// HuffingtonPostCustomExtractor provides the custom extraction rules for www.huffingtonpost.com
// JavaScript equivalent: export const WwwHuffingtonpostComExtractor = { ... }.
var HuffingtonPostCustomExtractor = &CustomExtractor{
	Domain: "www.huffingtonpost.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1.headline__title"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "span.author-card__details__name"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"div.entry__body"},
		},
		DisableDefaultCleaner: true,

		// Clean selectors - remove unwanted elements
		Clean: []string{
			".pull-quote",
			".tag-cloud",
			".embed-asset",
			".below-entry",
			".entry-corrections",
			"#suggested-story",
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"article:modified_time\"]", Attribute: "value"},
			{Selector: "meta[name=\"article:published_time\"]", Attribute: "value"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Dek: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h2.headline__subtitle"},
		},
	},
}

// GetHuffingtonPostExtractor returns the HuffingtonPost custom extractor.
func GetHuffingtonPostExtractor() *CustomExtractor {
	return HuffingtonPostCustomExtractor
}
