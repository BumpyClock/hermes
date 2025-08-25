// ABOUTME: 247Sports custom extractor with college sports content patterns and data-published date handling
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/247sports.com/index.js

package custom

// TwofortysevensportsComExtractor provides the custom extraction rules for 247sports.com
// JavaScript equivalent: export const twofortysevensportsComExtractor = { ... }
var TwofortysevensportsComExtractor = &CustomExtractor{
	Domain: "247sports.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"title",
			"article header h1",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			".article-cnt__author",
			".author",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"time[data-published]", "data-published"},
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
				".article-body",
				"section.body.article",
			},
		},
		
		// Transform functions (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors (empty in JavaScript)
		Clean: []string{},
	},
}

// GetTwofortysevensportsComExtractor returns the 247Sports custom extractor
func GetTwofortysevensportsComExtractor() *CustomExtractor {
	return TwofortysevensportsComExtractor
}