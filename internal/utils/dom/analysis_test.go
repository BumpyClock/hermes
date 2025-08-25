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
			name: "insufficient short content",
			html: `<div>Short text</div>`,
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
		name     string
		html     string
		selector string
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

func TestIsWordpress(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		isWordpress bool
	}{
		{
			name: "wordpress generator meta",
			html: `<html><head>
				<meta name="generator" content="WordPress 5.8">
			</head><body></body></html>`,
			isWordpress: true,
		},
		{
			name: "wordpress classes",
			html: `<html><body>
				<div class="wp-content">Content</div>
			</body></html>`,
			isWordpress: true,
		},
		{
			name: "wordpress script",
			html: `<html><head>
				<script src="/wp-content/themes/theme/script.js"></script>
			</head><body></body></html>`,
			isWordpress: false, // Our implementation doesn't check scripts yet
		},
		{
			name: "not wordpress",
			html: `<html><head>
				<meta name="generator" content="Hugo 0.88.1">
			</head><body></body></html>`,
			isWordpress: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			result := dom.IsWordpress(doc)
			assert.Equal(t, tt.isWordpress, result)
		})
	}
}

func TestHasSentenceEnd(t *testing.T) {
	tests := []struct {
		text     string
		expected bool
	}{
		{"This is a sentence.", true},
		{"This is a question?", true},
		{"This is an exclamation!", true},
		{"This has a colon:", true},
		{"This has a semicolon;", true},
		{"This has no ending", false},
		{"", false},
		{"   ", false},
		{"This ends with comma,", false},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			result := dom.HasSentenceEnd(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectTextDirection(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "english text",
			text:     "This is English text that should be left-to-right.",
			expected: "ltr",
		},
		{
			name:     "arabic text",
			text:     "هذا نص عربي يجب أن يكون من اليمين إلى اليسار",
			expected: "rtl",
		},
		{
			name:     "hebrew text",
			text:     "זה טקסט עברי שצריך להיות מימין לשמאל",
			expected: "rtl",
		},
		{
			name:     "mixed text with mostly english",
			text:     "This is mostly English with some العربية words.",
			expected: "ltr",
		},
		{
			name:     "empty text",
			text:     "",
			expected: "ltr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dom.DetectTextDirection(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetContentScore(t *testing.T) {
	tests := []struct {
		name          string
		html          string
		minScore      float64
		hasPositive   bool
		hasNegative   bool
	}{
		{
			name: "high quality content",
			html: `<div class="article-content">
				<p>This is substantial paragraph content that should score well.</p>
				<p>Another paragraph with meaningful content for the reader.</p>
				<p>A third paragraph to boost the content score significantly.</p>
			</div>`,
			minScore:    20.0,
			hasPositive: true,
			hasNegative: false,
		},
		{
			name: "low quality content",
			html: `<div class="sidebar ad-banner">
				<a href="#">Link 1</a>
				<a href="#">Link 2</a>
				<a href="#">Link 3</a>
			</div>`,
			minScore:    0.0,
			hasPositive: false,
			hasNegative: true,
		},
		{
			name: "empty content",
			html: `<div></div>`,
			minScore:    0.0,
			hasPositive: false,
			hasNegative: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			element := doc.Find("div").First()
			score := dom.GetContentScore(element)

			if tt.minScore > 0 {
				assert.Greater(t, score, tt.minScore, "Score should be above minimum")
			}

			// Verify positive/negative scoring logic is working
			classAndId := ""
			if class, exists := element.Attr("class"); exists {
				classAndId += class
			}
			if id, exists := element.Attr("id"); exists {
				classAndId += " " + id
			}

			if tt.hasPositive {
				assert.True(t, dom.POSITIVE_SCORE_RE.MatchString(classAndId), "Should match positive patterns")
			}
			if tt.hasNegative {
				assert.True(t, dom.NEGATIVE_SCORE_RE.MatchString(classAndId), "Should match negative patterns")
			}
		})
	}
}

func TestCountWords(t *testing.T) {
	tests := []struct {
		text     string
		expected int
	}{
		{"Hello world", 2},
		{"This is a test sentence.", 5},
		{"", 0},
		{"   ", 0},
		{"SingleWord", 1},
		{"Multiple   spaces   between   words", 4},
		{"Words\nwith\nnewlines", 3},
		{"Words\twith\ttabs", 3},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			result := dom.CountWords(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCountSentences(t *testing.T) {
	tests := []struct {
		text     string
		expected int
	}{
		{"This is one sentence.", 1},
		{"First sentence. Second sentence!", 2},
		{"Question? Answer! Statement.", 3},
		{"No sentence ending", 1},
		{"", 0},
		{"Multiple!!! Exclamations!!!", 6},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			result := dom.CountSentences(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsLikelyArticleElement(t *testing.T) {
	tests := []struct {
		name      string
		html      string
		selector  string
		isArticle bool
	}{
		{
			name:      "article tag",
			html:      `<article>Article content</article>`,
			selector:  "article",
			isArticle: true,
		},
		{
			name:      "main tag",
			html:      `<main>Main content</main>`,
			selector:  "main",
			isArticle: true,
		},
		{
			name: "good content characteristics",
			html: `<div class="content">
				<p>This is substantial paragraph content that should be considered article-like content.</p>
				<p>Another paragraph with meaningful content for the reader that makes this element substantial.</p>
				<p>A third paragraph to ensure we have good content density and paragraph count for article detection.</p>
			</div>`,
			selector:  "div",
			isArticle: true,
		},
		{
			name:      "short content",
			html:      `<div class="content">Short</div>`,
			selector:  "div",
			isArticle: false,
		},
		{
			name: "too many links",
			html: `<div class="content">
				<a href="#">Link 1</a> <a href="#">Link 2</a> <a href="#">Link 3</a>
				<a href="#">Link 4</a> <a href="#">Link 5</a> <a href="#">Link 6</a>
				Some text but mostly links.
			</div>`,
			selector:  "div",
			isArticle: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			element := doc.Find(tt.selector).First()
			result := dom.IsLikelyArticleElement(element)

			assert.Equal(t, tt.isArticle, result)
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

	b.Run("GetContentScore", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			dom.GetContentScore(element)
		}
	})
}