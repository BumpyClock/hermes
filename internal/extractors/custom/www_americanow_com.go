// ABOUTME: Custom extractor for www.americanow.com - American news and current events
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.americanow.com/index.js WwwAmericanowComExtractor

package custom

// GetWwwAmericanowComExtractor returns the custom extractor for www.americanow.com
func GetWwwAmericanowComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.americanow.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				".title",
				[]string{"meta[name=\"title\"]", "value"},
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				".byline",
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[name=\"publish_date\"]", "value"},
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
					// Multi-match selector: first try complex selector, then fallback
					[]string{".article-content", ".image", ".body"},
					".body",
				},
			},
			
			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},
			
			Clean: []string{
				".article-video-wrapper",
				".show-for-small-only",
			},
		},
	}
}