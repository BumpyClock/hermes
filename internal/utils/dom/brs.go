package dom

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// BrsToPs converts consecutive <br /> tags into <p /> tags
// JavaScript implementation: src/utils/dom/brs-to-ps.js
//
// Given goquery Document, convert consecutive <br /> tags into <p /> tags instead.
// The algorithm exactly matches JavaScript:
// 1. Iterate through all BR elements
// 2. If next element is also BR, set collapsing=true and remove current BR
// 3. If collapsing and current BR is NOT followed by another BR, call paragraphize on that last BR
func BrsToPs(doc *goquery.Document) *goquery.Document {
	collapsing := false
	
	// JavaScript: $('br').each((index, element) => {
	// We need to collect all BR elements first to avoid mutation issues during iteration
	var brElements []*goquery.Selection
	doc.Find("br").Each(func(index int, element *goquery.Selection) {
		brElements = append(brElements, element)
	})
	
	for _, element := range brElements {
		// Skip if element was already removed
		if element.Length() == 0 {
			continue
		}
		
		// JavaScript: const nextElement = $element.next().get(0);
		// We need to check the actual next sibling, not just next element sibling
		// because text nodes between BRs should break the consecutive chain
		isNextBr := false
		
		parent := element.Parent()
		if parent.Length() > 0 {
			// Find the position of this BR in parent's contents
			var brIndex = -1
			var allContents []*goquery.Selection
			parent.Contents().Each(func(i int, s *goquery.Selection) {
				allContents = append(allContents, s)
				if s.Get(0) == element.Get(0) {
					brIndex = i
				}
			})
			
			// Check if the immediate next sibling is a BR
			if brIndex != -1 && brIndex+1 < len(allContents) {
				nextSibling := allContents[brIndex+1]
				tagName := strings.ToLower(goquery.NodeName(nextSibling))
				
				// Only consider it consecutive if the next sibling is immediately a BR
				// or if it's whitespace-only text followed by a BR
				if tagName == "br" {
					isNextBr = true
				} else if tagName == "#text" {
					text := nextSibling.Text()
					if strings.TrimSpace(text) == "" {
						// Check if the sibling after the whitespace is a BR
						if brIndex+2 < len(allContents) {
							nextNextSibling := allContents[brIndex+2]
							if strings.ToLower(goquery.NodeName(nextNextSibling)) == "br" {
								isNextBr = true
							}
						}
					}
				}
			}
		}
		
		if isNextBr {
			// JavaScript: collapsing = true; $element.remove();
			collapsing = true
			element.Remove()
		} else if collapsing {
			// JavaScript: collapsing = false; paragraphize(element, $, true);
			collapsing = false
			paragraphize(element, true)
		}
		// Note: Single BRs are left alone (no action taken)
	}
	
	return doc
}

// paragraphize converts a BR element and following inline siblings into a paragraph
// JavaScript implementation: src/utils/dom/paragraphize.js
//
// When br=true:
// 1. Create new P element
// 2. Move all following inline siblings into the P until hitting a block element
// 3. Replace the BR with the P
func paragraphize(node *goquery.Selection, br bool) {
	if !br || node.Length() == 0 {
		return
	}
	
	// JavaScript: const p = $('<p></p>');
	// Create the paragraph before the BR
	node.BeforeHtml("<p></p>")
	p := node.Prev()
	
	if p.Length() == 0 || !p.Is("p") {
		return
	}
	
	// JavaScript: let sibling = node.nextSibling;
	// We need to work with parent.Contents() to get both element and text nodes
	parent := node.Parent()
	if parent.Length() == 0 {
		node.Remove()
		return
	}
	
	// Find the position of our BR node in the parent's contents
	var brIndex = -1
	var allContents []*goquery.Selection
	parent.Contents().Each(func(i int, s *goquery.Selection) {
		allContents = append(allContents, s)
		if s.Get(0) == node.Get(0) {
			brIndex = i
		}
	})
	
	if brIndex == -1 {
		node.Remove()
		return
	}
	
	// Collect following siblings (both text and element nodes)
	var contentParts []string
	
	// JavaScript: while (sibling && !(sibling.tagName && BLOCK_LEVEL_TAGS_RE.test(sibling.tagName)))
	for i := brIndex + 1; i < len(allContents); i++ {
		sibling := allContents[i]
		tagName := strings.ToLower(goquery.NodeName(sibling))
		
		// For text nodes, goquery.NodeName returns "#text"
		if tagName == "#text" {
			// Text content
			text := sibling.Text()
			if strings.TrimSpace(text) != "" {
				contentParts = append(contentParts, text)
				sibling.Remove()
			}
		} else if BLOCK_LEVEL_TAGS_RE.MatchString(tagName) {
			// Stop at block level elements
			break
		} else {
			// Element content (inline elements)
			html, err := goquery.OuterHtml(sibling)
			if err == nil && html != "" {
				contentParts = append(contentParts, html)
				sibling.Remove()
			}
		}
	}
	
	// Add all collected content to the paragraph
	if len(contentParts) > 0 {
		fullContent := strings.Join(contentParts, "")
		p.SetHtml(fullContent)
	} else {
		// If no content, add a space to create a visible paragraph
		p.SetText(" ")
	}
	
	// JavaScript: $node.replaceWith(p); $node.remove();
	// Remove the BR since the paragraph now replaces it
	node.Remove()
}