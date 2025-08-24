// ABOUTME: LA Times custom extractor with headline h1, standardBylineAuthorName, and page-article-body with trb_ar_la transforms
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.latimes.com/index.js WwwLatimesComExtractor

package custom

import (
	"github.com/PuerkitoBio/goquery"
)

// GetLATimesExtractor returns the custom extractor for www.latimes.com
func GetLATimesExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.latimes.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				"h1.headline",
				".trb_ar_hl",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				`a[data-click="standardBylineAuthorName"]`,
				[]string{`meta[name="author"]`, "value"},
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="article:published_time"]`, "value"},
				[]string{`meta[itemprop="datePublished"]`, "value"},
			},
		},
		
		LeadImageURL: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="og:image"]`, "value"},
			},
		},
		
		Content: &ContentExtractor{
			FieldExtractor: &FieldExtractor{
				Selectors: []interface{}{
					".page-article-body",
					".trb_ar_main",
				},
			},
			
			Transforms: map[string]TransformFunction{
				".trb_ar_la": &FunctionTransform{
					Fn: func(node *goquery.Selection) error {
						// Find figure element and replace node with it
						figure := node.Find("figure")
						if figure.Length() > 0 {
							figureHtml, _ := figure.Html()
							node.ReplaceWithHtml("<figure>" + figureHtml + "</figure>")
						}
						return nil
					},
				},
			},
			
			Clean: []string{
				".trb_ar_by",
				".trb_ar_cr",
			},
		},
	}
}