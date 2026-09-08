// ABOUTME: Daring Fireball custom extractor for John Gruber's blog
// ABOUTME: Handles linked posts and regular articles with proper content extraction
// ABOUTME: Addresses truncation issues with bullet points and full content capture

package custom

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// DaringFireballExtractor provides the custom extraction rules for daringfireball.net.
var DaringFireballExtractor = &CustomExtractor{
	Domain: "daringfireball.net",
	SupportedDomains: []string{
		"www.daringfireball.net",
	},

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "title"},
			{Selector: "h1"},
			{Selector: "h2.entry-title"},
			{Selector: "h1.entry-title"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "[name='author']"},
			{Selector: ".author"},
			{Selector: ".byline"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "time", Attribute: "datetime"},
			{Selector: "[datetime]", Attribute: "datetime"},
			{Selector: "p.smallprint em"}, // Format: "★ Saturday, 30 August 2025"
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			// Target the main content area specifically
			{"div#Main"}, // Daring Fireball uses div#Main for article content
			{".main-content"},
			{"main"},
			{"article"},
			{"body"}, // Last fallback
		},

		// Clean out navigation and non-content elements (conservative)
		Clean: []string{
			// Page furniture outside main content
			"div#Banner",
			"div#Sidebar",
			"div#Footer",
			"#SidebarMartini",

			// Footer/date blocks within content
			".smallprint",

			// Scripts and styles
			"script",
			"style",
			"noscript",

			// Ads/sponsored
			"[href*='apps.apple.com']",
			"img[src*='/martini/']",
			".ads",
			".advertisement",
			".sponsored",
		},

		// Basic transforms - preserve important formatting
		Transforms: map[string]TransformFunction{
			// Remove footer elements that contain specific text patterns.
			"p": &FunctionTransform{
				Fn: func(selection *goquery.Selection) error {
					text := selection.Text()
					if strings.Contains(text, "★") ||
						strings.Contains(text, "Display Preferences") ||
						strings.Contains(text, "Copyright") {
						selection.Remove()
					}
					return nil
				},
			},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[property='og:image']", Attribute: "content"},
			{Selector: "meta[name='twitter:image']", Attribute: "content"},
			{Selector: "meta[name='og:image']", Attribute: "content"},
		},
	},
}

// GetDaringFireballExtractor returns the Daring Fireball custom extractor.
func GetDaringFireballExtractor() *CustomExtractor {
	return DaringFireballExtractor
}
