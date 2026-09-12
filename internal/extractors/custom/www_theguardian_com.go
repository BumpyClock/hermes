// ABOUTME: The Guardian custom extractor with content headline, address byline, and standfirst dek
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.theguardian.com/index.js WwwTheguardianComExtractor

package custom

// GetTheGuardianExtractor returns the custom extractor for www.theguardian.com.
func GetTheGuardianExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.theguardian.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "h1"},
				{Selector: ".content__headline"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `address[data-link-name="byline"]`},
				{Selector: "p.byline"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="article:published_time"]`, Attribute: "value"},
			},
		},

		Dek: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `div[data-gu-name="standfirst"]`},
				{Selector: ".content__standfirst"},
			},
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="og:image"]`, Attribute: "value"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{"#maincontent"},
				{".content__article-body"},
			},

			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},

			Clean: []string{
				"aside",
				`[data-gu-name="after-content"]`,
				`[data-component="related-content"]`,
				".hide-on-mobile",
				".inline-icon",
			},
		},
	}
}
