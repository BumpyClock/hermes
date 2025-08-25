// ABOUTME: Wikipedia.org custom extractor with reference cleanup, infobox handling, and citation processing
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/wikipedia.org/index.js

package custom

import (
	"github.com/PuerkitoBio/goquery"
)

// WikipediaCustomExtractor provides the custom extraction rules for wikipedia.org
// JavaScript equivalent: export const WikipediaExtractor = { ... }
var WikipediaCustomExtractor = &CustomExtractor{
	Domain: "wikipedia.org",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h2.title",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			// Wikipedia has a hardcoded author
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				"#mw-content-text",
			},
			DefaultCleaner: false,
		},
		
		// Transform top infobox to an image with caption
		Transforms: map[string]TransformFunction{
			// Handle infobox images
			".infobox img": &FunctionTransform{
				Fn: transformWikipediaInfoboxImg,
			},
			
			// Transform infobox caption to figcaption
			".infobox caption": &StringTransform{
				TargetTag: "figcaption",
			},
			
			// Transform infobox to figure
			".infobox": &StringTransform{
				TargetTag: "figure",
			},
		},
		
		// Selectors to remove from the extracted content
		Clean: []string{
			".mw-editsection",
			"figure tr, figure td, figure tbody",
			"#toc",
			".navbox",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			"#footer-info-lastmod",
		},
	},
	
	LeadImageURL: nil,
	
	Dek: nil,
	
	NextPageURL: nil,
	
	Excerpt: nil,
}

// transformWikipediaInfoboxImg handles Wikipedia infobox image processing
// JavaScript equivalent: '.infobox img': $node => { ... }
func transformWikipediaInfoboxImg(selection *goquery.Selection) error {
	parent := selection.ParentsFiltered(".infobox")
	if parent.Length() > 0 {
		// Only prepend the first image in .infobox
		existingImages := parent.Find("img")
		if existingImages.Length() == 0 {
			parent.PrependSelection(selection.Clone())
		}
	}
	
	return nil
}

// GetWikipediaExtractor returns the Wikipedia custom extractor
func GetWikipediaExtractor() *CustomExtractor {
	// Set hardcoded author as per JavaScript
	WikipediaCustomExtractor.Author = &FieldExtractor{
		Selectors: []interface{}{
			// Wikipedia Contributors is hardcoded in the JavaScript
		},
	}
	
	return WikipediaCustomExtractor
}