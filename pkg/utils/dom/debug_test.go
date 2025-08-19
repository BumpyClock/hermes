package dom

import (
	"fmt"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestDebugConvertNodeTo(t *testing.T) {
	html := `<div class="test">Hello world</div>`
	
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	div := doc.Find("div")
	if div.Length() == 0 {
		t.Fatal("No div found")
	}

	fmt.Printf("Before conversion: %s\n", div.Get(0).Data)
	
	ConvertNodeTo(div, "p")
	
	p := doc.Find("p")
	if p.Length() == 0 {
		t.Fatal("No p found after conversion")
	}
	
	fmt.Printf("After conversion: %s\n", p.Get(0).Data)
	bodyHtml, _ := doc.Find("body").Html()
	fmt.Printf("Full HTML: %s\n", bodyHtml)
}

func TestDebugHtmlBodyConversion(t *testing.T) {
	html := `<html><body><div>Test content</div></body></html>`
	
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	fmt.Printf("Initial HTML structure:\n")
	fullHtml, _ := doc.Html()
	fmt.Printf("%s\n", fullHtml)
	
	htmlElem := doc.Find("html")
	bodyElem := doc.Find("body")
	
	fmt.Printf("HTML elements found: %d\n", htmlElem.Length())
	fmt.Printf("BODY elements found: %d\n", bodyElem.Length())
	
	if htmlElem.Length() > 0 {
		fmt.Printf("HTML tag name: %s\n", htmlElem.Get(0).Data)
		
		// Try to convert HTML to div
		fmt.Println("Converting HTML to div...")
		ConvertNodeTo(htmlElem, "div")
		
		finalHtml, _ := doc.Html()
		fmt.Printf("After HTML conversion:\n%s\n", finalHtml)
	}
}