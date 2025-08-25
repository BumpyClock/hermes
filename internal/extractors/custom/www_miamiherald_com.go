// ABOUTME: Miami Herald custom extractor with title h1, published-date p, and dateline-storybody content
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.miamiherald.com/index.js WwwMiamiheraldComExtractor

package custom

// GetMiamiHeraldExtractor returns the custom extractor for www.miamiherald.com
func GetMiamiHeraldExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.miamiherald.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				"h1.title",
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				"p.published-date",
			},
			// Note: timezone: 'America/New_York' is handled by date cleaner in Go version
		},
		
		LeadImageURL: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="og:image"]`, "value"},
			},
		},
		
		Content: &ContentExtractor{
			FieldExtractor: &FieldExtractor{
				Selectors: []interface{}{
					"div.dateline-storybody",
				},
			},
			
			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},
			
			Clean: []string{
				// No clean selectors in JavaScript version
			},
		},
		
		// Note: Author field is not present in the JavaScript version
	}
}