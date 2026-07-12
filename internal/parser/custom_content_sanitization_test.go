package parser

import (
	"strings"
	"testing"

	"github.com/BumpyClock/hermes/internal/extractors/custom"
	"github.com/PuerkitoBio/goquery"
)

func TestProcessCustomContentPreservesGothamistImageCaption(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`<html><body><div class="article-body"><p>Story text.</p><div class="image-none"><img src="https://gothamist.com/photo.jpg" alt="Photo" width="640" height="360"><i>Photo caption</i></div></div></body></html>`))
	if err != nil {
		t.Fatal(err)
	}
	content, err := processCustomContent(doc.Find(".article-body"), doc, custom.GetGothamistComExtractor().Content, "Gothamist headline", "https://gothamist.com/news/example")
	if err != nil {
		t.Fatal(err)
	}
	const expected = `<figure><img src="https://gothamist.com/photo.jpg" alt="Photo"/><figcaption>Photo caption</figcaption></figure>`
	if !strings.Contains(content, expected) {
		t.Fatalf("custom content missing transformed image caption: %q", content)
	}
	if formatted := formatContent(content, "html"); !strings.Contains(formatted, expected) {
		t.Fatalf("formatted content missing transformed image caption: %q", formatted)
	}
}
