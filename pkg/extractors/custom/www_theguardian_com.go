// ABOUTME: The Guardian custom extractor with content headline, address byline, and standfirst dek
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.theguardian.com/index.js WwwTheguardianComExtractor

package custom

// GetTheGuardianExtractor returns the custom extractor for www.theguardian.com
func GetTheGuardianExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.theguardian.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				"h1",
				".content__headline",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				`address[data-link-name="byline"]`,
				"p.byline",
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="article:published_time"]`, "value"},
			},
		},
		
		Dek: &FieldExtractor{
			Selectors: []interface{}{
				`div[data-gu-name="standfirst"]`,
				".content__standfirst",
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
					"#maincontent",
					".content__article-body",
				},
			},
			
			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},
			
			Clean: []string{
				".hide-on-mobile",
				".inline-icon",
			},
		},
	}
}