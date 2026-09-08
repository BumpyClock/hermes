// ABOUTME: Miami Herald custom extractor with title h1, published-date p, and dateline-storybody content
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.miamiherald.com/index.js WwwMiamiheraldComExtractor

package custom

// GetMiamiHeraldExtractor returns the custom extractor for www.miamiherald.com.
func GetMiamiHeraldExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.miamiherald.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "h1.title"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "p.published-date"},
			},
			// Note: timezone: 'America/New_York' is handled by date cleaner in Go version
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="og:image"]`, Attribute: "value"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{"div.dateline-storybody"},
			},

			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},

			Clean: []string{
				// No clean selectors in JavaScript version
			},
		},

		// Note: Author field is not present in the JavaScript version
	}
}
