// ABOUTME: The Verge custom extractor with noscript transforms and multi-match selectors
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.theverge.com/index.js

package custom

import (
	"github.com/PuerkitoBio/goquery"
)

// WwwThevergeComExtractor provides the custom extraction rules for www.theverge.com
// JavaScript equivalent: export const WwwThevergeComExtractor = { ... }
var WwwThevergeComExtractor = &CustomExtractor{
	Domain: "www.theverge.com",
	
	SupportedDomains: []string{"www.polygon.com"},
	
	Title: &FieldExtractor{
		Selectors: []interface{}{"h1"},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"author\"]", "value"},
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:published_time\"]", "value"},
		},
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			".p-dek",
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
				// feature template multi-match
				[]interface{}{".c-entry-hero .e-image", ".c-entry-intro", ".c-entry-content"},
				// regular post multi-match
				[]interface{}{".e-image--hero", ".c-entry-content"},
				// feature template fallback
				".l-wrapper .l-feature",
				// regular post fallback
				"div.c-entry-content",
			},
		},
		
		// Transform functions for The Verge-specific content
		Transforms: map[string]TransformFunction{
			// Transform lazy-loaded images
			"noscript": &FunctionTransform{
				Fn: func(selection *goquery.Selection) error {
					children := selection.Children()
					if children.Length() == 1 {
						firstChild := children.First()
						if goquery.NodeName(firstChild) == "img" {
							// Convert to span
							html, _ := children.Html()
							selection.ReplaceWithHtml("<span>" + html + "</span>")
						}
					}
					return nil
				},
			},
		},
		
		// Clean selectors - remove unwanted elements
		Clean: []string{
			".aside",
			"img.c-dynamic-image", // images come from noscript transform
		},
	},
}

// GetWwwThevergeComExtractor returns The Verge custom extractor
func GetWwwThevergeComExtractor() *CustomExtractor {
	return WwwThevergeComExtractor
}