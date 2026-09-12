// ABOUTME: Deadspin (Gawker Media) custom extractor with multi-domain support and YouTube transforms
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/deadspin.com/index.js

package custom

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// DeadspinComExtractor provides the custom extraction rules for deadspin.com and supported domains
// JavaScript equivalent: export const DeadspinExtractor = { ... }.
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
		Selectors: []SelectorEntry{
			{Selector: "header h1"},
			{Selector: "h1.headline"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "a[data-ga*=\"Author\"]"},
			{Selector: ".author"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"article:published_time\"]", Attribute: "value"},
			{Selector: "time.updated[datetime]", Attribute: "datetime"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{".js_post-content"},
			{".post-content"},
			{".entry-content"},
		},

		// Transform functions for Deadspin-specific content
		Transforms: map[string]TransformFunction{
			// Transform lazy-loaded YouTube iframes
			"iframe.lazyload[data-recommend-id^=\"youtube://\"]": &FunctionTransform{
				Fn: func(selection *goquery.Selection) error {
					id, exists := selection.Attr("id")
					if exists && strings.HasPrefix(id, "youtube-") {
						youtubeId := strings.TrimPrefix(id, "youtube-")
						selection.SetAttr("src", "https://www.youtube.com/embed/"+youtubeId)
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
}

// GetDeadspinComExtractor returns the Deadspin custom extractor.
func GetDeadspinComExtractor() *CustomExtractor {
	return DeadspinComExtractor
}
