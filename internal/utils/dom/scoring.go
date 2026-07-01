package dom

import (
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// JavaScript scoring functions from extractors/generic/content/scoring

// scoreCommas counts commas in text (more commas = better content quality)
// JavaScript: function scoreCommas(text).
func scoreCommas(text string) int {
	return strings.Count(text, ",")
}

// scoreLength gives bonus for text length in 50-character chunks
// JavaScript: function scoreLength(text, lengthBonus = 1).
func scoreLength(text string, lengthBonus int) int {
	if lengthBonus == 0 {
		lengthBonus = 1
	}
	return (len(text) / 50) * lengthBonus
}

// scoreParagraph provides multi-factor paragraph scoring
// JavaScript: function scoreParagraph(paragraph).
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
// JavaScript: function getOrInitScore(node, $, weightNodes = true).
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
// JavaScript: function setScore(node, $, score).
func setScore(element *goquery.Selection, score int) {
	element.SetAttr("data-content-score", itoa(score))
}

// getScore retrieves score from element
// JavaScript: function getScore(node, $).
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

// Helper functions - now using standard library.
func parseInt(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.Atoi(s)
}

func itoa(i int) string {
	return strconv.Itoa(i)
}

// Additional scoring functions to support cleanTags logic

// countParagraphs counts paragraph elements.
func countParagraphs(element *goquery.Selection) int {
	return element.Find("p").Length()
}

// countImages counts image elements.
func countImages(element *goquery.Selection) int {
	return element.Find("img").Length()
}

// countInputs counts input/form elements.
func countInputs(element *goquery.Selection) int {
	return element.Find("input, textarea, select, button").Length()
}

// countLists counts list elements.
func countLists(element *goquery.Selection) int {
	return element.Find("ul, ol, dl").Length()
}

// textLength gets text length with whitespace normalization.
func textLength(element *goquery.Selection) int {
	text := strings.TrimSpace(element.Text())
	// Normalize whitespace like JavaScript
	text = strings.Join(strings.Fields(text), " ")
	return len(text)
}

// addScore adds a score amount to a node
// JavaScript: export default function addScore($node, $, amount).
func addScore(element *goquery.Selection, amount int) *goquery.Selection {
	// JavaScript: const score = getOrInitScore($node, $) + amount;
	score := getOrInitScore(element, true) + amount
	// JavaScript: setScore($node, $, score);
	setScore(element, score)
	return element
}

// addToParent adds 1/4 of a child's score to its parent
// JavaScript: export default function addToParent(node, $, score).
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
// JavaScript: export default function scoreNode($node).
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
// JavaScript: function getWeight(node).
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
// JavaScript: export default function findTopCandidate($).
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

	// Sibling merging used to calculate extra candidates here, but returned the original candidate
	// on every path. Keep current behavior explicit and skip dead traversal work.
	return candidate
}

// linkDensityCompat provides link density calculation compatible with JavaScript tests.
func linkDensityCompat(element *goquery.Selection) float64 {
	return LinkDensity(element)
}
