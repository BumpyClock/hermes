package dom

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// StripUnlikelyCandidates loops through the provided document and removes any non-link nodes
// that are unlikely candidates for article content.
//
// Links are ignored because there are very often links to content
// that are identified as non-body-content, but may be inside
// article-like content.
//
// :param doc: a goquery Document to strip nodes from
// :return: the cleaned goquery Document
func StripUnlikelyCandidates(doc *goquery.Document) *goquery.Document {
	// Find all elements except links
	doc.Find("*").Not("a").Each(func(index int, node *goquery.Selection) {
		classes, classExists := node.Attr("class")
		id, idExists := node.Attr("id")
		
		// Skip if no class or id attributes
		if !classExists && !idExists {
			return
		}

		// Combine class and id for testing
		classAndId := ""
		if classExists {
			classAndId += classes
		}
		if idExists {
			if classAndId != "" {
				classAndId += " "
			}
			classAndId += id
		}

		// If it's empty, skip
		if strings.TrimSpace(classAndId) == "" {
			return
		}

		// Check against whitelist first - if it matches, keep it
		if CANDIDATES_WHITELIST.MatchString(classAndId) {
			return
		}

		// Check against blacklist - if it matches, remove it
		if CANDIDATES_BLACKLIST.MatchString(classAndId) {
			node.Remove()
		}
	})

	return doc
}