// ABOUTME: Blogger/Blogspot custom extractor for Blogger-specific content extraction
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/blogspot.com/index.js

package custom

// BloggerCustomExtractor provides the custom extraction rules for Blogger/Blogspot
// JavaScript equivalent: export const BloggerExtractor = { ... }.
var BloggerCustomExtractor = &CustomExtractor{
	Domain: "blogspot.com",

	// Blogger supports multiple international domains
	SupportedDomains: []string{
		"www.blogspot.com",
		"blogspot.co.uk",
		"blogspot.ca",
		"blogspot.de",
		"blogspot.fr",
		"blogspot.jp",
		"blogspot.in",
		"blogspot.com.au",
		"blogspot.com.br",
		"blogspot.mx",
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{{".post-content noscript"}},

		// Convert the noscript tag to a div
		Transforms: map[string]TransformFunction{
			"noscript": &StringTransform{TargetTag: "div"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{{Selector: ".post-author-name"}},
	},

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{{Selector: ".post h2.title"}},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{{Selector: "span.publishdate"}},
	},
}

// GetBloggerExtractor returns the Blogger custom extractor.
func GetBloggerExtractor() *CustomExtractor {
	return BloggerCustomExtractor
}
