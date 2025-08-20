// ABOUTME: Blogspot.com custom extractor with noscript content handling and template variations
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/blogspot.com/index.js

package custom

// BlogspotCustomExtractor provides the custom extraction rules for blogspot.com
// JavaScript equivalent: export const BloggerExtractor = { ... }
var BlogspotCustomExtractor = &CustomExtractor{
	Domain: "blogspot.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			".post h2.title",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			".post-author-name",
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				// Blogger is insane and does not load its content
				// initially in the page, but it's all there
				// in noscript
				".post-content noscript",
			},
		},
		
		// Convert the noscript tag to a div
		Transforms: map[string]TransformFunction{
			"noscript": &StringTransform{
				TargetTag: "div",
			},
		},
		
		// Selectors to remove from the extracted content
		Clean: []string{},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			"span.publishdate",
		},
	},
	
	LeadImageURL: nil,
	
	Dek: nil,
	
	NextPageURL: nil,
	
	Excerpt: nil,
}

// GetBlogspotExtractor returns the Blogspot custom extractor
func GetBlogspotExtractor() *CustomExtractor {
	return BlogspotCustomExtractor
}