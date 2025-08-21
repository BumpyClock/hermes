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
			[]string{"meta[name=\"parsely-author\"]", "value"},
			[]string{"meta[name=\"cse-authors\"]", "value"},
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
				// Modern Verge layout - just article body components (no follow section)
				".duet--article--article-body-component",
				// Backup selectors for content divs
				"div[id*='zephr-anchor']",
				// Generic content fallbacks
				"article",
				".article-content",
				// Legacy selectors as final fallback
				[]interface{}{".c-entry-hero .e-image", ".c-entry-intro", ".c-entry-content"},
				[]interface{}{".e-image--hero", ".c-entry-content"},
				".l-wrapper .l-feature",
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
			// Remove excessive image galleries to reduce character count
			".duet--article--image-gallery-two-up .kqz8fh5 .kqz8fh8 .kqz8fh7",
			".duet--article--image-gallery-two-up .kqz8fha .kqz8fh9", 
			"div[class*='image-gallery'] img[srcset]", // Remove srcset attributes
			".duet--media--content-warning", // Remove content warnings
			"._1etxtj1", // Remove image gallery navigation
			// Remove topic follow sections and related content
			".c-related-list",
			".c-entry-group-labels", 
			".c-follow-button",
			".tly2fw0", // Follow topics section
			"button", // Interactive buttons
			// Remove navigation elements
			".c-image-gallery__nav",
			"[class*='follow']", // Follow buttons/sections
		},
	},
}

// GetWwwThevergeComExtractor returns The Verge custom extractor
func GetWwwThevergeComExtractor() *CustomExtractor {
	return WwwThevergeComExtractor
}