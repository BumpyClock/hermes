// ABOUTME: ScanNetSecurity extractor for Japanese cybersecurity news
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/scan.netsecurity.ne.jp/index.js

package custom

// ScanNetsecurityNeJpExtractor provides the custom extraction rules for scan.netsecurity.ne.jp
// JavaScript equivalent: export const ScanNetsecurityNeJpExtractor = { ... }
var ScanNetsecurityNeJpExtractor = &CustomExtractor{
	Domain: "scan.netsecurity.ne.jp",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"header.arti-header h1.head",
		},
	},
	
	// Author is null in JavaScript
	Author: nil,
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:modified_time\"]", "value"},
		},
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			"header.arti-header p.arti-summary",
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
				"div.arti-content.arti-content--thumbnail",
			},
			DefaultCleaner: false,
		},
		
		// Transform functions (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors
		Clean: []string{
			"aside.arti-giga",
		},
	},
}

// GetScanNetsecurityNeJpExtractor returns the ScanNetSecurity custom extractor
func GetScanNetsecurityNeJpExtractor() *CustomExtractor {
	return ScanNetsecurityNeJpExtractor
}