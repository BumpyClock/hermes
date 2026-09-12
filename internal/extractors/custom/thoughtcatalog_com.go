// ABOUTME: ThoughtCatalog.com custom extractor with lifestyle content, writer profiles, and content cleaning
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/thoughtcatalog.com/index.js

package custom

// ThoughtCatalogCustomExtractor provides the custom extraction rules for thoughtcatalog.com
// JavaScript equivalent: export const ThoughtcatalogComExtractor = { ... }.
var ThoughtCatalogCustomExtractor = &CustomExtractor{
	Domain: "thoughtcatalog.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1.title"},
			{Selector: "meta[name=\"og:title\"]", Attribute: "value"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "cite a"},
			{Selector: "div.col-xs-12.article_header div.writer-container.writer-container-inline.writer-no-avatar h4.writer-name"},
			{Selector: "h1.writer-name"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{".entry.post"},
		},

		// Clean selectors - remove unwanted elements
		Clean: []string{
			".tc_mark",
			"figcaption",
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

// GetThoughtCatalogExtractor returns the ThoughtCatalog custom extractor.
func GetThoughtCatalogExtractor() *CustomExtractor {
	return ThoughtCatalogCustomExtractor
}
