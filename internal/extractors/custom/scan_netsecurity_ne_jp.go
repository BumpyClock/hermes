// ABOUTME: ScanNetSecurity extractor for Japanese cybersecurity news
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/scan.netsecurity.ne.jp/index.js

package custom

// ScanNetsecurityNeJpExtractor provides the custom extraction rules for scan.netsecurity.ne.jp
// JavaScript equivalent: export const ScanNetsecurityNeJpExtractor = { ... }.
var ScanNetsecurityNeJpExtractor = &CustomExtractor{
	Domain: "scan.netsecurity.ne.jp",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "header.arti-header h1.head"},
		},
	},

	// Author is null in JavaScript

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"article:modified_time\"]", Attribute: "value"},
		},
	},

	Dek: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "header.arti-header p.arti-summary"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"div.arti-content.arti-content--thumbnail"},
		},
		DisableDefaultCleaner: true,

		// Clean selectors
		Clean: []string{
			"aside.arti-giga",
		},
	},
}

// GetScanNetsecurityNeJpExtractor returns the ScanNetSecurity custom extractor.
func GetScanNetsecurityNeJpExtractor() *CustomExtractor {
	return ScanNetsecurityNeJpExtractor
}
