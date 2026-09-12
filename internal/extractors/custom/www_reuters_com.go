// ABOUTME: Reuters custom extractor with article-headline and ArticleBodyWrapper selectors
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.reuters.com/index.js WwwReutersComExtractor

package custom

// GetReutersExtractor returns the custom extractor for www.reuters.com.
func GetReutersExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.reuters.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `h1[data-testid="Heading"]`},
				{Selector: `h1[class*="ArticleHeader-headline-"]`},
				{Selector: "h1.article-headline"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `[data-testid="AuthorName"]`},
				{Selector: `meta[name="article:author"]`, Attribute: "value"},
				{Selector: `meta[name="og:article:author"]`, Attribute: "value"},
				{Selector: ".author"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="article:published_time"]`, Attribute: "value"},
				{Selector: `meta[name="og:article:published_time"]`, Attribute: "value"},
			},
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="og:image"]`, Attribute: "value"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{`[data-testid="ArticleBody"]`},
				{"div.ArticleBodyWrapper"},
				{"#article-text"},
			},

			Transforms: map[string]TransformFunction{
				".article-subtitle": &StringTransform{
					TargetTag: "h4",
				},
			},

			Clean: []string{
				`[data-testid="promo-box"]`,
				`[data-testid="ContextWidget"]`,
				`div[class^="ArticleBody-byline-container-"]`,
				"#article-byline .author",
			},
		},
	}
}
