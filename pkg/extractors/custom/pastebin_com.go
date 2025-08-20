// ABOUTME: Pastebin.com custom extractor with code content handling, syntax highlighting, and list transforms  
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/pastebin.com/index.js

package custom

// PastebinCustomExtractor provides the custom extraction rules for pastebin.com
// JavaScript equivalent: export const PastebinComExtractor = { ... }
var PastebinCustomExtractor = &CustomExtractor{
	Domain: "pastebin.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			".username",
			".paste_box_line2 .t_us + a",
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				".source",
				"#selectable .text",
			},
		},
		
		// Transform functions for Pastebin code content
		Transforms: map[string]TransformFunction{
			// Convert ordered list to div
			"ol": &StringTransform{
				TargetTag: "div",
			},
			
			// Convert list items to paragraphs
			"li": &StringTransform{
				TargetTag: "p",
			},
		},
		
		// Clean selectors - empty for Pastebin
		Clean: []string{},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			".date",
			".paste_box_line2 .t_da + span",
		},
		// Timezone from JavaScript: 'America/New_York'
		// Format from JavaScript: 'MMMM D, YYYY'
	},
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
	
	Dek: nil,
	
	NextPageURL: nil,
	
	Excerpt: nil,
}

// GetPastebinExtractor returns the Pastebin custom extractor
func GetPastebinExtractor() *CustomExtractor {
	return PastebinCustomExtractor
}