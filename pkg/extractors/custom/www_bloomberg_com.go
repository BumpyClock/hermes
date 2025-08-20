// ABOUTME: Bloomberg custom extractor with multiple template support (normal, graphics, news) and parsely meta
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.bloomberg.com/index.js WwwBloombergComExtractor

package custom

// GetBloombergExtractor returns the custom extractor for www.bloomberg.com
func GetBloombergExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.bloomberg.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				// normal articles
				".lede-headline",
				
				// /graphics/ template
				"h1.article-title",
				
				// /news/ template
				`h1[class^="headline"]`,
				"h1.lede-text-only__hed",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="parsely-author"]`, "value"},
				".byline-details__link",
				
				// /graphics/ template
				".bydek",
				
				// /news/ template
				".author",
				`p[class*="author"]`,
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"time.published-at", "datetime"},
				[]string{"time[datetime]", "datetime"},
				[]string{`meta[name="date"]`, "value"},
				[]string{`meta[name="parsely-pub-date"]`, "value"},
				[]string{`meta[name="parsely-pub-date"]`, "content"},
			},
		},
		
		Dek: &FieldExtractor{
			Selectors: []interface{}{},
		},
		
		LeadImageURL: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="og:image"]`, "value"},
				[]string{`meta[name="og:image"]`, "content"},
			},
		},
		
		Content: &ContentExtractor{
			FieldExtractor: &FieldExtractor{
				Selectors: []interface{}{
					".article-body__content",
					".body-content",
					
					// /graphics/ template
					"section.copy-block",
					
					// /news/ template
					".body-copy",
				},
			},
			
			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},
			
			Clean: []string{
				".inline-newsletter",
				".page-ad",
			},
		},
	}
}