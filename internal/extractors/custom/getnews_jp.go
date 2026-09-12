// ABOUTME: GetNews Japan news site custom extractor with article body content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/getnews.jp/index.js

package custom

// GetnewsJpExtractor provides the custom extraction rules for getnews.jp
// JavaScript equivalent: export const GetnewsJpExtractor = { ... }.
var GetnewsJpExtractor = &CustomExtractor{
	Domain: "getnews.jp",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "article h1"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"article:author\"]", Attribute: "value"},
			{Selector: "span.prof"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"article:published_time\"]", Attribute: "value"},
			{Selector: "ul.cattag-top time", Attribute: "datetime"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"div.post-bodycopy"},
		},
	},
}

// GetGetnewsJpExtractor returns the GetNews Japan custom extractor.
func GetGetnewsJpExtractor() *CustomExtractor {
	return GetnewsJpExtractor
}
