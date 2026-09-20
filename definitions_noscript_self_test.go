package hermes

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestDefinitionsNoscriptSelfRecoveryIsLocalAndOrdered(t *testing.T) {
	s := transformDefinitions(t, `    - target: descendants
      selector: noscript
      noscript.recover:
        source: {self: true}
        target: {self: true}
        position: replace
`)
	source := `<article><p>Opening recovered article.</p><div><noscript><picture><img src="/one.jpg" alt="One"><img src="/two.jpg" alt="Two"></picture></noscript></div><figure><noscript>&lt;picture&gt;&lt;img src="/three.jpg" alt="Three"&gt;&lt;/picture&gt;</noscript><figcaption>Nested parent caption.</figcaption></figure><noscript></noscript><p>Closing recovered article.</p></article>`
	const articleURL = "https://93.184.216.34/story"
	for _, format := range []string{"html", "markdown", "text"} {
		for _, op := range []string{"Parse", "ParseHTML"} {
			t.Run(format+"/"+op, func(t *testing.T) {
				c := New(WithDefinitions(s), WithContentType(format), WithTransport(definitionTransport(func(r *http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"text/html"}}, Body: io.NopCloser(strings.NewReader(source)), Request: r}, nil
				})))
				var result *Result
				var err error
				if op == "Parse" {
					result, err = c.Parse(context.Background(), articleURL)
				} else {
					result, err = c.ParseHTML(context.Background(), source, articleURL)
				}
				if err != nil {
					t.Fatal(err)
				}
				for _, value := range []string{"Opening recovered article.", "Nested parent caption.", "Closing recovered article."} {
					if strings.Count(result.Content, value) != 1 {
						t.Fatalf("%s lost or duplicated %q: %s", format, value, result.Content)
					}
				}
				if strings.Contains(result.Content, "<noscript") {
					t.Fatalf("%s retained noscript: %s", format, result.Content)
				}
				if format != "text" {
					previous := -1
					for _, image := range []string{"one.jpg", "two.jpg", "three.jpg"} {
						if strings.Count(result.Content, image) != 1 {
							t.Fatalf("%s image count for %q: %s", format, image, result.Content)
						}
						index := strings.Index(result.Content, image)
						if index <= previous {
							t.Fatalf("%s image order changed: %s", format, result.Content)
						}
						previous = index
					}
				}
			})
		}
	}
}
