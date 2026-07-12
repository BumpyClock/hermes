// ABOUTME: TechCrunch extractor scoped to the WordPress article body.
// ABOUTME: Excludes recirculation and promotional blocks outside entry content.

package custom

func GetTechCrunchExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain:           "techcrunch.com",
		SupportedDomains: []string{"www.techcrunch.com"},
		Title:            &FieldExtractor{Selectors: []interface{}{"article h1", "h1"}},
		Author: &FieldExtractor{Selectors: []interface{}{
			[]string{`meta[name="author"]`, "value"},
		}},
		DatePublished: &FieldExtractor{Selectors: []interface{}{
			[]string{`meta[name="article:published_time"]`, "value"},
		}},
		LeadImageURL: &FieldExtractor{Selectors: []interface{}{
			[]string{`meta[name="og:image"]`, "value"},
		}},
		Content: &ContentExtractor{
			FieldExtractor: &FieldExtractor{Selectors: []interface{}{
				".entry-content.wp-block-post-content",
				".entry-content",
			}},
			Clean: []string{
				`.wp-block-techcrunch-newsletter`,
				`[class*="newsletter"]`,
				`[class*="event-promo"]`,
			},
		},
	}
}
