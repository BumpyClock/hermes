// ABOUTME: TechCrunch extractor scoped to the WordPress article body.
// ABOUTME: Excludes recirculation and promotional blocks outside entry content.

package custom

func GetTechCrunchExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain:           "techcrunch.com",
		SupportedDomains: []string{"www.techcrunch.com"},
		Title:            &FieldExtractor{Selectors: []SelectorEntry{{Selector: "article h1"}, {Selector: "h1"}}},
		Author: &FieldExtractor{Selectors: []SelectorEntry{
			{Selector: `meta[name="author"]`, Attribute: "value"},
		}},
		DatePublished: &FieldExtractor{Selectors: []SelectorEntry{
			{Selector: `meta[name="article:published_time"]`, Attribute: "value"},
		}},
		LeadImageURL: &FieldExtractor{Selectors: []SelectorEntry{
			{Selector: `meta[name="og:image"]`, Attribute: "value"},
		}},
		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{".entry-content.wp-block-post-content"},
				{".entry-content"},
			},
			Clean: []string{
				`.wp-block-techcrunch-newsletter`,
				`[class*="newsletter"]`,
				`[class*="event-promo"]`,
			},
		},
	}
}
