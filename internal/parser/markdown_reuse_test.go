package parser

import (
	"fmt"
	"strings"
	"testing"
)

func TestMarkdownConverterReuseMatchesFresh(t *testing.T) {
	sources := []string{
		"<p>Inline\t<strong> bold </strong>\r\n<em> emphasis </em> tail</p>",
		`<ol start="7"><li>First<ul><li>Nested</li></ul></li><li>Second</li></ol>`,
		`<p><a href="https://example.com/a" title="First">A</a> <a href="/b">B</a></p>`,
		`<pre><code>&lt;literal&gt;&#10;second line</code></pre><p>&amp;copy; &lt;tag&gt;</p>`,
		`<table><caption>Data</caption><tr><td>One</td><td>Two</td></tr></table>`,
		`<p>日本<strong>語</strong>！ <em>&nbsp;Read&nbsp;</em></p>`,
		`<figure><img src="/image-{width}.jpg" alt="Image"><figcaption>Caption</figcaption></figure>`,
		"",
	}
	for worker := 0; worker < 16; worker++ {
		t.Run(fmt.Sprintf("worker-%d", worker), func(t *testing.T) {
			t.Parallel()
			for repeat := 0; repeat < 4; repeat++ {
				for _, source := range sources {
					source = fmt.Sprintf("<h2>Article %d/%d</h2>", worker, repeat) + source
					want, err := newMarkdownConverter().ConvertString(source)
					if err != nil {
						t.Fatal(err)
					}
					if got := convertToMarkdown(source); got != want {
						t.Fatalf("reused converter changed output: got %q, want %q", got, want)
					}
				}
			}
		})
	}
}

func BenchmarkMarkdownConverterReuse(b *testing.B) {
	source := strings.Repeat(`<p>Article <strong>words</strong> and <a href="/report">link</a>.</p>`, 10)
	for _, parallel := range []bool{false, true} {
		b.Run(fmt.Sprintf("parallel-%t", parallel), func(b *testing.B) {
			for _, reuse := range []bool{false, true} {
				b.Run(fmt.Sprintf("reuse-%t", reuse), func(b *testing.B) {
					b.ReportAllocs()
					convert := func() {
						if reuse {
							_ = convertToMarkdown(source)
						} else if _, err := newMarkdownConverter().ConvertString(source); err != nil {
							b.Fatal(err)
						}
					}
					if parallel {
						b.RunParallel(func(pb *testing.PB) {
							for pb.Next() {
								convert()
							}
						})
					} else {
						for i := 0; i < b.N; i++ {
							convert()
						}
					}
				})
			}
		})
	}
}
