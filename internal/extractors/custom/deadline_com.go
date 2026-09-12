// ABOUTME: Deadline.com custom extractor for entertainment industry formatting
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/deadline.com/index.js

package custom

import (
	"github.com/PuerkitoBio/goquery"
)

// DeadlineCustomExtractor provides the custom extraction rules for deadline.com
// JavaScript equivalent: export const DeadlineComExtractor = { ... }.
var DeadlineCustomExtractor = &CustomExtractor{
	Domain: "deadline.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "section.author h2"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"article:published_time\"]", Attribute: "value"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"div.a-article-grid__main.pmc-a-grid article.pmc-a-grid-item"},
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
}

// transformDeadlineTwitterEmbed replaces Twitter embeds with their inner HTML
// JavaScript equivalent: '.embed-twitter': $node => { ... }.
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

// GetDeadlineExtractor returns the Deadline.com custom extractor.
func GetDeadlineExtractor() *CustomExtractor {
	return DeadlineCustomExtractor
}
