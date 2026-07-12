package security

import (
	"strings"
	"testing"
)

func TestSanitizeHTMLPreservesSafeArticleMarkup(t *testing.T) {
	content := SanitizeHTML(`<figure><a href="https://gothamist.com/story"><img src="http://gothamist.com/photo.jpg" alt="Photo" width="640" height="360"></a><a href="http://gothamist.com/related">Related story</a><figcaption>Photo caption</figcaption></figure>`)

	for _, expected := range []string{
		`<figure>`,
		`<a href="https://gothamist.com/story"`,
		`<a href="http://gothamist.com/related"`,
		`<img src="http://gothamist.com/photo.jpg" alt="Photo" width="640" height="360">`,
		`<figcaption>Photo caption</figcaption>`,
	} {
		if !strings.Contains(content, expected) {
			t.Errorf("sanitized content missing %q: %q", expected, content)
		}
	}
}

func TestSanitizeHTMLRemovesUnsafeURLsAndSrcset(t *testing.T) {
	content := SanitizeHTML(`<a href="javascript:alert(1)">bad link</a><a href="data:text/html,bad">data link</a><a href="ftp://example.com/file">ftp link</a><img src="javascript:alert(1)"><img src="data:image/svg+xml,bad"><img src="ftp://example.com/image"><img srcset="javascript:alert(1) 1x, https://example.com/safe.jpg 2x">`)

	for _, unsafe := range []string{"javascript:", "data:", "ftp:", "srcset", "example.com"} {
		if strings.Contains(content, unsafe) {
			t.Errorf("sanitized content preserved unsafe value %q: %q", unsafe, content)
		}
	}
	if !strings.Contains(content, "bad link") || !strings.Contains(content, "data link") || !strings.Contains(content, "ftp link") {
		t.Fatalf("sanitized content removed safe text while stripping URLs: %q", content)
	}
}
