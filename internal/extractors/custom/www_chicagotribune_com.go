// ABOUTME: Chicago Tribune custom extractor with og:title meta, article_byline span, and article content
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.chicagotribune.com/index.js WwwChicagotribuneComExtractor

package custom

// GetChicagoTribuneExtractor returns the custom extractor for www.chicagotribune.com.
func GetChicagoTribuneExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.chicagotribune.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="og:title"]`, Attribute: "value"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "div.article_byline span:first-of-type"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "time"},
			},
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="og:image"]`, Attribute: "value"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{"article"},
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
