// ABOUTME: CNBC financial news extractor with comprehensive market coverage and business analysis
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.cnbc.com/index.js WwwCnbcComExtractor

package custom

// GetCNBCExtractor returns the custom extractor for www.cnbc.com.
func GetCNBCExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.cnbc.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "h1.title"},
				{Selector: "h1.ArticleHeader-headline"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="author"]`, Attribute: "value"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="article:published_time"]`, Attribute: "value"},
			},
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="og:image"]`, Attribute: "value"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{"div#article_body.content"},
				{"div.story"},
				{"div.ArticleBody-articleBody"},
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
