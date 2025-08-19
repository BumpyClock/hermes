package dom

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// TestRewriteTopLevel_FunctionalBehavior tests the actual expected behavior
// in the context of content processing (which is how this function is used)
func TestRewriteTopLevel_FunctionalBehavior(t *testing.T) {
	// This test mirrors the JavaScript test more accurately
	html := `<html><body><div><p><a href="">Wow how about that</a></p></div></body></html>`
	
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	result := RewriteTopLevel(doc)

	// The key test: verify we have the expected number of nested divs
	// Original structure: html > body > div > p > a
	// After rewrite: (wrapper html) > (wrapper body) > div > div > div > p > a
	// The important part is that we have 3 divs total (original + converted html + converted body)
	
	divs := result.Find("div")
	if divs.Length() != 3 {
		t.Errorf("Expected exactly 3 div elements (original + converted html + converted body), but found %d", divs.Length())
	}

	// Verify content is preserved
	links := result.Find("a")
	if links.Length() != 1 {
		t.Errorf("Expected 1 link, but found %d", links.Length())
	}
	
	if links.Length() > 0 {
		linkText := strings.TrimSpace(links.First().Text())
		if linkText != "Wow how about that" {
			t.Errorf("Expected link text 'Wow how about that', but got '%s'", linkText)
		}
		
		// The href attribute should be preserved
		_, hasHref := links.First().Attr("href")
		if !hasHref {
			t.Errorf("Expected href attribute to be preserved")
		}
	}

	// Verify paragraph is preserved  
	paragraphs := result.Find("p")
	if paragraphs.Length() != 1 {
		t.Errorf("Expected 1 paragraph, but found %d", paragraphs.Length())
	}
}

// TestRewriteTopLevel_AttributePreservation tests that attributes are preserved
func TestRewriteTopLevel_AttributePreservation(t *testing.T) {
	html := `<html lang="en"><body class="article"><div id="content">Test</div></body></html>`
	
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	result := RewriteTopLevel(doc)

	// Check that the converted html and body attributes appear on the divs
	divs := result.Find("div")
	
	// Find div with lang attribute (converted from html)
	langDiv := divs.FilterFunction(func(i int, s *goquery.Selection) bool {
		_, hasLang := s.Attr("lang")
		return hasLang
	})
	
	if langDiv.Length() != 1 {
		t.Errorf("Expected 1 div with lang attribute (converted from html), but found %d", langDiv.Length())
	}
	
	if langDiv.Length() > 0 {
		lang, _ := langDiv.First().Attr("lang")
		if lang != "en" {
			t.Errorf("Expected lang='en', but got lang='%s'", lang)
		}
	}

	// Find div with class attribute (converted from body)
	classDiv := divs.FilterFunction(func(i int, s *goquery.Selection) bool {
		class, hasClass := s.Attr("class")
		return hasClass && class == "article"
	})
	
	if classDiv.Length() != 1 {
		t.Errorf("Expected 1 div with class='article' (converted from body), but found %d", classDiv.Length())
	}
}

// TestRewriteTopLevel_NoHtmlBodyTags tests behavior when document has no html/body tags
func TestRewriteTopLevel_NoHtmlBodyTags(t *testing.T) {
	html := `<div><p>Regular content</p></div>`
	
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	originalDivCount := doc.Find("div").Length()
	result := RewriteTopLevel(doc)

	// When there are no html/body tags to convert, the content should remain the same
	// Note: goquery may still add wrapper html/body tags, but the internal structure should be preserved
	finalDivCount := result.Find("div").Length()
	
	if finalDivCount < originalDivCount {
		t.Errorf("Expected at least %d divs to be preserved, but found %d", originalDivCount, finalDivCount)
	}
	
	// Content should be preserved
	paragraphs := result.Find("p")
	if paragraphs.Length() != 1 {
		t.Errorf("Expected 1 paragraph, but found %d", paragraphs.Length())
	}
	
	if paragraphs.Length() > 0 && strings.TrimSpace(paragraphs.First().Text()) != "Regular content" {
		t.Errorf("Expected text 'Regular content', but got '%s'", paragraphs.First().Text())
	}
}

// TestRewriteTopLevel_EmptyElements tests behavior with empty html/body elements
func TestRewriteTopLevel_EmptyElements(t *testing.T) {
	html := `<html><body></body></html>`
	
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	result := RewriteTopLevel(doc)

	// Should have 2 empty divs (converted from html and body)
	divs := result.Find("div")
	if divs.Length() != 2 {
		t.Errorf("Expected 2 divs (converted from empty html and body), but found %d", divs.Length())
	}
	
	// All divs should be empty
	divs.Each(func(i int, div *goquery.Selection) {
		text := strings.TrimSpace(div.Text())
		if text != "" {
			t.Errorf("Expected empty div, but found text: '%s'", text)
		}
	})
}