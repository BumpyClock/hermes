// ABOUTME: Le Monde (www.lemonde.fr) custom extractor for French news content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.lemonde.fr/index.js

package custom

// WwwLemondeFrExtractor provides the custom extraction rules for www.lemonde.fr
// JavaScript equivalent: export const WwwLemondeFrExtractor = { ... }.
var WwwLemondeFrExtractor = &CustomExtractor{
	Domain: "www.lemonde.fr",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1"},
			{Selector: "h1.article__title"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".article__author-link"},
			{Selector: ".author__name"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{".article__content"},
		},

		// Clean selectors - remove unwanted elements
		Clean: []string{
			"figcaption",
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:article:published_time\"]", Attribute: "value"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Dek: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".article__desc"},
		},
	},
}

// GetWwwLemondeFrExtractor returns the Le Monde custom extractor.
func GetWwwLemondeFrExtractor() *CustomExtractor {
	return WwwLemondeFrExtractor
}
