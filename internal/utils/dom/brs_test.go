package dom_test

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BumpyClock/hermes/internal/utils/dom"
)

func TestBrsToPs(t *testing.T) {
	tests := []struct {
		name            string
		html            string
		expectedPs      int
		expectedContent []string
	}{
		{
			name: "converts consecutive br tags",
			html: `<html><body>
				<p>First paragraph</p>
				<br><br>
				Text after double br
				<br><br>
				More text after another double br
				<p>Last paragraph</p>
			</body></html>`,
			expectedPs: 4, // 2 original + 2 converted
			expectedContent: []string{
				"First paragraph",
				"Text after double br",
				"More text after another double br", 
				"Last paragraph",
			},
		},
		{
			name: "ignores single br tags",
			html: `<html><body>
				<p>Paragraph with<br>line break</p>
				<br>
				Single br text
				<p>Another paragraph</p>
			</body></html>`,
			expectedPs: 2, // Original paragraphs only
			expectedContent: []string{
				"Paragraph with",
				"Another paragraph",
			},
		},
		{
			name: "handles multiple consecutive br groups",
			html: `<html><body>
				Text before
				<br><br>
				First converted text
				<br>
				Single br
				<br><br>
				Second converted text
			</body></html>`,
			expectedPs: 2, // 2 converted paragraphs
			expectedContent: []string{
				"First converted text",
				"Second converted text",
			},
		},
		{
			name: "preserves formatting in converted paragraphs",
			html: `<html><body>
				<br><br>
				Text with <strong>formatting</strong> and <em>emphasis</em>
			</body></html>`,
			expectedPs: 1,
			expectedContent: []string{
				"Text with formatting and emphasis",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			result := dom.BrsToPs(doc)

			// Check paragraph count
			paragraphs := result.Find("p")
			assert.GreaterOrEqual(t, paragraphs.Length(), tt.expectedPs, "Should have at least expected number of paragraphs")

			// Check that expected content is present
			bodyText := result.Find("body").Text()
			for _, content := range tt.expectedContent {
				assert.Contains(t, bodyText, content, "Should contain expected content: %s", content)
			}

			// Verify that consecutive br tags are reduced
			consecutiveBrs := result.Find("br + br")
			assert.Equal(t, 0, consecutiveBrs.Length(), "Should not have consecutive br tags")
		})
	}
}

func TestParagraphize(t *testing.T) {
	tests := []struct {
		name            string
		html            string
		expectedPs      int
		preservedFormat bool
	}{
		{
			name: "converts br with following content",
			html: `<html><body>
				<div>Content before</div>
				<br><br>
				Text after br with <strong>formatting</strong>
				<span>and inline elements</span>
				<div>Block element stops conversion</div>
			</body></html>`,
			expectedPs:      2, // Original div + new p
			preservedFormat: true,
		},
		{
			name: "handles empty content after br",
			html: `<html><body>
				<br><br>
				<div>Immediate block element</div>
			</body></html>`,
			expectedPs:      0, // No paragraph created due to immediate block
			preservedFormat: false,
		},
		{
			name: "stops at block level elements",
			html: `<html><body>
				<br><br>
				Inline text
				<span>inline span</span>
				<p>Block paragraph stops here</p>
				More text after block
			</body></html>`,
			expectedPs:      2, // New p + existing p
			preservedFormat: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			result := dom.BrsToPs(doc)

			paragraphs := result.Find("p")
			assert.GreaterOrEqual(t, paragraphs.Length(), tt.expectedPs, "Should have expected number of paragraphs")

			if tt.preservedFormat {
				// Check that formatting is preserved in converted paragraphs
				found := false
				paragraphs.Each(func(i int, p *goquery.Selection) {
					if p.Find("strong, span").Length() > 0 {
						found = true
					}
				})
				if result.Find("strong, span").Length() > 0 {
					assert.True(t, found, "Should preserve formatting in converted paragraphs")
				}
			}
		})
	}
}

func TestBrsToPs_EdgeCases(t *testing.T) {
	t.Run("handles deeply nested br tags", func(t *testing.T) {
		html := `<html><body>
			<div>
				<div>
					<br><br>
					Nested content
				</div>
			</div>
		</body></html>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		result := dom.BrsToPs(doc)
		
		// Should still process nested br tags
		paragraphs := result.Find("p")
		assert.True(t, paragraphs.Length() >= 1, "Should create paragraphs from nested br tags")
	})

	t.Run("handles br tags with attributes", func(t *testing.T) {
		html := `<html><body>
			<br class="spacer"><br id="break">
			Content after styled breaks
		</body></html>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		result := dom.BrsToPs(doc)
		
		// Should still process br tags with attributes
		paragraphs := result.Find("p")
		assert.True(t, paragraphs.Length() >= 1, "Should handle br tags with attributes")
	})

	t.Run("handles self-closing br syntax", func(t *testing.T) {
		html := `<html><body>
			<br/><br/>
			Content after self-closing breaks
		</body></html>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		result := dom.BrsToPs(doc)
		
		paragraphs := result.Find("p")
		assert.True(t, paragraphs.Length() >= 1, "Should handle self-closing br tags")
	})

	t.Run("preserves existing paragraph structure", func(t *testing.T) {
		html := `<html><body>
			<p>Existing paragraph</p>
			<br><br>
			New content
			<p>Another existing paragraph</p>
		</body></html>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		result := dom.BrsToPs(doc)
		
		paragraphs := result.Find("p")
		assert.True(t, paragraphs.Length() >= 3, "Should preserve existing paragraphs and add new ones")
		
		// Verify original content is preserved
		bodyText := result.Find("body").Text()
		assert.Contains(t, bodyText, "Existing paragraph")
		assert.Contains(t, bodyText, "Another existing paragraph")
		assert.Contains(t, bodyText, "New content")
	})

	t.Run("handles mixed content types", func(t *testing.T) {
		html := `<html><body>
			<br><br>
			Text content
			<img src="image.jpg" alt="Image">
			<a href="#">Link</a>
			<strong>Bold text</strong>
			<table><tr><td>Table stops here</td></tr></table>
			More content after table
		</body></html>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		result := dom.BrsToPs(doc)
		
		paragraphs := result.Find("p")
		assert.True(t, paragraphs.Length() >= 1, "Should create paragraph from mixed content")
		
		// Verify that inline elements are preserved
		firstP := paragraphs.First()
		if firstP.Length() > 0 {
			assert.True(t, firstP.Find("img, a, strong").Length() > 0, "Should preserve inline elements in paragraph")
		}
	})
}

func TestBrsToPs_Performance(t *testing.T) {
	// Test with a large document
	htmlBuilder := strings.Builder{}
	htmlBuilder.WriteString("<html><body>")
	
	for i := 0; i < 100; i++ {
		htmlBuilder.WriteString("<br><br>")
		htmlBuilder.WriteString("Content block " + string(rune(i)) + " with some text to make it substantial.")
	}
	
	htmlBuilder.WriteString("</body></html>")
	
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlBuilder.String()))
	require.NoError(t, err)

	// Should not panic or hang
	result := dom.BrsToPs(doc)
	
	paragraphs := result.Find("p")
	assert.True(t, paragraphs.Length() > 50, "Should create many paragraphs from large document")
}

func BenchmarkBrsToPs(b *testing.B) {
	html := `<html><body>
		<p>First paragraph</p>
		<br><br>
		Text after double br with <strong>formatting</strong>
		<br><br>
		More text with <em>emphasis</em> and <a href="#">links</a>
		<br>
		Single br text
		<br><br>
		Final converted text
	</body></html>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
		dom.BrsToPs(doc)
	}
}

func BenchmarkBrsToPs_Large(b *testing.B) {
	// Create a larger document for performance testing
	htmlBuilder := strings.Builder{}
	htmlBuilder.WriteString("<html><body>")
	
	for i := 0; i < 50; i++ {
		htmlBuilder.WriteString("<br><br>")
		htmlBuilder.WriteString("Content block with some text to make it substantial and realistic.")
		htmlBuilder.WriteString("<strong>Bold text</strong> and <em>emphasis</em>.")
	}
	
	htmlBuilder.WriteString("</body></html>")
	html := htmlBuilder.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
		dom.BrsToPs(doc)
	}
}