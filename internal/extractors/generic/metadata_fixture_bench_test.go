package generic

import (
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// BenchmarkMetadataExtractFixtures covers both selector and meta tiers on real
// pages: cnet reaches the author, title, and date selector lists, while nytimes
// resolves author and date from meta tags and title from the selector lists.
func BenchmarkMetadataExtractFixtures(b *testing.B) {
	for _, fixture := range []string{"www.cnet.com", "www.nytimes.com"} {
		b.Run(fixture, func(b *testing.B) {
			source, err := os.ReadFile("../../fixtures/" + fixture + ".html") //nolint:gosec // Benchmark reads fixed local fixture paths.
			if err != nil {
				b.Fatal(err)
			}
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(source)))
			if err != nil {
				b.Fatal(err)
			}
			var metaCache []string
			seen := map[string]bool{}
			doc.Find("meta").Each(func(_ int, meta *goquery.Selection) {
				if name := meta.AttrOr("name", ""); name != "" && !seen[name] {
					metaCache = append(metaCache, name)
					seen[name] = true
				}
			})
			targetURL := "https://" + fixture + "/article"
			author := &GenericAuthorExtractor{}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				GenericTitleExtractor.Extract(doc.Selection, targetURL, metaCache)
				author.Extract(doc.Selection, metaCache)
				GenericDateExtractor.Extract(doc.Selection, targetURL, metaCache)
			}
		})
	}
}
