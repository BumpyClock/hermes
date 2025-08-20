// ABOUTME: Vox.com custom extractor with media-rich content handling and image transforms
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.vox.com/index.js

package custom

import (
	"github.com/PuerkitoBio/goquery"
)

// VoxCustomExtractor provides the custom extraction rules for www.vox.com
// JavaScript equivalent: export const WwwVoxComExtractor = { ... }
var VoxCustomExtractor = &CustomExtractor{
	Domain: "www.vox.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1.c-page-title",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"author\"]", "value"},
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"figure.e-image--hero", ".c-entry-content"},
				".c-entry-content",
			},
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
		
		// Clean selectors - empty for Vox
		Clean: []string{},
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
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			".p-dek",
		},
	},
	
	NextPageURL: nil,
	
	Excerpt: nil,
}

// transformVoxNoscriptImage handles Vox's lazy-loaded images
// JavaScript equivalent: 'figure .e-image__image noscript': $node => { ... }
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

// GetVoxExtractor returns the Vox custom extractor
func GetVoxExtractor() *CustomExtractor {
	return VoxCustomExtractor
}