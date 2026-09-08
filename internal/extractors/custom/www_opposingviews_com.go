// ABOUTME: Custom extractor for www.opposingviews.com - Political news and opinion site
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.opposingviews.com/index.js WwwOpposingviewsComExtractor

package custom

// GetWwwOpposingviewsComExtractor returns the custom extractor for www.opposingviews.com.
func GetWwwOpposingviewsComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.opposingviews.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "h1.m-detail-header--title"},
				{Selector: "h1.title"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"author\"]", Attribute: "value"},
				{Selector: "div.date span span a"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"published\"]", Attribute: "value"},
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
				{".m-detail--body"},
				{".article-content"},
			},

			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},

			Clean: []string{
				".show-for-small-only",
			},
		},
	}
}
