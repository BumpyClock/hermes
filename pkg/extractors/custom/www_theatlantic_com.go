// ABOUTME: The Atlantic custom extractor for long-form journalism articles
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.theatlantic.com/index.js

package custom

// TheAtlanticCustomExtractor provides the custom extraction rules for www.theatlantic.com
// JavaScript equivalent: export const TheAtlanticExtractor = { ... }
var TheAtlanticCustomExtractor = &CustomExtractor{
	Domain: "www.theatlantic.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1",
			".c-article-header__hed",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"author\"]", "value"},
			".c-byline__author",
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				"article",
				".article-body",
			},
		},
		
		// No transforms in original JavaScript (empty array)
		Transforms: map[string]TransformFunction{},
		
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
		Selectors: []interface{}{
			[]string{"meta[name=\"description\"]", "value"},
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"time[itemprop=\"datePublished\"]", "datetime"},
		},
	},
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
	
	// Next page URL and excerpt are null in original JavaScript
	NextPageURL: &FieldExtractor{
		Selectors: []interface{}{},
	},
	
	Excerpt: &FieldExtractor{
		Selectors: []interface{}{},
	},
}

// GetTheAtlanticExtractor returns the The Atlantic custom extractor
func GetTheAtlanticExtractor() *CustomExtractor {
	return TheAtlanticCustomExtractor
}