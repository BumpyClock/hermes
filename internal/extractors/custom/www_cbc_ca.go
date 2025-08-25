// ABOUTME: CBC (Canadian broadcaster) custom extractor with simple story-based extraction
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.cbc.ca/index.js

package custom

// WwwCbcCaExtractor provides the custom extraction rules for www.cbc.ca
// JavaScript equivalent: export const WwwCbcCaExtractor = { ... }
var WwwCbcCaExtractor = &CustomExtractor{
	Domain: "www.cbc.ca",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			".authorText",
			".bylineDetails",
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				".story",
			},
		},
		
		// Transform functions (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors (empty in JavaScript)
		Clean: []string{},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{".timeStamp[datetime]", "datetime"},
		},
	},
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			".deck",
		},
	},
	
	NextPageURL: nil,
	
	Excerpt: nil,
}

// GetWwwCbcCaExtractor returns the CBC custom extractor
func GetWwwCbcCaExtractor() *CustomExtractor {
	return WwwCbcCaExtractor
}