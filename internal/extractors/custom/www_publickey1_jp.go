// ABOUTME: Publickey1 Japan tech news site custom extractor with Japanese date format
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.publickey1.jp/index.js

package custom

// WwwPublickey1JpExtractor provides the custom extraction rules for www.publickey1.jp
// JavaScript equivalent: export const WwwPublickey1JpExtractor = { ... }
var WwwPublickey1JpExtractor = &CustomExtractor{
	Domain: "www.publickey1.jp",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			".bloggerinchief p:first-of-type",
			"#subcol p:has(img)",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			"div.pubdate",
		},
		
		// format: 'YYYY年MM月DD日' in JavaScript - note: Go implementation handles format in date cleaner
		Format:   "YYYY年MM月DD日",
		
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
				"#maincol",
			},
		},
		
		// defaultCleaner: false in JavaScript
		DefaultCleaner: false,
		
		// transforms: {} (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean navigation and ads
		Clean: []string{
			"#breadcrumbs",
			"div.sbm",
			"div.ad_footer",
		},
	},
}

// GetWwwPublickey1JpExtractor returns the Publickey1 Japan custom extractor
func GetWwwPublickey1JpExtractor() *CustomExtractor {
	return WwwPublickey1JpExtractor
}