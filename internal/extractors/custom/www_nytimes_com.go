// ABOUTME: New York Times custom extractor with headline selectors, author meta, and g-blocks content
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.nytimes.com/index.js NYTimesExtractor

package custom

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// GetNYTimesExtractor returns the custom extractor for www.nytimes.com.
func GetNYTimesExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.nytimes.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `h1[data-testid="headline"]`},
				{Selector: "h1.g-headline"},
				{Selector: `h1[itemprop="headline"]`},
				{Selector: "h1.headline"},
				{Selector: "h1 .balancedHeadline"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="author"]`, Attribute: "value"},
				{Selector: ".g-byline"},
				{Selector: ".byline"},
				{Selector: `meta[name="byl"]`, Attribute: "value"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{"div.g-blocks"},
				{`section[name="articleBody"]`},
				{"article#story"},
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
			Selectors: []SelectorEntry{
				{Selector: `meta[name="article:published_time"]`, Attribute: "value"},
				{Selector: `meta[name="article:published"]`, Attribute: "value"},
			},
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: `meta[name="og:image"]`, Attribute: "value"},
			},
		},
	}
}
