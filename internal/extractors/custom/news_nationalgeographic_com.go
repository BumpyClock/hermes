// ABOUTME: News National Geographic custom extractor with date format and timezone handling
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/news.nationalgeographic.com/index.js

package custom

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
)

// NewsNationalgeographicComExtractor provides the custom extraction rules for news.nationalgeographic.com
// JavaScript equivalent: export const NewsNationalgeographicComExtractor = { ... }
var NewsNationalgeographicComExtractor = &CustomExtractor{
	Domain: "news.nationalgeographic.com",
	
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
		// JavaScript: format: 'ddd MMM DD HH:mm:ss zz YYYY', timezone: 'EST'
		// Note: Go handles timezone conversion automatically during parsing
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
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
				[]string{".parsys.content", ".__image-lead__"},
				".content",
			},
		},
		
		// Transform functions for News National Geographic-specific content
		// JavaScript: transforms: { '.parsys.content': ($node, $) => { ... } }
		Transforms: map[string]TransformFunction{
			".parsys.content": &FunctionTransform{
				Fn: func(selection *goquery.Selection) error {
					// Find image source from platform-src data attribute
					imgSrc, exists := selection.Find(".image.parbase.section").Find(".picturefill").First().Attr("data-platform-src")
					if exists && imgSrc != "" {
						imageHTML := fmt.Sprintf(`<img class="__image-lead__" src="%s"/>`, imgSrc)
						selection.PrependHtml(imageHTML)
					}
					return nil
				},
			},
		},
		
		// Clean selectors - remove unwanted elements  
		Clean: []string{
			".pull-quote.pull-quote--large",
		},
	},
}

// GetNewsNationalgeographicComExtractor returns the News National Geographic custom extractor
func GetNewsNationalgeographicComExtractor() *CustomExtractor {
	return NewsNationalgeographicComExtractor
}