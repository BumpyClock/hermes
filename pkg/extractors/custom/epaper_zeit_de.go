// ABOUTME: Zeit.de e-paper custom extractor with string-based transforms for layout elements
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/epaper.zeit.de/index.js

package custom

// EpaperZeitDeExtractor provides the custom extraction rules for epaper.zeit.de
// JavaScript equivalent: export const EpaperZeitDeExtractor = { ... }
var EpaperZeitDeExtractor = &CustomExtractor{
	Domain: "epaper.zeit.de",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"p.title",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			".article__author",
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				".article",
			},
		},
		
		// String-based transform functions for layout elements
		Transforms: map[string]TransformFunction{
			"p.title":           &StringTransform{"h1"},
			".article__author":  &StringTransform{"p"},
			"byline":           &StringTransform{"p"},
			"linkbox":          &StringTransform{"p"},
		},
		
		// Clean selectors - remove unwanted elements
		Clean: []string{
			"image-credits",
			"box[type=citation]",
		},
	},
	
	// JavaScript: date_published: null
	DatePublished: nil,
	
	// JavaScript: lead_image_url: null
	LeadImageURL: nil,
	
	// JavaScript: dek: null - but has excerpt
	Dek: nil,
	
	NextPageURL: nil,
	
	Excerpt: &FieldExtractor{
		Selectors: []interface{}{
			"subtitle",
		},
	},
}

// GetEpaperZeitDeExtractor returns the Zeit.de e-paper custom extractor
func GetEpaperZeitDeExtractor() *CustomExtractor {
	return EpaperZeitDeExtractor
}