// ABOUTME: NY Daily News custom extractor with headline h1, article_byline span, and article with ra-related cleaning
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.nydailynews.com/index.js WwwNydailynewsComExtractor

package custom

// GetNYDailyNewsExtractor returns the custom extractor for www.nydailynews.com.
func GetNYDailyNewsExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.nydailynews.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "h1.headline"},
				{Selector: "h1#ra-headline"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: ".article_byline span"},
				{Selector: `meta[name="parsely-author"]`, Attribute: "value"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "time"},
				{Selector: `meta[name="sailthru.date"]`, Attribute: "value"},
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
				{"article#ra-body"},
			},

			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},

			Clean: []string{
				"dl#ra-tags",
				".ra-related",
				"a.ra-editor",
				"dl#ra-share-bottom",
			},
		},
	}
}
