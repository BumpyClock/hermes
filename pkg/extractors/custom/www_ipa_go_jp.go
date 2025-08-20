// ABOUTME: IPA (Information-technology Promotion Agency, Japan) government extractor
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.ipa.go.jp/index.js

package custom

// WwwIpaGoJpExtractor provides the custom extraction rules for www.ipa.go.jp
// JavaScript equivalent: export const WwwIpaGoJpExtractor = { ... }
var WwwIpaGoJpExtractor = &CustomExtractor{
	Domain: "www.ipa.go.jp",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1",
		},
	},
	
	// Author is null in JavaScript
	Author: nil,
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			"p.ipar_text_right",
		},
		// JavaScript: format: 'YYYY年M月D日', timezone: 'Asia/Tokyo'
		// Go handles Japanese date formats and timezone automatically
	},
	
	// Dek is null in JavaScript
	Dek: nil,
	
	// Lead image URL is null in JavaScript
	LeadImageURL: nil,
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				"#ipar_main",
			},
			DefaultCleaner: false,
		},
		
		// Transform functions (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors
		Clean: []string{
			"p.ipar_text_right",
		},
	},
}

// GetWwwIpaGoJpExtractor returns the IPA Japan custom extractor
func GetWwwIpaGoJpExtractor() *CustomExtractor {
	return WwwIpaGoJpExtractor
}