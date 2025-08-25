// ABOUTME: CNN Money extractor for financial news with dek field and storytext content processing
// ABOUTME: JavaScript equivalent: src/extractors/custom/money.cnn.com/index.js MoneyCnnComExtractor

package custom

// GetMoneyCNNExtractor returns the custom extractor for money.cnn.com
func GetMoneyCNNExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "money.cnn.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				".article-title",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="date"]`, "value"},
				".byline a",
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="date"]`, "value"},
			},
			// JavaScript equivalent: timezone: 'GMT'
			// Note: Timezone handling would be implemented in date parsing logic
		},
		
		Dek: &FieldExtractor{
			Selectors: []interface{}{
				"#storytext h2",
			},
		},
		
		LeadImageURL: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="og:image"]`, "value"},
			},
		},
		
		Content: &ContentExtractor{
			FieldExtractor: &FieldExtractor{
				Selectors: []interface{}{
					"#storytext",
				},
			},
			
			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},
			
			Clean: []string{
				".inStoryHeading",
			},
		},
	}
}