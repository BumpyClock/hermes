package dom

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestDebugParent(t *testing.T) {
	html := `<div><p>Test</p></div>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	
	div := doc.Find("div").First()
	p := doc.Find("p").First()
	
	t.Logf("Div parent length: %d", div.Parent().Length())
	t.Logf("P parent length: %d", p.Parent().Length())
	
	if p.Parent().Length() > 0 {
		t.Logf("P parent is: %s", goquery.NodeName(p.Parent()))
	}
	
	// What should happen with getOrInitScore(p)?
	t.Logf("P scoreNode: %d", scoreNode(p))
	t.Logf("P GetWeight: %d", GetWeight(p))
	
	// Before calling getOrInitScore
	t.Logf("Before getOrInitScore - Div: %d, P: %d", getScore(div), getScore(p))
	
	result := getOrInitScore(p, true)
	t.Logf("getOrInitScore(p, true) returned: %d", result)
	
	// After calling getOrInitScore  
	t.Logf("After getOrInitScore - Div: %d, P: %d", getScore(div), getScore(p))
}