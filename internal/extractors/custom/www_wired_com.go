// ABOUTME: Wired.com custom extractor with article content patterns 
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.wired.com/index.js

package custom

// WwwWiredComExtractor provides the custom extraction rules for www.wired.com
// JavaScript equivalent: export const WiredExtractor = { ... }
var WwwWiredComExtractor = &CustomExtractor{
	Domain: "www.wired.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1[data-testId=\"ContentHeaderHed\"]",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:author\"]", "value"},
			"a[rel=\"author\"]",
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				"article.article.main-content",
				"article.content",
			},
		},
		
		// Transform functions (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors - remove unwanted elements
		Clean: []string{
			".visually-hidden",
			"figcaption img.photo",
			".alert-message",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:published_time\"]", "value"},
		},
	},
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			// Empty in JavaScript
		},
	},
	
	NextPageURL: nil,
	
	Excerpt: nil,
}

// GetWwwWiredComExtractor returns the Wired.com custom extractor
func GetWwwWiredComExtractor() *CustomExtractor {
	return WwwWiredComExtractor
}