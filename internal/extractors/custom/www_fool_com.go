// ABOUTME: Custom extractor for www.fool.com - The Motley Fool investment and finance site
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.fool.com/index.js WwwFoolComExtractor

package custom

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
)

// GetWwwFoolComExtractor returns the custom extractor for www.fool.com.
func GetWwwFoolComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.fool.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "h1"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"author\"]", Attribute: "value"},
				{Selector: ".author-inline .author-name"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"date\"]", Attribute: "value"},
			},
		},

		Dek: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"og:description\"]", Attribute: "value"},
				{Selector: "header h2"},
			},
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{".tailwind-article-body"},
				{".article-content"},
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
// JavaScript equivalent: '.caption img': $node => { ... }.
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
