// ABOUTME: Custom extractor for www.al.com - Alabama news and local coverage
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.al.com/index.js WwwAlComExtractor

package custom

// GetWwwAlComExtractor returns the custom extractor for www.al.com.
func GetWwwAlComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.al.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"title\"]", Attribute: "value"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"article_author\"]", Attribute: "value"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"article_date_original\"]", Attribute: "value"},
			},
			// Note: timezone: 'EST' is handled by date cleaner in Go version
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{".entry-content"},
			},

			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},

			Clean: []string{
				// No clean selectors in JavaScript version
			},
		},
	}
}
