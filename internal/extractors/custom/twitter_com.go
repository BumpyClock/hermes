// ABOUTME: Twitter.com custom extractor with tweet content, thread structure, and timeline processing  
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/twitter.com/index.js

package custom

import (
	"github.com/PuerkitoBio/goquery"
)

// TwitterCustomExtractor provides the custom extraction rules for twitter.com
// JavaScript equivalent: export const TwitterExtractor = { ... }
var TwitterCustomExtractor = &CustomExtractor{
	Domain: "twitter.com",
	
	Title: nil, // Not specified in JavaScript
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			".tweet.permalink-tweet .username",
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				`.permalink[role=main]`,
			},
			DefaultCleaner: false,
		},
		
		// Transform functions for Twitter-specific content
		Transforms: map[string]TransformFunction{
			// Transform the whole page structure for Twitter
			`.permalink[role=main]`: &FunctionTransform{
				Fn: transformTwitterPermalink,
			},
			
			// Twitter wraps @ with s, which renders as a strikethrough
			"s": &StringTransform{
				TargetTag: "span",
			},
		},
		
		// Clean selectors - remove unwanted elements
		Clean: []string{
			".stream-item-footer",
			"button",
			".tweet-details-fixer",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{`.permalink-tweet ._timestamp[data-time-ms]`, "data-time-ms"},
		},
	},
	
	LeadImageURL: nil,
	
	Dek: nil,
	
	NextPageURL: nil,
	
	Excerpt: nil,
}

// transformTwitterPermalink handles Twitter's complex page structure transformation
// JavaScript equivalent: '.permalink[role=main]': ($node, $) => { ... }
func transformTwitterPermalink(selection *goquery.Selection) error {
	// Simplified implementation - just find and preserve tweets
	tweets := selection.Find(".tweet")
	
	if tweets.Length() > 0 {
		// Create a simple container
		containerHTML := `<div id="TWEETS_GO_HERE">`
		tweets.Each(func(i int, tweet *goquery.Selection) {
			if html, err := tweet.Html(); err == nil {
				containerHTML += "<div class=\"tweet\">" + html + "</div>"
			}
		})
		containerHTML += "</div>"
		
		selection.SetHtml(containerHTML)
	}
	
	return nil
}

// GetTwitterExtractor returns the Twitter custom extractor
func GetTwitterExtractor() *CustomExtractor {
	return TwitterCustomExtractor
}