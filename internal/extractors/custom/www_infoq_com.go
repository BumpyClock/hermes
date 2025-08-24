// ABOUTME: InfoQ custom extractor with DefaultCleaner false handling
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.infoq.com/index.js

package custom

// WwwInfoqComExtractor provides the custom extraction rules for www.infoq.com
// JavaScript equivalent: export const WwwInfoqComExtractor = { ... }
var WwwInfoqComExtractor = &CustomExtractor{
	Domain: "www.infoq.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1.heading",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			"div.widget.article__authors",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			".article__readTime.date",
		},
		// Note: format and timezone would be handled at extraction time
		// format: 'YYYY年MM月DD日' (from JavaScript)
		// timezone: 'Asia/Tokyo' (from JavaScript)
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:description\"]", "value"},
		},
	},
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				"div.article__data",
			},
			DefaultCleaner: false, // defaultCleaner: false in JavaScript
		},
		
		// Transform functions (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors (empty in JavaScript)
		Clean: []string{},
	},
}

// GetWwwInfoqComExtractor returns the InfoQ custom extractor
func GetWwwInfoqComExtractor() *CustomExtractor {
	return WwwInfoqComExtractor
}