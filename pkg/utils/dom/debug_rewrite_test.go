package dom

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestDebugRewriteStringReplacement(t *testing.T) {
	html := `<html><body><div><p><a href="">Test</a></p></div></body></html>`
	
	t.Logf("Original HTML: %s", html)
	
	// Test the string replacement function
	htmlReplaced := replaceHtmlTag(html, "html")
	t.Logf("After html replacement: %s", htmlReplaced)
	
	bodyReplaced := replaceHtmlTag(htmlReplaced, "body")
	t.Logf("After body replacement: %s", bodyReplaced)
	
	// Check if goquery can parse the result
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(bodyReplaced))
	if err != nil {
		t.Fatalf("Failed to parse modified HTML: %v", err)
	}
	
	// Check final structure
	finalHTML, _ := doc.Html()
	t.Logf("Final parsed HTML: %s", finalHTML)
	
	htmlCount := doc.Find("html").Length()
	bodyCount := doc.Find("body").Length()
	divCount := doc.Find("div").Length()
	
	t.Logf("Final counts - HTML: %d, BODY: %d, DIV: %d", htmlCount, bodyCount, divCount)
}