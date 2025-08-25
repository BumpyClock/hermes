// ABOUTME: Custom extractor for www.aol.com - AOL news and entertainment articles
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.aol.com/index.js WwwAolComExtractor

package custom

// GetWwwAolComExtractor returns the custom extractor for www.aol.com
func GetWwwAolComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.aol.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				"h1.p-article__title",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[name=\"author\"]", "value"},
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				".p-article__byline__date",
			},
			// Note: timezone: 'America/New_York' is handled by date cleaner in Go version
		},
		
		LeadImageURL: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[name=\"og:image\"]", "value"},
			},
		},
		
		Content: &ContentExtractor{
			FieldExtractor: &FieldExtractor{
				Selectors: []interface{}{
					".article-content",
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