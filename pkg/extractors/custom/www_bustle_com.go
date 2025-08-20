// ABOUTME: Bustle custom extractor for fashion/lifestyle content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.bustle.com/index.js

package custom

// BustleCustomExtractor provides the custom extraction rules for www.bustle.com
// JavaScript equivalent: export const WwwBustleComExtractor = { ... }
var BustleCustomExtractor = &CustomExtractor{
	Domain: "www.bustle.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1",
			"h1.post-page__title",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			"a[href*=\"profile\"]",
			"div.content-meta__author",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"time", "datetime"},
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
				"article",
				".post-page__body",
			},
		},
		
		// No transforms in original JavaScript (empty object)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors - empty array in original JavaScript
		Clean: []string{},
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

// GetBustleExtractor returns the Bustle custom extractor
func GetBustleExtractor() *CustomExtractor {
	return BustleCustomExtractor
}