// ABOUTME: Rewrite top-level DOM elements (html, body) to avoid multiple body tag complications
// ABOUTME: Converts html and body elements to div tags while preserving their content and attributes
package dom

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// RewriteTopLevel rewrites the tag name to div if it's a top level node like body or
// html to avoid later complications with multiple body tags.
// This is a faithful port of the JavaScript rewriteTopLevel function.
func RewriteTopLevel(doc *goquery.Document) *goquery.Document {
	// Get the HTML string from the document
	htmlContent, err := doc.Html()
	if err != nil {
		return doc // Return original if we can't get HTML
	}

	// Convert html and body tags to divs using string replacement
	// This approach works around goquery's limitations with root element manipulation
	
	// Replace opening html tags with div tags (preserving attributes)
	htmlContent = replaceHtmlTag(htmlContent, "html")
	
	// Replace opening body tags with div tags (preserving attributes)  
	htmlContent = replaceHtmlTag(htmlContent, "body")

	// Create a new document from the modified HTML
	newDoc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return doc // Return original if parsing fails
	}

	return newDoc
}

// replaceHtmlTag replaces opening and closing tags of the specified type with div tags
// while preserving attributes
func replaceHtmlTag(htmlContent, tagName string) string {
	// Use regex-like approach but safer with strings package
	// Replace opening tags like <html> or <html class="foo"> with <div> or <div class="foo">
	
	// First, find and replace opening tags
	result := htmlContent
	
	// Simple approach: look for opening tags and replace them
	openTagStart := "<" + tagName
	openTagEnd := ">"
	
	for {
		start := strings.Index(result, openTagStart)
		if start == -1 {
			break
		}
		
		end := strings.Index(result[start:], openTagEnd)
		if end == -1 {
			break
		}
		
		end += start
		
		// Extract the tag content including attributes
		fullTag := result[start:end+1]
		
		// Replace tag name with div while preserving attributes
		var newTag string
		if strings.Contains(fullTag, " ") {
			// Has attributes: <html class="foo"> -> <div class="foo">
			parts := strings.SplitN(fullTag, " ", 2)
			newTag = "<div " + parts[1]
		} else {
			// No attributes: <html> -> <div>
			newTag = "<div>"
		}
		
		result = result[:start] + newTag + result[end+1:]
	}
	
	// Replace closing tags
	closingTag := "</" + tagName + ">"
	result = strings.ReplaceAll(result, closingTag, "</div>")
	
	return result
}