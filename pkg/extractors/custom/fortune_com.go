// ABOUTME: Fortune magazine business content extractor with financial news and executive interviews
// ABOUTME: JavaScript equivalent: src/extractors/custom/fortune.com/index.js FortuneComExtractor

package custom

// GetFortuneComExtractor returns the custom extractor for fortune.com
func GetFortuneComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "fortune.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				"h1",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="author"]`, "value"},
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				".MblGHNMJ",
			},
			// JavaScript equivalent: timezone: 'UTC'
			// Note: Timezone handling would be implemented in date parsing logic
		},
		
		LeadImageURL: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="og:image"]`, "value"},
			},
		},
		
		Content: &ContentExtractor{
			FieldExtractor: &FieldExtractor{
				Selectors: []interface{}{
					[]string{"picture", "article.row"}, // Multi-match selector: [picture, article.row]
					"article.row",
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