// ABOUTME: Custom extractor for www.slate.com - Slate magazine news and opinion articles
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.slate.com/index.js WwwSlateComExtractor

package custom

// GetWwwSlateComExtractor returns the custom extractor for www.slate.com
func GetWwwSlateComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.slate.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				".hed",
				"h1",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				"a[rel=author]",
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				".pub-date",
			},
			// Note: timezone: 'America/New_York' is handled by date cleaner in Go version
		},
		
		Dek: &FieldExtractor{
			Selectors: []interface{}{
				".dek",
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
					".body",
				},
			},
			
			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},
			
			Clean: []string{
				".about-the-author",
				".pullquote",
				".newsletter-signup-component",
				".top-comment",
			},
		},
	}
}