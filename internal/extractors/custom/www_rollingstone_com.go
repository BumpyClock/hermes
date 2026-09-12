// ABOUTME: Rolling Stone custom extractor for music site with album reviews
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.rollingstone.com/index.js

package custom

// RollingStoneCustomExtractor provides the custom extraction rules for www.rollingstone.com
// JavaScript equivalent: export const WwwRollingstoneComExtractor = { ... }.
var RollingStoneCustomExtractor = &CustomExtractor{
	Domain: "www.rollingstone.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1.l-article-header__row--title"},
			{Selector: "h1.content-title"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "a.c-byline__link"},
			{Selector: "a.content-author.tracked-offpage"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"article:published_time\"]", Attribute: "value"},
			{Selector: "time.content-published-date"},
		},
	},

	Dek: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h2.l-article-header__row--lead"},
			{Selector: ".content-description"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{".l-article-content"},
			{".lead-container", ".article-content"},
			{".article-content"},
		},

		// Clean selectors - remove unwanted elements
		Clean: []string{
			".c-related-links-wrapper",
			".module-related",
		},
	},
}

// GetRollingStoneExtractor returns the Rolling Stone custom extractor.
func GetRollingStoneExtractor() *CustomExtractor {
	return RollingStoneCustomExtractor
}
