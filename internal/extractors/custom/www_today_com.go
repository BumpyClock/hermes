// ABOUTME: Custom extractor for www.today.com - NBC Today Show news and lifestyle
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.today.com/index.js WwwTodayComExtractor

package custom

// GetWwwTodayComExtractor returns the custom extractor for www.today.com
func GetWwwTodayComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.today.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				"h1.article-hero-headline__htag",
				"h1.entry-headline",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				"span.byline-name",
				[]string{"meta[name=\"author\"]", "value"},
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				"time[datetime]",
				[]string{"meta[name=\"DC.date.issued\"]", "value"},
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
					"div.article-body__content",
					".entry-container",
				},
			},
			
			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},
			
			Clean: []string{
				".label-comment",
			},
		},
	}
}