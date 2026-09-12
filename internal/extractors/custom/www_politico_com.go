// ABOUTME: Politico custom extractor with og:title meta, story-text content, and timezone support
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.politico.com/index.js PoliticoExtractor

package custom

// GetPoliticoExtractor returns the custom extractor for www.politico.com.
func GetPoliticoExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.politico.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="og:title"]`, Attribute: "value"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `div[itemprop="author"] meta[itemprop="name"]`, Attribute: "value"},
				{Selector: ".story-meta__authors .vcard"},
				{Selector: ".story-main-content .byline .vcard"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{".story-text"},
				{".story-main-content"},
				{".story-core"},
			},

			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version (transforms: [] is empty)
			},

			Clean: []string{
				"figcaption",
				".story-meta",
				".ad",
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `time[itemprop="datePublished"]`, Attribute: "datetime"},
				{Selector: `.story-meta__details time[datetime]`, Attribute: "datetime"},
				{Selector: `.story-main-content .timestamp time[datetime]`, Attribute: "datetime"},
			},
			// Note: timezone: 'America/New_York' is handled by date cleaner in Go version
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="og:image"]`, Attribute: "value"},
			},
		},

		Dek: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="og:description"]`, Attribute: "value"},
			},
		},
	}
}
