// ABOUTME: US Magazine custom extractor for celebrity magazine content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.usmagazine.com/index.js

package custom

// USMagazineCustomExtractor provides the custom extraction rules for www.usmagazine.com
// JavaScript equivalent: export const WwwUsmagazineComExtractor = { ... }
var USMagazineCustomExtractor = &CustomExtractor{
	Domain: "www.usmagazine.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"header h1",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			"a.author",
			"a.article-byline.tracked-offpage",
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
				"div.article-content",
			},
		},
		
		// No transforms in original JavaScript (empty object)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors - remove unwanted elements
		Clean: []string{
			".module-related",
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

// GetUSMagazineExtractor returns the US Magazine custom extractor
func GetUSMagazineExtractor() *CustomExtractor {
	return USMagazineCustomExtractor
}