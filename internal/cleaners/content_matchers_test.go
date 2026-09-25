package cleaners

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
)

func TestPreserveMatchersKeepDescendantsAndLaterMatches(t *testing.T) {
	const source = `<article><h1>Title</h1><p>Enough article text to retain.</p>` +
		`<div class="media"><span><img src="/one.jpg" width="1" height="1"></span></div>` +
		`<div class="media"><img src="/two.jpg" width="1" height="1"></div>` +
		`<script>hidden</script><div class="junk"><a href="/ad">Advertisement</a></div></article>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExtractCleanNode(doc.Find("article"), doc, ContentCleanOptions{
		CleanConditionally: true,
		PreserveMatchers:   []goquery.Matcher{cascadia.MustCompile(".media"), cascadia.MustCompile("script")},
	}).Html()
	if err != nil {
		t.Fatal(err)
	}
	for _, image := range []string{"one.jpg", "two.jpg"} {
		if !strings.Contains(result, image) {
			t.Fatalf("preservation lost a descendant or later match: %s", result)
		}
	}
	if strings.Contains(result, "hidden") || strings.Contains(result, "Advertisement") {
		t.Fatalf("preservation bypassed cleanup: %s", result)
	}
}
