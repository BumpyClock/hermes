package cleaners

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
)

func TestPreserveMatchersMatchStringSelectors(t *testing.T) {
	const source = `<article><h1>Title</h1><p>Enough article text to retain.</p>` +
		`<div class="media"><span><img src="/one.jpg" width="1" height="1"></span></div>` +
		`<div class="media"><img src="/two.jpg" width="1" height="1"></div>` +
		`<script>hidden</script><div class="junk"><a href="/ad">Advertisement</a></div></article>`
	clean := func(options ContentCleanOptions) string {
		t.Helper()
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(source))
		if err != nil {
			t.Fatal(err)
		}
		result, err := ExtractCleanNode(doc.Find("article"), doc, options).Html()
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	stringResult := clean(ContentCleanOptions{CleanConditionally: true, Preserve: []string{".media", "script", "["}})
	compiledResult := clean(ContentCleanOptions{
		CleanConditionally: true,
		PreserveMatchers:   []goquery.Matcher{cascadia.MustCompile(".media"), cascadia.MustCompile("script")},
	})
	if compiledResult != stringResult {
		t.Fatalf("compiled preserve changed output:\n%s\n%s", compiledResult, stringResult)
	}
	for _, image := range []string{"one.jpg", "two.jpg"} {
		if !strings.Contains(compiledResult, image) {
			t.Fatalf("preservation lost a descendant or later match: %s", compiledResult)
		}
	}
	if strings.Contains(compiledResult, "hidden") || strings.Contains(compiledResult, "Advertisement") {
		t.Fatalf("preservation bypassed cleanup: %s", compiledResult)
	}
}
