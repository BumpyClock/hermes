// ABOUTME: ZDNet Japan custom extractor with cXenseParse:author meta pattern
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/japan.zdnet.com/index.js

package custom

// JapanZdnetComExtractor provides the custom extraction rules for japan.zdnet.com
// JavaScript equivalent: export const JapanZdnetComExtractor = { ... }
var JapanZdnetComExtractor = &CustomExtractor{
	Domain: "japan.zdnet.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"cXenseParse:author\"]", "value"},
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:published_time\"]", "value"},
		},
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

// GetJapanZdnetComExtractor returns the ZDNet Japan custom extractor
func GetJapanZdnetComExtractor() *CustomExtractor {
	return JapanZdnetComExtractor
}