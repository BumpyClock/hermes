// ABOUTME: Medium.com custom extractor with transforms, selectors, and content cleaning
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/medium.com/index.js

package custom

import (
	"regexp"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

// MediumCustomExtractor provides the custom extraction rules for Medium.com
// JavaScript equivalent: export const MediumExtractor = { ... }
var MediumCustomExtractorFixed = &CustomExtractor{
	Domain: "medium.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1",
			[]string{"meta[name=\"og:title\"]", "value"},
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"author\"]", "value"},
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{"article"},
		},
		
		// Clean selectors - remove unwanted elements
		Clean: []string{"span a", "svg"},
		
		// Transform functions for Medium-specific content
		Transforms: map[string]TransformFunction{
			// Allow drop cap character
			"section span:first-of-type": &FunctionTransform{
				Fn: func(selection *goquery.Selection) error {
					text := selection.Text()
					if len(text) == 1 && regexp.MustCompile(`^[a-zA-Z()]+$`).MatchString(text) {
						selection.ReplaceWith(text)
					}
					return nil
				},
			},
			
			// Remove smaller images (author photo 48px, leading sentence images 79px, etc.)
			"img": &FunctionTransform{
				Fn: transformMediumImageFixed,
			},
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:published_time\"]", "value"},
		},
	},
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
	
	// Dek is null in original JavaScript
	Dek: nil,
	
	NextPageURL: &FieldExtractor{
		Selectors: []interface{}{
			// Empty in JavaScript original
		},
	},
	
	Excerpt: &FieldExtractor{
		Selectors: []interface{}{
			// Empty in JavaScript original
		},
	},
}

// transformMediumImageFixed removes smaller images that didn't get caught by generic cleaner
// JavaScript equivalent: img: $node => { ... }
func transformMediumImageFixed(selection *goquery.Selection) error {
	widthStr, exists := selection.Attr("width")
	if !exists {
		return nil
	}
	
	width, err := strconv.Atoi(widthStr)
	if err != nil {
		return nil
	}
	
	if width < 100 {
		selection.Remove()
	}
	
	return nil
}

// GetMediumExtractorFixed returns the Medium custom extractor
func GetMediumExtractorFixed() *CustomExtractor {
	return MediumCustomExtractorFixed
}