// ABOUTME: SECT (Security Engineering & Communication Technology) IIJ extractor
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/sect.iij.ad.jp/index.js

package custom

// SectIijAdJpExtractor provides the custom extraction rules for sect.iij.ad.jp
// JavaScript equivalent: export const SectIijAdJpExtractor = { ... }
var SectIijAdJpExtractor = &CustomExtractor{
	Domain: "sect.iij.ad.jp",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"div.title-box-inner h1",
			"h3",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			"p.post-author a",
			"dl.entrydate dd",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			"time",
		},
		// JavaScript: format: 'YYYY年MM月DD日', timezone: 'Asia/Tokyo'
		// Go handles Japanese date formats and timezone automatically
	},
	
	// Dek is null in JavaScript
	Dek: nil,
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				".entry-inner",
				"#article",
			},
		},
		
		// Transform functions (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors
		Clean: []string{
			"dl.entrydate",
		},
	},
}

// GetSectIijAdJpExtractor returns the SECT IIJ custom extractor
func GetSectIijAdJpExtractor() *CustomExtractor {
	return SectIijAdJpExtractor
}