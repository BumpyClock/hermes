// ABOUTME: LittleThings custom extractor for lifestyle content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.littlethings.com/index.js

package custom

// LittleThingsCustomExtractor provides the custom extraction rules for www.littlethings.com
// JavaScript equivalent: export const LittleThingsExtractor = { ... }.
var LittleThingsCustomExtractor = &CustomExtractor{
	Domain: "www.littlethings.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1[class*=\"PostHeader\"]"},
			{Selector: "h1.post-title"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "div[class^=\"PostHeader__ScAuthorNameSection\"]"},
			{Selector: "meta[name=\"author\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"section[class*=\"PostMainArticle\"]"},
			{".mainContentIntro"},
			{".content-wrapper"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},
}

// GetLittleThingsExtractor returns the LittleThings custom extractor.
func GetLittleThingsExtractor() *CustomExtractor {
	return LittleThingsCustomExtractor
}
