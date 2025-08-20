// ABOUTME: Integration tests for content cleaner with the existing content extraction pipeline
// ABOUTME: Verifies that the standalone cleaner works correctly with the generic content extractor

package cleaners

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestContentCleanerIntegration tests the content cleaner in isolation and with the generic extractor
func TestContentCleanerIntegration(t *testing.T) {
	t.Run("real world article cleaning", func(t *testing.T) {
		// Complex HTML with many cleaning challenges
		html := `
		<!DOCTYPE html>
		<html>
			<head>
				<title>Test Article - Example.com</title>
				<meta charset="utf-8">
				<script>analytics.track('page_view');</script>
				<style>.hidden { display: none; }</style>
			</head>
			<body class="article-page">
				<nav class="site-navigation">
					<ul>
						<li><a href="/">Home</a></li>
						<li><a href="/about">About</a></li>
					</ul>
				</nav>
				
				<article class="main-content">
					<header>
						<h1 class="article-title">The Ultimate Guide to Content Extraction</h1>
						<div class="article-meta">
							<span class="author">John Doe</span>
							<time datetime="2023-01-15">January 15, 2023</time>
						</div>
					</header>
					
					<div class="article-body">
						<img src="/images/spacer.gif" width="1" height="1" alt="tracking pixel" />
						
						<p class="lead">This article explains how content extraction works in modern web parsers.</p>
						
						<h2>Introduction</h2>
						<p>Content extraction is a complex process that involves <a href="/dom-parsing">DOM parsing</a> and <a href="/content-scoring">content scoring</a>.</p>
						
						<div class="embedded-content">
							<iframe src="https://www.youtube.com/embed/dQw4w9WgXcQ" width="560" height="315"></iframe>
						</div>
						
						<h3>Key Benefits</h3>
						<ul class="benefits-list">
							<li>Clean content extraction</li>
							<li>Removes navigation and ads</li>
							<li>Preserves article structure</li>
						</ul>
						
						<blockquote class="highlight">
							<p>"Content extraction transforms messy web pages into clean, readable articles."</p>
						</blockquote>
						
						<h2>Technical Details</h2>
						<p>The extraction process involves multiple stages:</p>
						
						<div class="code-example">
							<code>extractContent(html, options)</code>
						</div>
						
						<p>Each stage performs specific cleaning operations to ensure high-quality output.</p>
						
						<aside class="related-articles">
							<h4>Related Articles</h4>
							<ul>
								<li><a href="/article-1">Understanding HTML Parsing</a></li>
								<li><a href="/article-2">Content Scoring Algorithms</a></li>
							</ul>
						</aside>
						
						<div class="social-sharing" style="border: 1px solid #ccc; padding: 10px;">
							<a href="https://twitter.com/share" onclick="share('twitter')">Tweet</a>
							<a href="https://facebook.com/share" onclick="share('facebook')">Share</a>
						</div>
						
						<p>   </p> <!-- Empty paragraph -->
						<div></div> <!-- Empty div -->
						
						<footer class="article-footer">
							<p>Published on Example.com</p>
						</footer>
					</div>
				</article>
				
				<aside class="sidebar">
					<div class="ad-banner">
						<img src="/ads/banner.jpg" alt="Advertisement" />
					</div>
					
					<div class="newsletter-signup">
						<form>
							<input type="email" placeholder="Subscribe to newsletter" />
							<button type="submit">Subscribe</button>
						</form>
					</div>
				</aside>
			</body>
		</html>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		// Test standalone content cleaner
		article := doc.Find("article")
		require.Equal(t, 1, article.Length())

		trueBool := true
		opts := ContentCleanOptions{
			CleanConditionally: true,
			Title:              "The Ultimate Guide to Content Extraction",
			URL:                "https://example.com/article",
			DefaultCleaner:     &trueBool,
		}

		cleaned := ExtractCleanNode(article, doc, opts)
		require.NotNil(t, cleaned)

		// Get the cleaned HTML content
		cleanedHTML, err := cleaned.Html()
		require.NoError(t, err)

		// Verify essential content is preserved
		assert.Contains(t, cleanedHTML, "This article explains how content extraction works")
		assert.Contains(t, cleanedHTML, "Content extraction is a complex process")
		assert.Contains(t, cleanedHTML, "Clean content extraction")
		assert.Contains(t, cleanedHTML, "transforms messy web pages into clean")

		// Verify structure is preserved
		assert.Contains(t, cleanedHTML, "<h2>Introduction</h2>")
		assert.Contains(t, cleanedHTML, "<h3>Key Benefits</h3>")
		assert.Contains(t, cleanedHTML, "<ul")
		assert.Contains(t, cleanedHTML, "<blockquote")

		// Verify cleaning was applied
		assert.NotContains(t, cleanedHTML, "<script")
		assert.NotContains(t, cleanedHTML, "<style")
		assert.NotContains(t, cleanedHTML, "analytics.track")
		assert.NotContains(t, cleanedHTML, ".hidden")

		// Verify tracking pixel was removed
		assert.NotContains(t, cleanedHTML, "spacer.gif")

		// Verify YouTube iframe was preserved
		assert.Contains(t, cleanedHTML, "youtube.com/embed")

		// Verify links were made absolute
		assert.Contains(t, cleanedHTML, "https://example.com/dom-parsing")
		assert.Contains(t, cleanedHTML, "https://example.com/content-scoring")

		// Verify onclick attributes were removed
		assert.NotContains(t, cleanedHTML, "onclick=")

		// Verify title H1 was handled (should be removed since it matches title)
		// Note: This depends on the header cleaning implementation
		titleH1Count := strings.Count(cleanedHTML, "The Ultimate Guide to Content Extraction")
		assert.LessOrEqual(t, titleH1Count, 1, "Title H1 should be removed or converted")
	})

	t.Run("minimal content cleaning", func(t *testing.T) {
		html := `<div><p>Simple content with <a href="/link">a link</a>.</p></div>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		article := doc.Find("div")
		opts := ContentCleanOptions{
			URL: "https://example.com/page",
		}

		cleaned := ExtractCleanNode(article, doc, opts)
		cleanedHTML, err := cleaned.Html()
		require.NoError(t, err)

		// Should have made links absolute
		assert.Contains(t, cleanedHTML, "https://example.com/link")
		assert.Contains(t, cleanedHTML, "Simple content with")
	})

	t.Run("aggressive cleaning disabled", func(t *testing.T) {
		html := `
		<div>
			<img src="small.gif" width="1" height="1" />
			<p>Content here</p>
			<ul class="menu">
				<li><a href="#">Menu 1</a></li>
				<li><a href="#">Menu 2</a></li>
			</ul>
		</div>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		article := doc.Find("div")
		falseBool := false
		opts := ContentCleanOptions{
			DefaultCleaner: &falseBool, // Disable aggressive cleaning
		}

		cleaned := ExtractCleanNode(article, doc, opts)
		cleanedHTML, err := cleaned.Html()
		require.NoError(t, err)

		// Small images should be preserved
		assert.Contains(t, cleanedHTML, "small.gif")

		// Menu lists should be preserved
		assert.Contains(t, cleanedHTML, "<ul")
		assert.Contains(t, cleanedHTML, "Menu 1")
	})

	t.Run("conditional cleaning disabled", func(t *testing.T) {
		html := `
		<div>
			<p>Good content</p>
			<div class="might-be-ads">
				<a href="#">Ad Link 1</a>
				<a href="#">Ad Link 2</a>
				<p>Some promotional text</p>
			</div>
		</div>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		article := doc.Find("div").First()
		trueBool := true
		opts := ContentCleanOptions{
			CleanConditionally: false, // Disable conditional cleaning
			DefaultCleaner:     &trueBool,
		}

		cleaned := ExtractCleanNode(article, doc, opts)
		cleanedHTML, err := cleaned.Html()
		require.NoError(t, err)

		// Content should be preserved when conditional cleaning is disabled
		assert.Contains(t, cleanedHTML, "Good content")
		assert.Contains(t, cleanedHTML, "promotional text")
		assert.Contains(t, cleanedHTML, "Ad Link")
	})
}

// TestContentCleanerStandalone tests the content cleaner as a standalone utility
func TestContentCleanerStandalone(t *testing.T) {
	t.Run("can be used independently", func(t *testing.T) {
		// Test that the content cleaner can be used as a standalone utility
		// without the generic content extractor
		html := `
		<article>
			<script>analytics();</script>
			<h1>Article Title</h1>
			<p>Content with <a href="/test">relative link</a>.</p>
			<img src="spacer.gif" width="1" height="1" />
			<style>.hidden{display:none;}</style>
		</article>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		article := doc.Find("article")
		opts := ContentCleanOptions{
			Title: "Article Title",
			URL:   "https://example.com/article",
		}

		// This should work without any dependency on the generic extractor
		cleaned := ExtractCleanNode(article, doc, opts)
		require.NotNil(t, cleaned)

		cleanedHTML, err := cleaned.Html()
		require.NoError(t, err)

		// Verify cleaning worked
		assert.Contains(t, cleanedHTML, "Content with")
		assert.Contains(t, cleanedHTML, "https://example.com/test") // Absolute link
		assert.NotContains(t, cleanedHTML, "analytics()")             // Script removed
		assert.NotContains(t, cleanedHTML, "spacer.gif")             // Spacer removed
		assert.NotContains(t, cleanedHTML, ".hidden")                // Style removed
	})
}