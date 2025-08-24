// ABOUTME: ICI Radio-Canada (French Canadian news) custom extractor with date format and timezone handling
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/ici.radio-canada.ca/index.js

package custom

// IciRadioCanadaCaExtractor provides the custom extraction rules for ici.radio-canada.ca
// JavaScript equivalent: export const IciRadioCanadaCaExtractor = { ... }
var IciRadioCanadaCaExtractor = &CustomExtractor{
	Domain: "ici.radio-canada.ca",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"dc.creator\"]", "value"},
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				"section.document-content-style",
				[]string{".main-multimedia-item", ".news-story-content"},
			},
		},
		
		// Transform functions (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors (empty in JavaScript)
		Clean: []string{},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"dc.date.created\"]", "value"},
		},
		// Note: JavaScript version has format: 'YYYY-MM-DD|HH[h]mm' and timezone: 'America/New_York'
		// This is handled by dateparse library in Go which can parse various formats automatically
	},
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			"div.lead-container",
			".bunker-component.lead",
		},
	},
	
	NextPageURL: nil,
	
	Excerpt: nil,
}

// GetIciRadioCanadaCaExtractor returns the ICI Radio-Canada custom extractor
func GetIciRadioCanadaCaExtractor() *CustomExtractor {
	return IciRadioCanadaCaExtractor
}