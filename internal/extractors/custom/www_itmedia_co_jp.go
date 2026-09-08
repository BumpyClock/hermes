// ABOUTME: ITmedia Japan tech news site custom extractor with multiple supported domains
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.itmedia.co.jp/index.js

package custom

// WwwItmediaCoJpExtractor provides the custom extraction rules for www.itmedia.co.jp and related domains
// JavaScript equivalent: export const WwwItmediaCoJpExtractor = { ... }.
var WwwItmediaCoJpExtractor = &CustomExtractor{
	Domain: "www.itmedia.co.jp",

	// JavaScript equivalent: supportedDomains: ['www.atmarkit.co.jp', 'techtarget.itmedia.co.jp', 'nlab.itmedia.co.jp']
	SupportedDomains: []string{
		"www.atmarkit.co.jp",
		"techtarget.itmedia.co.jp",
		"nlab.itmedia.co.jp",
	},

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "#cmsTitle h1"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "#byline"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"article:modified_time\"]", Attribute: "value"},
		},
	},

	Dek: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "#cmsAbstract h2"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"#cmsBody"},
		},

		// defaultCleaner: false in JavaScript
		DisableDefaultCleaner: true,

		// Clean social sharing content
		Clean: []string{
			"#snsSharebox",
		},
	},
}

// GetWwwItmediaCoJpExtractor returns the ITmedia Japan custom extractor.
func GetWwwItmediaCoJpExtractor() *CustomExtractor {
	return WwwItmediaCoJpExtractor
}
