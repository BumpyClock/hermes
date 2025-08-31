// ABOUTME: Daring Fireball custom extractor for John Gruber's blog
// ABOUTME: Handles linked posts and regular articles with proper content extraction
// ABOUTME: Addresses truncation issues with bullet points and full content capture

package custom

import (
	"strings"
	
	"github.com/PuerkitoBio/goquery"
)

// DaringFireballExtractor provides the custom extraction rules for daringfireball.net
var DaringFireballExtractor = &CustomExtractor{
	Domain: "daringfireball.net",
	SupportedDomains: []string{
		"www.daringfireball.net",
	},
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"title",
			"h1",
			"h2.entry-title",
			"h1.entry-title",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			"[name='author']",
			".author",
			".byline",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"time", "datetime"},
			[]string{"[datetime]", "datetime"},
			"p.smallprint em", // Format: "★ Saturday, 30 August 2025"
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				// Target the main content area specifically
				"div#Main", // Daring Fireball uses div#Main for article content
				".main-content",
				"main",
				"article",
				"body", // Last fallback
			},
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
            "a[href='/preferences/']",

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
			// Remove footer elements that contain specific text patterns
            "p": &FunctionTransform{
                Fn: func(selection *goquery.Selection) error {
                    text := selection.Text()
                    // Remove paragraphs containing footer text
                    if strings.Contains(text, "★") ||
                       strings.Contains(text, "Display Preferences") ||
                       strings.Contains(text, "Copyright") {
                        selection.Remove()
                    }
                    return nil
                },
            },
			// Remove links to preferences
			"a": &FunctionTransform{
				Fn: func(selection *goquery.Selection) error {
					href, exists := selection.Attr("href")
					if exists && strings.Contains(href, "/preferences/") {
						selection.Remove()
					}
					return nil
				},
			},
		},
	},
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[property='og:image']", "content"},
			[]string{"meta[name='twitter:image']", "content"},
			[]string{"meta[name='og:image']", "content"},
		},
	},
}

// GetDaringFireballExtractor returns the Daring Fireball custom extractor
func GetDaringFireballExtractor() *CustomExtractor {
	return DaringFireballExtractor
}
