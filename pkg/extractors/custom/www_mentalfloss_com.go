// ABOUTME: Custom extractor for www.mentalfloss.com - General interest trivia and knowledge site
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.mentalfloss.com/index.js WwwMentalflossComExtractor

package custom

// GetWwwMentalflossComExtractor returns the custom extractor for www.mentalfloss.com
func GetWwwMentalflossComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.mentalfloss.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[name=\"og:title\"]", "value"},
				"h1.title",
				".title-group",
				".inner",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				"a[data-vars-label*=\"authors\"]",
				".field-name-field-enhanced-authors",
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[name=\"article:published_time\"]", "value"},
				".date-display-single",
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
					"article main",
					"div.field.field-name-body",
				},
			},
			
			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},
			
			Clean: []string{
				"small",
			},
		},
	}
}