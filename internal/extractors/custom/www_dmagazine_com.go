// ABOUTME: Custom extractor for www.dmagazine.com - Dallas magazine lifestyle and local news
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.dmagazine.com/index.js WwwDmagazineComExtractor

package custom

// GetWwwDmagazineComExtractor returns the custom extractor for www.dmagazine.com.
func GetWwwDmagazineComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.dmagazine.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "h1.story__title"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: ".story__info .story__info__item:first-child"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: ".story__info"},
			},
			// Note: timezone: 'America/Chicago' and format: 'MMMM D, YYYY h:mm a' are handled by date cleaner in Go version
		},

		Dek: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: ".story__subhead"},
			},
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "article figure a:first-child", Attribute: "href"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{".story__content"},
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
