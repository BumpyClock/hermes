// ABOUTME: Yomiuri Shimbun (major Japanese newspaper) custom extractor with article content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.yomiuri.co.jp/index.js

package custom

// WwwYomiuriCoJpExtractor provides the custom extraction rules for www.yomiuri.co.jp
// JavaScript equivalent: export const WwwYomiuriCoJpExtractor = { ... }.
var WwwYomiuriCoJpExtractor = &CustomExtractor{
	Domain: "www.yomiuri.co.jp",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1.title-article.c-article-title"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"article:published_time\"]", Attribute: "value"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"div.p-main-contents"},
		},
	},
}

// GetWwwYomiuriCoJpExtractor returns the Yomiuri Shimbun custom extractor.
func GetWwwYomiuriCoJpExtractor() *CustomExtractor {
	return WwwYomiuriCoJpExtractor
}
