// ABOUTME: BBC News extractor scoped to the semantic article boundary.
// ABOUTME: Avoids generic main-content scoring selecting recommendation cards.

package custom

func GetBBCExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain:           "www.bbc.com",
		SupportedDomains: []string{"bbc.com"},
		Title: &FieldExtractor{Selectors: []interface{}{
			"article h1",
			`[data-component="headline-block"] h1`,
		}},
		Author: &FieldExtractor{Selectors: []interface{}{
			`[data-testid="byline-contributors"]`,
		}},
		DatePublished: &FieldExtractor{Selectors: []interface{}{
			[]string{"article time", "datetime"},
		}},
		LeadImageURL: &FieldExtractor{Selectors: []interface{}{
			[]string{`meta[name="og:image"]`, "value"},
		}},
		Content: &ContentExtractor{
			FieldExtractor: &FieldExtractor{Selectors: []interface{}{"article"}},
			Clean: []string{
				"h1",
				`[data-testid="byline"]`,
				`[data-testid="share-tools"]`,
				`[data-component="recommendations"]`,
			},
		},
	}
}
