// ABOUTME: Cleans H1 tags from article content based on count threshold analysis.
// ABOUTME: Removes H1s if less than 3, converts to H2s if 3 or more to preserve content.
package dom

import "github.com/PuerkitoBio/goquery"

// CleanHOnes processes H1 tags in a document based on their count.
// H1 tags are typically the article title, which should be extracted
// by the title extractor instead. If there's less than 3 of them (<3),
// strip them. Otherwise, turn them into H2s.
//
// This preserves content structure when there are multiple H1s that
// likely represent section headers rather than the main title.
//
// :param doc: A goquery Document to process
// :return: The modified goquery Document (returned for convenience, mutation is in-place)
func CleanHOnes(doc *goquery.Document) *goquery.Document {
	// Find all H1 elements in the document
	hOnes := doc.Find("h1")
	hOnesCount := hOnes.Length()

	if hOnesCount < 3 {
		// Remove H1s if there are fewer than 3
		hOnes.Each(func(index int, node *goquery.Selection) {
			node.Remove()
		})
	} else {
		// Convert H1s to H2s if there are 3 or more
		hOnes.Each(func(index int, node *goquery.Selection) {
			ConvertNodeTo(node, "h2")
		})
	}

	return doc
}