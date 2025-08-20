// ABOUTME: PopSugar custom extractor for lifestyle with shopping content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.popsugar.com/index.js

package custom

// PopSugarCustomExtractor provides the custom extraction rules for www.popsugar.com
// JavaScript equivalent: export const WwwPopsugarComExtractor = { ... }
var PopSugarCustomExtractor = &CustomExtractor{
	Domain: "www.popsugar.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h2.post-title",
			"title-text",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:author\"]", "value"},
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:published_time\"]", "value"},
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
				"#content",
			},
		},
		
		// No transforms in original JavaScript (empty object)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors - remove unwanted elements
		Clean: []string{
			".share-copy-title",
			".post-tags",
			".reactions",
		},
	},
	
	// No selectors in original JavaScript for these fields
	Dek: &FieldExtractor{
		Selectors: []interface{}{},
	},
	
	NextPageURL: &FieldExtractor{
		Selectors: []interface{}{},
	},
	
	Excerpt: &FieldExtractor{
		Selectors: []interface{}{},
	},
}

// GetPopSugarExtractor returns the PopSugar custom extractor
func GetPopSugarExtractor() *CustomExtractor {
	return PopSugarCustomExtractor
}