// ABOUTME: MOONGIFT Japan open source/tech site custom extractor with timezone support
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.moongift.jp/index.js

package custom

// WwwMoongiftJpExtractor provides the custom extraction rules for www.moongift.jp
// JavaScript equivalent: export const WwwMoongiftJpExtractor = { ... }.
var WwwMoongiftJpExtractor = &CustomExtractor{
	Domain: "www.moongift.jp",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1.title a"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "ul.meta li:not(.social):first-of-type"},
		},

		// timezone: 'Asia/Tokyo' in JavaScript - note: Go implementation handles timezone in date cleaner
		Timezone: "Asia/Tokyo",
	},

	Dek: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:description\"]", Attribute: "value"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"#main"},
		},

		// Clean service promotion content
		Clean: []string{
			"ul.mg_service.cf",
		},
	},
}

// GetWwwMoongiftJpExtractor returns the MOONGIFT Japan custom extractor.
func GetWwwMoongiftJpExtractor() *CustomExtractor {
	return WwwMoongiftJpExtractor
}
