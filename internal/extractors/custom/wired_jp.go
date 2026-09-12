// ABOUTME: Wired Japan custom extractor with URL.resolve pattern for data-original images
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/wired.jp/index.js

package custom

import (
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

// WiredJpExtractor provides the custom extraction rules for wired.jp
// JavaScript equivalent: export const WiredJpExtractor = { ... }.
var WiredJpExtractor = &CustomExtractor{
	Domain: "wired.jp",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1[data-testid=\"ContentHeaderHed\"]"},
			{Selector: "h1.post-title"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"article:author\"]", Attribute: "value"},
			{Selector: "p[itemprop=\"author\"]"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"article:published_time\"]", Attribute: "value"},
			{Selector: "time", Attribute: "datetime"},
		},
	},

	Dek: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "div[class^=\"ContentHeaderDek\"]"},
			{Selector: ".post-intro"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"div[data-attribute-verso-pattern=\"article-body\"]"},
			{"article.article-detail"},
		},

		// Transform functions for Wired Japan-specific content
		Transforms: map[string]TransformFunction{
			"img[data-original]": &FunctionTransform{
				Fn: func(selection *goquery.Selection) error {
					dataOriginal, hasDataOriginal := selection.Attr("data-original")
					src, hasSrc := selection.Attr("src")

					if hasDataOriginal && hasSrc {
						// Resolve URL like URL.resolve(src, dataOriginal) in JavaScript
						base, err := url.Parse(src)
						if err != nil {
							return err
						}

						ref, err := url.Parse(dataOriginal)
						if err != nil {
							return err
						}

						resolved := base.ResolveReference(ref)
						selection.SetAttr("src", resolved.String())
					}
					return nil
				},
			},
		},

		// Clean selectors
		Clean: []string{
			".post-category",
			"time",
			"h1.post-title",
			".social-area-syncer",
		},
	},
}

// GetWiredJpExtractor returns the Wired Japan custom extractor.
func GetWiredJpExtractor() *CustomExtractor {
	return WiredJpExtractor
}
