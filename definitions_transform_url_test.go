package hermes

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestDefinitionsURLBuildPreservesEscapedSegments(t *testing.T) {
	tests := []struct {
		name, base, path, want string
	}{
		{
			"slash-bearing segment followed by another",
			"https://media.example.com",
			"[{literal: 'a/'}, {literal: b}]",
			"https://media.example.com/a%2F/b",
		},
		{
			"multiple segment slashes followed by another",
			"https://media.example.com",
			"[{literal: 'a/b/'}, {literal: c}]",
			"https://media.example.com/a%2Fb%2F/c",
		},
		{
			"attribute segment followed by another",
			"https://media.example.com",
			"[{attribute: title, required: true}, {literal: c}]",
			"https://media.example.com/a%2Fb%2F/c",
		},
		{
			"base ending with escaped slash",
			"https://media.example.com/prefix%2F",
			"[{literal: next}]",
			"https://media.example.com/prefix%2F/next",
		},
		{
			"base escaped slash before directory separator",
			"https://media.example.com/prefix%2F/",
			"[{literal: next}]",
			"https://media.example.com/prefix%2F/next",
		},
		{
			"only segment ends with slash",
			"https://media.example.com",
			"[{literal: 'a/'}]",
			"https://media.example.com/a%2F",
		},
		{
			"last segment ends with slash",
			"https://media.example.com",
			"[{literal: a}, {literal: 'b/'}]",
			"https://media.example.com/a/b%2F",
		},
		{
			"ordinary base directory separator",
			"https://media.example.com/prefix/",
			"[{literal: a}, {literal: b}]",
			"https://media.example.com/prefix/a/b",
		},
	}
	const source = `<article><p>Read the <a title="a/b/">article resource</a> for further reporting.</p></article>`
	const articleURL = "https://93.184.216.34/story"
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := transformDefinitions(t, `    - target: descendants
      selector: a
      url.build:
        attribute: href
        base: `+test.base+`
        path: `+test.path+"\n")
			client := New(WithDefinitions(s), WithTransport(definitionTransport(func(r *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"text/html"}}, Body: io.NopCloser(strings.NewReader(source)), Request: r}, nil
			})))
			for _, op := range []string{"Parse", "ParseHTML"} {
				t.Run(op, func(t *testing.T) {
					var result *Result
					var err error
					if op == "Parse" {
						result, err = client.Parse(context.Background(), articleURL)
					} else {
						result, err = client.ParseHTML(context.Background(), source, articleURL)
					}
					if err != nil {
						t.Fatal(err)
					}
					doc, err := goquery.NewDocumentFromReader(strings.NewReader(result.Content))
					if err != nil {
						t.Fatal(err)
					}
					if got := doc.Find("a").AttrOr("href", ""); got != test.want {
						t.Fatalf("href = %q, want %q", got, test.want)
					}
				})
			}
		})
	}
}
