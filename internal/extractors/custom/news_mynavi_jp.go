// ABOUTME: MyNavi News Japan tech news site custom extractor with data-original image transform
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/news.mynavi.jp/index.js

package custom

import (
	"github.com/PuerkitoBio/goquery"
)

// NewsMynaviJpExtractor provides the custom extraction rules for news.mynavi.jp
// JavaScript equivalent: export const NewsMynaviJpExtractor = { ... }
var NewsMynaviJpExtractor = &CustomExtractor{
	Domain: "news.mynavi.jp",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:title\"]", "value"},
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			"a.articleHeader_name",
			"main div.article-author a.article-author__name",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:published_time\"]", "value"},
		},
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:description\"]", "value"},
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
				"div.article-body",
				"main article div",
			},
		},
		
		// Transform functions for MyNavi News-specific lazy loaded images
		// JavaScript equivalent: img: $node => { const src = $node.attr('data-original'); if (src !== '') { $node.attr('src', src); } }
		Transforms: map[string]TransformFunction{
			"img": &FunctionTransform{
				Fn: func(selection *goquery.Selection) error {
					dataOriginal, exists := selection.Attr("data-original")
					if exists && dataOriginal != "" {
						selection.SetAttr("src", dataOriginal)
					}
					return nil
				},
			},
		},
		
		// clean: [] (empty in JavaScript)
		Clean: []string{},
	},
}

// GetNewsMynaviJpExtractor returns the MyNavi News Japan custom extractor
func GetNewsMynaviJpExtractor() *CustomExtractor {
	return NewsMynaviJpExtractor
}