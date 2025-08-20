// ABOUTME: Chicago Tribune custom extractor with og:title meta, article_byline span, and article content
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.chicagotribune.com/index.js WwwChicagotribuneComExtractor

package custom

// GetChicagoTribuneExtractor returns the custom extractor for www.chicagotribune.com
func GetChicagoTribuneExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.chicagotribune.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="og:title"]`, "value"},
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				"div.article_byline span:first-of-type",
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				"time",
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
					"article",
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