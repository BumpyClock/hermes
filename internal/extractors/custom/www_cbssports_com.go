// ABOUTME: CBS Sports custom extractor with UTC timezone and article content patterns
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.cbssports.com/index.js

package custom

// WwwCbssportsComExtractor provides the custom extraction rules for www.cbssports.com
// JavaScript equivalent: export const WwwCbssportsComExtractor = { ... }.
var WwwCbssportsComExtractor = &CustomExtractor{
	Domain: "www.cbssports.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".Article-headline"},
			{Selector: ".article-headline"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".ArticleAuthor-nameText"},
			{Selector: ".author-name"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[itemprop=\"datePublished\"]", Attribute: "value"},
		},
		Timezone: "UTC",
	},

	Dek: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".Article-subline"},
			{Selector: ".article-subline"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{".article"},
		},
	},
}

// GetWwwCbssportsComExtractor returns the CBS Sports custom extractor.
func GetWwwCbssportsComExtractor() *CustomExtractor {
	return WwwCbssportsComExtractor
}
