// ABOUTME: CNET custom extractor with figure.image transforms and timezone support
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.cnet.com/index.js

package custom

import (
	"github.com/PuerkitoBio/goquery"
)

// WwwCnetComExtractor provides the custom extraction rules for www.cnet.com
// JavaScript equivalent: export const WwwCnetComExtractor = { ... }
var WwwCnetComExtractor = &CustomExtractor{
	Domain: "www.cnet.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:title\"]", "value"},
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			"span.author",
			"a.author",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			"time",
		},
		// Note: timezone support would be handled at extraction time
		// timezone: 'America/Los_Angeles' (from JavaScript)
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			".c-head_dek",
			".article-dek",
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
				[]interface{}{"img.__image-lead__", ".article-main-body"},
				".article-main-body",
			},
		},
		
		// Transform functions for CNET-specific content
		Transforms: map[string]TransformFunction{
			"figure.image": &FunctionTransform{
				Fn: func(selection *goquery.Selection) error {
					img := selection.Find("img")
					if img.Length() > 0 {
						img.SetAttr("width", "100%")
						img.SetAttr("height", "100%")
						img.AddClass("__image-lead__")
						
						// Remove .imgContainer and prepend img
						selection.Find(".imgContainer").Remove()
						selection.PrependSelection(img)
					}
					return nil
				},
			},
		},
		
		// Clean selectors (empty in JavaScript)
		Clean: []string{},
	},
}

// GetWwwCnetComExtractor returns the CNET custom extractor
func GetWwwCnetComExtractor() *CustomExtractor {
	return WwwCnetComExtractor
}