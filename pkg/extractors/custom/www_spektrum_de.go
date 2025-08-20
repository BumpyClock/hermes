// ABOUTME: Spektrum.de (German science magazine) custom extractor with image selector patterns
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.spektrum.de/index.js

package custom

// WwwSpektrumDeExtractor provides the custom extraction rules for www.spektrum.de
// JavaScript equivalent: export const SpektrumExtractor = { ... }
var WwwSpektrumDeExtractor = &CustomExtractor{
	Domain: "www.spektrum.de",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			".content__title",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			".content__author__info__name",
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				"article.content",
			},
		},
		
		// Transform functions (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors - remove unwanted elements
		Clean: []string{
			".breadcrumbs",
			".hide-for-print",
			"aside",
			"header h2",
			".image__article__top",
			".content__author",
			".copyright",
			".callout-box",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			".content__meta__date",
		},
		// Note: JavaScript version has timezone: 'Europe/Berlin' - this is handled by dateparse in Go
	},
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			// This is how the meta tag appears in the original source code.
			[]string{"meta[name=\"og:image\"]", "value"},
			// This is how the meta tag appears in the DOM in Chrome.
			// The selector is included here to make the code work within the browser as well.
			[]string{"meta[property=\"og:image\"]", "content"},
			// This is the image that is shown on the page.
			// It can be slightly cropped compared to the original in the meta tag.
			".image__article__top img",
		},
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			".content__intro",
		},
	},
	
	NextPageURL: nil,
	
	Excerpt: nil,
}

// GetWwwSpektrumDeExtractor returns the Spektrum.de custom extractor
func GetWwwSpektrumDeExtractor() *CustomExtractor {
	return WwwSpektrumDeExtractor
}