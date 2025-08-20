// ABOUTME: NBC News custom extractor with article-hero-headline h1, byline-name span, and article-body__content
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.nbcnews.com/index.js WwwNbcnewsComExtractor

package custom

// GetNBCNewsExtractor returns the custom extractor for www.nbcnews.com
func GetNBCNewsExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.nbcnews.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				"div.article-hero-headline h1",
				"div.article-hed h1",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				"div.article-inline-byline span.byline-name",
				"span.byline_author",
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="article:published"]`, "value"},
				[]string{`.flag_article-wrapper time.timestamp_article[datetime]`, "datetime"},
				".flag_article-wrapper time",
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
					"div.article-body__content",
					"div.article-body",
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