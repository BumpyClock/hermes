// ABOUTME: LittleThings custom extractor for lifestyle content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.littlethings.com/index.js

package custom

// LittleThingsCustomExtractor provides the custom extraction rules for www.littlethings.com
// JavaScript equivalent: export const LittleThingsExtractor = { ... }
var LittleThingsCustomExtractor = &CustomExtractor{
	Domain: "www.littlethings.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1[class*=\"PostHeader\"]",
			"h1.post-title",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			"div[class^=\"PostHeader__ScAuthorNameSection\"]",
			[]string{"meta[name=\"author\"]", "value"},
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				"section[class*=\"PostMainArticle\"]",
				".mainContentIntro",
				".content-wrapper",
			},
		},
		
		// No transforms in original JavaScript (empty array)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors - empty array in original JavaScript
		Clean: []string{},
	},
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
	
	// Next page URL and excerpt are null in original JavaScript
	NextPageURL: &FieldExtractor{
		Selectors: []interface{}{},
	},
	
	Excerpt: &FieldExtractor{
		Selectors: []interface{}{},
	},
	
	// No selectors in original JavaScript for these fields
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{},
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{},
	},
}

// GetLittleThingsExtractor returns the LittleThings custom extractor
func GetLittleThingsExtractor() *CustomExtractor {
	return LittleThingsCustomExtractor
}