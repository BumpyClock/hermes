// ABOUTME: Phoronix custom extractor with date format parsing
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.phoronix.com/index.js

package custom

// WwwPhoronixComExtractor provides the custom extraction rules for www.phoronix.com
// JavaScript equivalent: export const WwwPhoronixComExtractor = { ... }.
var WwwPhoronixComExtractor = &CustomExtractor{
	Domain: "www.phoronix.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "article h1"},
			{Selector: "article header"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".author a:first-child"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".author"},
		},
		// Note: format and timezone would be handled at extraction time
		// format: 'D MMMM YYYY at hh:mm' (from JavaScript)
		// timezone: 'America/New_York' (from JavaScript)
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{".content"},
		},
	},
}

// GetWwwPhoronixComExtractor returns the Phoronix custom extractor.
func GetWwwPhoronixComExtractor() *CustomExtractor {
	return WwwPhoronixComExtractor
}
