package dom_test

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BumpyClock/parser-go/pkg/utils/dom"
)

func TestStripUnlikelyCandidates(t *testing.T) {
	tests := []struct {
		name           string
		html           string
		expectedRemain []string
		expectedGone   []string
	}{
		{
			name: "removes sidebar elements",
			html: `<html><body>
				<div class="content">Main content</div>
				<div class="sidebar">Sidebar content</div>
				<div id="nav-menu">Navigation</div>
			</body></html>`,
			expectedRemain: []string{".content"},
			expectedGone:   []string{".sidebar", "#nav-menu"},
		},
		{
			name: "preserves content elements",
			html: `<html><body>
				<div class="article-content">Article text</div>
				<div class="entry-content">Entry text</div>
				<div class="advert">Advertisement</div>
			</body></html>`,
			expectedRemain: []string{".article-content", ".entry-content"},
			expectedGone:   []string{".advert"},
		},
		{
			name: "preserves links regardless of class",
			html: `<html><body>
				<a href="#" class="sidebar">Link in sidebar class</a>
				<div class="sidebar">Div with sidebar class</div>
			</body></html>`,
			expectedRemain: []string{"a"},
			expectedGone:   []string{"div.sidebar"},
		},
		{
			name: "whitelist overrides blacklist",
			html: `<html><body>
				<div class="sidebar content">Mixed classes</div>
				<div class="sidebar">Pure sidebar</div>
			</body></html>`,
			expectedRemain: []string{".sidebar.content"},
			expectedGone:   []string{".sidebar:not(.content)"},
		},
		{
			name: "handles elements without class or id",
			html: `<html><body>
				<div>No class or id</div>
				<p class="sidebar">Paragraph with sidebar class</p>
			</body></html>`,
			expectedRemain: []string{"div"},
			expectedGone:   []string{".sidebar"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			// Apply the strip function
			result := dom.StripUnlikelyCandidates(doc)

			// Check that expected elements remain
			for _, selector := range tt.expectedRemain {
				elements := result.Find(selector)
				assert.True(t, elements.Length() > 0, "Expected element should remain: %s", selector)
			}

			// Check that expected elements are gone
			for _, selector := range tt.expectedGone {
				elements := result.Find(selector)
				assert.Equal(t, 0, elements.Length(), "Expected element should be removed: %s", selector)
			}
		})
	}
}

func TestStripUnlikelyCandidates_SpecificCases(t *testing.T) {
	t.Run("preserves important content patterns", func(t *testing.T) {
		html := `<html><body>
			<div class="article">Article content</div>
			<div class="post-content">Post content</div>
			<div class="entry">Entry content</div>
			<div class="main">Main content</div>
		</body></html>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		result := dom.StripUnlikelyCandidates(doc)

		// All should remain
		assert.Equal(t, 1, result.Find(".article").Length())
		assert.Equal(t, 1, result.Find(".post-content").Length())
		assert.Equal(t, 1, result.Find(".entry").Length())
		assert.Equal(t, 1, result.Find(".main").Length())
	})

	t.Run("removes common junk patterns", func(t *testing.T) {
		html := `<html><body>
			<div class="footer">Footer</div>
			<div class="navigation">Nav</div>
			<div class="comment">Comment</div>
			<div class="advertisement">Ad</div>
			<div id="sidebar">Sidebar</div>
		</body></html>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		result := dom.StripUnlikelyCandidates(doc)

		// All should be removed
		assert.Equal(t, 0, result.Find(".footer").Length())
		assert.Equal(t, 0, result.Find(".navigation").Length())
		assert.Equal(t, 0, result.Find(".comment").Length())
		assert.Equal(t, 0, result.Find(".advertisement").Length())
		assert.Equal(t, 0, result.Find("#sidebar").Length())
	})

	t.Run("handles mixed case", func(t *testing.T) {
		html := `<html><body>
			<div class="SIDEBAR">Uppercase sidebar</div>
			<div class="Article">Mixed case article</div>
		</body></html>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		result := dom.StripUnlikelyCandidates(doc)

		// Regex should be case-insensitive
		assert.Equal(t, 0, result.Find(".SIDEBAR").Length(), "Should remove uppercase SIDEBAR")
		assert.Equal(t, 1, result.Find(".Article").Length(), "Should keep mixed case Article")
	})
}

func BenchmarkStripUnlikelyCandidates(b *testing.B) {
	html := `<html><body>
		<div class="header">Header</div>
		<div class="navigation">Nav</div>
		<div class="content">Main content here</div>
		<div class="sidebar">Sidebar content</div>
		<div class="footer">Footer</div>
		<div class="advertisement">Ad</div>
		<div class="comment">Comment</div>
		<div class="article">Article content</div>
	</body></html>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a fresh copy for each iteration
		freshDoc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
		dom.StripUnlikelyCandidates(freshDoc)
	}
}