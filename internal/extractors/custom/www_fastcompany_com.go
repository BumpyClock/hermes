// ABOUTME: Custom extractor for www.fastcompany.com - Business innovation and technology magazine
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.fastcompany.com/index.js WwwFastcompanyComExtractor

package custom

// GetWwwFastcompanyComExtractor returns the custom extractor for www.fastcompany.com.
func GetWwwFastcompanyComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.fastcompany.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "h1"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"author\"]", Attribute: "value"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"article:published_time\"]", Attribute: "value"},
			},
		},

		Dek: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: ".post__deck"},
			},
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{".post__article"},
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
