// ABOUTME: Apple Newsroom extractor for press releases and updates.
// ABOUTME: Uses the visible page body and excludes duplicated download/contact content.

package custom

func GetAppleNewsroomExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain:           "www.apple.com",
		SupportedDomains: []string{"apple.com"},
		Title: &FieldExtractor{Selectors: []SelectorEntry{
			{Selector: "h1.hero-headline"},
			{Selector: "article h1"},
		}},
		LeadImageURL: &FieldExtractor{Selectors: []SelectorEntry{
			{Selector: `meta[name="og:image"]`, Attribute: "value"},
		}},
		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{"article .pagebody"},
				{".pagebody"},
			},
			Clean: []string{
				".docsanddownloads",
				".presscontacts",
				"aside.legal-info",
			},
		},
	}
}
