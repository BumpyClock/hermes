// ABOUTME: Deadline.com custom extractor for entertainment industry formatting
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/deadline.com/index.js

package custom

import (
	"github.com/PuerkitoBio/goquery"
)

// DeadlineCustomExtractor provides the custom extraction rules for deadline.com
// JavaScript equivalent: export const DeadlineComExtractor = { ... }
var DeadlineCustomExtractor = &CustomExtractor{
	Domain: "deadline.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			"section.author h2",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:published_time\"]", "value"},
		},
	},
	
	// Dek is null in original JavaScript
	Dek: nil,
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				"div.a-article-grid__main.pmc-a-grid article.pmc-a-grid-item",
			},
		},
		
		// Transform functions for Deadline-specific content
		Transforms: map[string]TransformFunction{
			".embed-twitter": &FunctionTransform{
				Fn: transformDeadlineTwitterEmbed,
			},
		},
		
		// Clean selectors - remove figcaptions
		Clean: []string{
			"figcaption",
		},
	},
	
	// No selectors in original JavaScript for these fields
	NextPageURL: &FieldExtractor{
		Selectors: []interface{}{},
	},
	
	Excerpt: &FieldExtractor{
		Selectors: []interface{}{},
	},
}

// transformDeadlineTwitterEmbed replaces Twitter embeds with their inner HTML
// JavaScript equivalent: '.embed-twitter': $node => { ... }
func transformDeadlineTwitterEmbed(selection *goquery.Selection) error {
	// Get inner HTML
	innerHtml, err := selection.Html()
	if err != nil {
		return nil
	}
	
	// Replace with inner HTML
	selection.ReplaceWithHtml(innerHtml)
	
	return nil
}

// GetDeadlineExtractor returns the Deadline.com custom extractor
func GetDeadlineExtractor() *CustomExtractor {
	return DeadlineCustomExtractor
}