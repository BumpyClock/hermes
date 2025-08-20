// ABOUTME: GetNews Japan news site custom extractor with article body content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/getnews.jp/index.js

package custom

// GetnewsJpExtractor provides the custom extraction rules for getnews.jp
// JavaScript equivalent: export const GetnewsJpExtractor = { ... }
var GetnewsJpExtractor = &CustomExtractor{
	Domain: "getnews.jp",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"article h1",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:author\"]", "value"},
			"span.prof",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:published_time\"]", "value"},
			[]string{"ul.cattag-top time", "datetime"},
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
				"div.post-bodycopy",
			},
		},
		
		// transforms: {} (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// clean: [] (empty in JavaScript)
		Clean: []string{},
	},
}

// GetGetnewsJpExtractor returns the GetNews Japan custom extractor
func GetGetnewsJpExtractor() *CustomExtractor {
	return GetnewsJpExtractor
}