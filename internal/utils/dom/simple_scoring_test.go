package dom

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestSimpleScoringSteps(t *testing.T) {
	// Test a minimal case to understand the exact behavior
	html := `<div class="hentry"><p class="entry-content">Short text with comma, here.</p></div>`
	
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}
	
	div := doc.Find("div").First()
	p := doc.Find("p").First()
	
	t.Logf("Initial scores - Div: %d, P: %d", getScore(div), getScore(p))
	
	// Manual step 1: hNews boost
	t.Logf("Div GetWeight: %d", GetWeight(div))
	addScore(div, 80)
	t.Logf("After hNews boost - Div: %d", getScore(div))
	
	// Manual step 2: First scorePs
	// This should:
	// 1. For each p/pre not already scored
	// 2. score = getOrInitScore(p, true) = scoreNode(p) + getWeight(p) + addToParent(p, score)
	// 3. setScore(p, score) - but getOrInitScore doesn't call setScore
	// 4. rawScore = scoreNode(p)
	// 5. addScoreTo(parent, rawScore)
	// 6. addScoreTo(grandparent, rawScore/2)
	
	pText := p.Text()
	expectedScoreNode := scoreNode(p)
	expectedPWeight := GetWeight(p)
	expectedPScore := expectedScoreNode + expectedPWeight
	
	t.Logf("P text: '%s'", pText)
	t.Logf("P scoreNode: %d", expectedScoreNode)
	t.Logf("P GetWeight: %d", expectedPWeight) 
	t.Logf("Expected P score: %d", expectedPScore)
	
	// Simulate first scorePs call manually
	scorePs(doc, true)
	t.Logf("After first scorePs - Div: %d, P: %d", getScore(div), getScore(p))
	
	// Simulate second scorePs call
	scorePs(doc, true) 
	t.Logf("After second scorePs - Div: %d, P: %d", getScore(div), getScore(p))
	
	// Now compare to full ScoreContent
	doc2, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	ScoreContent(doc2, true)
	t.Logf("ScoreContent result - Div: %d, P: %d", 
		getScore(doc2.Find("div").First()), 
		getScore(doc2.Find("p").First()))
}

func TestUnderstandAddToParent(t *testing.T) {
	// Test addToParent behavior in isolation
	html := `<div><p>Test</p></div>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	
	div := doc.Find("div").First()
	p := doc.Find("p").First()
	
	t.Logf("Before addToParent - Div: %d, P: %d", getScore(div), getScore(p))
	
	// addToParent should add 25% of child score to parent
	addToParent(p, 100)
	t.Logf("After addToParent(p, 100) - Div: %d, P: %d", getScore(div), getScore(p))
	
	// addToParent again 
	addToParent(p, 100)
	t.Logf("After second addToParent(p, 100) - Div: %d, P: %d", getScore(div), getScore(p))
}