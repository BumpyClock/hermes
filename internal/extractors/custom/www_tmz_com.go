// ABOUTME: TMZ custom extractor for celebrity content and photo galleries
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.tmz.com/index.js

package custom

// TMZCustomExtractor provides the custom extraction rules for www.tmz.com
// JavaScript equivalent: export const WwwTmzComExtractor = { ... }.
var TMZCustomExtractor = &CustomExtractor{
	Domain: "www.tmz.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".post-title-breadcrumb"},
			{Selector: "h1"},
			{Selector: ".headline"},
		},
	},

	// Author is a static string in original JavaScript
	Author: &FieldExtractor{
		Selectors: []SelectorEntry{{Selector: "TMZ STAFF"}},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".article__published-at"},
			{Selector: ".article-posted-date"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{".article__blocks"},
			{".article-content"},
			{".all-post-body"},
		},

		// Clean selectors - remove unwanted elements
		Clean: []string{
			".lightbox-link",
		},
	},
}

// GetTMZExtractor returns the TMZ custom extractor.
func GetTMZExtractor() *CustomExtractor {
	return TMZCustomExtractor
}
