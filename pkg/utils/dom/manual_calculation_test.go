package dom

import (
	"fmt"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestManualCalculation(t *testing.T) {
	// Let's manually calculate what the JavaScript should produce
	html := `
		<div class="hentry">
			<p class="entry-content">Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis, sem. Nulla consequat massa quis enim. Donec pede justo, fringilla vel, aliquet nec, vulputate eget, arcu. In enim justo, rhoncus ut, imperdiet a, venenatis vitae, justo. Nullam dictum felis eu pede mollis pretium. Integer tincidunt. Cras dapibus. Vivamus elementum semper nisi. Aenean vulputate eleifend tellus. Aenean leo ligula, porttitor eu.</p>
		</div>
	`
	
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}
	
	div := doc.Find("div").First()
	p := doc.Find("p").First()
	
	fmt.Println("=== Manual Calculation ===")
	
	// Step 1: Check GetWeight for div.hentry
	divWeight := GetWeight(div)
	fmt.Printf("Div.hentry GetWeight: %d\n", divWeight)
	
	// Step 2: Check GetWeight for p.entry-content  
	pWeight := GetWeight(p)
	fmt.Printf("P.entry-content GetWeight: %d\n", pWeight)
	
	// Step 3: Calculate paragraph score
	pText := strings.TrimSpace(p.Text())
	fmt.Printf("P text length: %d\n", len(pText))
	fmt.Printf("P comma count: %d\n", strings.Count(pText, ","))
	
	pScore := scoreCommas(pText) + scoreLength(pText, 1)
	if len(pText) < 20 {
		pScore -= 10
	}
	if len(pText) >= 50 && len(pText) <= 200 {
		pScore += 5
	}
	fmt.Printf("P calculated scoreParagraph: %d\n", pScore)
	fmt.Printf("P actual scoreParagraph: %d\n", scoreParagraph(p))
	
	// Step 4: What should happen in scorePs first call
	fmt.Println("\n=== First scorePs call simulation ===")
	// For p element:
	// 1. getOrInitScore(p, true) = scoreNode(p) + getWeight(p) + addToParent(p, score)
	//    scoreNode(p) calls scoreParagraph(p) = 27  
	//    getWeight(p) for "entry-content" = ?
	//    addToParent adds 1/4 of (27+weight) to div
	// 2. setScore(p, score)
	// 3. rawScore = scoreNode(p) = 27
	// 4. addScoreTo(div, 27) - but div is already scored from step 1
	// 5. addScoreTo(grandparent, 27/2) - no grandparent
	
	fmt.Printf("Expected p scoreNode result: %d\n", scoreNode(p))
	fmt.Printf("Expected p getWeight result: %d\n", GetWeight(p))
	
	// What should div get from hNews boost: +80
	// What should div get from addToParent when p is initialized: +25% of p's total
	// What should div get from addScoreTo when p is processed: +full p score
	
	expectedPInitialScore := scoreNode(p) + GetWeight(p)  // 27 + 25 = 52
	expectedDivFromAddToParent := expectedPInitialScore / 4  // 52/4 = 13
	expectedDivFromAddScoreTo := scoreNode(p)  // 27
	expectedDivFromHNews := 80
	
	fmt.Printf("\nExpected calculations:")
	fmt.Printf("P initial score: %d\n", expectedPInitialScore)
	fmt.Printf("Div from hNews: %d\n", expectedDivFromHNews)
	fmt.Printf("Div from addToParent: %d\n", expectedDivFromAddToParent)
	fmt.Printf("Div from addScoreTo: %d\n", expectedDivFromAddScoreTo)
	fmt.Printf("Expected div total after 1st scorePs: %d\n", expectedDivFromHNews + expectedDivFromAddToParent + expectedDivFromAddScoreTo)
}