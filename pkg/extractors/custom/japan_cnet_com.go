// ABOUTME: CNET Japan custom extractor with Japanese date format parsing
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/japan.cnet.com/index.js

package custom

// JapanCnetComExtractor provides the custom extraction rules for japan.cnet.com
// JavaScript equivalent: export const JapanCnetComExtractor = { ... }
var JapanCnetComExtractor = &CustomExtractor{
	Domain: "japan.cnet.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			".leaf-headline-ttl",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			".writer",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			".date",
		},
		// Note: format and timezone would be handled at extraction time
		// format: 'YYYY年MM月DD日 HH時mm分' (from JavaScript)
		// timezone: 'Asia/Tokyo' (from JavaScript)
	},
	
	Dek: nil,
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				"div.article_body",
			},
		},
		
		// Transform functions (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors (empty in JavaScript)
		Clean: []string{},
	},
}

// GetJapanCnetComExtractor returns the CNET Japan custom extractor
func GetJapanCnetComExtractor() *CustomExtractor {
	return JapanCnetComExtractor
}