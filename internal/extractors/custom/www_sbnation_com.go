// ABOUTME: SB Nation custom extractor with Vox Media content patterns 
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.sbnation.com/index.js

package custom

// WwwSbnationComExtractor provides the custom extraction rules for www.sbnation.com
// JavaScript equivalent: export const WwwSbnationComExtractor = { ... }
var WwwSbnationComExtractor = &CustomExtractor{
	Domain: "www.sbnation.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1.c-page-title",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"author\"]", "value"},
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:published_time\"]", "value"},
		},
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			"p.c-entry-summary.p-dek",
			"h2.c-entry-summary.p-dek",
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
				"div.c-entry-content",
			},
		},
		
		// Transform functions (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors (empty in JavaScript)
		Clean: []string{},
	},
}

// GetWwwSbnationComExtractor returns the SB Nation custom extractor
func GetWwwSbnationComExtractor() *CustomExtractor {
	return WwwSbnationComExtractor
}