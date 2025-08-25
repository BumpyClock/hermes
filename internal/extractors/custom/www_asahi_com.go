// ABOUTME: Asahi Shimbun (major Japanese newspaper) custom extractor with main content selection
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.asahi.com/index.js

package custom

// WwwAsahiComExtractor provides the custom extraction rules for www.asahi.com
// JavaScript equivalent: export const WwwAsahiComExtractor = { ... }
var WwwAsahiComExtractor = &CustomExtractor{
	Domain: "www.asahi.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"main h1",
			".ArticleTitle h1",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:author\"]", "value"},
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"pubdate\"]", "value"},
		},
	},
	
	Dek: nil,
	
	Excerpt: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:description\"]", "value"},
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
				"main",
			},
		},
		
		// defaultCleaner: false in JavaScript
		DefaultCleaner: false,
		
		// transforms: {} (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors to remove ads and unwanted content
		Clean: []string{
			"div.AdMod",
			"div.LoginSelectArea",
			"time",
			"div.notPrint",
		},
	},
}

// GetWwwAsahiComExtractor returns the Asahi Shimbun custom extractor
func GetWwwAsahiComExtractor() *CustomExtractor {
	return WwwAsahiComExtractor
}