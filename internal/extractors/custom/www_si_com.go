// ABOUTME: Sports Illustrated custom extractor with timezone support and noscript transforms 
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.si.com/index.js

package custom

import (
	"github.com/PuerkitoBio/goquery"
)

// WwwSiComExtractor provides the custom extraction rules for www.si.com
// JavaScript equivalent: export const WwwSiComExtractor = { ... }
var WwwSiComExtractor = &CustomExtractor{
	Domain: "www.si.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1",
			"h1.headline",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"author\"]", "value"},
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"published\"]", "value"},
		},
		Timezone: "America/New_York",
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			".m-detail-header--dek",
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
				".m-detail--body",
				[]interface{}{"p", ".marquee_large_2x", ".component.image"},
			},
		},
		
		// Transform functions for SI-specific content
		Transforms: map[string]TransformFunction{
			// Transform noscript elements with single img child to figure
			"noscript": &FunctionTransform{
				Fn: func(selection *goquery.Selection) error {
					children := selection.Children()
					if children.Length() == 1 {
						firstChild := children.First()
						if goquery.NodeName(firstChild) == "img" {
							// Convert to figure
							html, _ := children.Html()
							selection.ReplaceWithHtml("<figure>" + html + "</figure>")
						}
					}
					return nil
				},
			},
		},
		
		// Clean selectors - remove unwanted elements
		Clean: []string{
			".inline-thumb",
			".primary-message", 
			".description",
			".instructions",
		},
	},
}

// GetWwwSiComExtractor returns the Sports Illustrated custom extractor
func GetWwwSiComExtractor() *CustomExtractor {
	return WwwSiComExtractor
}