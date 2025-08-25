// ABOUTME: National Geographic custom extractor with complex image transform handling
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.nationalgeographic.com/index.js

package custom

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
)

// WwwNationalgeographicComExtractor provides the custom extraction rules for www.nationalgeographic.com
// JavaScript equivalent: export const WwwNationalgeographicComExtractor = { ... }
var WwwNationalgeographicComExtractor = &CustomExtractor{
	Domain: "www.nationalgeographic.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1",
			"h1.main-title",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			".byline-component__contributors b span",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:published_time\"]", "value"},
		},
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			".Article__Headline__Desc",
			".article__deck",
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
				"section.Article__Content",
				[]string{".parsys.content", ".__image-lead__"},
				".content",
			},
		},
		
		// Transform functions for National Geographic-specific content
		// JavaScript: transforms: { '.parsys.content': ($node, $) => { ... } }
		Transforms: map[string]TransformFunction{
			".parsys.content": &FunctionTransform{
				Fn: func(selection *goquery.Selection) error {
					// Check if first child is imageGroup
					imageParent := selection.Children().First()
					if imageParent.HasClass("imageGroup") {
						// Complex image data extraction from platform data attributes
						dataAttrContainer := imageParent.Find(".media--medium__container").Children().First()
						imgPath1, exists1 := dataAttrContainer.Attr("data-platform-image1-path")
						imgPath2, exists2 := dataAttrContainer.Attr("data-platform-image2-path")
						
						if exists1 && exists2 && imgPath1 != "" && imgPath2 != "" {
							// Prepend both images as lead content
							imageHTML := fmt.Sprintf(`<div class="__image-lead__">
								<img src="%s"/>
								<img src="%s"/>
							</div>`, imgPath1, imgPath2)
							selection.PrependHtml(imageHTML)
						}
					} else {
						// Find single image source from platform-src data
						imgSrc, exists := selection.Find(".image.parbase.section").Find(".picturefill").First().Attr("data-platform-src")
						if exists && imgSrc != "" {
							imageHTML := fmt.Sprintf(`<img class="__image-lead__" src="%s"/>`, imgSrc)
							selection.PrependHtml(imageHTML)
						}
					}
					return nil
				},
			},
		},
		
		// Clean selectors - remove unwanted elements
		Clean: []string{
			".pull-quote.pull-quote--small",
		},
	},
}

// GetWwwNationalgeographicComExtractor returns the National Geographic custom extractor
func GetWwwNationalgeographicComExtractor() *CustomExtractor {
	return WwwNationalgeographicComExtractor
}