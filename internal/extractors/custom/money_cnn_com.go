// ABOUTME: CNN Money extractor for financial news with dek field and storytext content processing
// ABOUTME: JavaScript equivalent: src/extractors/custom/money.cnn.com/index.js MoneyCnnComExtractor

package custom

// GetMoneyCNNExtractor returns the custom extractor for money.cnn.com.
func GetMoneyCNNExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "money.cnn.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: ".article-title"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="date"]`, Attribute: "value"},
				{Selector: ".byline a"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="date"]`, Attribute: "value"},
			},
			// JavaScript equivalent: timezone: 'GMT'
			// Note: Timezone handling would be implemented in date parsing logic
		},

		Dek: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "#storytext h2"},
			},
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="og:image"]`, Attribute: "value"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{"#storytext"},
			},

			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},

			Clean: []string{
				".inStoryHeading",
			},
		},
	}
}
