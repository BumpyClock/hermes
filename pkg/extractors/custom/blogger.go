// ABOUTME: Blogger/Blogspot custom extractor for Blogger-specific content extraction
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/blogspot.com/index.js

package custom

// BloggerCustomExtractor provides the custom extraction rules for Blogger/Blogspot
// JavaScript equivalent: export const BloggerExtractor = { ... }
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
		FieldExtractor: &FieldExtractor{
			// Blogger is insane and does not load its content
			// initially in the page, but it's all there in noscript
			Selectors: []interface{}{".post-content noscript"},
		},
		
		// Selectors to remove from the extracted content
		Clean: []string{},
		
		// Convert the noscript tag to a div
		Transforms: map[string]TransformFunction{
			"noscript": &StringTransform{TargetTag: "div"},
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{".post-author-name"},
	},
	
	Title: &FieldExtractor{
		Selectors: []interface{}{".post h2.title"},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{"span.publishdate"},
	},
}

// GetBloggerExtractor returns the Blogger custom extractor
func GetBloggerExtractor() *CustomExtractor {
	return BloggerCustomExtractor
}