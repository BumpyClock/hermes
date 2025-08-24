// ABOUTME: Weekly ASCII Japan tech magazine custom extractor with Japanese date format and timezone
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/weekly.ascii.jp/index.js

package custom

// WeeklyAsciiJpExtractor provides the custom extraction rules for weekly.ascii.jp
// JavaScript equivalent: export const WeeklyAsciiJpExtractor = { ... }
var WeeklyAsciiJpExtractor = &CustomExtractor{
	Domain: "weekly.ascii.jp",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"article h1",
			"h1[itemprop=\"headline\"]",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			"p.author",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			"p.date",
			[]string{"meta[name=\"odate\"]", "value"},
		},
		
		// format: 'YYYY年MM月DD日 HH:mm' in JavaScript - note: Go implementation handles format in date cleaner
		Format:   "YYYY年MM月DD日 HH:mm",
		
		// timezone: 'Asia/Tokyo' in JavaScript - note: Go implementation handles timezone in date cleaner
		Timezone: "Asia/Tokyo",
	},
	
	Dek: nil,
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				"div#contents_detail",
				"div.article",
			},
		},
		
		// transforms: {} (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// clean: [] (empty in JavaScript)
		Clean: []string{},
	},
}

// GetWeeklyAsciiJpExtractor returns the Weekly ASCII Japan custom extractor
func GetWeeklyAsciiJpExtractor() *CustomExtractor {
	return WeeklyAsciiJpExtractor
}