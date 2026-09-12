// ABOUTME: Bloomberg custom extractor with multiple template support (normal, graphics, news) and parsely meta
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.bloomberg.com/index.js WwwBloombergComExtractor

package custom

// GetBloombergExtractor returns the custom extractor for www.bloomberg.com.
func GetBloombergExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.bloomberg.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				// normal articles
				{Selector: ".lede-headline"},

				// /graphics/ template
				{Selector: "h1.article-title"},

				// /news/ template
				{Selector: `h1[class^="headline"]`},
				{Selector: "h1.lede-text-only__hed"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="parsely-author"]`, Attribute: "value"},
				{Selector: ".byline-details__link"},

				// /graphics/ template
				{Selector: ".bydek"},

				// /news/ template
				{Selector: ".author"},
				{Selector: `p[class*="author"]`},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "time.published-at", Attribute: "datetime"},
				{Selector: "time[datetime]", Attribute: "datetime"},
				{Selector: `meta[name="date"]`, Attribute: "value"},
				{Selector: `meta[name="parsely-pub-date"]`, Attribute: "value"},
				{Selector: `meta[name="parsely-pub-date"]`, Attribute: "content"},
			},
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="og:image"]`, Attribute: "value"},
				{Selector: `meta[name="og:image"]`, Attribute: "content"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{".article-body__content"},
				{".body-content"},

				// /graphics/ template
				{"section.copy-block"},

				// /news/ template
				{".body-copy"},
			},

			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},

			Clean: []string{
				".inline-newsletter",
				".page-ad",
			},
		},
	}
}
