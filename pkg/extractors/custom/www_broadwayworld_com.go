// ABOUTME: Custom extractor for www.broadwayworld.com - Theater and entertainment news
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.broadwayworld.com/index.js BroadwayWorldExtractor

package custom

// GetWwwBroadwayworldComExtractor returns the custom extractor for www.broadwayworld.com
func GetWwwBroadwayworldComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.broadwayworld.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				"h1[itemprop=headline]",
				"h1.article-title",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				"span[itemprop=author]",
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[itemprop=datePublished]", "value"},
			},
		},
		
		LeadImageURL: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[name=\"og:image\"]", "value"},
			},
		},
		
		Dek: &FieldExtractor{
			Selectors: []interface{}{
				// Empty array in JavaScript version
			},
		},
		
		Content: &ContentExtractor{
			FieldExtractor: &FieldExtractor{
				Selectors: []interface{}{
					"div[itemprop=articlebody]",
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