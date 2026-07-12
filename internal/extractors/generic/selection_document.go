package generic

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func documentFromSelection(selection *goquery.Selection) *goquery.Document {
	html, err := selection.Html()
	if err != nil {
		return nil
	}

	if !strings.Contains(strings.ToLower(html), "<html") {
		html = "<html>" + html + "</html>"
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil
	}
	return doc
}
