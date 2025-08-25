// ABOUTME: Score content implementation for content scoring orchestration
// ABOUTME: Handles span conversion, score propagation, and hNews boosting

package dom

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// convertSpanToDivForScoring converts span elements to divs to improve scoring
// JavaScript: function convertSpans($node, $)
func convertSpanToDivForScoring(element *goquery.Selection) {
	if element.Length() > 0 {
		tagName := strings.ToLower(goquery.NodeName(element))
		if tagName == "span" {
			// convert spans to divs
			ConvertNodeTo(element, "div")
		}
	}
}

// addScoreTo adds score to a node after converting spans
// JavaScript: function addScoreTo($node, $, score)
func addScoreTo(element *goquery.Selection, score int) {
	if element != nil && element.Length() > 0 {
		convertSpanToDivForScoring(element)
		addScore(element, score)
	}
}

// scorePs scores paragraph and pre elements, propagating scores to parents
// JavaScript: function scorePs($, weightNodes)
func scorePs(doc *goquery.Document, weightNodes bool) {
	// JavaScript: $('p, pre').not('[score]').each((index, node) => {
	doc.Find("p, pre").Not("[score]").Each(func(index int, element *goquery.Selection) {
		// The raw score for this paragraph, before we add any parent/child scores
		// JavaScript: let $node = $(node);
		// JavaScript: $node = setScore($node, $, getOrInitScore($node, $, weightNodes));
		score := getOrInitScore(element, weightNodes)
		setScore(element, score)

		// JavaScript: const $parent = $node.parent();
		parent := element.Parent()
		// JavaScript: const rawScore = scoreNode($node);
		rawScore := scoreNode(element)

		// JavaScript: addScoreTo($parent, $, rawScore, weightNodes);
		addScoreTo(parent, rawScore)
		
		if parent.Length() > 0 {
			// Add half of the individual content score to the grandparent
			// JavaScript: addScoreTo($parent.parent(), $, rawScore / 2, weightNodes);
			grandparent := parent.Parent()
			addScoreTo(grandparent, rawScore/2)
		}
	})
}

// ScoreContent orchestrates the entire content scoring process
// JavaScript: export default function scoreContent($, weightNodes = true)
func ScoreContent(doc *goquery.Document, weightNodes bool) {
	// First, look for special hNews based selectors and give them a big
	// boost, if they exist
	// JavaScript: HNEWS_CONTENT_SELECTORS.forEach(([parentSelector, childSelector]) => {
	for _, selectors := range HNEWS_CONTENT_SELECTORS {
		parentSelector := selectors[0]
		childSelector := selectors[1]
		
		// JavaScript: $(`${parentSelector} ${childSelector}`).each((index, node) => {
		combinedSelector := parentSelector + " " + childSelector
		doc.Find(combinedSelector).Each(func(index int, element *goquery.Selection) {
			// JavaScript: addScore($(node).parent(parentSelector), $, 80);
			parent := element.ParentsFiltered(parentSelector).First()
			addScore(parent, 80)
		})
	}

	// Doubling this again
	// Previous solution caused a bug
	// in which parents weren't retaining
	// scores. This is not ideal, and
	// should be fixed.
	// JavaScript: scorePs($, weightNodes);
	scorePs(doc, weightNodes)
	// JavaScript: scorePs($, weightNodes);
	scorePs(doc, weightNodes)
}