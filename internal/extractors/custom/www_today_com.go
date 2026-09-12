// ABOUTME: Custom extractor for www.today.com - NBC Today Show news and lifestyle
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.today.com/index.js WwwTodayComExtractor

package custom

// GetWwwTodayComExtractor returns the custom extractor for www.today.com.
func GetWwwTodayComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.today.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "h1.article-hero-headline__htag"},
				{Selector: "h1.entry-headline"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "span.byline-name"},
				{Selector: "meta[name=\"author\"]", Attribute: "value"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "time[datetime]"},
				{Selector: "meta[name=\"DC.date.issued\"]", Attribute: "value"},
			},
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{"div.article-body__content"},
				{".entry-container"},
			},

			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},

			Clean: []string{
				".label-comment",
			},
		},
	}
}
