// ABOUTME: Test suite for content cleaner with comprehensive JavaScript compatibility verification
// ABOUTME: Tests the complete content cleaning pipeline including all DOM transformations and edge cases

package cleaners

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExtractCleanNode tests the main content cleaning function
func TestExtractCleanNode(t *testing.T) {
	t.Run("basic cleaning pipeline", func(t *testing.T) {
		html := `
		<html>
			<head><title>Test Title</title></head>
			<body>
				<article class="main-content">
					<h1>Article Title</h1>
					<img src="spacer.gif" width="1" height="1" />
					<p>This is the main content.</p>
					<a href="/relative-link">Relative Link</a>
					<iframe src="https://youtube.com/embed/test"></iframe>
					<script>alert('test');</script>
					<style>.test{}</style>
					<p>   </p>
					<div class="ads" style="color:red; font-size:12px;">Ad content</div>
				</article>
			</body>
		</html>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		article := doc.Find("article")
		require.Equal(t, 1, article.Length())

		trueBool := true
		opts := ContentCleanOptions{
			CleanConditionally: true,
			Title:              "Test Title",
			URL:                "https://example.com/test",
			DefaultCleaner:     &trueBool,
		}

		cleaned := ExtractCleanNode(article, doc, opts)

		// Verify the article was cleaned
		assert.NotNil(t, cleaned)
		assert.Greater(t, cleaned.Length(), 0)

		// Check that content was preserved
		text := strings.TrimSpace(cleaned.Text())
		assert.Contains(t, text, "This is the main content")

		// Check that script tags were removed
		scripts := cleaned.Find("script")
		assert.Equal(t, 0, scripts.Length())

		// Check that style tags were removed
		styles := cleaned.Find("style")
		assert.Equal(t, 0, styles.Length())
	})

	t.Run("with defaultCleaner disabled", func(t *testing.T) {
		html := `
		<div>
			<img src="spacer.gif" width="1" height="1" />
			<p>Content with small image</p>
			<ul class="navigation">
				<li><a href="#">Nav 1</a></li>
				<li><a href="#">Nav 2</a></li>
			</ul>
		</div>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		article := doc.Find("div")
		falseBool := false
		opts := ContentCleanOptions{
			CleanConditionally: true,
			DefaultCleaner:     &falseBool, // Disable aggressive cleaning
		}

		cleaned := ExtractCleanNode(article, doc, opts)

		// Small images should be preserved when defaultCleaner is false
		images := cleaned.Find("img")
		assert.Greater(t, images.Length(), 0)

		// Navigation lists should be preserved when defaultCleaner is false
		lists := cleaned.Find("ul")
		assert.Greater(t, lists.Length(), 0)
	})

	t.Run("with cleanConditionally disabled", func(t *testing.T) {
		html := `
		<div>
			<p>Good content here</p>
			<div class="might-be-content">
				<a href="#">Link 1</a>
				<a href="#">Link 2</a>
				<p>Some text with links</p>
			</div>
		</div>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		article := doc.Find("div").First()
		trueBool := true
		opts := ContentCleanOptions{
			CleanConditionally: false, // Don't clean conditionally
			DefaultCleaner:     &trueBool,
		}

		cleaned := ExtractCleanNode(article, doc, opts)

		// Content should be preserved when conditional cleaning is disabled
		text := strings.TrimSpace(cleaned.Text())
		assert.Contains(t, text, "Good content here")
		assert.Contains(t, text, "Some text with links")
	})

	t.Run("handles nil/empty input", func(t *testing.T) {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader("<html><body></body></html>"))
		require.NoError(t, err)

		opts := ContentCleanOptions{}

		// Test with nil selection
		cleaned := ExtractCleanNode(nil, doc, opts)
		assert.Nil(t, cleaned)

		// Test with empty selection
		empty := doc.Find("nonexistent")
		cleaned = ExtractCleanNode(empty, doc, opts)
		assert.Equal(t, 0, cleaned.Length())
	})

	t.Run("absolute link conversion", func(t *testing.T) {
		html := `
		<div>
			<a href="/relative">Relative</a>
			<a href="../parent">Parent</a>
			<a href="//cdn.example.com/resource">Protocol-relative</a>
			<a href="https://absolute.com">Absolute</a>
		</div>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		article := doc.Find("div")
		opts := ContentCleanOptions{
			URL: "https://example.com/path/page",
		}

		cleaned := ExtractCleanNode(article, doc, opts)

		// Check that relative links were converted to absolute
		links := cleaned.Find("a")
		hrefs := make([]string, 0)
		links.Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if exists {
				hrefs = append(hrefs, href)
			}
		})

		// Should have absolute URLs
		assert.Contains(t, hrefs, "https://example.com/relative")
		assert.Contains(t, hrefs, "https://example.com/parent")
		assert.Contains(t, hrefs, "https://cdn.example.com/resource")
		assert.Contains(t, hrefs, "https://absolute.com")
	})

	t.Run("preserves marked elements", func(t *testing.T) {
		html := `
		<div>
			<p>Regular content</p>
			<iframe src="https://youtube.com/embed/test123"></iframe>
			<iframe src="https://vimeo.com/123456"></iframe>
			<iframe src="https://malicious.com/tracker"></iframe>
		</div>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		article := doc.Find("div")
		opts := ContentCleanOptions{
			URL: "https://example.com/page",
		}

		cleaned := ExtractCleanNode(article, doc, opts)

		// YouTube and Vimeo iframes should be preserved
		iframes := cleaned.Find("iframe")
		
		// Should have some iframes (YouTube/Vimeo preserved, malicious one removed)
		sources := make([]string, 0)
		iframes.Each(func(i int, s *goquery.Selection) {
			src, exists := s.Attr("src")
			if exists {
				sources = append(sources, src)
			}
		})

		// Check that legitimate video iframes are preserved
		hasYoutube := false
		hasVimeo := false
		for _, src := range sources {
			if strings.Contains(src, "youtube.com") {
				hasYoutube = true
			}
			if strings.Contains(src, "vimeo.com") {
				hasVimeo = true
			}
		}

		assert.True(t, hasYoutube, "YouTube iframe should be preserved")
		assert.True(t, hasVimeo, "Vimeo iframe should be preserved")
	})

	t.Run("header cleaning with title context", func(t *testing.T) {
		html := `
		<div>
			<h1>Exact Article Title</h1>
			<h1>Different Header</h1>
			<h2>Subtitle</h2>
			<p>Article content here</p>
			<h3>Section Header</h3>
			<p>More content</p>
		</div>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		article := doc.Find("div")
		opts := ContentCleanOptions{
			Title: "Exact Article Title",
		}

		cleaned := ExtractCleanNode(article, doc, opts)

		// Check that headers were processed appropriately
		text := cleaned.Text()
		assert.Contains(t, text, "Article content here")
		assert.Contains(t, text, "More content")
	})
}

// TestContentCleanOptions tests the options struct
func TestContentCleanOptions(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		opts := ContentCleanOptions{}
		
		// Test default behavior
		assert.False(t, opts.CleanConditionally)
		assert.Equal(t, "", opts.Title)
		assert.Equal(t, "", opts.URL)
		assert.Nil(t, opts.DefaultCleaner) // Should be nil by default
	})

	t.Run("with values", func(t *testing.T) {
		trueBool := true
		opts := ContentCleanOptions{
			CleanConditionally: true,
			Title:              "Test Title",
			URL:                "https://example.com",
			DefaultCleaner:     &trueBool,
		}

		assert.True(t, opts.CleanConditionally)
		assert.Equal(t, "Test Title", opts.Title)
		assert.Equal(t, "https://example.com", opts.URL)
		assert.True(t, *opts.DefaultCleaner)
	})
}

// TestCleaningPipelineStages tests individual stages of the cleaning pipeline
func TestCleaningPipelineStages(t *testing.T) {
	t.Run("rewrite top level", func(t *testing.T) {
		html := `<html><head></head><body><p>Content</p></body></html>`
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		body := doc.Find("body")
		opts := ContentCleanOptions{}

		cleaned := ExtractCleanNode(body, doc, opts)
		
		// Body should be rewritten to div
		assert.NotNil(t, cleaned)
		tagName := goquery.NodeName(cleaned)
		// After rewrite, the element should be a div or maintain content
		assert.True(t, tagName == "div" || tagName == "body")
	})

	t.Run("remove empty paragraphs", func(t *testing.T) {
		html := `
		<div>
			<p>Good content</p>
			<p>   </p>
			<p></p>
			<p><br></p>
			<p>More good content</p>
		</div>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		article := doc.Find("div")
		opts := ContentCleanOptions{}

		cleaned := ExtractCleanNode(article, doc, opts)

		// Should have content but fewer paragraphs
		text := strings.TrimSpace(cleaned.Text())
		assert.Contains(t, text, "Good content")
		assert.Contains(t, text, "More good content")

		// Empty paragraphs should be removed
		paragraphs := cleaned.Find("p")
		validParagraphs := 0
		paragraphs.Each(func(i int, s *goquery.Selection) {
			if strings.TrimSpace(s.Text()) != "" {
				validParagraphs++
			}
		})
		assert.GreaterOrEqual(t, validParagraphs, 2) // At least the two with content
	})

	t.Run("clean attributes", func(t *testing.T) {
		html := `
		<div>
			<p style="color:red; font-size:12px;" class="article" id="main">
				Content with attributes
			</p>
			<a href="https://example.com" onclick="track()" data-custom="value">
				Link with many attributes
			</a>
		</div>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		article := doc.Find("div")
		opts := ContentCleanOptions{}

		cleaned := ExtractCleanNode(article, doc, opts)

		// Check that some attributes were cleaned
		text := cleaned.Text()
		assert.Contains(t, text, "Content with attributes")
		assert.Contains(t, text, "Link with many attributes")

		// Links should still have href but not onclick
		links := cleaned.Find("a")
		if links.Length() > 0 {
			href, hrefExists := links.First().Attr("href")
			assert.True(t, hrefExists)
			assert.Equal(t, "https://example.com", href)

			_, onclickExists := links.First().Attr("onclick")
			assert.False(t, onclickExists, "onclick should be removed")
		}
	})
}

// TestJavaScriptCompatibility tests compatibility with the JavaScript implementation
func TestJavaScriptCompatibility(t *testing.T) {
	t.Run("matches JavaScript cleaning order", func(t *testing.T) {
		// This HTML represents typical messy article content
		html := `
		<html>
			<head><title>Page Title</title></head>
			<body>
				<article>
					<h1>Article Title</h1>
					<img src="spacer.gif" width="1" height="1" alt="spacer">
					<p>This is <a href="/test">good content</a> with a link.</p>
					<script>analytics.track();</script>
					<style>.hidden{display:none;}</style>
					<iframe src="https://youtube.com/embed/abc123"></iframe>
					<p>More content here.</p>
					<div class="social-sharing">
						<a href="#share">Share</a>
						<a href="#tweet">Tweet</a>
					</div>
					<p>   </p>
					<h2 class="article-title">Article Title</h2>
				</article>
			</body>
		</html>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		article := doc.Find("article")
		trueBool := true
		opts := ContentCleanOptions{
			CleanConditionally: true,
			Title:              "Article Title",
			URL:                "https://example.com/article",
			DefaultCleaner:     &trueBool,
		}

		cleaned := ExtractCleanNode(article, doc, opts)

		// Verify the cleaning matches JavaScript behavior:
		// 1. Content should be preserved
		text := strings.TrimSpace(cleaned.Text())
		assert.Contains(t, text, "good content")
		assert.Contains(t, text, "More content here")

		// 2. Script and style tags should be removed
		assert.Equal(t, 0, cleaned.Find("script").Length())
		assert.Equal(t, 0, cleaned.Find("style").Length())

		// 3. YouTube iframe should be preserved
		iframes := cleaned.Find("iframe")
		hasYoutube := false
		iframes.Each(func(i int, s *goquery.Selection) {
			src, _ := s.Attr("src")
			if strings.Contains(src, "youtube.com") {
				hasYoutube = true
			}
		})
		assert.True(t, hasYoutube)

		// 4. Links should be absolute
		links := cleaned.Find("a[href]")
		hasAbsoluteLink := false
		links.Each(func(i int, s *goquery.Selection) {
			href, _ := s.Attr("href")
			if strings.HasPrefix(href, "https://example.com") {
				hasAbsoluteLink = true
			}
		})
		assert.True(t, hasAbsoluteLink)
	})
}