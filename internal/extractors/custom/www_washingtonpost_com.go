// ABOUTME: Washington Post custom extractor with h1 title selectors, pb-author-name, and article-body content
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.washingtonpost.com/index.js WwwWashingtonpostComExtractor

package custom

import (
	"github.com/PuerkitoBio/goquery"
)

// GetWashingtonPostExtractor returns the custom extractor for www.washingtonpost.com.
func GetWashingtonPostExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.washingtonpost.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "#topper-text-elems h1"},
				{Selector: "article header h1"},
				{Selector: "h1"},
				{Selector: "#topper-headline-wrapper"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="author"]`, Attribute: "value"},
				{Selector: ".pb-author-name"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `.author-timestamp[itemprop="datePublished"]`, Attribute: "content"},
			},
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="og:image"]`, Attribute: "value"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{".article-body"},
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
