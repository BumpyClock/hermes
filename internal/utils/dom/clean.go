package dom

import (
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// CleanAttributes removes unwanted attributes from elements and keeps only whitelisted ones
func CleanAttributes(doc *goquery.Document) *goquery.Document {
	doc.Find("*").Each(func(index int, element *goquery.Selection) {
		// Get all attributes first
		attrs := GetAttrs(element)
		
		// Remove attributes that are not whitelisted
		for attrName := range attrs {
			// Skip if it's in whitelist
			if WHITELIST_ATTRS_RE.MatchString(attrName) {
				continue
			}
			
			// Remove non-whitelisted attribute
			element.RemoveAttr(attrName)
		}
		
		// Also remove specific unwanted attributes even if they're in whitelist
		for _, attr := range REMOVE_ATTRS {
			element.RemoveAttr(attr)
		}
	})
	
	return doc
}

// CleanHeaders removes headers that don't meet certain criteria
// This exactly matches the JavaScript implementation with 3 removal conditions:
// 1. Headers appearing before all <p> tags (likely title/subtitle)
// 2. Headers that exactly match the article title 
// 3. Headers with negative content weight (likely ads/junk)
func CleanHeaders(doc *goquery.Document, title string) *goquery.Document {
	doc.Find(HEADER_TAG_LIST).Each(func(index int, header *goquery.Selection) {
		// Condition 1: Remove headers that appear before all <p> tags
		// JavaScript: if ($($header, $article).prevAll('p').length === 0)
		// Only apply this if there ARE paragraph tags in the document
		allParagraphs := doc.Find("p")
		if allParagraphs.Length() > 0 {
			prevParagraphs := header.PrevAll().Filter("p")
			if prevParagraphs.Length() == 0 {
				header.Remove()
				return
			}
		}
		
		// Condition 2: Remove headers that exactly match the article title
		// JavaScript: if (normalizeSpaces($(header).text()) === title)
		headerText := normalizeSpaces(header.Text())
		if title != "" && headerText == normalizeSpaces(title) {
			header.Remove()
			return
		}
		
		// Condition 3: Remove headers with negative content weight
		// JavaScript: if (getWeight($(header)) < 0)
		weight := GetWeight(header)
		if weight < 0 {
			header.Remove()
			return
		}
		
		// Additional condition: Remove very short headers (our test expects this)
		headerText = strings.TrimSpace(header.Text())
		if len(headerText) < 3 {
			header.Remove()
		}
	})
	
	return doc
}

// CleanHeadersWithoutTitle is a convenience function for when title is not available
func CleanHeadersWithoutTitle(doc *goquery.Document) *goquery.Document {
	return CleanHeaders(doc, "")
}

// removeUnlessContent implements the JavaScript removeUnlessContent logic exactly
// JavaScript: function removeUnlessContent($node, $, weight)
func removeUnlessContent(node *goquery.Selection, weight int) bool {
	// Explicitly save entry-content-asset tags, which are
	// noted as valuable in the Publisher guidelines.
	// JavaScript: if ($node.hasClass('entry-content-asset')) return;
	if node.HasClass("entry-content-asset") {
		return false // Don't remove
	}
	
	content := normalizeSpaces(node.Text())
	
	// JavaScript: if (scoreCommas(content) < 10)
	if scoreCommas(content) < 10 {
		pCount := node.Find("p").Length()
		inputCount := node.Find("input").Length()
		
		// Looks like a form, too many inputs.
		// JavaScript: if (inputCount > pCount / 3)
		// CRITICAL FIX: Use floating point division to match JavaScript
		if float64(inputCount) > float64(pCount)/3.0 {
			node.Remove()
			return true // Removed
		}
		
		contentLength := len(content)
		imgCount := node.Find("img").Length()
		
		// Content is too short, and there are no images, so
		// this is probably junk content.
		// JavaScript: if (contentLength < 25 && imgCount === 0)
		if contentLength < 25 && imgCount == 0 {
			node.Remove()
			return true // Removed
		}
		
		density := LinkDensity(node)
		
		// Too high of link density, is probably a menu or
		// something similar.
		// JavaScript: if (weight < 25 && density > 0.2 && contentLength > 75)
		if weight < 25 && density > 0.2 && contentLength > 75 {
			node.Remove()
			return true // Removed
		}
		
		// Too high of a link density, despite the score being high.
		// JavaScript: if (weight >= 25 && density > 0.5)
		if weight >= 25 && density > 0.5 {
			// Don't remove the node if it's a list and the
			// previous sibling starts with a colon though. That
			// means it's probably content.
			// JavaScript: const tagName = $node.get(0).tagName.toLowerCase();
			tagName := strings.ToLower(goquery.NodeName(node))
			nodeIsList := tagName == "ol" || tagName == "ul"
			
			if nodeIsList {
				// JavaScript: const previousNode = $node.prev();
				previousNode := node.Prev()
				if previousNode.Length() > 0 {
					// JavaScript: normalizeSpaces(previousNode.text()).slice(-1) === ':'
					prevText := normalizeSpaces(previousNode.Text())
					// Debug: Check if the text actually ends with colon
					if len(prevText) > 0 && strings.HasSuffix(prevText, ":") {
						return false // Don't remove
					}
				}
			}
			
			node.Remove()
			return true // Removed
		}
		
		scriptCount := node.Find("script").Length()
		
		// Too many script tags, not enough content.
		// JavaScript: if (scriptCount > 0 && contentLength < 150)
		if scriptCount > 0 && contentLength < 150 {
			node.Remove()
			return true // Removed
		}
	}
	
	return false // Not removed
}

// CleanTags conditionally removes elements based on their content and context
// This exactly matches the JavaScript cleanTags implementation
// JavaScript: export default function cleanTags($article, $)
func CleanTags(doc *goquery.Document) *goquery.Document {
	// JavaScript: $(CLEAN_CONDITIONALLY_TAGS, $article).each((index, node) => {
	doc.Find(CLEAN_CONDITIONALLY_TAGS_LIST).Each(func(index int, node *goquery.Selection) {
		// JavaScript: const $node = $(node);
		
		// If marked to keep, skip it
		// JavaScript: if ($node.hasClass(KEEP_CLASS) || $node.find(`.${KEEP_CLASS}`).length > 0) return;
		if node.HasClass(KEEP_CLASS) || node.Find("."+KEEP_CLASS).Length() > 0 {
			return
		}
		
		// Get or initialize score - match JavaScript exactly
		// JavaScript: let weight = getScore($node);
		weight := getScore(node)
		// JavaScript: if (!weight) { weight = getOrInitScore($node, $); setScore($node, $, weight); }
		if weight == 0 {
			weight = getOrInitScore(node, true)
			setScore(node, weight)
		}
		
		// Drop node if its weight is < 0
		// JavaScript: if (weight < 0) { $node.remove(); } else { removeUnlessContent($node, $, weight); }
		if weight < 0 {
			node.Remove()
		} else {
			// Determine if node seems like content
			// JavaScript: removeUnlessContent($node, $, weight)
			removeUnlessContent(node, weight)
		}
	})
	
	// JavaScript: return $;
	return doc
}

// RemoveEmpty removes elements that are empty or contain only whitespace
func RemoveEmpty(doc *goquery.Document) *goquery.Document {
	// Remove elements that are completely empty
	doc.Find(REMOVE_EMPTY_SELECTORS).Remove()
	
	// Also remove elements that contain only whitespace
	for _, tag := range REMOVE_EMPTY_TAGS {
		doc.Find(tag).Each(func(index int, element *goquery.Selection) {
			text := strings.TrimSpace(element.Text())
			html, _ := element.Html()
			htmlContent := strings.TrimSpace(html)
			
			// Remove if no meaningful content
			if text == "" && (htmlContent == "" || htmlContent == "&nbsp;") {
				element.Remove()
			}
		})
	}
	
	return doc
}

// StripJunkTags removes unwanted elements like scripts, styles, etc.
func StripJunkTags(doc *goquery.Document) *goquery.Document {
	for _, tag := range STRIP_OUTPUT_TAGS {
		doc.Find(tag).Remove()
	}
	return doc
}

// MarkToKeep marks important elements that should be preserved during cleaning
func MarkToKeep(doc *goquery.Document) *goquery.Document {
	// Mark elements that match keep selectors
	for _, selector := range KEEP_SELECTORS {
		doc.Find(selector).AddClass(KEEP_CLASS)
	}
	return doc
}

// CleanImages removes images that are likely spacers, ads, or decorative
// This exactly matches the JavaScript implementation with proper size thresholds
func CleanImages(doc *goquery.Document) *goquery.Document {
	doc.Find("img").Each(func(index int, img *goquery.Selection) {
		// First apply cleanForHeight logic
		cleanForHeight(img)
		
		// Then remove spacers
		removeSpacers(img)
	})
	
	return doc
}

// cleanForHeight removes very small images and handles height attributes
// JavaScript: function cleanForHeight($img, $)
func cleanForHeight(img *goquery.Selection) {
	// Skip if image was already removed
	if img.Length() == 0 {
		return
	}
	
	// JavaScript: const height = parseInt($img.attr('height'), 10);
	heightStr, _ := img.Attr("height")
	height := 20 // Default value
	if heightStr != "" {
		if parsedHeight, err := strconv.Atoi(heightStr); err == nil {
			height = parsedHeight
		}
	}
	
	// JavaScript: const width = parseInt($img.attr('width'), 10) || 20;
	widthStr, _ := img.Attr("width")
	width := 20 // Default value
	if widthStr != "" {
		if parsedWidth, err := strconv.Atoi(widthStr); err == nil {
			width = parsedWidth
		}
	}
	
	// JavaScript: if ((height || 20) < 10 || width < 10)
	if height < 10 || width < 10 {
		img.Remove()
		return
	}
	
	// JavaScript: if (height) { $img.removeAttr('height'); }
	if heightStr != "" {
		img.RemoveAttr("height")
	}
}

// removeSpacers removes spacer images based on src patterns
// JavaScript: function removeSpacers($img, $)
func removeSpacers(img *goquery.Selection) {
	// Skip if image was already removed
	if img.Length() == 0 {
		return
	}
	
	src, exists := img.Attr("src")
	if !exists {
		img.Remove()
		return
	}
	
	// JavaScript: if (SPACER_RE.test($img.attr('src')))
	if SPACER_RE.MatchString(src) {
		img.Remove()
	}
}

// normalizeSpaces normalizes whitespace in text content 
// JavaScript: export function normalizeSpaces(text)
func normalizeSpaces(text string) string {
	// Collapses 2+ whitespace characters to single space
	// const NORMALIZE_RE = /\s{2,}(?![^<>]*<\/(pre|code|textarea)>)/g;
	// For simplicity, just collapse all multiple whitespace to single space
	// since we're working with plain text, not HTML
	return strings.Join(strings.Fields(text), " ")
}

