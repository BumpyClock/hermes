// ABOUTME: CNBC financial news extractor with comprehensive market coverage and business analysis
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.cnbc.com/index.js WwwCnbcComExtractor

package custom

// GetCNBCExtractor returns the custom extractor for www.cnbc.com
func GetCNBCExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.cnbc.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				"h1.title",
				"h1.ArticleHeader-headline",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="author"]`, "value"},
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="article:published_time"]`, "value"},
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
					"div#article_body.content",
					"div.story",
					"div.ArticleBody-articleBody",
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