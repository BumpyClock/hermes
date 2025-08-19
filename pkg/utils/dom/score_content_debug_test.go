package dom

import (
	"fmt"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestScoreContentDebugHNews(t *testing.T) {
	// Debug the hNews scoring step by step
	html := `
		<div class="hentry">
			<p class="entry-content">Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis, sem. Nulla consequat massa quis enim. Donec pede justo, fringilla vel, aliquet nec, vulputate eget, arcu. In enim justo, rhoncus ut, imperdiet a, venenatis vitae, justo. Nullam dictum felis eu pede mollis pretium. Integer tincidunt. Cras dapibus. Vivamus elementum semper nisi. Aenean vulputate eleifend tellus. Aenean leo ligula, porttitor eu.</p>
		</div>
	`
	
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}
	
	fmt.Println("=== Initial state ===")
	div := doc.Find("div").First()
	p := doc.Find("p").First()
	fmt.Printf("Div initial score: %d\n", getScore(div))
	fmt.Printf("P initial score: %d\n", getScore(p))
	
	// Step 1: Test hNews boost
	fmt.Println("\n=== After hNews boost ===")
	for _, selectors := range HNEWS_CONTENT_SELECTORS {
		parentSelector := selectors[0]
		childSelector := selectors[1]
		
		combinedSelector := parentSelector + " " + childSelector
		fmt.Printf("Checking selector: %s\n", combinedSelector)
		
		doc.Find(combinedSelector).Each(func(index int, element *goquery.Selection) {
			fmt.Printf("Found element: %s with class %s\n", goquery.NodeName(element), element.AttrOr("class", "none"))
			parent := element.ParentsFiltered(parentSelector).First()
			fmt.Printf("Parent: %s with class %s\n", goquery.NodeName(parent), parent.AttrOr("class", "none"))
			
			beforeScore := getScore(parent)
			addScore(parent, 80)
			afterScore := getScore(parent)
			fmt.Printf("Parent score: %d -> %d\n", beforeScore, afterScore)
		})
	}
	
	fmt.Printf("Div after hNews boost: %d\n", getScore(div))
	
	// Step 2: Test scorePs
	fmt.Println("\n=== After first scorePs ===")
	scorePs(doc, true)
	fmt.Printf("Div after first scorePs: %d\n", getScore(div))
	fmt.Printf("P after first scorePs: %d\n", getScore(p))
	
	fmt.Println("\n=== After second scorePs ===")  
	scorePs(doc, true)
	fmt.Printf("Div after second scorePs: %d\n", getScore(div))
	fmt.Printf("P after second scorePs: %d\n", getScore(p))
	
	// Also test the scoring calculation manually
	fmt.Println("\n=== Manual scoring calculation ===")
	pText := p.Text()
	fmt.Printf("P text length: %d\n", len(pText))
	fmt.Printf("P comma count: %d\n", strings.Count(pText, ","))
	fmt.Printf("P scoreNode result: %d\n", scoreNode(p))
	fmt.Printf("P scoreParagraph result: %d\n", scoreParagraph(p))
}

func TestScoreContentDebugNonHNews(t *testing.T) {
	// Debug the non-hNews scoring
	html := `
		<div class="">
			<p class="entry-content">Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis, sem. Nulla consequat massa quis enim. Donec pede justo, fringilla vel, aliquet nec, vulputate eget, arcu. In enim justo, rhoncus ut, imperdiet a, venenatis vitae, justo. Nullam dictum felis eu pede mollis pretium. Integer tincidunt. Cras dapibus. Vivamus elementum semper nisi. Aenean vulputate eleifend tellus. Aenean leo ligula, porttitor eu.</p>
			<p class="entry-content">Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis, sem. Nulla consequat massa quis enim. Donec pede justo, fringilla vel, aliquet nec, vulputate eget, arcu. In enim justo, rhoncus ut, imperdiet a, venenatis vitae, justo. Nullam dictum felis eu pede mollis pretium. Integer tincidunt. Cras dapibus. Vivamus elementum semper nisi. Aenean vulputate eleifend tellus. Aenean leo ligula, porttitor eu.</p>
		</div>
	`
	
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}
	
	fmt.Println("=== Non-hNews Debug ===")
	div := doc.Find("div").First()
	
	fmt.Printf("Div initial score: %d\n", getScore(div))
	
	// Run full scoring
	ScoreContent(doc, true)
	
	fmt.Printf("Div final score: %d (expected 65)\n", getScore(div))
	
	// Check each paragraph
	doc.Find("p").Each(func(index int, p *goquery.Selection) {
		fmt.Printf("P%d score: %d, scoreNode: %d, scoreParagraph: %d\n", 
			index, getScore(p), scoreNode(p), scoreParagraph(p))
	})
}