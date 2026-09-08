// ABOUTME: NBC News custom extractor with article-hero-headline h1, byline-name span, and article-body__content
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.nbcnews.com/index.js WwwNbcnewsComExtractor

package custom

// GetNBCNewsExtractor returns the custom extractor for www.nbcnews.com.
func GetNBCNewsExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.nbcnews.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "div.article-hero-headline h1"},
				{Selector: "div.article-hed h1"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "div.article-inline-byline span.byline-name"},
				{Selector: "span.byline_author"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="article:published"]`, Attribute: "value"},
				{Selector: `.flag_article-wrapper time.timestamp_article[datetime]`, Attribute: "datetime"},
				{Selector: ".flag_article-wrapper time"},
			},
			// Note: timezone: 'America/New_York' is handled by date cleaner in Go version
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="og:image"]`, Attribute: "value"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{"div.article-body__content"},
				{"div.article-body"},
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
