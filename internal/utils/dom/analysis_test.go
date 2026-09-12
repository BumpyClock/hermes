package dom_test

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BumpyClock/hermes/internal/utils/dom"
)

func TestLinkDensity(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected float64
	}{
		{
			name:     "no links",
			html:     `<div>This is pure text content without any links.</div>`,
			expected: 0.0,
		},
		{
			name:     "all links",
			html:     `<div><a href="#">Link one</a> <a href="#">Link two</a></div>`,
			expected: 1.0,
		},
		{
			name:     "half links",
			html:     `<div>Regular text <a href="#">Link text</a> more text</div>`,
			expected: 0.25, // "Link text" is 9 chars out of ~36 total
		},
		{
			name:     "nested links",
			html:     `<div>Text <p><a href="#">Nested link</a></p> more text</div>`,
			expected: 0.4, // More accurate calculation
		},
		{
			name:     "empty element",
			html:     `<div></div>`,
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			element := doc.Find("div").First()
			density := dom.LinkDensity(element)

			assert.InDelta(t, tt.expected, density, 0.1, "Link density should be close to expected")
		})
	}
}

func TestNodeIsSufficient(t *testing.T) {
	tests := []struct {
		name       string
		html       string
		sufficient bool
	}{
		{
			name: "sufficient content with paragraphs",
			html: `<div>
				<p>This is a substantial paragraph with enough content to be considered sufficient for an article. It has meaningful text that provides value.</p>
				<p>This is another paragraph that adds to the content quality and makes this a good candidate for article content.</p>
			</div>`,
			sufficient: true,
		},
		{
			name:       "insufficient short content",
			html:       `<div>Short text</div>`,
			sufficient: false,
		},
		{
			name: "too many links",
			html: `<div>
				<a href="#">Link 1</a> <a href="#">Link 2</a> <a href="#">Link 3</a>
				<a href="#">Link 4</a> <a href="#">Link 5</a> minimal text
			</div>`,
			sufficient: false,
		},
		{
			name: "good content without paragraphs but enough elements",
			html: `<div>
				<section>This is substantial content in a section element that provides meaningful information and creates a significant amount of text content.</section>
				<article>Another substantial piece of content that adds value to the reader experience and provides extensive information.</article>
				<div>Additional content that makes this element quite substantial and worthy of consideration with lots of detail.</div>
				<section>Even more content to ensure we have sufficient text length and element count.</section>
			</div>`,
			sufficient: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			element := doc.Find("div").First()
			result := dom.NodeIsSufficient(element)

			assert.Equal(t, tt.sufficient, result)
		})
	}
}

func TestWithinComment(t *testing.T) {
	tests := []struct {
		name      string
		html      string
		selector  string
		inComment bool
	}{
		{
			name: "element in comment section",
			html: `<div class="comments">
				<div class="comment">
					<p>This is a comment</p>
				</div>
			</div>`,
			selector:  ".comment p",
			inComment: true,
		},
		{
			name: "element not in comment",
			html: `<div class="content">
				<p>This is article content</p>
			</div>`,
			selector:  "p",
			inComment: false,
		},
		{
			name: "element in disqus",
			html: `<div id="disqus_thread">
				<div class="post">Comment post</div>
			</div>`,
			selector:  ".post",
			inComment: true,
		},
		{
			name: "nested comment detection",
			html: `<div class="main">
				<div class="comment-section">
					<div class="individual-comment">
						<span>Reply text</span>
					</div>
				</div>
			</div>`,
			selector:  "span",
			inComment: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			element := doc.Find(tt.selector).First()
			result := dom.WithinComment(element)

			assert.Equal(t, tt.inComment, result)
		})
	}
}

func BenchmarkAnalysisFunctions(b *testing.B) {
	html := `<div class="article-content">
		<p>This is a substantial paragraph with meaningful content.</p>
		<p>Another paragraph with <a href="#">some links</a> and more text.</p>
		<p>A third paragraph to provide good content density.</p>
	</div>`

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	element := doc.Find("div").First()

	b.Run("LinkDensity", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			dom.LinkDensity(element)
		}
	})

	b.Run("NodeIsSufficient", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			dom.NodeIsSufficient(element)
		}
	})

}
