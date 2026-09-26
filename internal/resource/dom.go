package resource

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// NormalizeMetaTags normalizes meta tags for easier extraction
// - Converts 'content' attribute to 'value'
// - Converts 'property' attribute to 'name'
// This matches the JavaScript normalizeMetaTags function.
func NormalizeMetaTags(doc *goquery.Document) *goquery.Document {
	// Convert content -> value
	doc.Find("meta[content]").Each(func(i int, s *goquery.Selection) {
		content, exists := s.Attr("content")
		if exists {
			s.SetAttr("value", content)
			s.RemoveAttr("content")
		}
	})

	// Convert property -> name
	doc.Find("meta[property]").Each(func(i int, s *goquery.Selection) {
		property, exists := s.Attr("property")
		if exists {
			s.SetAttr("name", property)
			s.RemoveAttr("property")
		}
	})

	return doc
}

// ConvertLazyLoadedImages converts lazy-loaded images into normal images
// Many sites have img tags with no source, or placeholders in src attribute
// We need to properly fill in the src attribute from data-* attributes.
// Attributes are visited in source order, so when several attributes match,
// the last one sets src or srcset.
func ConvertLazyLoadedImages(doc *goquery.Document) *goquery.Document {
	doc.Find("img").Each(func(i int, img *goquery.Selection) {
		// Iterate a snapshot because SetAttr can append to the live attributes.
		for _, attr := range slices.Clone(img.Nodes[0].Attr) {
			attrName, value := attr.Key, attr.Val
			// Skip srcset attribute for srcset handling
			if attrName != "srcset" && IS_LINK_RE.MatchString(value) && IS_SRCSET_RE.MatchString(value) {
				img.SetAttr("srcset", value)
			} else if attrName != "src" && attrName != "srcset" &&
				IS_LINK_RE.MatchString(value) && IS_IMAGE_RE.MatchString(value) {
				// Check if value is JSON and extract src
				if src := extractSrcFromJSON(value); src != "" {
					img.SetAttr("src", src)
				} else {
					img.SetAttr("src", value)
				}
			}
		}
	})

	return doc
}

// extractSrcFromJSON attempts to extract src from JSON string.
func extractSrcFromJSON(str string) string {
	var data struct {
		Src string `json:"src"`
	}

	if err := json.Unmarshal([]byte(str), &data); err == nil {
		return data.Src
	}

	return ""
}

// Clean removes unwanted elements from the DOM
// Removes scripts, styles, forms, and comments.
func Clean(doc *goquery.Document) *goquery.Document {
	// Remove unwanted tags
	tagsList := strings.Split(TAGS_TO_REMOVE, ",")
	for _, tag := range tagsList {
		doc.Find(strings.TrimSpace(tag)).Remove()
	}

	// Remove comments - this is more complex in goquery
	// We need to traverse and find comment nodes
	cleanComments(doc)

	return doc
}

// cleanComments removes HTML comments from the document.
func cleanComments(doc *goquery.Document) {
	// In goquery, we need to traverse the DOM and remove comment nodes
	// This is less elegant than the jQuery version but achieves the same result
	doc.Find("*").Each(func(i int, s *goquery.Selection) {
		if len(s.Nodes) > 0 {
			node := s.Nodes[0]

			// Check child nodes for comments
			for child := node.FirstChild; child != nil; {
				next := child.NextSibling
				if child.Type == 8 { // Comment node type
					node.RemoveChild(child)
				}
				child = next
			}
		}
	})
}
