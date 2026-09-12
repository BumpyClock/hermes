// ABOUTME: TechLog IIJ (Internet Initiative Japan) extractor for technical blog
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/techlog.iij.ad.jp/index.js

package custom

// TechlogIijAdJpExtractor provides the custom extraction rules for techlog.iij.ad.jp
// JavaScript equivalent: export const TechlogIijAdJpExtractor = { ... }.
var TechlogIijAdJpExtractor = &CustomExtractor{
	Domain: "techlog.iij.ad.jp",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1.entry-title"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "a[rel=\"author\"]"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "time.entry-date", Attribute: "datetime"},
		},
	},

	// Dek is null in JavaScript

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"div.entry-content"},
		},
		DisableDefaultCleaner: true,

		// Clean selectors
		Clean: []string{
			".wp_social_bookmarking_light",
		},
	},
}

// GetTechlogIijAdJpExtractor returns the TechLog IIJ custom extractor.
func GetTechlogIijAdJpExtractor() *CustomExtractor {
	return TechlogIijAdJpExtractor
}
