// ABOUTME: Washington Post custom extractor with h1 title selectors, pb-author-name, and article-body content
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.washingtonpost.com/index.js WwwWashingtonpostComExtractor

package custom

import (
	"github.com/PuerkitoBio/goquery"
)

// GetWashingtonPostExtractor returns the custom extractor for www.washingtonpost.com
func GetWashingtonPostExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.washingtonpost.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				"h1",
				"#topper-headline-wrapper",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				".pb-author-name",
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`.author-timestamp[itemprop="datePublished"]`, "content"},
			},
		},
		
		Dek: &FieldExtractor{
			Selectors: []interface{}{},
		},
		
		LeadImageURL: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="og:image"]`, "value"},
			},
		},
		
		Content: &ContentExtractor{
			FieldExtractor: &FieldExtractor{
				Selectors: []interface{}{
					".article-body",
				},
			},
			
			Transforms: map[string]TransformFunction{
				"div.inline-content": &FunctionTransform{
					Fn: func(node *goquery.Selection) error {
						// Check if node contains img, iframe, or video elements
						if node.Find("img,iframe,video").Length() > 0 {
							// Convert to figure element
							content, _ := node.Html()
							node.ReplaceWithHtml("<figure>" + content + "</figure>")
							return nil
						}
						
						// Remove node if it doesn't contain media
						node.Remove()
						return nil
					},
				},
				".pb-caption": &StringTransform{
					TargetTag: "figcaption",
				},
			},
			
			Clean: []string{
				".interstitial-link",
				".newsletter-inline-unit",
			},
		},
	}
}