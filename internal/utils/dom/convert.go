package dom

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ConvertToParagraphs loops through the provided doc, and converts any p-like elements to
// actual paragraph tags.
//
// Things fitting this criteria:
// * Multiple consecutive <br /> tags.
// * <div /> tags without block level elements inside of them
// * <span /> tags who are not children of <p /> or <div /> tags.
//
// :param doc: A goquery Document to search
// :return: goquery Document with new p elements
// (By-reference mutation, though. Returned just for convenience.)
func ConvertToParagraphs(doc *goquery.Document) *goquery.Document {
	doc = BrsToPs(doc)
	doc = convertDivs(doc)
	doc = convertSpans(doc)

	return doc
}

// convertDivs converts div elements that don't contain block-level elements to paragraphs
func convertDivs(doc *goquery.Document) *goquery.Document {
	doc.Find("div").Each(func(index int, div *goquery.Selection) {
		// Check if this div contains any block-level elements
		convertible := div.Find(DIV_TO_P_BLOCK_TAGS_LIST).Length() == 0

		if convertible {
			ConvertNodeTo(div, "p")
		}
	})

	return doc
}

// convertSpans converts span elements that are not children of p, div, li, or figcaption to paragraphs
func convertSpans(doc *goquery.Document) *goquery.Document {
	doc.Find("span").Each(func(index int, span *goquery.Selection) {
		// Check if this span has parent p, div, li, or figcaption elements
		convertible := span.ParentsFiltered("p, div, li, figcaption").Length() == 0
		if convertible {
			ConvertNodeTo(span, "p")
		}
	})

	return doc
}

// ConvertNodeTo converts a node to a different tag type while preserving attributes and content
func ConvertNodeTo(node *goquery.Selection, tag string) {
	if node.Length() == 0 {
		return
	}

	// Get all attributes from the original node
	attrs := GetAttrs(node)

	// Build attribute string
	var attribParts []string
	for key, value := range attrs {
		if value != "" {
			attribParts = append(attribParts, fmt.Sprintf(`%s="%s"`, key, value))
		} else {
			attribParts = append(attribParts, key)
		}
	}
	attribString := strings.Join(attribParts, " ")

	// Get the HTML content
	html, err := node.Html()
	if err != nil {
		// Fallback to text content if HTML parsing fails
		html = node.Text()
	}

	// Create the replacement HTML
	var replacement string
	if attribString != "" {
		replacement = fmt.Sprintf("<%s %s>%s</%s>", tag, attribString, html, tag)
	} else {
		replacement = fmt.Sprintf("<%s>%s</%s>", tag, html, tag)
	}

	// Replace the node
	node.ReplaceWithHtml(replacement)
}

