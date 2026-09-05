package generic

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func BenchmarkMetadataExtraction(b *testing.B) {
	markup := `<html><head><meta name="dc.title" content="Example article"><meta name="dc.author" content="Example Author"><meta name="article:published_time" content="2025-01-02T03:04:05Z"></head><body>` +
		strings.Repeat(`<p>Article text with <a href="/related">related content</a> and additional detail.</p>`, 256) + `</body></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(markup))
	if err != nil {
		b.Fatal(err)
	}
	cache := []string{"dc.title", "dc.author", "article:published_time"}
	author := &GenericAuthorExtractor{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenericTitleExtractor.Extract(doc.Selection, "https://example.com/article", cache)
		author.Extract(doc.Selection, cache)
		GenericDateExtractor.Extract(doc.Selection, "https://example.com/article", cache)
	}
}
