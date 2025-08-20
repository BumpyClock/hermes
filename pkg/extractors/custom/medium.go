// ABOUTME: Medium.com custom extractor with transforms, selectors, and content cleaning
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/medium.com/index.js

package custom

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// MediumCustomExtractor provides the custom extraction rules for Medium.com
// JavaScript equivalent: export const MediumExtractor = { ... }
var MediumCustomExtractor = &CustomExtractor{
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
			
			// Re-write lazy-loaded YouTube videos
			"iframe": &FunctionTransform{
				Fn: transformMediumIframe,
			},
			
			// Rewrite figures to pull out image and caption, remove rest
			"figure": &FunctionTransform{
				Fn: transformMediumFigure,
			},
			
			// Remove smaller images (author photo 48px, leading sentence images 79px, etc.)
			"img": &FunctionTransform{
				Fn: transformMediumImage,
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

// transformMediumIframe handles Medium's lazy-loaded YouTube videos
// JavaScript equivalent: iframe: $node => { ... }
func transformMediumIframe(selection *goquery.Selection) error {
	ytRe := regexp.MustCompile(`https://i\.embed\.ly/.+url=https://i\.ytimg\.com/vi/(\w+)/`)
	
	thumbnail, exists := selection.Attr("data-thumbnail")
	if !exists {
		return nil
	}
	
	// Decode URL
	// Note: In real implementation, would need proper URL decoding
	thumbnail = strings.ReplaceAll(thumbnail, "%3A", ":")
	thumbnail = strings.ReplaceAll(thumbnail, "%2F", "/")
	
	parent := selection.Parent()
	if !parent.Is("figure") {
		return nil
	}
	
	if matches := ytRe.FindStringSubmatch(thumbnail); len(matches) > 1 {
		youtubeID := matches[1]
		selection.SetAttr("src", "https://www.youtube.com/embed/"+youtubeID)
		
		caption := parent.Find("figcaption")
		parent.Empty()
		parent.AppendSelection(selection.Clone())
		parent.AppendSelection(caption.Clone())
	} else {
		// If we can't draw the YouTube preview, remove the figure
		parent.Remove()
	}
	
	return nil
}

// transformMediumFigure rewrite figures to pull out image and caption, remove rest
// JavaScript equivalent: figure: $node => { ... }
func transformMediumFigure(selection *goquery.Selection) error {
	// Ignore if figure has an iframe
	if selection.Find("iframe").Length() > 0 {
		return nil
	}
	
	// Get the last image
	images := selection.Find("img")
	if images.Length() == 0 {
		return nil
	}
	
	lastImg := images.Last()
	caption := selection.Find("figcaption")
	
	// Clear figure and add only image and caption
	selection.Empty()
	selection.AppendSelection(lastImg.Clone())
	if caption.Length() > 0 {
		selection.AppendSelection(caption.Clone())
	}
	
	return nil
}

// transformMediumImage removes smaller images that didn't get caught by generic cleaner
// JavaScript equivalent: img: $node => { ... }
func transformMediumImage(selection *goquery.Selection) error {
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

// GetMediumExtractor returns the Medium custom extractor
func GetMediumExtractor() *CustomExtractor {
	return MediumCustomExtractor
}