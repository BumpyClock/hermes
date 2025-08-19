package dom

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestDebugMath(t *testing.T) {
	t.Logf("int(float64(100)*0.25) = %d", int(float64(100)*0.25))
	
	// Test what happens when we call addScore on a fresh div
	html := `<div><p>Test</p></div>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	
	div := doc.Find("div").First()
	
	t.Logf("Div GetWeight: %d", GetWeight(div))
	t.Logf("Div getScore before: %d", getScore(div))
	
	// Call addScore with 25
	addScore(div, 25)
	t.Logf("After addScore(div, 25): %d", getScore(div))
	
	// Reset and try again
	doc2, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	div2 := doc2.Find("div").First()
	
	t.Logf("Fresh div getOrInitScore(div, true): %d", getOrInitScore(div2, true))
	t.Logf("Div2 score after getOrInitScore: %d", getScore(div2))
}