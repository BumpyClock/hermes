// ABOUTME: TechLog IIJ (Internet Initiative Japan) extractor for technical blog
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/techlog.iij.ad.jp/index.js

package custom

// TechlogIijAdJpExtractor provides the custom extraction rules for techlog.iij.ad.jp
// JavaScript equivalent: export const TechlogIijAdJpExtractor = { ... }
var TechlogIijAdJpExtractor = &CustomExtractor{
	Domain: "techlog.iij.ad.jp",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1.entry-title",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			"a[rel=\"author\"]",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"time.entry-date", "datetime"},
		},
	},
	
	// Dek is null in JavaScript
	Dek: nil,
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				"div.entry-content",
			},
			DefaultCleaner: false,
		},
		
		// Transform functions (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors
		Clean: []string{
			".wp_social_bookmarking_light",
		},
	},
}

// GetTechlogIijAdJpExtractor returns the TechLog IIJ custom extractor
func GetTechlogIijAdJpExtractor() *CustomExtractor {
	return TechlogIijAdJpExtractor
}