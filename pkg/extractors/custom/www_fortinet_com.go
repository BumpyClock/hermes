// ABOUTME: Fortinet cybersecurity company extractor with noscript image transforms
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.fortinet.com/index.js

package custom

import (
	"github.com/PuerkitoBio/goquery"
)

// WwwFortinetComExtractor provides the custom extraction rules for www.fortinet.com
// JavaScript equivalent: export const WwwFortinetComExtractor = { ... }
var WwwFortinetComExtractor = &CustomExtractor{
	Domain: "www.fortinet.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			".b15-blog-meta__author",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:published_time\"]", "value"},
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
				"div.responsivegrid.aem-GridColumn.aem-GridColumn--default--12",
			},
		},
		
		// Transform functions for Fortinet-specific content
		// JavaScript: transforms: { noscript: $node => { ... } }
		Transforms: map[string]TransformFunction{
			"noscript": &FunctionTransform{
				Fn: func(selection *goquery.Selection) error {
					// Check if noscript has a single img child
					children := selection.Children()
					if children.Length() == 1 {
						firstChild := children.First()
						if firstChild.Is("img") {
							// Convert noscript with single img to figure
							selection.ReplaceWithSelection(firstChild.WrapInner("<figure>").Parent())
						}
					}
					return nil
				},
			},
		},
		
		// Clean selectors (none in JavaScript)
		Clean: []string{},
	},
}

// GetWwwFortinetComExtractor returns the Fortinet custom extractor
func GetWwwFortinetComExtractor() *CustomExtractor {
	return WwwFortinetComExtractor
}