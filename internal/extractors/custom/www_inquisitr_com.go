// ABOUTME: Custom extractor for www.inquisitr.com - Alternative news and opinion site
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.inquisitr.com/index.js WwwInquisitrComExtractor

package custom

// GetWwwInquisitrComExtractor returns the custom extractor for www.inquisitr.com.
func GetWwwInquisitrComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.inquisitr.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "h1.entry-title.story--header--title"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "div.story--header--author"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"datePublished\"]", Attribute: "value"},
			},
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{"article.story"},
				{".entry-content"},
			},

			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},

			Clean: []string{
				".post-category",
				".story--header--socials",
				".story--header--content",
			},
		},
	}
}
