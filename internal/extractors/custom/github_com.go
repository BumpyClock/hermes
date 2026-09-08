// ABOUTME: GitHub custom extractor with README content and relative-time selectors
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/github.com/index.js

package custom

// GithubComExtractor provides the custom extraction rules for github.com
// JavaScript equivalent: export const GithubComExtractor = { ... }.
var GithubComExtractor = &CustomExtractor{
	Domain: "github.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:title\"]", Attribute: "value"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "relative-time[datetime]", Attribute: "datetime"},
			{Selector: "span[itemprop=\"dateModified\"] relative-time", Attribute: "datetime"},
		},
	},

	Dek: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"description\"]", Attribute: "value"},
			{Selector: "span[itemprop=\"about\"]"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"#readme article"},
		},
	},
}

// GetGithubComExtractor returns the GitHub custom extractor.
func GetGithubComExtractor() *CustomExtractor {
	return GithubComExtractor
}
