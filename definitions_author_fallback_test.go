package hermes

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestDefinitionsAuthorFallbackIgnoresSidebarNavigation(t *testing.T) {
	snapshot, _ := localDefinitions(t, `schema: 1
site: absent-author
hosts: [93.184.216.34]
metadata:
  author: [{text: ".post-author"}]
content:
  groups: [[article]]
`)
	const source = `<html><head><title>Original article without a byline</title></head><body>
<article><p>This original article supplies enough meaningful prose to exercise metadata fallback without an article author.</p></article>
<div id="sidebar_top"><div class="widget Profile"><div class="profile-info">
<a class="profile-link visit-profile pill-button" rel="author" href="/profile">Visit profile</a>
</div></div></div></body></html>`
	for _, format := range []string{"html", "markdown", "text"} {
		client := New(WithDefinitions(snapshot), WithContentType(format),
			WithTransport(definitionTransport(func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": {"text/html"}},
					Body:       io.NopCloser(strings.NewReader(source)),
					Request:    r,
				}, nil
			})))
		for _, method := range []string{"Parse", "ParseHTML"} {
			var result *Result
			var err error
			if method == "Parse" {
				result, err = client.Parse(context.Background(), "https://93.184.216.34/article")
			} else {
				result, err = client.ParseHTML(context.Background(), source, "https://93.184.216.34/article")
			}
			if err != nil {
				t.Fatal(err)
			}
			if result.Author != "" {
				t.Fatalf("%s/%s attributed sidebar navigation as author: %q", method, format, result.Author)
			}
			if !strings.Contains(result.Content, "This original article supplies enough meaningful prose") {
				t.Fatalf("%s/%s lost the article body", method, format)
			}
		}
	}
}
