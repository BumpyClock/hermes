// ABOUTME: Politico custom extractor with og:title meta, story-text content, and timezone support
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.politico.com/index.js PoliticoExtractor

package custom

// GetPoliticoExtractor returns the custom extractor for www.politico.com
func GetPoliticoExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.politico.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="og:title"]`, "value"},
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`div[itemprop="author"] meta[itemprop="name"]`, "value"},
				".story-meta__authors .vcard",
				".story-main-content .byline .vcard",
			},
		},
		
		Content: &ContentExtractor{
			FieldExtractor: &FieldExtractor{
				Selectors: []interface{}{
					".story-text",
					".story-main-content",
					".story-core",
				},
			},
			
			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version (transforms: [] is empty)
			},
			
			Clean: []string{
				"figcaption",
				".story-meta",
				".ad",
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`time[itemprop="datePublished"]`, "datetime"},
				[]string{`.story-meta__details time[datetime]`, "datetime"},
				[]string{`.story-main-content .timestamp time[datetime]`, "datetime"},
			},
			// Note: timezone: 'America/New_York' is handled by date cleaner in Go version
		},
		
		LeadImageURL: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="og:image"]`, "value"},
			},
		},
		
		Dek: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="og:description"]`, "value"},
			},
		},
	}
}