// ABOUTME: Gizmodo Japan custom extractor with image src replacement transforms
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.gizmodo.jp/index.js

package custom

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// WwwGizmodoJpExtractor provides the custom extraction rules for www.gizmodo.jp
// JavaScript equivalent: export const WwwGizmodoJpExtractor = { ... }
var WwwGizmodoJpExtractor = &CustomExtractor{
	Domain: "www.gizmodo.jp",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1.p-post-title",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			"li.p-post-AssistAuthor",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"li.p-post-AssistTime time", "datetime"},
		},
	},
	
	Dek: nil,
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				"article.p-post",
			},
		},
		
		// Transform functions for Gizmodo Japan-specific content
		Transforms: map[string]TransformFunction{
			"img.p-post-thumbnailImage": &FunctionTransform{
				Fn: func(selection *goquery.Selection) error {
					src, exists := selection.Attr("src")
					if exists {
						// Replace URL pattern: remove ^.*=%27 and %27;$
						src = strings.ReplaceAll(src, "%27", "'")
						if idx := strings.LastIndex(src, "='"); idx >= 0 {
							src = src[idx+2:]
						}
						if strings.HasSuffix(src, "';") {
							src = src[:len(src)-2]
						}
						selection.SetAttr("src", src)
					}
					return nil
				},
			},
		},
		
		// Clean selectors
		Clean: []string{
			"h1.p-post-title",
			"ul.p-post-Assist",
		},
	},
}

// GetWwwGizmodoJpExtractor returns the Gizmodo Japan custom extractor
func GetWwwGizmodoJpExtractor() *CustomExtractor {
	return WwwGizmodoJpExtractor
}