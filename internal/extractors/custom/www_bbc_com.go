// ABOUTME: BBC News extractor scoped to the semantic article boundary.
// ABOUTME: Avoids generic main-content scoring selecting recommendation cards.

package custom

func GetBBCExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain:           "www.bbc.com",
		SupportedDomains: []string{"bbc.com"},
		Title: &FieldExtractor{Selectors: []SelectorEntry{
			{Selector: "article h1"},
			{Selector: `[data-component="headline-block"] h1`},
		}},
		Author: &FieldExtractor{Selectors: []SelectorEntry{
			{Selector: `[data-testid="byline-contributors"]`},
		}},
		DatePublished: &FieldExtractor{Selectors: []SelectorEntry{
			{Selector: "article time", Attribute: "datetime"},
		}},
		LeadImageURL: &FieldExtractor{Selectors: []SelectorEntry{
			{Selector: `meta[name="og:image"]`, Attribute: "value"},
		}},
		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{{"article"}},
			Clean: []string{
				"h1",
				`[data-testid="byline"]`,
				`[data-testid="share-tools"]`,
				`[data-component="recommendations"]`,
			},
		},
	}
}
