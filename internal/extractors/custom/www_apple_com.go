// ABOUTME: Apple Newsroom extractor for press releases and updates.
// ABOUTME: Uses the visible page body and excludes duplicated download/contact content.

package custom

func GetAppleNewsroomExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain:           "www.apple.com",
		SupportedDomains: []string{"apple.com"},
		Title: &FieldExtractor{Selectors: []interface{}{
			"h1.hero-headline",
			"article h1",
		}},
		LeadImageURL: &FieldExtractor{Selectors: []interface{}{
			[]string{`meta[name="og:image"]`, "value"},
		}},
		Content: &ContentExtractor{
			FieldExtractor: &FieldExtractor{Selectors: []interface{}{
				"article .pagebody",
				".pagebody",
			}},
			Clean: []string{
				".docsanddownloads",
				".presscontacts",
				"aside.legal-info",
			},
		},
	}
}
