// ABOUTME: Custom extractor for www.ndtv.com - New Delhi Television news site
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.ndtv.com/index.js WwwNdtvComExtractor

package custom

import (
	"github.com/PuerkitoBio/goquery"
)

// GetWwwNdtvComExtractor returns the custom extractor for www.ndtv.com
func GetWwwNdtvComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.ndtv.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[name=\"og:title\"]", "value"},
				"h1.entry-title",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				"span[itemprop=\"author\"] span[itemprop=\"name\"]",
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"span[itemprop=\"dateModified\"]", "content"},
			},
		},
		
		Dek: &FieldExtractor{
			Selectors: []interface{}{
				"h2",
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
					"div[itemprop=\"articleBody\"]",
				},
			},
			
			Transforms: map[string]TransformFunction{
				// Complex transform for dateline handling
				".place_cont": &FunctionTransform{
					Fn: transformNDTVDateline,
				},
			},
			
			Clean: []string{
				".highlghts_Wdgt",
				".ins_instory_dv_caption",
				"input",
				"._world-wrapper .mt20",
			},
		},
	}
}

// transformNDTVDateline moves dateline from b tag to first paragraph
// JavaScript equivalent: '.place_cont': $node => { ... }
func transformNDTVDateline(selection *goquery.Selection) error {
	// Check if the node is not already inside a paragraph
	parents := selection.ParentsFiltered("p")
	if parents.Length() == 0 {
		// Find next sibling paragraph
		nextSibling := selection.Next()
		if nextSibling.Length() > 0 {
			// Get the HTML content of the current node
			nodeHtml, err := selection.Html()
			if err != nil {
				return nil
			}
			
			// Remove the current node
			selection.Remove()
			
			// Prepend to the next paragraph
			nextSibling.PrependHtml(nodeHtml)
		}
	}
	
	return nil
}