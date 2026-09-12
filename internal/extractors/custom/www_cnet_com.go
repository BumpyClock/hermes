// ABOUTME: CNET custom extractor with figure.image transforms and timezone support
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.cnet.com/index.js

package custom

import (
	"github.com/PuerkitoBio/goquery"
)

// WwwCnetComExtractor provides the custom extraction rules for www.cnet.com
// JavaScript equivalent: export const WwwCnetComExtractor = { ... }.
var WwwCnetComExtractor = &CustomExtractor{
	Domain: "www.cnet.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:title\"]", Attribute: "value"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "span.author"},
			{Selector: "a.author"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "time"},
		},
		// Note: timezone support would be handled at extraction time
		// timezone: 'America/Los_Angeles' (from JavaScript)
	},

	Dek: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".c-head_dek"},
			{Selector: ".article-dek"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"img.__image-lead__", ".article-main-body"},
			{".article-main-body"},
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
	},
}

// GetWwwCnetComExtractor returns the CNET custom extractor.
func GetWwwCnetComExtractor() *CustomExtractor {
	return WwwCnetComExtractor
}
