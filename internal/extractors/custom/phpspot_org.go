// ABOUTME: PHPSpot Japan PHP/development site custom extractor with Japanese date format
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/phpspot.org/index.js

package custom

// PhpspotOrgExtractor provides the custom extraction rules for phpspot.org
// JavaScript equivalent: export const PhpspotOrgExtractor = { ... }.
var PhpspotOrgExtractor = &CustomExtractor{
	Domain: "phpspot.org",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h3.hl"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h4.hl"},
		},

		// format: 'YYYY年MM月DD日' in JavaScript - note: Go implementation handles format in date cleaner
		Format: "YYYY年MM月DD日",

		// timezone: 'Asia/Tokyo' in JavaScript - note: Go implementation handles timezone in date cleaner
		Timezone: "Asia/Tokyo",
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"div.entrybody"},
		},

		// defaultCleaner: false in JavaScript
		DisableDefaultCleaner: true,
	},
}

// GetPhpspotOrgExtractor returns the PHPSpot Japan custom extractor.
func GetPhpspotOrgExtractor() *CustomExtractor {
	return PhpspotOrgExtractor
}
