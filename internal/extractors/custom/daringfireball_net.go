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
		
		// Clean out navigation and non-content elements
		Clean: []string{
			// Remove Daring Fireball specific structure elements
			"div#Banner",    // Top banner/logo
			"div#Sidebar",   // Left sidebar navigation
			"div#Footer",    // Footer
			"#SidebarMartini", // Sponsored app section
			
			// Remove navigation elements
			"nav",
			"ul",            // Navigation menu lists
			"div#Sidebar ul", // Remove sidebar navigation list specifically
			
			// Remove header/logo elements
			"a[title*='Daring Fireball']", // Logo links
			"img[alt*='Daring Fireball']", // Logo images
			
			// Remove byline and author info in header
			"p:contains('By John Gruber')", // Author byline in header
			
			// Remove metadata and footer elements
			".smallprint", // Date stamps and footer info that appear at end
			"div#Footer",  // Footer content
			"[href='/preferences/']", // Footer preference links  
			"a[href='/preferences/']", // Display Preferences links
			"em", // Date stamps are typically in <em> tags
			"p:last-child", // Last paragraph often contains copyright
			"div#Main > p:last-child", // Last paragraph in main content
			"div#Main > p:last-of-type", // Last paragraph of its type
			
			// Remove advertisements
			"[href*='apps.apple.com']", // App store links
			"img[src*='/martini/']", // Martini ad images
			"a:contains('Walk the World')", // Specific app ads
			
			// Remove scripts and styles
			"script",
			"style",
			"noscript",
			
			// Remove ads and sponsored content
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
					if strings.Contains(text, "★ _") || 
					   strings.Contains(text, "Display Preferences") || 
					   strings.Contains(text, "Copyright ©") {
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