// ABOUTME: Custom extractor for www.msn.com - Microsoft Network news portal
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.msn.com/index.js MSNExtractor

package custom

// GetWwwMsnComExtractor returns the custom extractor for www.msn.com
func GetWwwMsnComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.msn.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				"h1",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				"span.authorname-txt",
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				"span.time",
			},
		},
		
		LeadImageURL: &FieldExtractor{
			Selectors: []interface{}{
				// Empty selectors array in JavaScript version
			},
		},
		
		Content: &ContentExtractor{
			FieldExtractor: &FieldExtractor{
				Selectors: []interface{}{
					"div.richtext",
				},
			},
			
			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},
			
			Clean: []string{
				"span.caption",
			},
		},
	}
}