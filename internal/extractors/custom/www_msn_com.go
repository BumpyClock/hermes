// ABOUTME: Custom extractor for www.msn.com - Microsoft Network news portal
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.msn.com/index.js MSNExtractor

package custom

// GetWwwMsnComExtractor returns the custom extractor for www.msn.com.
func GetWwwMsnComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.msn.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "h1"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "span.authorname-txt"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "span.time"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{"div.richtext"},
			},

			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},

			Clean: []string{
				"span.caption",
			},
		},
	}
}
