// ABOUTME: Mashable custom extractor with string transforms (.image-credit to figcaption)
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/mashable.com/index.js

package custom

// MashableComExtractor provides the custom extraction rules for mashable.com
// JavaScript equivalent: export const MashableComExtractor = { ... }
var MashableComExtractor = &CustomExtractor{
	Domain: "mashable.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"header h1",
			"h1.title",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:author\"]", "value"},
			"span.author_name a",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:published_time\"]", "value"},
		},
	},
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				"#article",
				"section.article-content.blueprint",
			},
		},
		
		// Transform functions for Mashable-specific content
		Transforms: map[string]TransformFunction{
			".image-credit": &StringTransform{
				TargetTag: "figcaption",
			},
		},
		
		// Clean selectors (empty in JavaScript)
		Clean: []string{},
	},
}

// GetMashableComExtractor returns the Mashable custom extractor
func GetMashableComExtractor() *CustomExtractor {
	return MashableComExtractor
}