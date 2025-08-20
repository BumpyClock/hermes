// ABOUTME: Reuters custom extractor with article-headline and ArticleBodyWrapper selectors
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.reuters.com/index.js WwwReutersComExtractor

package custom

// GetReutersExtractor returns the custom extractor for www.reuters.com
func GetReutersExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.reuters.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				`h1[class*="ArticleHeader-headline-"]`,
				"h1.article-headline",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="og:article:author"]`, "value"},
				".author",
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="og:article:published_time"]`, "value"},
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
					"div.ArticleBodyWrapper",
					"#article-text",
				},
			},
			
			Transforms: map[string]TransformFunction{
				".article-subtitle": &StringTransform{
					TargetTag: "h4",
				},
			},
			
			Clean: []string{
				`div[class^="ArticleBody-byline-container-"]`,
				"#article-byline .author",
			},
		},
	}
}