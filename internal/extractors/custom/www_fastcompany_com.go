// ABOUTME: Custom extractor for www.fastcompany.com - Business innovation and technology magazine
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.fastcompany.com/index.js WwwFastcompanyComExtractor

package custom

// GetWwwFastcompanyComExtractor returns the custom extractor for www.fastcompany.com
func GetWwwFastcompanyComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.fastcompany.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				"h1",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[name=\"author\"]", "value"},
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[name=\"article:published_time\"]", "value"},
			},
		},
		
		Dek: &FieldExtractor{
			Selectors: []interface{}{
				".post__deck",
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
					".post__article",
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