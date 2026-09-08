// ABOUTME: Custom extractor for www.westernjournalism.com - Conservative news and political commentary
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.westernjournalism.com/index.js WwwWesternjournalismComExtractor

package custom

// GetWwwWesternjournalismComExtractor returns the custom extractor for www.westernjournalism.com.
func GetWwwWesternjournalismComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.westernjournalism.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "title"},
				{Selector: "h1.entry-title"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"author\"]", Attribute: "value"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"DC.date.issued\"]", Attribute: "value"},
			},
		},

		Dek: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: ".subtitle"},
			},
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{"div.article-sharing.top + div"},
			},

			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},

			Clean: []string{
				".ad-notice-small",
			},
		},
	}
}
