// ABOUTME: NY Daily News custom extractor with headline h1, article_byline span, and article with ra-related cleaning
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.nydailynews.com/index.js WwwNydailynewsComExtractor

package custom

// GetNYDailyNewsExtractor returns the custom extractor for www.nydailynews.com
func GetNYDailyNewsExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.nydailynews.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				"h1.headline",
				"h1#ra-headline",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				".article_byline span",
				[]string{`meta[name="parsely-author"]`, "value"},
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				"time",
				[]string{`meta[name="sailthru.date"]`, "value"},
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
					"article#ra-body",
				},
			},
			
			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},
			
			Clean: []string{
				"dl#ra-tags",
				".ra-related",
				"a.ra-editor",
				"dl#ra-share-bottom",
			},
		},
	}
}