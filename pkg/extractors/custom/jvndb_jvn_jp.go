// ABOUTME: JVNDB (Japan Vulnerability Notes Database) extractor
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/jvndb.jvn.jp/index.js

package custom

// JvndbJvnJpExtractor provides the custom extraction rules for jvndb.jvn.jp
// JavaScript equivalent: export const JvndbJvnJpExtractor = { ... }
var JvndbJvnJpExtractor = &CustomExtractor{
	Domain: "jvndb.jvn.jp",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"title",
		},
	},
	
	// Author is null in JavaScript
	Author: nil,
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			"div.modifytxt:nth-child(2)",
		},
		// JavaScript: format: 'YYYY/MM/DD', timezone: 'Asia/Tokyo'
		// Go handles date formats and timezone automatically
	},
	
	// Dek is null in JavaScript
	Dek: nil,
	
	// Lead image URL is null in JavaScript
	LeadImageURL: nil,
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				"#news-list",
			},
			DefaultCleaner: false,
		},
		
		// Transform functions (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors (empty in JavaScript)
		Clean: []string{},
	},
}

// GetJvndbJvnJpExtractor returns the JVNDB custom extractor
func GetJvndbJvnJpExtractor() *CustomExtractor {
	return JvndbJvnJpExtractor
}