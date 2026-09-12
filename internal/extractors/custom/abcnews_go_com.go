// ABOUTME: ABC News custom extractor with Article_main__body h1, ShareByline byline, and article content
// ABOUTME: JavaScript equivalent: src/extractors/custom/abcnews.go.com/index.js AbcnewsGoComExtractor

package custom

// GetABCNewsExtractor returns the custom extractor for abcnews.go.com.
func GetABCNewsExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "abcnews.go.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `div[class*="Article_main__body"] h1`},
				{Selector: ".article-header h1"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: ".ShareByline span:nth-child(2)"},
				{Selector: ".authors"},
			},
			// Note: clean: ['.author-overlay', '.by-text'] is handled differently in Go
			// The JavaScript version applies clean to author field specifically
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: ".ShareByline"},
				{Selector: ".timestamp"},
			},
			// Note: format: 'MMMM D, YYYY h:mm a' and timezone: 'America/New_York'
			// are handled by date cleaner in Go version
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="og:image"]`, Attribute: "value"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{"article"},
				{".article-copy"},
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
