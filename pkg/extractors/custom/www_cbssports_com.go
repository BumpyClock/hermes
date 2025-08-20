// ABOUTME: CBS Sports custom extractor with UTC timezone and article content patterns 
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.cbssports.com/index.js

package custom

// WwwCbssportsComExtractor provides the custom extraction rules for www.cbssports.com
// JavaScript equivalent: export const WwwCbssportsComExtractor = { ... }
var WwwCbssportsComExtractor = &CustomExtractor{
	Domain: "www.cbssports.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			".Article-headline",
			".article-headline",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			".ArticleAuthor-nameText",
			".author-name",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[itemprop=\"datePublished\"]", "value"},
		},
		Timezone: "UTC",
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			".Article-subline",
			".article-subline",
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
				".article",
			},
		},
		
		// Transform functions (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors (empty in JavaScript)
		Clean: []string{},
	},
}

// GetWwwCbssportsComExtractor returns the CBS Sports custom extractor
func GetWwwCbssportsComExtractor() *CustomExtractor {
	return WwwCbssportsComExtractor
}