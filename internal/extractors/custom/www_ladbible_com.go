// ABOUTME: Custom extractor for www.ladbible.com - UK entertainment and viral content site
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.ladbible.com/index.js WwwLadbibleComExtractor

package custom

// GetWwwLadbibleComExtractor returns the custom extractor for www.ladbible.com
func GetWwwLadbibleComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.ladbible.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				"h1",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				"[class*=Byline]",
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				"time",
			},
			// Note: timezone: 'Europe/London' is handled by date cleaner in Go version
		},
		
		LeadImageURL: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[name=\"og:image\"]", "value"},
			},
		},
		
		Content: &ContentExtractor{
			FieldExtractor: &FieldExtractor{
				Selectors: []interface{}{
					"[class*=ArticleContainer]",
				},
			},
			
			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},
			
			Clean: []string{
				"time",
				"source",
				"a[href^=\"https://www.ladbible.com/\"]",
				"picture",
				"[class*=StyledCardBlock]",
			},
		},
	}
}