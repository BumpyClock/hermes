// ABOUTME: BuzzFeed.com custom extractor with list article support, quiz content, and buzz-specific transforms
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.buzzfeed.com/index.js

package custom

import (
	"github.com/PuerkitoBio/goquery"
)

// BuzzFeedCustomExtractor provides the custom extraction rules for www.buzzfeed.com
// JavaScript equivalent: export const BuzzfeedExtractor = { ... }
var BuzzFeedCustomExtractor = &CustomExtractor{
	Domain: "www.buzzfeed.com",
	
	SupportedDomains: []string{"www.buzzfeednews.com"},
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1.embed-headline-title",
			// enter title selectors
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			`a[data-action="user/username"]`,
			"byline__author",
			[]string{"meta[name=\"author\"]", "value"},
			// enter author selectors
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				[]string{`div[class^="featureimage_featureImageWrapper"]`, ".js-subbuzz-wrapper"},
				[]string{".js-subbuzz-wrapper"},
			},
			DefaultCleaner: false,
		},
		
		// Transform functions for BuzzFeed-specific content
		Transforms: map[string]TransformFunction{
			// Transform h2 to b for BuzzFeed styling
			"h2": &StringTransform{
				TargetTag: "b",
			},
			
			// Handle BuzzFeed longform custom header media
			"div.longform_custom_header_media": &FunctionTransform{
				Fn: transformBuzzFeedHeaderMedia,
			},
			
			// Transform longform header image source to figcaption
			"figure.longform_custom_header_media .longform_header_image_source": &StringTransform{
				TargetTag: "figcaption",
			},
		},
		
		// Clean selectors - remove unwanted elements
		Clean: []string{
			".instapaper_ignore",
			".suplist_list_hide .buzz_superlist_item .buzz_superlist_number_inline",
			".share-box",
			".print",
			".js-inline-share-bar",
			".js-ad-placement",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"time[datetime]", "datetime"},
		},
	},
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			".embed-headline-description",
		},
	},
	
	NextPageURL: nil,
	
	Excerpt: nil,
}

// transformBuzzFeedHeaderMedia handles BuzzFeed longform custom header media
// JavaScript equivalent: 'div.longform_custom_header_media': $node => { ... }
func transformBuzzFeedHeaderMedia(selection *goquery.Selection) error {
	hasImg := selection.Find("img").Length() > 0
	hasSource := selection.Find(".longform_header_image_source").Length() > 0
	
	if hasImg && hasSource {
		// Convert to figure element
		html, err := selection.Html()
		if err != nil {
			return err
		}
		
		selection.ReplaceWithHtml("<figure>" + html + "</figure>")
		return nil
	}
	
	// Return null equivalent - remove the element
	selection.Remove()
	return nil
}

// GetBuzzFeedExtractor returns the BuzzFeed custom extractor
func GetBuzzFeedExtractor() *CustomExtractor {
	return BuzzFeedCustomExtractor
}