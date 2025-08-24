// ABOUTME: Custom extractor for www.rawstory.com - Progressive news and politics site
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.rawstory.com/index.js WwwRawstoryComExtractor

package custom

// GetWwwRawstoryComExtractor returns the custom extractor for www.rawstory.com
func GetWwwRawstoryComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.rawstory.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[name=\"og:title\"]", "value"},
				".blog-title",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				"div.main-post-head .social-author__name",
				".blog-author a:first-of-type",
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[name=\"article:published_time\"]", "value"},
				".blog-author a:last-of-type",
			},
			// Note: timezone: 'EST' is handled by date cleaner in Go version
		},
		
		LeadImageURL: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[name=\"og:image\"]", "value"},
			},
		},
		
		Content: &ContentExtractor{
			FieldExtractor: &FieldExtractor{
				Selectors: []interface{}{
					".post-body",
					".blog-content",
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