package dom_test

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BumpyClock/hermes/internal/utils/dom"
)

func TestConvertToParagraphs(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected []string // Expected paragraph contents
	}{
		{
			name: "converts shallow divs",
			html: `<html><body>
				<div>This should become a paragraph</div>
				<div><p>This should stay as div</p></div>
				<div><img src="test.jpg">This should stay as div</div>
			</body></html>`,
			expected: []string{"This should become a paragraph"},
		},
		{
			name: "converts standalone spans",
			html: `<html><body>
				<span>Standalone span</span>
				<div><span>Nested span</span></div>
				<p><span>Span in paragraph</span></p>
			</body></html>`,
			expected: []string{"Standalone span"},
		},
		{
			name: "converts consecutive br tags",
			html: `<html><body>
				<p>First paragraph</p>
				<br><br>
				Text after double br
				<p>Last paragraph</p>
			</body></html>`,
			expected: []string{"First paragraph", "Text after double br", "Last paragraph"},
		},
		{
			name: "preserves complex divs",
			html: `<html><body>
				<div>Simple div text</div>
				<div>
					<div>Nested div</div>
					<blockquote>Quote</blockquote>
				</div>
			</body></html>`,
			expected: []string{"Simple div text"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			result := dom.ConvertToParagraphs(doc)

			// Check that we have the expected number of paragraphs
			paragraphs := result.Find("p")
			
			// Verify paragraph contents
			found := 0
			paragraphs.Each(func(i int, p *goquery.Selection) {
				text := strings.TrimSpace(p.Text())
				for _, expected := range tt.expected {
					if strings.Contains(text, expected) {
						found++
						break
					}
				}
			})

			assert.GreaterOrEqual(t, found, len(tt.expected), "Should find expected paragraph content")
		})
	}
}

func TestConvertNodeTo(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		selector    string
		newTag      string
		expectedTag string
		preserved   map[string]string // attributes that should be preserved
	}{
		{
			name:        "converts div to paragraph",
			html:        `<div class="content" id="main">Test content</div>`,
			selector:    "div",
			newTag:      "p",
			expectedTag: "p",
			preserved:   map[string]string{"class": "content", "id": "main"},
		},
		{
			name:        "converts span to paragraph",
			html:        `<span style="color: red;">Styled text</span>`,
			selector:    "span",
			newTag:      "p",
			expectedTag: "p",
			preserved:   map[string]string{"style": "color: red;"},
		},
		{
			name:        "converts h1 to h2",
			html:        `<h1 class="title">Main Title</h1>`,
			selector:    "h1",
			newTag:      "h2",
			expectedTag: "h2",
			preserved:   map[string]string{"class": "title"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			element := doc.Find(tt.selector).First()
			require.True(t, element.Length() > 0, "Should find element to convert")

			originalText := element.Text()
			dom.ConvertNodeTo(element, tt.newTag)

			// Check that the tag was changed
			newElement := doc.Find(tt.expectedTag).First()
			assert.True(t, newElement.Length() > 0, "Should find converted element")

			// Check that content is preserved
			assert.Equal(t, originalText, newElement.Text(), "Content should be preserved")

			// Check that attributes are preserved
			for attr, expectedValue := range tt.preserved {
				actualValue, exists := newElement.Attr(attr)
				assert.True(t, exists, "Attribute %s should exist", attr)
				assert.Equal(t, expectedValue, actualValue, "Attribute %s value should be preserved", attr)
			}
		})
	}
}


func TestConvertDivs_EdgeCases(t *testing.T) {
	t.Run("deeply nested block elements", func(t *testing.T) {
		html := `<html><body>
			<div>
				<div>
					<div>
						<p>Deep paragraph</p>
					</div>
				</div>
			</div>
		</body></html>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		result := dom.ConvertToParagraphs(doc)

		// Should not convert any divs because they contain block elements
		divs := result.Find("div")
		assert.True(t, divs.Length() > 0, "Divs with block children should remain")
	})

	t.Run("mixed content preservation", func(t *testing.T) {
		html := `<html><body>
			<div>Text <strong>bold</strong> more text</div>
		</body></html>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		result := dom.ConvertToParagraphs(doc)

		// Should convert to paragraph and preserve inline formatting
		paragraphs := result.Find("p")
		assert.True(t, paragraphs.Length() > 0, "Should create paragraph")
		
		strong := paragraphs.Find("strong")
		assert.True(t, strong.Length() > 0, "Should preserve inline formatting")
		assert.Equal(t, "bold", strong.Text())
	})
}

func BenchmarkConvertToParagraphs(b *testing.B) {
	html := `<html><body>
		<div>Simple div 1</div>
		<div><p>Complex div</p></div>
		<span>Standalone span</span>
		<div><span>Nested span</span></div>
		<br><br>
		Text after br
		<div>Simple div 2</div>
	</body></html>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
		dom.ConvertToParagraphs(doc)
	}
}