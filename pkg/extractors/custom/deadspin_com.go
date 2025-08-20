// ABOUTME: Deadspin (Gawker Media) custom extractor with multi-domain support and YouTube transforms 
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/deadspin.com/index.js

package custom

import (
	"strings"
	"github.com/PuerkitoBio/goquery"
)

// DeadspinComExtractor provides the custom extraction rules for deadspin.com and supported domains
// JavaScript equivalent: export const DeadspinExtractor = { ... }
var DeadspinComExtractor = &CustomExtractor{
	Domain: "deadspin.com",
	
	SupportedDomains: []string{
		"jezebel.com",
		"lifehacker.com",
		"kotaku.com",
		"gizmodo.com",
		"jalopnik.com",
		"kinja.com",
		"avclub.com",
		"clickhole.com",
		"splinternews.com",
		"theonion.com",
		"theroot.com",
		"thetakeout.com",
		"theinventory.com",
	},
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"header h1",
			"h1.headline",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			"a[data-ga*=\"Author\"]",
			".author",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:published_time\"]", "value"},
			[]string{"time.updated[datetime]", "datetime"},
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
				".js_post-content",
				".post-content", 
				".entry-content",
			},
		},
		
		// Transform functions for Deadspin-specific content
		Transforms: map[string]TransformFunction{
			// Transform lazy-loaded YouTube iframes
			"iframe.lazyload[data-recommend-id^=\"youtube://\"]": &FunctionTransform{
				Fn: func(selection *goquery.Selection) error {
					id, exists := selection.Attr("id")
					if exists && strings.HasPrefix(id, "youtube-") {
						youtubeId := strings.TrimPrefix(id, "youtube-")
						selection.SetAttr("src", "https://www.youtube.com/embed/" + youtubeId)
					}
					return nil
				},
			},
		},
		
		// Clean selectors - remove unwanted elements
		Clean: []string{
			".magnifier",
			".lightbox",
		},
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			// Empty selectors in JavaScript
		},
	},
	
	NextPageURL: nil, // Empty in JavaScript
	
	Excerpt: nil, // Empty in JavaScript
}

// GetDeadspinComExtractor returns the Deadspin custom extractor
func GetDeadspinComExtractor() *CustomExtractor {
	return DeadspinComExtractor
}