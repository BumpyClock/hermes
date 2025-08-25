// ABOUTME: Ma.ttias.be (Belgian tech blog) custom extractor with complex header transforms
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/ma.ttias.be/index.js

package custom

import (
	"github.com/PuerkitoBio/goquery"
)

// MaTtiasBeExtractor provides the custom extraction rules for ma.ttias.be
// JavaScript equivalent: export const MaTtiasBeExtractor = { ... }
var MaTtiasBeExtractor = &CustomExtractor{
	Domain: "ma.ttias.be",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"twitter:title\"]", "value"},
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"author\"]", "value"},
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				".content",
			},
		},
		
		// Complex transform functions for ma.ttias.be-specific content structure
		Transforms: map[string]TransformFunction{
			"h2": &FunctionTransform{
				Fn: func(selection *goquery.Selection) error {
					// Remove the "id" attribute to avoid low scores and element removal
					selection.RemoveAttr("id")
					
					// h1 elements will be demoted to h2, so demote h2 elements to h3
					selection.Get(0).Data = "h3"
					return nil
				},
			},
			"h1": &FunctionTransform{
				Fn: func(selection *goquery.Selection) error {
					// Remove the "id" attribute to avoid low scores and element removal
					selection.RemoveAttr("id")
					
					// Add empty paragraph after h1 to prevent h2 removal
					selection.AfterHtml("<p></p>")
					return nil
				},
			},
			"ul": &FunctionTransform{
				Fn: func(selection *goquery.Selection) error {
					// Add class to avoid lists being incorrectly removed as navigation
					selection.AddClass("entry-content-asset")
					return nil
				},
			},
		},
		
		// No clean selectors in JavaScript
		Clean: []string{},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:published_time\"]", "value"},
		},
	},
	
	// No explicit selectors for other fields in JavaScript
	LeadImageURL: nil,
	Dek:          nil,
	NextPageURL:  nil,
	Excerpt:      nil,
}

// GetMaTtiasBeExtractor returns the ma.ttias.be custom extractor
func GetMaTtiasBeExtractor() *CustomExtractor {
	return MaTtiasBeExtractor
}