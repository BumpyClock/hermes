// ABOUTME: CNN custom extractor with pg-headline title, zn-body-text content with paragraph transforms
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.cnn.com/index.js WwwCnnComExtractor

package custom

import (
	"strings"
	
	"github.com/PuerkitoBio/goquery"
)

// GetCNNExtractor returns the custom extractor for www.cnn.com
func GetCNNExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.cnn.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				"h1.pg-headline",
				"h1",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="author"]`, "value"},
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`meta[name="article:published_time"]`, "value"},
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
					// More specific selector to grab lead image and body
					[]string{".media__video--thumbnail", ".zn-body-text"},
					// Fallback for the above  
					".zn-body-text",
					`div[itemprop="articleBody"]`,
				},
			},
			
			Transforms: map[string]TransformFunction{
				".zn-body__paragraph, .el__leafmedia--sourced-paragraph": &FunctionTransform{
					Fn: func(node *goquery.Selection) error {
						html, _ := node.Html()
						if strings.TrimSpace(html) != "" {
							// Convert to paragraph
							node.ReplaceWithHtml("<p>" + html + "</p>")
							return nil
						}
						return nil
					},
				},
				
				// Clean short, all-link sections linking to related content
				".zn-body__paragraph": &FunctionTransform{
					Fn: func(node *goquery.Selection) error {
						if node.Find("a").Length() > 0 {
							nodeText := strings.TrimSpace(node.Text())
							linkText := strings.TrimSpace(node.Find("a").Text())
							
							// Remove if paragraph text equals link text (likely a related link)
							if nodeText == linkText {
								node.Remove()
							}
						}
						return nil
					},
				},
				
				".media__video--thumbnail": &StringTransform{
					TargetTag: "figure",
				},
			},
			
			Clean: []string{
				// No specific clean selectors in JavaScript version
			},
		},
	}
}