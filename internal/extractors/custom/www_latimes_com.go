// ABOUTME: LA Times custom extractor with headline h1, standardBylineAuthorName, and page-article-body with trb_ar_la transforms
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.latimes.com/index.js WwwLatimesComExtractor

package custom

import (
	"github.com/PuerkitoBio/goquery"
)

// GetLATimesExtractor returns the custom extractor for www.latimes.com.
func GetLATimesExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.latimes.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "h1.headline"},
				{Selector: ".trb_ar_hl"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `a[data-click="standardBylineAuthorName"]`},
				{Selector: `meta[name="author"]`, Attribute: "value"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="article:published_time"]`, Attribute: "value"},
				{Selector: `meta[itemprop="datePublished"]`, Attribute: "value"},
			},
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="og:image"]`, Attribute: "value"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{".page-article-body"},
				{".trb_ar_main"},
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
