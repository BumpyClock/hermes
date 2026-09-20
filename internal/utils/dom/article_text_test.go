package dom

import (
	"strings"
	"testing"
)

func TestStripTagsWithBlockBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name, source, want string
	}{
		{"paragraphs", `<p>First</p><p>Second</p>`, "First Second"},
		{"line breaks", `<p>First<br>Second<br/>Third</p>`, "First Second Third"},
		{"list items", `<ul><li>First</li><li>Second</li></ul>`, "First Second"},
		{"table cells and caption", `<table><caption>Data</caption><tr><th>First</th><th>Second</th></tr><tr><td>One</td><td>Two</td></tr></table>`, "Data First Second One Two"},
		{"figure caption", `<figure><img src="/image"><figcaption>Caption</figcaption></figure><p>Body</p>`, "Caption Body"},
		{"inline adjacency", `<p>pre<span>fix</span>, <strong>mid</strong><em>dle</em>.</p>`, "prefix, middle."},
		{"literal markup", `<p>&lt;tag&gt; &amp;amp; &lt;/tag&gt;</p>`, "<tag> &amp; </tag>"},
		{"nonvisible nodes", `<p>Visible</p><style>hidden</style><script>hidden</script><template>hidden</template><!--hidden-->`, "Visible"},
		{"nonvisible only", `<script>hidden</script><template>hidden</template>`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := strings.Join(strings.Fields(StripTagsWithBlockBoundaries(tc.source)), " ")
			if got != tc.want {
				t.Fatalf("article text = %q, want %q", got, tc.want)
			}
		})
	}
	if got := StripTags(`<span>pre</span><span>fix</span>`); got != "prefix" {
		t.Fatalf("flat metadata extraction changed: %q", got)
	}
}
