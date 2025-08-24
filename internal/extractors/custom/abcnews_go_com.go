// ABOUTME: ABC News custom extractor with Article_main__body h1, ShareByline byline, and article content
// ABOUTME: JavaScript equivalent: src/extractors/custom/abcnews.go.com/index.js AbcnewsGoComExtractor

package custom

// GetABCNewsExtractor returns the custom extractor for abcnews.go.com
func GetABCNewsExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "abcnews.go.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				`div[class*="Article_main__body"] h1`,
				".article-header h1",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				".ShareByline span:nth-child(2)",
				".authors",
			},
			// Note: clean: ['.author-overlay', '.by-text'] is handled differently in Go
			// The JavaScript version applies clean to author field specifically
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				".ShareByline",
				".timestamp",
			},
			// Note: format: 'MMMM D, YYYY h:mm a' and timezone: 'America/New_York' 
			// are handled by date cleaner in Go version
		},
		
		LeadImageURL: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="og:image"]`, "value"},
			},
		},
		
		Content: &ContentExtractor{
			FieldExtractor: &FieldExtractor{
				Selectors: []interface{}{
					"article",
					".article-copy",
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