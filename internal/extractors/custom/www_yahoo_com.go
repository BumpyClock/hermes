// ABOUTME: Custom extractor for www.yahoo.com - Yahoo News and content portal
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.yahoo.com/index.js YahooExtractor

package custom

// GetWwwYahooComExtractor returns the custom extractor for www.yahoo.com
func GetWwwYahooComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.yahoo.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				"header.canvas-header",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				"span.provider-name",
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"time.date[datetime]", "datetime"},
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
					".content-canvas",
				},
			},
			
			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},
			
			Clean: []string{
				".figure-caption",
			},
		},
	}
}