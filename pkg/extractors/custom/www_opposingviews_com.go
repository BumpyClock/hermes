// ABOUTME: Custom extractor for www.opposingviews.com - Political news and opinion site
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.opposingviews.com/index.js WwwOpposingviewsComExtractor

package custom

// GetWwwOpposingviewsComExtractor returns the custom extractor for www.opposingviews.com
func GetWwwOpposingviewsComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.opposingviews.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				"h1.m-detail-header--title",
				"h1.title",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[name=\"author\"]", "value"},
				"div.date span span a",
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[name=\"published\"]", "value"},
				[]string{"meta[name=\"publish_date\"]", "value"},
			},
		},
		
		Dek: &FieldExtractor{
			Selectors: []interface{}{
				// Empty array in JavaScript version
			},
		},
		
		LeadImageURL: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[name=\"og:image\"]", "value"},
			},
		},
		
		Content: &ContentExtractor{
			FieldExtractor: &FieldExtractor{
				Selectors: []interface{}{
					".m-detail--body",
					".article-content",
				},
			},
			
			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},
			
			Clean: []string{
				".show-for-small-only",
			},
		},
	}
}