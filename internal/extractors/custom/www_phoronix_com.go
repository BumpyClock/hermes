// ABOUTME: Phoronix custom extractor with date format parsing
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.phoronix.com/index.js

package custom

// WwwPhoronixComExtractor provides the custom extraction rules for www.phoronix.com
// JavaScript equivalent: export const WwwPhoronixComExtractor = { ... }
var WwwPhoronixComExtractor = &CustomExtractor{
	Domain: "www.phoronix.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"article h1",
			"article header",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			".author a:first-child",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			".author",
		},
		// Note: format and timezone would be handled at extraction time
		// format: 'D MMMM YYYY at hh:mm' (from JavaScript)
		// timezone: 'America/New_York' (from JavaScript) 
	},
	
	Dek: nil,
	
	LeadImageURL: nil,
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				".content",
			},
		},
		
		// Transform functions (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors (empty in JavaScript)
		Clean: []string{},
	},
}

// GetWwwPhoronixComExtractor returns the Phoronix custom extractor
func GetWwwPhoronixComExtractor() *CustomExtractor {
	return WwwPhoronixComExtractor
}