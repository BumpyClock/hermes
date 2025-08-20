// ABOUTME: Android Central custom extractor with simple meta selector patterns
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.androidcentral.com/index.js

package custom

// WwwAndroidcentralComExtractor provides the custom extraction rules for www.androidcentral.com
// JavaScript equivalent: export const WwwAndroidcentralComExtractor = { ... }
var WwwAndroidcentralComExtractor = &CustomExtractor{
	Domain: "www.androidcentral.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1",
			"h1.main-title",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"parsely-author\"]", "value"},
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:published_time\"]", "value"},
		},
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
				"#article-body",
			},
		},
		
		// Transform functions (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors
		Clean: []string{
			".intro",
			"blockquote",
		},
	},
}

// GetWwwAndroidcentralComExtractor returns the Android Central custom extractor
func GetWwwAndroidcentralComExtractor() *CustomExtractor {
	return WwwAndroidcentralComExtractor
}