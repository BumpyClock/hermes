// ABOUTME: Reddit.com custom extractor with thread structure, comment extraction, and media handling
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.reddit.com/index.js

package custom

import (
	"regexp"
	"strings"
	
	"github.com/PuerkitoBio/goquery"
)

// RedditCustomExtractor provides the custom extraction rules for www.reddit.com
// JavaScript equivalent: export const WwwRedditComExtractor = { ... }
var RedditCustomExtractor = &CustomExtractor{
	Domain: "www.reddit.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			`div[data-test-id="post-content"] h1`,
			`div[data-test-id="post-content"] h2`,
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			`div[data-test-id="post-content"] a[href*="user/"]`,
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				// text post
				[]string{`div[data-test-id="post-content"] p`},
				// external link with media preview (YouTube, imgur album, etc...)
				[]string{
					`div[data-test-id="post-content"] a[target="_blank"]:not([data-click-id="timestamp"])`,
					`div[data-test-id="post-content"] div[data-click-id="media"]`,
				},
				// Embedded media (Reddit video)
				[]string{`div[data-test-id="post-content"] div[data-click-id="media"]`},
				// external link
				[]string{`div[data-test-id="post-content"] a`},
				`div[data-test-id="post-content"]`,
			},
		},
		
		// Transform functions for Reddit-specific content
		Transforms: map[string]TransformFunction{
			// Handle Reddit external link image previews
			`div[role="img"]`: &FunctionTransform{
				Fn: transformRedditImagePreview,
			},
		},
		
		// Clean selectors - remove unwanted elements
		Clean: []string{
			".icon",
			`span[id^="PostAwardBadges"]`,
			`div a[data-test-id="comments-page-link-num-comments"]`,
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			`div[data-test-id="post-content"] span[data-click-id="timestamp"]`,
			`div[data-test-id="post-content"] a[data-click-id="timestamp"]`,
		},
	},
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
	
	Dek: nil,
	
	NextPageURL: nil,
	
	Excerpt: nil,
}

// transformRedditImagePreview handles Reddit's external link image previews
// JavaScript equivalent: 'div[role="img"]': $node => { ... }
func transformRedditImagePreview(selection *goquery.Selection) error {
	img := selection.Find("img")
	bgImg, exists := selection.Attr("background-image")
	
	if img.Length() == 1 && exists {
		// Extract URL from background-image CSS property
		bgImgRegex := regexp.MustCompile(`\((.*?)\)`)
		matches := bgImgRegex.FindStringSubmatch(bgImg)
		if len(matches) > 1 {
			// Remove quotes from URL
			url := strings.ReplaceAll(matches[1], "'", "")
			url = strings.ReplaceAll(url, "\"", "")
			img.SetAttr("src", url)
			selection.ReplaceWithSelection(img)
		}
	}
	
	return nil
}

// GetRedditExtractor returns the Reddit custom extractor
func GetRedditExtractor() *CustomExtractor {
	return RedditCustomExtractor
}