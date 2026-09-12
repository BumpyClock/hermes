// ABOUTME: Lifehacker Japan lifestyle site custom extractor with complex image URL transform
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.lifehacker.jp/index.js

package custom

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// WwwLifehackerJpExtractor provides the custom extraction rules for www.lifehacker.jp
// JavaScript equivalent: export const WwwLifehackerJpExtractor = { ... }.
var WwwLifehackerJpExtractor = &CustomExtractor{
	Domain: "www.lifehacker.jp",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1[class^=\"article_pArticle_Title\"]"},
			{Selector: "h1.lh-summary-title"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"author\"]", Attribute: "value"},
			{Selector: "p.lh-entryDetailInner--credit"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"article:published_time\"]", Attribute: "value"},
			{Selector: "div.lh-entryDetail-header time", Attribute: "datetime"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"div[class^=\"article_pArticle_Body__\"]"},
			{"div.lh-entryDetail-body"},
		},

		// Transform functions for Lifehacker Japan-specific lazy loaded images
		// JavaScript equivalent: 'img.lazyload': $node => { const src = $node.attr('src'); $node.attr('src', src.replace(/^.*=%27/, '').replace(/%27;$/, '')); }
		Transforms: map[string]TransformFunction{
			"img.lazyload": &FunctionTransform{
				Fn: func(selection *goquery.Selection) error {
					src, exists := selection.Attr("src")
					if exists {
						// Replace pattern: remove ^.*=%27 and %27;$
						// This handles encoded URL patterns used by Lifehacker Japan
						src = strings.ReplaceAll(src, "%27", "'")

						// Remove everything up to ='
						if idx := strings.LastIndex(src, "='"); idx >= 0 {
							src = src[idx+2:]
						}

						// Remove trailing ';
						src = strings.TrimSuffix(src, "';")

						selection.SetAttr("src", src)
					}
					return nil
				},
			},
		},

		// Clean author credit lines
		Clean: []string{
			"p.lh-entryDetailInner--credit",
		},
	},
}

// GetWwwLifehackerJpExtractor returns the Lifehacker Japan custom extractor.
func GetWwwLifehackerJpExtractor() *CustomExtractor {
	return WwwLifehackerJpExtractor
}
