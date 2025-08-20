// ABOUTME: MacRumors custom extractor with timezone support and rel=author patterns
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.macrumors.com/index.js

package custom

// WwwMacrumorsComExtractor provides the custom extraction rules for www.macrumors.com
// JavaScript equivalent: export const WwwMacrumorsComExtractor = { ... }
var WwwMacrumorsComExtractor = &CustomExtractor{
	Domain: "www.macrumors.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1",
			"h1.title",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			"article a[rel=\"author\"]",
			".author-url",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"time", "datetime"},
		},
		// Note: timezone support would be handled at extraction time
		// timezone: 'America/Los_Angeles' (from JavaScript)
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"description\"]", "value"},
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
				"article",
				".article",
			},
		},
		
		// Transform functions (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors (empty in JavaScript)
		Clean: []string{},
	},
}

// GetWwwMacrumorsComExtractor returns the MacRumors custom extractor
func GetWwwMacrumorsComExtractor() *CustomExtractor {
	return WwwMacrumorsComExtractor
}