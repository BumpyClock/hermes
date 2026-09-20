package hermes

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestCapturedPublicationDates(t *testing.T) {
	snapshot, _ := localDefinitions(t, `schema: 1
site: date-capture
hosts: [93.184.216.34]
metadata:
  date_published:
    - text_capture:
        selector: .author
        pattern: '([0-9]{1,2} [A-Za-z]+ [0-9]{4}|[0-9]{4}/[0-9]{1,2}/[0-9]{1,2}|[0-9]{4}年[0-9]{1,2}月[0-9]{1,2}日)'
        group: 1
content:
  groups: [[article]]
`)
	for _, tc := range []struct {
		source, want string
	}{
		{"17 September 2026", "2026-09-17"},
		{"1 June 2019", "2019-06-01"},
		{"17 Sep 2026", "2026-09-17"},
		{"2019/3/5", "2019-03-05"},
		{"2019/03/05", "2019-03-05"},
		{"2024/12/9", "2024-12-09"},
		{"2019年4月4日", "2019-04-04"},
		{"2024年12月09日", "2024-12-09"},
	} {
		t.Run(tc.source, func(t *testing.T) {
			source := `<title>Synthetic publication-date report</title><article>` +
				`<div class="author">Written by <a href="/author">Ada Reporter</a> on ` + tc.source + ` at 08:00.</div>` +
				`<p>The original synthetic article provides enough substantive reporting to exercise its date extraction.</p></article>`
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
					if result.DatePublished == nil || result.DatePublished.Format("2006-01-02") != tc.want {
						t.Fatalf("%s/%s date = %v, want %s", method, format, result.DatePublished, tc.want)
					}
				}
			}
		})
	}
}

func TestUTCPublicationDateAttribute(t *testing.T) {
	snapshot, _ := localDefinitions(t, `schema: 1
site: utc-date
hosts: [93.184.216.34]
metadata:
  date_published:
    - attribute: {selector: time, name: datetime}
content:
  groups: [[article]]
`)
	source := `<title>Synthetic live report</title><article>` +
		`<time datetime="2026-09-19 02:21:00 UTC">Publication time</time>` +
		`<p>This original report exercises a publisher timestamp with an explicit UTC suffix.</p></article>`
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
			if result.DatePublished == nil || result.DatePublished.Format(time.RFC3339) != "2026-09-19T02:21:00Z" {
				t.Fatalf("%s/%s publication time = %v", method, format, result.DatePublished)
			}
		}
	}
}
