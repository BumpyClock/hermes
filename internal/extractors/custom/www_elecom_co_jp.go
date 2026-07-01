// ABOUTME: Custom extractor for www.elecom.co.jp - Japanese electronics company press releases
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.elecom.co.jp/index.js WwwElecomCoJpExtractor

package custom

import (
	"github.com/PuerkitoBio/goquery"
)

// GetWwwElecomCoJpExtractor returns the custom extractor for www.elecom.co.jp.
func GetWwwElecomCoJpExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.elecom.co.jp",

		Title: &FieldExtractor{
			Selectors: []interface{}{
				"title",
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []interface{}{
				"p.section-last",
			},
			// Note: format: 'YYYY.MM.DD' and timezone: 'Asia/Tokyo' are handled by date cleaner in Go version
		},

		Content: &ContentExtractor{
			FieldExtractor: &FieldExtractor{
				Selectors: []interface{}{
					"td.TableMain2",
				},
				DefaultCleaner: false, // Explicit defaultCleaner: false from JavaScript
			},

			Transforms: map[string]TransformFunction{
				// Transform table to set width=auto
				"table": &FunctionTransform{
					Fn: transformElecomTableWidth,
				},
			},
		},
	}
}

// transformElecomTableWidth sets table width to auto
// JavaScript equivalent: table: $node => { $node.attr('width', 'auto'); }.
func transformElecomTableWidth(selection *goquery.Selection) error {
	selection.SetAttr("width", "auto")
	return nil
}
