// ABOUTME: Uproxx custom extractor for music and entertainment content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/uproxx.com/index.js

package custom

// UproxxCustomExtractor provides the custom extraction rules for uproxx.com
// JavaScript equivalent: export const UproxxComExtractor = { ... }
var UproxxCustomExtractor = &CustomExtractor{
	Domain: "uproxx.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"div.entry-header h1",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"qc:author\"]", "value"},
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
				".entry-content",
			},
		},
		
		// Transform functions for Uproxx-specific content
		Transforms: map[string]TransformFunction{
			"div.image": &StringTransform{TargetTag: "figure"},
			"div.image .wp-media-credit": &StringTransform{TargetTag: "figcaption"},
		},
		
		// Clean selectors - empty array in original JavaScript
		Clean: []string{},
	},
	
	// No selectors in original JavaScript for these fields
	Dek: &FieldExtractor{
		Selectors: []interface{}{},
	},
	
	NextPageURL: &FieldExtractor{
		Selectors: []interface{}{},
	},
	
	Excerpt: &FieldExtractor{
		Selectors: []interface{}{},
	},
}

// GetUproxxExtractor returns the Uproxx custom extractor
func GetUproxxExtractor() *CustomExtractor {
	return UproxxCustomExtractor
}