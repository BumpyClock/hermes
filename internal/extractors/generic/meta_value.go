package generic

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func firstMetaValue(selection *goquery.Selection, names []string, valid func(string) bool) string {
	for _, name := range names {
		for _, selector := range []string{
			`meta[name="` + name + `"]`,
			`meta[property="` + name + `"]`,
		} {
			if value := firstMetaAttr(selection, selector, "value", valid); value != "" {
				return value
			}
			if value := firstMetaAttr(selection, selector, "content", valid); value != "" {
				return value
			}
		}
	}
	return ""
}

func firstMetaAttr(selection *goquery.Selection, selector, attr string, valid func(string) bool) string {
	value := strings.TrimSpace(selection.Find(selector).AttrOr(attr, ""))
	if value == "" {
		return ""
	}
	if valid != nil && !valid(value) {
		return ""
	}
	return value
}
