// ABOUTME: Fandom Wikia.com custom extractor with wiki content and community-driven articles
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/fandom.wikia.com/index.js

package custom

// FandomWikiaCustomExtractor provides the custom extraction rules for fandom.wikia.com
// JavaScript equivalent: export const WikiaExtractor = { ... }
var FandomWikiaCustomExtractor = &CustomExtractor{
	Domain: "fandom.wikia.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1.entry-title",
			// enter title selectors
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			".author vcard",
			".fn",
			// enter author selectors
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				".grid-content",
				".entry-content",
				// enter content selectors
			},
		},
		
		// No transforms needed for Fandom Wikia
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors - empty for Fandom Wikia
		Clean: []string{},
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
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			// Empty selectors array
		},
	},
	
	NextPageURL: nil,
	
	Excerpt: nil,
}

// GetFandomWikiaExtractor returns the Fandom Wikia custom extractor
func GetFandomWikiaExtractor() *CustomExtractor {
	return FandomWikiaCustomExtractor
}