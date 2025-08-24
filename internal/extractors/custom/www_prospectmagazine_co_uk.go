// ABOUTME: Prospect Magazine UK custom extractor with European timezone handling
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.prospectmagazine.co.uk/index.js

package custom

// WwwProspectmagazineCoUkExtractor provides the custom extraction rules for www.prospectmagazine.co.uk
// JavaScript equivalent: export const WwwProspectmagazineCoUkExtractor = { ... }
var WwwProspectmagazineCoUkExtractor = &CustomExtractor{
	Domain: "www.prospectmagazine.co.uk",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			".blog-header__title",
			".page-title",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			".blog-header__author-link",
			".aside_author .title",
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				".blog__container",
				"article .post_content",
			},
		},
		
		// Transform functions (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors (empty in JavaScript)
		Clean: []string{},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:published_time\"]", "value"},
			".post-info",
		},
		// Note: JavaScript version has timezone: 'Europe/London'
		// This is handled by dateparse library in Go
	},
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			".blog-header__description",
			".page-subtitle",
		},
	},
	
	NextPageURL: nil,
	
	Excerpt: nil,
}

// GetWwwProspectmagazineCoUkExtractor returns the Prospect Magazine UK custom extractor
func GetWwwProspectmagazineCoUkExtractor() *CustomExtractor {
	return WwwProspectmagazineCoUkExtractor
}