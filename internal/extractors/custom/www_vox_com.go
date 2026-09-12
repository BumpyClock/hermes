// ABOUTME: Vox.com custom extractor with media-rich content handling and image transforms
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.vox.com/index.js

package custom

import (
	"github.com/PuerkitoBio/goquery"
)

// VoxCustomExtractor provides the custom extraction rules for www.vox.com
// JavaScript equivalent: export const WwwVoxComExtractor = { ... }.
var VoxCustomExtractor = &CustomExtractor{
	Domain: "www.vox.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1[class*=\"h74scy\"]"},
			{Selector: "h1.c-page-title"}, // Legacy fallback
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"author\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{".duet--article--article-body-component"},
			{"div[id*='zephr-anchor']"},
			{".duet--layout--entry-body"},
			// Legacy selectors as fallback
			{"figure.e-image--hero", ".c-entry-content"},
			{".c-entry-content"},
		},

		// Transform functions for Vox-specific content
		Transforms: map[string]TransformFunction{
			// Handle Vox noscript image loading
			"figure .e-image__image noscript": &FunctionTransform{
				Fn: transformVoxNoscriptImage,
			},

			// Transform image meta to figcaption
			"figure .e-image__meta": &StringTransform{
				TargetTag: "figcaption",
			},
		},

		// Clean selectors - remove unwanted elements
		Clean: []string{
			".duet--article--block-placement",   // Ads and promotional blocks
			".duet--article--related",           // Related articles
			".duet--cta--newsletter",            // Newsletter signup forms
			"form",                              // All forms
			".duet--article--share-buttons",     // Share buttons
			".duet--article--article-pullquote", // Pull quotes (duplicate content)
			".duet--media--caption",             // Image captions (can duplicate)
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"article:published_time\"]", Attribute: "value"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Dek: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "p[class*=\"h74scyi\"]"}, // Modern Vox subtitle
			{Selector: ".p-dek"},                // Legacy fallback
		},
	},
}

// transformVoxNoscriptImage handles Vox's lazy-loaded images
// JavaScript equivalent: 'figure .e-image__image noscript': $node => { ... }.
func transformVoxNoscriptImage(selection *goquery.Selection) error {
	imgHtml, err := selection.Html()
	if err != nil {
		return err
	}

	// Find the parent .e-image__image and replace .c-dynamic-image with the noscript content
	imageParent := selection.ParentsFiltered(".e-image__image")
	if imageParent.Length() > 0 {
		dynamicImage := imageParent.Find(".c-dynamic-image")
		if dynamicImage.Length() > 0 {
			dynamicImage.ReplaceWithHtml(imgHtml)
		}
	}

	return nil
}

// GetVoxExtractor returns the Vox custom extractor.
func GetVoxExtractor() *CustomExtractor {
	return VoxCustomExtractor
}
