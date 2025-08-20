// ABOUTME: ITmedia Japan tech news site custom extractor with multiple supported domains
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.itmedia.co.jp/index.js

package custom

// WwwItmediaCoJpExtractor provides the custom extraction rules for www.itmedia.co.jp and related domains
// JavaScript equivalent: export const WwwItmediaCoJpExtractor = { ... }
var WwwItmediaCoJpExtractor = &CustomExtractor{
	Domain: "www.itmedia.co.jp",
	
	// JavaScript equivalent: supportedDomains: ['www.atmarkit.co.jp', 'techtarget.itmedia.co.jp', 'nlab.itmedia.co.jp']
	SupportedDomains: []string{
		"www.atmarkit.co.jp",
		"techtarget.itmedia.co.jp",
		"nlab.itmedia.co.jp",
	},
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"#cmsTitle h1",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			"#byline",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:modified_time\"]", "value"},
		},
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			"#cmsAbstract h2",
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
				"#cmsBody",
			},
		},
		
		// defaultCleaner: false in JavaScript
		DefaultCleaner: false,
		
		// transforms: {} (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean social sharing content
		Clean: []string{
			"#snsSharebox",
		},
	},
}

// GetWwwItmediaCoJpExtractor returns the ITmedia Japan custom extractor
func GetWwwItmediaCoJpExtractor() *CustomExtractor {
	return WwwItmediaCoJpExtractor
}