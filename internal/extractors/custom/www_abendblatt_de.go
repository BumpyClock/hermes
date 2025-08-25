// ABOUTME: Abendblatt.de (German newspaper) custom extractor with obfuscated text transforms
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.abendblatt.de/index.js

package custom

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// DeobfuscateAbendblattText handles the complex obfuscation transform for Abendblatt.de
// JavaScript equivalent: complex function in transforms.p and transforms.div
func DeobfuscateAbendblattText(selection *goquery.Selection) *goquery.Selection {
	// Check if element has 'obfuscated' class
	if !selection.HasClass("obfuscated") {
		return selection
	}
	
	text := selection.Text()
	var output strings.Builder
	
	// Port of the JavaScript obfuscation algorithm
	for _, char := range text {
		r := int(char)
		switch r {
		case 177:
			output.WriteString("%")
		case 178:
			output.WriteString("!")
		case 180:
			output.WriteString(";")
		case 181:
			output.WriteString("=")
		case 32:
			output.WriteString(" ")
		case 10:
			output.WriteString("\n")
		default:
			if r > 33 {
				output.WriteRune(rune(r - 1))
			}
		}
	}
	
	// Update the element HTML, remove 'obfuscated' class and add 'deobfuscated' class
	selection.SetHtml(output.String())
	selection.RemoveClass("obfuscated")
	selection.AddClass("deobfuscated")
	
	return selection
}

// WwwAbendblattDeExtractor provides the custom extraction rules for www.abendblatt.de
// JavaScript equivalent: export const WwwAbendblattDeExtractor = { ... }
var WwwAbendblattDeExtractor = &CustomExtractor{
	Domain: "www.abendblatt.de",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h2.article__header__headline",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			"span.author-info__name-text",
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				"div.article__body",
			},
		},
		
		// Complex transform functions for obfuscated content
		Transforms: map[string]TransformFunction{
			"p": &FunctionTransform{
				Fn: func(selection *goquery.Selection) error {
					DeobfuscateAbendblattText(selection)
					return nil
				},
			},
			"div": &FunctionTransform{
				Fn: func(selection *goquery.Selection) error {
					DeobfuscateAbendblattText(selection)
					return nil
				},
			},
		},
		
		// Clean selectors - empty in JavaScript
		Clean: []string{},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"time.teaser-stream-time", "datetime"},
			[]string{"time.article__header__date", "datetime"},
		},
	},
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"description\"]", "value"},
		},
	},
	
	NextPageURL: nil,
	
	Excerpt: nil,
}

// GetWwwAbendblattDeExtractor returns the Abendblatt.de custom extractor
func GetWwwAbendblattDeExtractor() *CustomExtractor {
	return WwwAbendblattDeExtractor
}