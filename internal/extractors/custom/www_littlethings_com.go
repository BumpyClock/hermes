// ABOUTME: LittleThings custom extractor for lifestyle content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.littlethings.com/index.js

package custom

// LittleThingsCustomExtractor provides the custom extraction rules for www.littlethings.com
// JavaScript equivalent: export const LittleThingsExtractor = { ... }.
var LittleThingsCustomExtractor = &CustomExtractor{
	Domain: "www.littlethings.com",

	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1[class*=\"PostHeader\"]",
			"h1.post-title",
		},
	},

	Author: &FieldExtractor{
		Selectors: []interface{}{
			"div[class^=\"PostHeader__ScAuthorNameSection\"]",
			[]string{"meta[name=\"author\"]", "value"},
		},
	},

	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				"section[class*=\"PostMainArticle\"]",
				".mainContentIntro",
				".content-wrapper",
			},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
}

// GetLittleThingsExtractor returns the LittleThings custom extractor.
func GetLittleThingsExtractor() *CustomExtractor {
	return LittleThingsCustomExtractor
}
