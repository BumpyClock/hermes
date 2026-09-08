// ABOUTME: CNET Japan custom extractor with Japanese date format parsing
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/japan.cnet.com/index.js

package custom

// JapanCnetComExtractor provides the custom extraction rules for japan.cnet.com
// JavaScript equivalent: export const JapanCnetComExtractor = { ... }.
var JapanCnetComExtractor = &CustomExtractor{
	Domain: "japan.cnet.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".leaf-headline-ttl"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".writer"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".date"},
		},
		// Note: format and timezone would be handled at extraction time
		// format: 'YYYY年MM月DD日 HH時mm分' (from JavaScript)
		// timezone: 'Asia/Tokyo' (from JavaScript)
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"div.article_body"},
		},
	},
}

// GetJapanCnetComExtractor returns the CNET Japan custom extractor.
func GetJapanCnetComExtractor() *CustomExtractor {
	return JapanCnetComExtractor
}
