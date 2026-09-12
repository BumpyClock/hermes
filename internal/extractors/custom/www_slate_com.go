// ABOUTME: Custom extractor for www.slate.com - Slate magazine news and opinion articles
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.slate.com/index.js WwwSlateComExtractor

package custom

// GetWwwSlateComExtractor returns the custom extractor for www.slate.com.
func GetWwwSlateComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.slate.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: ".hed"},
				{Selector: "h1"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "a[rel=author]"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: ".pub-date"},
			},
			// Note: timezone: 'America/New_York' is handled by date cleaner in Go version
		},

		Dek: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: ".dek"},
			},
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{".body"},
			},

			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},

			Clean: []string{
				".about-the-author",
				".pullquote",
				".newsletter-signup-component",
				".top-comment",
			},
		},
	}
}
