// ABOUTME: Fortune magazine business content extractor with financial news and executive interviews
// ABOUTME: JavaScript equivalent: src/extractors/custom/fortune.com/index.js FortuneComExtractor

package custom

// GetFortuneComExtractor returns the custom extractor for fortune.com.
func GetFortuneComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "fortune.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "h1"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="author"]`, Attribute: "value"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: ".MblGHNMJ"},
			},
			// JavaScript equivalent: timezone: 'UTC'
			// Note: Timezone handling would be implemented in date parsing logic
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="og:image"]`, Attribute: "value"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{"picture", "article.row"}, // Multi-match selector: [picture, article.row]
				{"article.row"},
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
