package dom

import (
	"fmt"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// SimpleRewriteTopLevel is a minimal version for testing
func SimpleRewriteTopLevel(doc *goquery.Document) *goquery.Document {
	// Convert html tags to divs first
	doc.Find("html").Each(func(index int, html *goquery.Selection) {
		// For simple test, just replace with a basic approach
		if html.Length() > 0 {
			// Get inner content
			inner, _ := html.Html()
			// Replace with div wrapper
			html.ReplaceWithHtml(fmt.Sprintf("<div>%s</div>", inner))
		}
	})
	
	// Convert body tags to divs
	doc.Find("body").Each(func(index int, body *goquery.Selection) {
		if body.Length() > 0 {
			// Get inner content
			inner, _ := body.Html()
			// Replace with div wrapper  
			body.ReplaceWithHtml(fmt.Sprintf("<div>%s</div>", inner))
		}
	})

	return doc
}

func TestSimpleRewrite(t *testing.T) {
	html := `<div><p>Simple test</p></div>`
	
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	result := SimpleRewriteTopLevel(doc)
	
	// Should work fine with no html/body tags
	if result.Find("p").Length() != 1 {
		t.Errorf("Expected 1 paragraph, got %d", result.Find("p").Length())
	}
}

func TestSimpleRewriteWithBody(t *testing.T) {
	html := `<body><p>Body test</p></body>`
	
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	bodyCount := doc.Find("body").Length()
	fmt.Printf("Initial body count: %d\n", bodyCount)
	
	result := SimpleRewriteTopLevel(doc)
	
	finalBodyCount := result.Find("body").Length()
	fmt.Printf("Final body count: %d\n", finalBodyCount)
	
	// Verify content is still there
	if result.Find("p").Length() != 1 {
		t.Errorf("Expected 1 paragraph, got %d", result.Find("p").Length())
	}
	
	fullHTML, _ := result.Html()
	fmt.Printf("Final HTML: %s\n", fullHTML)
}