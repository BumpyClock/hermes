// ABOUTME: Custom extractor for www.inquisitr.com - Alternative news and opinion site
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.inquisitr.com/index.js WwwInquisitrComExtractor

package custom

// GetWwwInquisitrComExtractor returns the custom extractor for www.inquisitr.com
func GetWwwInquisitrComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.inquisitr.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				"h1.entry-title.story--header--title",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				"div.story--header--author",
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[name=\"datePublished\"]", "value"},
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
					"article.story",
					".entry-content",
				},
			},
			
			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},
			
			Clean: []string{
				".post-category",
				".story--header--socials",
				".story--header--content",
			},
		},
	}
}