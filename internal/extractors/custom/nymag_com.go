// ABOUTME: NY Magazine custom extractor for magazine-style content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/nymag.com/index.js

package custom

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// NYMagCustomExtractor provides the custom extraction rules for nymag.com
// JavaScript equivalent: export const NYMagExtractor = { ... }
var NYMagCustomExtractor = &CustomExtractor{
	Domain: "nymag.com",
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			// Order by most likely. Extractor will stop on first occurrence
			Selectors: []interface{}{
				"div.article-content",
				"section.body",
				"article.article",
			},
		},
		
		// Selectors to remove from the extracted content
		Clean: []string{
			".ad",
			".single-related-story",
		},
		
		// Object of transformations to make on matched elements
		Transforms: map[string]TransformFunction{
			// Convert h1s to h2s
			"h1": &StringTransform{TargetTag: "h2"},
			
			// Convert lazy-loaded noscript images to figures
			"noscript": &FunctionTransform{
				Fn: transformNYMagNoscript,
			},
		},
	},
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1.lede-feature-title",
			"h1.headline-primary",
			"h1",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			".by-authors",
			".lede-feature-author",
		},
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			".lede-feature-teaser",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"time.article-timestamp[datetime]", "datetime"},
			"time.article-timestamp",
		},
	},
	
	// No selectors in original JavaScript for these fields
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{},
	},
	
	NextPageURL: &FieldExtractor{
		Selectors: []interface{}{},
	},
	
	Excerpt: &FieldExtractor{
		Selectors: []interface{}{},
	},
}

// transformNYMagNoscript converts lazy-loaded noscript images to figures
// JavaScript equivalent: noscript: ($node, $) => { ... }
func transformNYMagNoscript(selection *goquery.Selection) error {
	// Get the text content of noscript element which should contain HTML
	noscriptText := selection.Text()
	
	// Parse the noscript content as HTML
	noscriptDoc, err := goquery.NewDocumentFromReader(strings.NewReader(noscriptText))
	if err != nil {
		return nil // Return nil to not perform transformation
	}
	
	// Check if there's exactly one img element
	imgElements := noscriptDoc.Find("img")
	if imgElements.Length() == 1 {
		// Replace the noscript with figure containing the img
		imgHtml, err := imgElements.Html()
		if err != nil {
			return nil
		}
		
		figureHtml := "<figure>" + imgHtml + "</figure>"
		selection.ReplaceWithHtml(figureHtml)
	}
	
	return nil
}

// GetNYMagExtractor returns the NY Magazine custom extractor
func GetNYMagExtractor() *CustomExtractor {
	return NYMagCustomExtractor
}