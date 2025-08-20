// ABOUTME: Custom extractor for www.westernjournalism.com - Conservative news and political commentary
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.westernjournalism.com/index.js WwwWesternjournalismComExtractor

package custom

// GetWwwWesternjournalismComExtractor returns the custom extractor for www.westernjournalism.com
func GetWwwWesternjournalismComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.westernjournalism.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				"title",
				"h1.entry-title",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[name=\"author\"]", "value"},
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[name=\"DC.date.issued\"]", "value"},
			},
		},
		
		Dek: &FieldExtractor{
			Selectors: []interface{}{
				".subtitle",
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
					"div.article-sharing.top + div",
				},
			},
			
			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},
			
			Clean: []string{
				".ad-notice-small",
			},
		},
	}
}