// ABOUTME: Yomiuri Shimbun (major Japanese newspaper) custom extractor with article content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.yomiuri.co.jp/index.js

package custom

// WwwYomiuriCoJpExtractor provides the custom extraction rules for www.yomiuri.co.jp
// JavaScript equivalent: export const WwwYomiuriCoJpExtractor = { ... }
var WwwYomiuriCoJpExtractor = &CustomExtractor{
	Domain: "www.yomiuri.co.jp",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1.title-article.c-article-title",
		},
	},
	
	Author: nil,
	
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
				"div.p-main-contents",
			},
		},
		
		// transforms: {} (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// clean: [] (empty in JavaScript)
		Clean: []string{},
	},
}

// GetWwwYomiuriCoJpExtractor returns the Yomiuri Shimbun custom extractor
func GetWwwYomiuriCoJpExtractor() *CustomExtractor {
	return WwwYomiuriCoJpExtractor
}