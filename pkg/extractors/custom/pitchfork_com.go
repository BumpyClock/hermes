// ABOUTME: Pitchfork custom extractor for music reviews and embedded media
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/pitchfork.com/index.js

package custom

// PitchforkCustomExtractor provides the custom extraction rules for pitchfork.com
// JavaScript equivalent: export const PitchforkComExtractor = { ... }
var PitchforkCustomExtractor = &CustomExtractor{
	Domain: "pitchfork.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:title\"]", "value"},
			"title",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:author\"]", "value"},
			".authors-detail__display-name",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			"div[class^=\"InfoSliceWrapper-\"]",
			[]string{".pub-date", "datetime"},
		},
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:description\"]", "value"},
			".review-detail__abstract",
		},
	},
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
			[]string{".single-album-tombstone__art img", "src"},
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				"div.body__inner-container",
				".review-detail__text",
			},
		},
		
		// No transforms in original JavaScript
		Transforms: map[string]TransformFunction{},
		
		// No clean selectors in original JavaScript
		Clean: []string{},
	},
	
	// Extended fields for music review scores
	Extend: map[string]*FieldExtractor{
		"score": {
			Selectors: []interface{}{
				"p[class*=\"Rating\"]",
				".score",
			},
		},
	},
	
	// No selectors in original JavaScript for these fields
	NextPageURL: &FieldExtractor{
		Selectors: []interface{}{},
	},
	
	Excerpt: &FieldExtractor{
		Selectors: []interface{}{},
	},
}

// GetPitchforkExtractor returns the Pitchfork custom extractor
func GetPitchforkExtractor() *CustomExtractor {
	return PitchforkCustomExtractor
}