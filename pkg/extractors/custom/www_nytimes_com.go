// ABOUTME: New York Times custom extractor with headline selectors, author meta, and g-blocks content
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.nytimes.com/index.js NYTimesExtractor

package custom

import (
	"strings"
	
	"github.com/PuerkitoBio/goquery"
)

// GetNYTimesExtractor returns the custom extractor for www.nytimes.com
func GetNYTimesExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.nytimes.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				`h1[data-testid="headline"]`,
				"h1.g-headline",
				`h1[itemprop="headline"]`,
				"h1.headline",
				"h1 .balancedHeadline",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="author"]`, "value"},
				".g-byline",
				".byline",
				[]string{`meta[name="byl"]`, "value"},
			},
		},
		
		Content: &ContentExtractor{
			FieldExtractor: &FieldExtractor{
				Selectors: []interface{}{
					"div.g-blocks",
					`section[name="articleBody"]`,
					"article#story",
				},
			},
			
			Transforms: map[string]TransformFunction{
				"img.g-lazy": &FunctionTransform{
					Fn: func(node *goquery.Selection) error {
						src, exists := node.Attr("src")
						if !exists {
							return nil
						}
						
						// Replace {{size}} placeholder with 640px width
						width := "640"
						src = strings.ReplaceAll(src, "{{size}}", width)
						node.SetAttr("src", src)
						
						return nil
					},
				},
			},
			
			Clean: []string{
				".ad",
				"header#story-header", 
				".story-body-1 .lede.video",
				".visually-hidden",
				"#newsletter-promo",
				".promo",
				".comments-button",
				".hidden",
				".comments",
				".supplemental",
				".nocontent",
				".story-footer-links",
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="article:published_time"]`, "value"},
				[]string{`meta[name="article:published"]`, "value"},
			},
		},
		
		LeadImageURL: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="og:image"]`, "value"},
			},
		},
	}
}