package dom

import (
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

// JavaScript scoring functions from extractors/generic/content/scoring

// scoreCommas counts commas in text (more commas = better content quality)
// JavaScript: function scoreCommas(text) 
func scoreCommas(text string) int {
	return strings.Count(text, ",")
}

// scoreLength gives bonus for text length in 50-character chunks
// JavaScript: function scoreLength(text, lengthBonus = 1)
func scoreLength(text string, lengthBonus int) int {
	if lengthBonus == 0 {
		lengthBonus = 1
	}
	return (len(text) / 50) * lengthBonus
}

// scoreParagraph provides multi-factor paragraph scoring  
// JavaScript: function scoreParagraph(paragraph)
func scoreParagraph(paragraph *goquery.Selection) int {
	text := strings.TrimSpace(paragraph.Text())
	if text == "" {
		return 0
	}
	
	score := 0
	
	// Base score from commas (content quality indicator)
	score += scoreCommas(text)
	
	// Length bonus (50 chars = 1 point)
	score += scoreLength(text, 1)
	
	// Penalty for short paragraphs  
	if len(text) < 20 {
		score -= 10
	}
	
	// Bonus for medium-length content
	if len(text) >= 50 && len(text) <= 200 {
		score += 5
	}
	
	return score
}

// getOrInitScore gets existing score or initializes with weight
// JavaScript: function getOrInitScore(node, $, weightNodes = true)
func getOrInitScore(element *goquery.Selection, weightNodes bool) int {
	// JavaScript: let score = getScore($node);
	score := getScore(element)
	
	// JavaScript: if (score) { return score; }
	if score != 0 {
		return score
	}
	
	// JavaScript: score = scoreNode($node);
	score = scoreNode(element)
	
	if weightNodes {
		// JavaScript: score += getWeight($node);
		score += GetWeight(element)
	}
	
	// JavaScript: addToParent($node, $, score);
	addToParent(element, score)
	
	// Note: JavaScript getOrInitScore does NOT call setScore
	// That's handled by addScore
	return score
}

// setScore stores score on element
// JavaScript: function setScore(node, $, score)
func setScore(element *goquery.Selection, score int) {
	element.SetAttr("data-content-score", itoa(score))
}

// getScore retrieves score from element
// JavaScript: function getScore(node, $)
func getScore(element *goquery.Selection) int {
	// First check for data-content-score (our internal scoring)
	if scoreStr, exists := element.Attr("data-content-score"); exists {
		if score, err := parseInt(scoreStr); err == nil {
			return score
		}
	}
	
	// Also check for 'score' attribute (used in tests and some JS implementations)
	if scoreStr, exists := element.Attr("score"); exists {
		if score, err := parseInt(scoreStr); err == nil {
			return score
		}
	}
	
	return 0
}

// Helper functions
func parseInt(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	
	result := 0
	negative := false
	start := 0
	
	// Handle negative sign
	if s[0] == '-' {
		negative = true
		start = 1
	} else if s[0] == '+' {
		start = 1
	}
	
	for i := start; i < len(s); i++ {
		r := rune(s[i])
		if !unicode.IsDigit(r) {
			break
		}
		result = result*10 + int(r-'0')
	}
	
	if negative {
		result = -result
	}
	
	return result, nil
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	
	result := ""
	negative := i < 0
	if negative {
		i = -i
	}
	
	for i > 0 {
		result = string(rune('0'+i%10)) + result
		i /= 10
	}
	
	if negative {
		result = "-" + result
	}
	
	return result
}

// Additional scoring functions to support cleanTags logic

// countParagraphs counts paragraph elements
func countParagraphs(element *goquery.Selection) int {
	return element.Find("p").Length()
}

// countImages counts image elements  
func countImages(element *goquery.Selection) int {
	return element.Find("img").Length()
}

// countInputs counts input/form elements
func countInputs(element *goquery.Selection) int {
	return element.Find("input, textarea, select, button").Length()
}

// countLists counts list elements
func countLists(element *goquery.Selection) int {
	return element.Find("ul, ol, dl").Length()
}

// textLength gets text length with whitespace normalization
func textLength(element *goquery.Selection) int {
	text := strings.TrimSpace(element.Text())
	// Normalize whitespace like JavaScript
	text = strings.Join(strings.Fields(text), " ")
	return len(text)
}

// addScore adds a score amount to a node
// JavaScript: export default function addScore($node, $, amount)
func addScore(element *goquery.Selection, amount int) *goquery.Selection {
	// JavaScript: const score = getOrInitScore($node, $) + amount;
	score := getOrInitScore(element, true) + amount
	// JavaScript: setScore($node, $, score);
	setScore(element, score)
	return element
}

// addToParent adds 1/4 of a child's score to its parent
// JavaScript: export default function addToParent(node, $, score)
func addToParent(element *goquery.Selection, score int) *goquery.Selection {
	// JavaScript: const parent = node.parent();
	parent := element.Parent()
	if parent.Length() > 0 {
		// JavaScript: addScore(parent, $, score * 0.25);
		addScore(parent, int(float64(score)*0.25))
	}
	return element
}

// scoreNode scores an individual node based on tag type
// JavaScript: export default function scoreNode($node)
func scoreNode(element *goquery.Selection) int {
	// JavaScript: const { tagName } = $node.get(0);
	tagName := strings.ToLower(goquery.NodeName(element))
	
	// JavaScript: if (PARAGRAPH_SCORE_TAGS.test(tagName))
	if PARAGRAPH_SCORE_TAGS.MatchString(tagName) {
		// JavaScript: return scoreParagraph($node);
		return scoreParagraph(element)
	}
	
	// JavaScript: if (tagName.toLowerCase() === 'div')
	if tagName == "div" {
		return 5
	}
	
	// JavaScript: if (CHILD_CONTENT_TAGS.test(tagName))
	if CHILD_CONTENT_TAGS.MatchString(tagName) {
		return 3
	}
	
	// JavaScript: if (BAD_TAGS.test(tagName))
	if BAD_TAGS.MatchString(tagName) {
		return -3
	}
	
	// JavaScript: if (tagName.toLowerCase() === 'th')
	if tagName == "th" {
		return -5
	}
	
	return 0
}

// GetWeight scores a node based on its className and id
// JavaScript: function getWeight(node)
func GetWeight(element *goquery.Selection) int {
	// JavaScript: const classes = node.attr('class');
	classes, _ := element.Attr("class")
	// JavaScript: const id = node.attr('id');
	id, _ := element.Attr("id")
	score := 0
	
	if id != "" {
		// JavaScript: if (POSITIVE_SCORE_RE.test(id))
		if POSITIVE_SCORE_RE.MatchString(id) {
			score += 25
		}
		// JavaScript: if (NEGATIVE_SCORE_RE.test(id))
		if NEGATIVE_SCORE_RE.MatchString(id) {
			score -= 25
		}
	}
	
	if classes != "" {
		if score == 0 {
			// JavaScript: if (POSITIVE_SCORE_RE.test(classes))
			if POSITIVE_SCORE_RE.MatchString(classes) {
				score += 25
			}
			// JavaScript: if (NEGATIVE_SCORE_RE.test(classes))
			if NEGATIVE_SCORE_RE.MatchString(classes) {
				score -= 25
			}
		}
		
		// JavaScript: if (PHOTO_HINTS_RE.test(classes))
		if PHOTO_HINTS_RE.MatchString(classes) {
			score += 10
		}
		
		// JavaScript: if (READABILITY_ASSET.test(classes))
		if READABILITY_ASSET.MatchString(classes) {
			score += 25
		}
	}
	
	return score
}

// FindTopCandidate finds the element with the highest score after calculating all scores
// After we've calculated scores, loop through all of the possible candidate nodes we found and find the one with the highest score.
// JavaScript: export default function findTopCandidate($)
func FindTopCandidate(doc *goquery.Document) *goquery.Selection {
	var candidate *goquery.Selection
	topScore := 0
	
	// JavaScript: $('[score]').each((index, node) => {
	// Look for elements with either score or data-content-score attributes
	doc.Find("[score], [data-content-score]").Each(func(index int, element *goquery.Selection) {
		// JavaScript: if (NON_TOP_CANDIDATE_TAGS_RE.test(node.tagName)) { return; }
		tagName := strings.ToLower(goquery.NodeName(element))
		if NON_TOP_CANDIDATE_TAGS_RE.MatchString(tagName) {
			return
		}
		
		// JavaScript: const score = getScore($node);
		score := getScore(element)
		
		// JavaScript: if (score > topScore) { topScore = score; $candidate = $node; }
		if score > topScore {
			topScore = score
			candidate = element
		}
	})
	
	// JavaScript: if (!$candidate) { return $('body') || $('*').first(); }
	if candidate == nil {
		// Try to find body element first
		body := doc.Find("body")
		if body.Length() > 0 {
			return body
		}
		// Fall back to first element
		all := doc.Find("*")
		if all.Length() > 0 {
			return all.First()
		}
		// Return empty selection if no elements found
		return doc.Find("")
	}
	
	// JavaScript: $candidate = mergeSiblings($candidate, topScore, $);
	candidate = MergeSiblings(candidate, topScore, doc)
	
	// JavaScript: return $candidate;
	return candidate
}

// MergeSiblings merges sibling elements that may be part of the main content
// Now that we have a top_candidate, look through the siblings of it to see if any of them are decently scored.
// JavaScript: export default function mergeSiblings($candidate, topScore, $)
func MergeSiblings(candidate *goquery.Selection, topScore int, doc *goquery.Document) *goquery.Selection {
	// JavaScript: if (!$candidate.parent().length) { return $candidate; }
	if candidate.Parent().Length() == 0 {
		return candidate
	}
	
	// JavaScript: const siblingScoreThreshold = Math.max(10, topScore * 0.25);
	siblingScoreThreshold := 10
	if threshold := int(float64(topScore) * 0.25); threshold > siblingScoreThreshold {
		siblingScoreThreshold = threshold
	}
	
	// JavaScript: const wrappingDiv = $('<div></div>');
	// Create a temporary div container (we'll simulate this by collecting elements)
	var mergedElements []*goquery.Selection
	
	// JavaScript: $candidate.parent().children().each((index, sibling) => {
	candidate.Parent().Children().Each(func(index int, sibling *goquery.Selection) {
		// JavaScript: if (NON_TOP_CANDIDATE_TAGS_RE.test(sibling.tagName)) { return null; }
		tagName := strings.ToLower(goquery.NodeName(sibling))
		if NON_TOP_CANDIDATE_TAGS_RE.MatchString(tagName) {
			return
		}
		
		// JavaScript: const siblingScore = getScore($sibling);
		siblingScore := getScore(sibling)
		
		// JavaScript: if (siblingScore) {
		if siblingScore > 0 {
			// JavaScript: if ($sibling.get(0) === $candidate.get(0)) { wrappingDiv.append($sibling); }
			if isSameElement(sibling, candidate) {
				mergedElements = append(mergedElements, sibling)
			} else {
				// JavaScript: let contentBonus = 0;
				contentBonus := 0
				
				// JavaScript: const density = linkDensity($sibling);
				density := LinkDensity(sibling)
				
				// JavaScript: if (density < 0.05) { contentBonus += 20; }
				if density < 0.05 {
					contentBonus += 20
				}
				
				// JavaScript: if (density >= 0.5) { contentBonus -= 20; }
				if density >= 0.5 {
					contentBonus -= 20
				}
				
				// JavaScript: if ($sibling.attr('class') === $candidate.attr('class')) { contentBonus += topScore * 0.2; }
				siblingClass, _ := sibling.Attr("class")
				candidateClass, _ := candidate.Attr("class")
				if siblingClass != "" && siblingClass == candidateClass {
					contentBonus += int(float64(topScore) * 0.2)
				}
				
				// JavaScript: const newScore = siblingScore + contentBonus;
				newScore := siblingScore + contentBonus
				
				// JavaScript: if (newScore >= siblingScoreThreshold) { return wrappingDiv.append($sibling); }
				if newScore >= siblingScoreThreshold {
					mergedElements = append(mergedElements, sibling)
					return
				}
				
				// JavaScript: if (sibling.tagName === 'p') {
				if tagName == "p" {
					// JavaScript: const siblingContent = $sibling.text();
					siblingContent := sibling.Text()
					
					// JavaScript: const siblingContentLength = textLength(siblingContent);
					siblingContentLength := textLengthString(siblingContent)
					
					// JavaScript: if (siblingContentLength > 80 && density < 0.25) { return wrappingDiv.append($sibling); }
					if siblingContentLength > 80 && density < 0.25 {
						mergedElements = append(mergedElements, sibling)
						return
					}
					
					// JavaScript: if (siblingContentLength <= 80 && density === 0 && hasSentenceEnd(siblingContent)) { return wrappingDiv.append($sibling); }
					if siblingContentLength <= 80 && density == 0 && HasSentenceEnd(siblingContent) {
						mergedElements = append(mergedElements, sibling)
						return
					}
				}
			}
		}
	})
	
	// JavaScript: if (wrappingDiv.children().length === 1 && wrappingDiv.children().first().get(0) === $candidate.get(0)) { return $candidate; }
	if len(mergedElements) == 1 && isSameElement(mergedElements[0], candidate) {
		return candidate
	}
	
	// If we have merged multiple elements, we need to create a wrapper
	// For now, we'll return the candidate as-is since creating a proper wrapper
	// requires more complex DOM manipulation in goquery
	// TODO: In a full implementation, we'd create a div wrapper containing all merged elements
	if len(mergedElements) > 1 {
		// Return the candidate for now - this is a limitation of our current approach
		// In the full implementation, we'd need to create a wrapper div with all merged content
		return candidate
	}
	
	// JavaScript: return wrappingDiv;
	return candidate
}

// Helper function to compare if two selections refer to the same DOM element
func isSameElement(sel1, sel2 *goquery.Selection) bool {
	if sel1.Length() == 0 || sel2.Length() == 0 {
		return false
	}
	
	// Compare by getting the underlying nodes - this is a simplified approach
	// In a full implementation, we'd compare the actual DOM node references
	node1 := sel1.Get(0)
	node2 := sel2.Get(0)
	return node1 == node2
}

// textLengthString calculates text length with whitespace normalization (for compatibility with existing tests)
func textLengthString(text string) int {
	// Normalize whitespace like JavaScript - trim and collapse multiple spaces
	text = strings.TrimSpace(text)
	text = strings.Join(strings.Fields(text), " ")
	return len(text)
}

// linkDensityCompat provides link density calculation compatible with JavaScript tests
func linkDensityCompat(element *goquery.Selection) float64 {
	return LinkDensity(element)
}