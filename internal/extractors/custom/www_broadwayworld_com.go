// ABOUTME: Custom extractor for www.broadwayworld.com - Theater and entertainment news
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.broadwayworld.com/index.js BroadwayWorldExtractor

package custom

// GetWwwBroadwayworldComExtractor returns the custom extractor for www.broadwayworld.com.
func GetWwwBroadwayworldComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.broadwayworld.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "h1[itemprop=headline]"},
				{Selector: "h1.article-title"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "span[itemprop=author]"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[itemprop=datePublished]", Attribute: "value"},
			},
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{"div[itemprop=articlebody]"},
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
