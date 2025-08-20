// ABOUTME: Custom extractor for www.fool.com - The Motley Fool investment and finance site
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.fool.com/index.js WwwFoolComExtractor

package custom

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
)

// GetWwwFoolComExtractor returns the custom extractor for www.fool.com
func GetWwwFoolComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.fool.com",
		
		Title: &FieldExtractor{
			Selectors: []interface{}{
				"h1",
			},
		},
		
		Author: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[name=\"author\"]", "value"},
				".author-inline .author-name",
			},
		},
		
		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[name=\"date\"]", "value"},
			},
		},
		
		Dek: &FieldExtractor{
			Selectors: []interface{}{
				[]string{"meta[name=\"og:description\"]", "value"},
				"header h2",
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
					".tailwind-article-body",
					".article-content",
				},
			},
			
			Transforms: map[string]TransformFunction{
				// Complex transform for caption images -> figure
				".caption img": &FunctionTransform{
					Fn: transformFoolCaptionImg,
				},
				// Simple transform for captions -> figcaptions
				".caption": &StringTransform{TargetTag: "figcaption"},
			},
			
			Clean: []string{
				"#pitch",
			},
		},
	}
}

// transformFoolCaptionImg converts .caption img to figure with img
// JavaScript equivalent: '.caption img': $node => { ... }
func transformFoolCaptionImg(selection *goquery.Selection) error {
	src, exists := selection.Attr("src")
	if !exists {
		return nil
	}
	
	figureHtml := fmt.Sprintf(`<figure><img src="%s"/></figure>`, src)
	parent := selection.Parent()
	parent.ReplaceWithHtml(figureHtml)
	
	return nil
}