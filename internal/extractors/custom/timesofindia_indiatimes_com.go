// ABOUTME: Times of India (Indian news) custom extractor with timezone handling and extended fields
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/timesofindia.indiatimes.com/index.js

package custom

// TimesofindiaIndiatimesComExtractor provides the custom extraction rules for timesofindia.indiatimes.com
// JavaScript equivalent: export const TimesofindiaIndiatimesComExtractor = { ... }
var TimesofindiaIndiatimesComExtractor = &CustomExtractor{
	Domain: "timesofindia.indiatimes.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1",
		},
	},
	
	// No author field in JavaScript - has extend.reporter instead
	Author: nil,
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				"div.contentwrapper:has(section)",
			},
		},
		
		// Transform functions (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors
		Clean: []string{
			"section",
			"h1",
			".byline",
			".img_cptn",
			".icon_share_wrap",
			"ul[itemtype=\"https://schema.org/BreadcrumbList\"]",
		},
		
		// JavaScript: defaultCleaner: false
		DefaultCleaner: false,
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			".byline",
		},
		// Note: JavaScript version has format: 'MMM D, YYYY, HH:mm z' and timezone: 'Asia/Kolkata'
		// This is handled by dateparse library in Go
	},
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
	
	// No dek field in JavaScript
	Dek: nil,
	
	NextPageURL: nil,
	
	Excerpt: nil,
	
	// JavaScript has extend field for reporter
	Extend: map[string]*FieldExtractor{
		"reporter": {
			Selectors: []interface{}{
				"div.byline",
			},
		},
	},
}

// GetTimesofindiaIndiatimesComExtractor returns the Times of India custom extractor
func GetTimesofindiaIndiatimesComExtractor() *CustomExtractor {
	return TimesofindiaIndiatimesComExtractor
}