// ABOUTME: MOONGIFT Japan open source/tech site custom extractor with timezone support
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.moongift.jp/index.js

package custom

// WwwMoongiftJpExtractor provides the custom extraction rules for www.moongift.jp
// JavaScript equivalent: export const WwwMoongiftJpExtractor = { ... }
var WwwMoongiftJpExtractor = &CustomExtractor{
	Domain: "www.moongift.jp",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1.title a",
		},
	},
	
	Author: nil,
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			"ul.meta li:not(.social):first-of-type",
		},
		
		// timezone: 'Asia/Tokyo' in JavaScript - note: Go implementation handles timezone in date cleaner
		Timezone: "Asia/Tokyo",
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:description\"]", "value"},
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
				"#main",
			},
		},
		
		// transforms: {} (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean service promotion content
		Clean: []string{
			"ul.mg_service.cf",
		},
	},
}

// GetWwwMoongiftJpExtractor returns the MOONGIFT Japan custom extractor
func GetWwwMoongiftJpExtractor() *CustomExtractor {
	return WwwMoongiftJpExtractor
}