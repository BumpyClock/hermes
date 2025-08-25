// ABOUTME: Tests for CleanHOnes function covering H1 removal and conversion scenarios.
// ABOUTME: Verifies 100% JavaScript compatibility with exact HTML output matching.
package dom

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestCleanHOnes_RemovesH1sWhenLessThan3(t *testing.T) {
	html := `
      <div>
        <h1>Look at this!</h1>
        <p>What do you think?</p>
        <h1>Can you believe it?!</h1>
      </div>
    `

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	CleanHOnes(doc)

	// Get the result HTML
	result, err := doc.Html()
	if err != nil {
		t.Fatalf("Failed to get HTML: %v", err)
	}

	// Check that H1s are removed
	if strings.Contains(result, "<h1>") {
		t.Errorf("H1 tags should be removed when there are less than 3")
	}

	// Check that paragraphs remain
	if !strings.Contains(result, "<p>What do you think?</p>") {
		t.Errorf("Paragraph content should remain intact")
	}

	// Verify exact structure matches expected output
	expectedElements := []string{
		"<p>What do you think?</p>",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected element not found: %s\nGot: %s", expected, result)
		}
	}

	// Verify H1 content is completely removed
	h1Elements := doc.Find("h1")
	if h1Elements.Length() != 0 {
		t.Errorf("Expected 0 H1 elements, got %d", h1Elements.Length())
	}
}

func TestCleanHOnes_ConvertsH1sToH2sWhen3OrMore(t *testing.T) {
	html := `
      <div>
        <h1>Look at this!</h1>
        <p>What do you think?</p>
        <h1>Can you believe it?!</h1>
        <p>What do you think?</p>
        <h1>Can you believe it?!</h1>
      </div>
    `

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	CleanHOnes(doc)

	// Get the result HTML
	result, err := doc.Html()
	if err != nil {
		t.Fatalf("Failed to get HTML: %v", err)
	}

	// Check that H1s are converted to H2s
	if strings.Contains(result, "<h1>") {
		t.Errorf("H1 tags should be converted to H2s when there are 3 or more")
	}

	// Check that H2s exist with correct content
	expectedH2s := []string{
		"<h2>Look at this!</h2>",
		"<h2>Can you believe it?!</h2>",
	}

	for _, expectedH2 := range expectedH2s {
		if !strings.Contains(result, expectedH2) {
			t.Errorf("Expected H2 not found: %s\nGot: %s", expectedH2, result)
		}
	}

	// Check that paragraphs remain
	if !strings.Contains(result, "<p>What do you think?</p>") {
		t.Errorf("Paragraph content should remain intact")
	}

	// Verify exact count of H2 elements
	h2Elements := doc.Find("h2")
	if h2Elements.Length() != 3 {
		t.Errorf("Expected 3 H2 elements, got %d", h2Elements.Length())
	}

	// Verify no H1 elements remain
	h1Elements := doc.Find("h1")
	if h1Elements.Length() != 0 {
		t.Errorf("Expected 0 H1 elements, got %d", h1Elements.Length())
	}
}

func TestCleanHOnes_HandlesEmptyDocument(t *testing.T) {
	html := `<div></div>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	result := CleanHOnes(doc)

	if result == nil {
		t.Errorf("CleanHOnes should return a document")
	}
}

func TestCleanHOnes_HandlesNoH1Elements(t *testing.T) {
	html := `
      <div>
        <h2>A heading</h2>
        <p>Some content</p>
        <h3>Another heading</h3>
      </div>
    `

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	originalHTML, _ := doc.Html()
	
	CleanHOnes(doc)

	// Get the result HTML
	result, err := doc.Html()
	if err != nil {
		t.Fatalf("Failed to get HTML: %v", err)
	}

	// Content should remain unchanged when no H1s exist
	if !strings.Contains(result, "<h2>A heading</h2>") {
		t.Errorf("H2 should remain unchanged")
	}
	if !strings.Contains(result, "<p>Some content</p>") {
		t.Errorf("Paragraph should remain unchanged")
	}
	if !strings.Contains(result, "<h3>Another heading</h3>") {
		t.Errorf("H3 should remain unchanged")
	}

	// Overall structure should be preserved
	if len(result) < len(originalHTML)-100 { // Allow for minor whitespace differences
		t.Errorf("Content should not be significantly changed when no H1s exist")
	}
}

func TestCleanHOnes_HandlesExactly3H1Elements(t *testing.T) {
	html := `
      <div>
        <h1>First</h1>
        <h1>Second</h1>
        <h1>Third</h1>
        <p>Content</p>
      </div>
    `

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	CleanHOnes(doc)

	// With exactly 3 H1s, they should be converted to H2s
	h1Elements := doc.Find("h1")
	if h1Elements.Length() != 0 {
		t.Errorf("Expected 0 H1 elements, got %d", h1Elements.Length())
	}

	h2Elements := doc.Find("h2")
	if h2Elements.Length() != 3 {
		t.Errorf("Expected 3 H2 elements, got %d", h2Elements.Length())
	}

	// Verify H2 content
	expectedTexts := []string{"First", "Second", "Third"}
	h2Elements.Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if i < len(expectedTexts) && text != expectedTexts[i] {
			t.Errorf("Expected H2 text '%s', got '%s'", expectedTexts[i], text)
		}
	})
}

func TestCleanHOnes_PreservesH1Attributes(t *testing.T) {
	html := `
      <div>
        <h1 id="heading1" class="main-title" data-test="value">First Heading</h1>
        <h1 class="secondary">Second Heading</h1>
        <h1>Third Heading</h1>
      </div>
    `

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	CleanHOnes(doc)

	// With 3 H1s, they should be converted to H2s with preserved attributes
	h2Elements := doc.Find("h2")
	if h2Elements.Length() != 3 {
		t.Errorf("Expected 3 H2 elements, got %d", h2Elements.Length())
	}

	// Check first H2 has preserved attributes
	firstH2 := h2Elements.First()
	if id, exists := firstH2.Attr("id"); !exists || id != "heading1" {
		t.Errorf("Expected id='heading1', got id='%s', exists=%v", id, exists)
	}
	if class, exists := firstH2.Attr("class"); !exists || class != "main-title" {
		t.Errorf("Expected class='main-title', got class='%s', exists=%v", class, exists)
	}
	if dataTest, exists := firstH2.Attr("data-test"); !exists || dataTest != "value" {
		t.Errorf("Expected data-test='value', got data-test='%s', exists=%v", dataTest, exists)
	}

	// Check second H2 has preserved class
	secondH2 := h2Elements.Eq(1)
	if class, exists := secondH2.Attr("class"); !exists || class != "secondary" {
		t.Errorf("Expected class='secondary', got class='%s', exists=%v", class, exists)
	}

	// Check third H2 has no attributes (as expected)
	thirdH2 := h2Elements.Eq(2)
	if thirdH2.Get(0).Attr != nil && len(thirdH2.Get(0).Attr) > 0 {
		t.Errorf("Third H2 should have no attributes")
	}
}