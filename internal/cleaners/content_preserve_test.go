package cleaners

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
)

func TestExtractCleanNodePreservesConfiguredMedia(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`<article><p>Article context remains visible.</p><div class="media"><img src="/photo.jpg" width="1" height="1" alt="Article image"></div><script>alert(1)</script></article>`))
	if err != nil {
		t.Fatal(err)
	}
	cleaned := ExtractCleanNode(doc.Find("article"), doc, ContentCleanOptions{
		CleanConditionally: true,
		PreserveMatchers:   []goquery.Matcher{cascadia.MustCompile(".media"), cascadia.MustCompile("script")},
	})
	content, err := cleaned.Html()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(content, "photo.jpg") || strings.Contains(content, "script") {
		t.Fatalf("configured media preservation or mandatory stripping failed: %s", content)
	}
}
