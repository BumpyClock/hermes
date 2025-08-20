// ABOUTME: Custom extractor for www.al.com - Alabama news and local coverage
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.al.com/index.js WwwAlComExtractor

package custom

// GetWwwAlComExtractor returns the custom extractor for www.al.com
func GetWwwAlComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.al.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[name=\"title\"]", "value"},
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[name=\"article_author\"]", "value"},
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[name=\"article_date_original\"]", "value"},
			},
			// Note: timezone: 'EST' is handled by date cleaner in Go version
		},
		
		LeadImageURL: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[name=\"og:image\"]", "value"},
			},
		},
		
		Content: &ContentExtractor{
			FieldExtractor: &FieldExtractor{
				Selectors: []interface{}{
					".entry-content",
				},
			},
			
			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},
			
			Clean: []string{
				// No clean selectors in JavaScript version
			},
		},
	}
}