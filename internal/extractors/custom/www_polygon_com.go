// ABOUTME: Polygon.com custom extractor for gaming and entertainment content
// ABOUTME: 100% JavaScript-compatible port optimized for Polygon's WordPress-based structure

package custom

import (
	"github.com/PuerkitoBio/goquery"
)

// PolygonExtractor provides the custom extraction rules for www.polygon.com
var PolygonExtractor = &CustomExtractor{
	Domain: "www.polygon.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1.article-header-title",
			"h1[class*='article']",
			"h1",
			// Meta tag fallback
			[]string{"meta[property=\"og:title\"]", "content"},
			[]string{"meta[name=\"twitter:title\"]", "content"},
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			".article-author",
			".meta_txt.article-author",
			".w-author-name .article-author", 
			// Meta tag fallbacks
			[]string{"meta[name=\"author\"]", "content"},
			[]string{"meta[property=\"article:author\"]", "content"},
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[property=\"article:published_time\"]", "content"},
			[]string{"meta[property=\"og:published_time\"]", "content"},
			".article-date",
			".meta_txt.article-date",
		},
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			"header p", // Polygon's subtitle in header
			".article-excerpt",
			[]string{"meta[name=\"description\"]", "content"},
		},
	},
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[property=\"og:image\"]", "content"},
			[]string{"meta[name=\"twitter:image\"]", "content"},
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				// Multi-selector approach: get content elements + container as fallback
				[]interface{}{
					"#article-body .content-block-regular", 
					"#article-body > p",
					"#article-body > h1", 
					"#article-body > h2",
					"#article-body > h3",
					"#article-body > h4", 
					"#article-body > figure",
					"#article-body > blockquote",
					"#article-body > img",
				},
				
				// Fallback to full container
				"#article-body",
				".article-body",
				
				// Article element fallbacks
				"article.w-article .article-body",
				"main .w-article",
				
				// WordPress content selectors
				".entry-content", 
				".post-content",
			},
		},
		
		// Transform functions for Polygon-specific content
		Transforms: map[string]TransformFunction{
			// Handle lazy-loaded images in noscript tags
			"noscript": &FunctionTransform{
				Fn: func(selection *goquery.Selection) error {
					html, err := selection.Html()
					if err != nil {
						return err
					}
					
					// Replace noscript with a span containing the image
					if html != "" {
						selection.ReplaceWithHtml("<span>" + html + "</span>")
					}
					return nil
				},
			},
		},
		
		// Clean selectors - remove unwanted elements
		Clean: []string{
			// Critical navigation elements to remove
			".w-directory-warning",
			"nav.article-directory-sidenav",
			".sidenav-level",
			".sidenav-item", 
			".directory-warning",
			"a.directory-warning",
			
			// Article footer navigation and pagination
			".article-footer-nav",
			".pagination-nav", 
			".article-nav",
			
			// Ads and promotional content
			"[class*='ad-']",
			"[id*='ad-']", 
			".advertisement",
			".promo",
			
			// Social media and sharing
			".social-share",
			".share-buttons", 
			".w-sharing-copy",
			
			// Navigation and UI elements
			".follow-container", 
			".w-follow-btn",
			".w-like-btn",
			".option-btn",
			".btn-fab",
			".disqus-load-btn",
			
			// Related content and sidebar elements  
			".w-related-content",
			".w-header-related-feed",
			".section-header",
			".section-title",
			".display-card-title",
			".display-card",
			".w-display-card-content",
			".article-header-complementary", 
			".sidebar-tabs",
			".tabs-ul",
			".tabs-header", 
			".tab-content",
			".sidebar-el-content",
			".related-articles",
			".newsletter-signup",
			".email-signup",
			
			// User interface elements
			".w-heading-options",
			".w-header-user-box",
			".user-box-title",
			".article-header-data",
			
			// Comments and interactive elements
			".comments-section",
			".comment-form",
			"#disqus_thread",
			".article-comments",
			
			// Trending and popular content
			"[class*='trending']",
			"[class*='popular']", 
			"[id*='trending']",
			"[id*='popular']",
			".article-header-complementary",
			
			// Login and subscription prompts
			".w-login",
			".valnet-login",
			".w-valnet-login",
			"[id*='login']",
			
			// Remove specific Polygon UI elements
			".article-header-bg",
			".thread-option",
			".fab-label",
			
			// Scripts and tracking
			"script",
			"noscript", // After transformation
			"style",
			
			// Images that are likely duplicates
			"img.c-dynamic-image", // Lazy loading placeholders
		},
	},
}

// GetPolygonExtractor returns the Polygon custom extractor
func GetPolygonExtractor() *CustomExtractor {
	return PolygonExtractor
}