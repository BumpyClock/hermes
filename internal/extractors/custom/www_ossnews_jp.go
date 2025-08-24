// ABOUTME: OSS News Japan open source news site custom extractor with Japanese date-time format
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.ossnews.jp/index.js

package custom

// WwwOssnewsJpExtractor provides the custom extraction rules for www.ossnews.jp
// JavaScript equivalent: export const WwwOssnewsJpExtractor = { ... }
var WwwOssnewsJpExtractor = &CustomExtractor{
	Domain: "www.ossnews.jp",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"#alpha-block h1.hxnewstitle",
		},
	},
	
	Author: nil,
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			"p.fs12",
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
				"#alpha-block .section:has(h1.hxnewstitle)",
			},
		},
		
		// defaultCleaner: false in JavaScript
		DefaultCleaner: false,
		
		// transforms: {} (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// clean: [] (empty in JavaScript)
		Clean: []string{},
	},
}

// GetWwwOssnewsJpExtractor returns the OSS News Japan custom extractor
func GetWwwOssnewsJpExtractor() *CustomExtractor {
	return WwwOssnewsJpExtractor
}