// ABOUTME: TMZ custom extractor for celebrity content and photo galleries
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.tmz.com/index.js

package custom

// TMZCustomExtractor provides the custom extraction rules for www.tmz.com
// JavaScript equivalent: export const WwwTmzComExtractor = { ... }
var TMZCustomExtractor = &CustomExtractor{
	Domain: "www.tmz.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			".post-title-breadcrumb",
			"h1",
			".headline",
		},
	},
	
	// Author is a static string in original JavaScript
	Author: &FieldExtractor{
		Selectors: []interface{}{"TMZ STAFF"},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			".article__published-at",
			".article-posted-date",
		},
	},
	
	Dek: &FieldExtractor{
		// Empty selectors in original JavaScript
		Selectors: []interface{}{},
	},
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				".article__blocks",
				".article-content",
				".all-post-body",
			},
		},
		
		// No transforms in original JavaScript (empty object)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors - remove unwanted elements
		Clean: []string{
			".lightbox-link",
		},
	},
	
	// No selectors in original JavaScript for these fields
	NextPageURL: &FieldExtractor{
		Selectors: []interface{}{},
	},
	
	Excerpt: &FieldExtractor{
		Selectors: []interface{}{},
	},
}

// GetTMZExtractor returns the TMZ custom extractor
func GetTMZExtractor() *CustomExtractor {
	return TMZCustomExtractor
}