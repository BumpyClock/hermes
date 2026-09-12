// ABOUTME: Custom extractor for www.yahoo.com - Yahoo News and content portal
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.yahoo.com/index.js YahooExtractor

package custom

// GetWwwYahooComExtractor returns the custom extractor for www.yahoo.com.
func GetWwwYahooComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.yahoo.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "header.canvas-header"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "span.provider-name"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "time.date[datetime]", Attribute: "datetime"},
			},
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{".content-canvas"},
			},

			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},

			Clean: []string{
				".figure-caption",
			},
		},
	}
}
