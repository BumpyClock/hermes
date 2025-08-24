// ABOUTME: RBB TODAY Japan tech news site custom extractor with article content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.rbbtoday.com/index.js

package custom

// WwwRbbtodayComExtractor provides the custom extraction rules for www.rbbtoday.com
// JavaScript equivalent: export const WwwRbbtodayComExtractor = { ... }
var WwwRbbtodayComExtractor = &CustomExtractor{
	Domain: "www.rbbtoday.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			".writer.writer-name",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"header time", "datetime"},
		},
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"description\"]", "value"},
			".arti-summary",
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
				".arti-content",
			},
		},
		
		// transforms: {} (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean promotional content
		Clean: []string{
			".arti-giga",
		},
	},
}

// GetWwwRbbtodayComExtractor returns the RBB TODAY Japan custom extractor
func GetWwwRbbtodayComExtractor() *CustomExtractor {
	return WwwRbbtodayComExtractor
}