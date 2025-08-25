// ABOUTME: BuzzAP Japan news site custom extractor with entry content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/buzzap.jp/index.js

package custom

// BuzzapJpExtractor provides the custom extraction rules for buzzap.jp
// JavaScript equivalent: export const BuzzapJpExtractor = { ... }
var BuzzapJpExtractor = &CustomExtractor{
	Domain: "buzzap.jp",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1.entry-title",
		},
	},
	
	Author: nil,
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"time.entry-date", "datetime"},
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
				"div.ctiframe",
			},
		},
		
		// defaultCleaner: false in JavaScript
		DefaultCleaner: false,
		
		// transforms: {} (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// clean: [] (empty in JavaScript)
		Clean: []string{},
	},
}

// GetBuzzapJpExtractor returns the BuzzAP Japan custom extractor
func GetBuzzapJpExtractor() *CustomExtractor {
	return BuzzapJpExtractor
}