// ABOUTME: Custom extractor for www.mentalfloss.com - General interest trivia and knowledge site
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.mentalfloss.com/index.js WwwMentalflossComExtractor

package custom

// GetWwwMentalflossComExtractor returns the custom extractor for www.mentalfloss.com.
func GetWwwMentalflossComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.mentalfloss.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"og:title\"]", Attribute: "value"},
				{Selector: "h1.title"},
				{Selector: ".title-group"},
				{Selector: ".inner"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "a[data-vars-label*=\"authors\"]"},
				{Selector: ".field-name-field-enhanced-authors"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"article:published_time\"]", Attribute: "value"},
				{Selector: ".date-display-single"},
			},
			// Note: timezone: 'America/New_York' is handled by date cleaner in Go version
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{"article main"},
				{"div.field.field-name-body"},
			},

			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},

			Clean: []string{
				"small",
			},
		},
	}
}
