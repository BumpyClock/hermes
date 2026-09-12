// ABOUTME: US Magazine custom extractor for celebrity magazine content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.usmagazine.com/index.js

package custom

// USMagazineCustomExtractor provides the custom extraction rules for www.usmagazine.com
// JavaScript equivalent: export const WwwUsmagazineComExtractor = { ... }.
var USMagazineCustomExtractor = &CustomExtractor{
	Domain: "www.usmagazine.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "header h1"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "a.author"},
			{Selector: "a.article-byline.tracked-offpage"},
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

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"div.article-content"},
		},

		// Clean selectors - remove unwanted elements
		Clean: []string{
			".module-related",
		},
	},
}

// GetUSMagazineExtractor returns the US Magazine custom extractor.
func GetUSMagazineExtractor() *CustomExtractor {
	return USMagazineCustomExtractor
}
