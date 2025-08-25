// ABOUTME: HTML sanitization utilities for preventing XSS attacks
// ABOUTME: Uses bluemonday for safe HTML content processing

package security

import (
	"github.com/microcosm-cc/bluemonday"
)

var (
	// StrictSanitizer allows only basic text formatting tags
	StrictSanitizer = bluemonday.StrictPolicy()
	
	// ArticleSanitizer allows common article formatting but removes dangerous elements
	ArticleSanitizer = createArticlePolicy()
	
	// UGCSanitizer for user-generated content with moderate restrictions
	UGCSanitizer = bluemonday.UGCPolicy()
)

// createArticlePolicy creates a policy suitable for article content
func createArticlePolicy() *bluemonday.Policy {
	p := bluemonday.NewPolicy()
	
	// Allow common article formatting
	p.AllowElements("p", "br", "strong", "b", "em", "i", "u", "h1", "h2", "h3", "h4", "h5", "h6")
	p.AllowElements("ul", "ol", "li", "blockquote", "pre", "code")
	p.AllowElements("img", "a", "span", "div")
	
	// Allow links with href
	p.AllowAttrs("href").OnElements("a")
	p.RequireNoReferrerOnLinks(true)
	
	// Allow images with src, alt, width, height
	p.AllowAttrs("src", "alt", "width", "height", "srcset", "sizes").OnElements("img")
	
	// Allow basic styling classes (but sanitize the actual CSS)
	p.AllowAttrs("class").OnElements("div", "span", "p", "img", "a")
	
	// Allow id for anchor links
	p.AllowAttrs("id").OnElements("h1", "h2", "h3", "h4", "h5", "h6", "div", "span")
	
	return p
}

// SanitizeHTML sanitizes HTML content for safe display
func SanitizeHTML(html string) string {
	return ArticleSanitizer.Sanitize(html)
}

// SanitizeHTMLStrict uses strict sanitization (text only)
func SanitizeHTMLStrict(html string) string {
	return StrictSanitizer.Sanitize(html)
}

// SanitizeUserContent sanitizes user-generated content
func SanitizeUserContent(html string) string {
	return UGCSanitizer.Sanitize(html)
}