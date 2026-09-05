package dom_test

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"

	"github.com/BumpyClock/hermes/internal/utils/dom"
)

func BenchmarkResponsiveImageLinks(b *testing.B) {
	markup := `<html><body>` + strings.Repeat(`<img class="photo" alt="Example" src="/image.jpg" srcset="/small.jpg 320w, /large.jpg 1280w">`, 32) + `</body></html>`
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(markup))
		if err != nil {
			b.Fatal(err)
		}
		dom.MakeLinksAbsolute(doc, "https://example.com/article")
	}
}

func BenchmarkWithinComment(b *testing.B) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`<html><body><main class="ARTICLE" id="story"><div class="ENTRY"><span class="AUTHOR" id="byline">Example Author</span></div></main></body></html>`))
	if err != nil {
		b.Fatal(err)
	}
	author := doc.Find("span")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dom.WithinComment(author)
	}
}
