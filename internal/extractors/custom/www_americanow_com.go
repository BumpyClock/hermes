// ABOUTME: Custom extractor for www.americanow.com - American news and current events
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.americanow.com/index.js WwwAmericanowComExtractor

package custom

// GetWwwAmericanowComExtractor returns the custom extractor for www.americanow.com.
func GetWwwAmericanowComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.americanow.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: ".title"},
				{Selector: "meta[name=\"title\"]", Attribute: "value"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: ".byline"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"publish_date\"]", Attribute: "value"},
			},
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				// Multi-match selector: first try complex selector, then fallback
				{".article-content", ".image", ".body"},
				{".body"},
			},

			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},

			Clean: []string{
				".article-video-wrapper",
				".show-for-small-only",
			},
		},
	}
}
