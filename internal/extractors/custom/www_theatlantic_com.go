// ABOUTME: The Atlantic custom extractor for long-form journalism articles
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.theatlantic.com/index.js

package custom

// TheAtlanticCustomExtractor provides the custom extraction rules for www.theatlantic.com
// JavaScript equivalent: export const TheAtlanticExtractor = { ... }.
var TheAtlanticCustomExtractor = &CustomExtractor{
	Domain: "www.theatlantic.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1"},
			{Selector: ".c-article-header__hed"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"author\"]", Attribute: "value"},
			{Selector: ".c-byline__author"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"article"},
			{".article-body"},
		},

		// Clean selectors - remove unwanted elements
		Clean: []string{
			".partner-box",
			".callout",
			".c-article-writer__image",
			".c-article-writer__content",
			".c-letters-cta__text",
			".c-footer__logo",
			".c-recirculation-link",
			".twitter-tweet",
		},
	},

	Dek: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"description\"]", Attribute: "value"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "time[itemprop=\"datePublished\"]", Attribute: "datetime"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},
}

// GetTheAtlanticExtractor returns the The Atlantic custom extractor.
func GetTheAtlanticExtractor() *CustomExtractor {
	return TheAtlanticCustomExtractor
}
