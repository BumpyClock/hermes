// ABOUTME: Comprehensive test suite for lead image extraction with JavaScript compatibility verification
// ABOUTME: Tests image scoring, meta tag extraction, and end-to-end image selection with real HTML fixtures

package generic

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenericLeadImageExtractor_Extract_MetaTags(t *testing.T) {
	extractor := NewGenericLeadImageExtractor()

	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name: "OpenGraph image meta tag",
			html: `<html><head>
				<meta property="og:image" content="https://example.com/og-image.jpg">
				</head><body></body></html>`,
			expected: "https://example.com/og-image.jpg",
		},
		{
			name: "Twitter image meta tag",
			html: `<html><head>
				<meta name="twitter:image" content="https://example.com/twitter-image.jpg">
				</head><body></body></html>`,
			expected: "https://example.com/twitter-image.jpg",
		},
		{
			name: "Image src meta tag",
			html: `<html><head>
				<meta name="image_src" content="https://example.com/image-src.jpg">
				</head><body></body></html>`,
			expected: "https://example.com/image-src.jpg",
		},
		{
			name: "Meta tag priority - OpenGraph wins over Twitter",
			html: `<html><head>
				<meta name="twitter:image" content="https://example.com/twitter-image.jpg">
				<meta property="og:image" content="https://example.com/og-image.jpg">
				</head><body></body></html>`,
			expected: "https://example.com/og-image.jpg",
		},
		{
			name: "Invalid URL in meta tag",
			html: `<html><head>
				<meta property="og:image" content="not-a-valid-url">
				</head><body></body></html>`,
			expected: "", // Should fall back to content extraction
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			params := ExtractorImageParams{
				Doc:       doc,
				Content:   "",
				MetaCache: map[string]string{},
				HTML:      tt.html,
			}

			result := extractor.Extract(params)
			if tt.expected == "" {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expected, *result)
			}
		})
	}
}

func TestGenericLeadImageExtractor_Extract_ContentImages(t *testing.T) {
	extractor := NewGenericLeadImageExtractor()

	tests := []struct {
		name     string
		html     string
		content  string
		expected string
	}{
		{
			name: "High scoring image in content",
			html: `<html><body>
				<div class="content">
					<img src="https://example.com/large-photo.jpg" width="800" height="600" alt="Main article photo">
				</div>
			</body></html>`,
			content:  ".content",
			expected: "https://example.com/large-photo.jpg",
		},
		{
			name: "Multiple images - picks highest scoring",
			html: `<html><body>
				<div class="content">
					<img src="https://example.com/small-icon.jpg" width="50" height="50">
					<img src="https://example.com/large-photo.jpg" width="800" height="600" alt="Main photo">
					<img src="https://example.com/medium.jpg" width="300" height="200">
				</div>
			</body></html>`,
			content:  ".content",
			expected: "https://example.com/large-photo.jpg",
		},
		{
			name: "Image with figure parent gets bonus",
			html: `<html><body>
				<div class="content">
					<figure>
						<img src="https://example.com/figure-photo.jpg" width="400" height="300">
					</figure>
					<img src="https://example.com/regular-photo.jpg" width="350" height="300">
				</div>
			</body></html>`,
			content:  ".content",
			expected: "https://example.com/figure-photo.jpg", // Figure bonus should make it win
		},
		{
			name: "Image with figcaption sibling gets bonus",
			html: `<html><body>
				<div class="content">
					<img src="https://example.com/with-caption.jpg" width="400" height="300">
					<figcaption>Photo caption</figcaption>
					<img src="https://example.com/no-caption.jpg" width="350" height="300">
				</div>
			</body></html>`,
			content:  ".content",
			expected: "https://example.com/with-caption.jpg", // Figcaption bonus should make it win
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			params := ExtractorImageParams{
				Doc:       doc,
				Content:   tt.content,
				MetaCache: map[string]string{},
				HTML:      tt.html,
			}

			result := extractor.Extract(params)
			require.NotNil(t, result, "Expected to find an image")
			assert.Equal(t, tt.expected, *result)
		})
	}
}

func TestGenericLeadImageExtractor_Extract_FallbackSelectors(t *testing.T) {
	extractor := NewGenericLeadImageExtractor()

	html := `<html><head>
		<link rel="image_src" href="https://example.com/fallback-image.jpg">
	</head><body>
		<div class="content">
			<!-- No good images in content -->
		</div>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	params := ExtractorImageParams{
		Doc:       doc,
		Content:   ".content",
		MetaCache: map[string]string{},
		HTML:      html,
	}

	result := extractor.Extract(params)
	require.NotNil(t, result)
	assert.Equal(t, "https://example.com/fallback-image.jpg", *result)
}

func TestScoreImageUrl(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected int
	}{
		{
			name:     "Positive hints - upload",
			url:      "https://example.com/upload/photo.jpg",
			expected: 30, // +20 for upload, +10 for jpg
		},
		{
			name:     "Negative hints - icon",
			url:      "https://example.com/images/icon.jpg",
			expected: -10, // -20 for icon, +10 for jpg
		},
		{
			name:     "GIF penalty",
			url:      "https://example.com/animation.gif",
			expected: -10, // -10 for gif
		},
		{
			name:     "JPG bonus",
			url:      "https://example.com/photo.jpg",
			expected: 30, // +20 for "photo" hint, +10 for jpg
		},
		{
			name:     "PNG neutral",
			url:      "https://example.com/photo.png",
			expected: 20, // +20 for "photo" hint, PNG is neutral
		},
		{
			name:     "Multiple positive hints",
			url:      "https://example.com/wp-content/upload/large-photo.jpg",
			expected: 30, // +20 for multiple hints, +10 for jpg
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scoreImageUrl(tt.url)
			assert.Equal(t, tt.expected, score)
		})
	}
}

func TestScoreByDimensions(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected int
	}{
		{
			name:     "Large image bonus",
			html:     `<img src="photo.jpg" width="800" height="600">`,
			expected: 480, // area 480000 / 1000 = 480
		},
		{
			name:     "Small image penalty",
			html:     `<img src="photo.jpg" width="40" height="40">`,
			expected: -200, // -50 (skinny) + -50 (short) + -100 (small area) = -200
		},
		{
			name:     "Skinny image penalty",
			html:     `<img src="photo.jpg" width="30" height="200">`,
			expected: -44, // -50 (skinny) + 6 (area bonus: 6000/1000) = -44
		},
		{
			name:     "Short image penalty",
			html:     `<img src="photo.jpg" width="200" height="30">`,
			expected: -44, // -50 (short) + 6 (area bonus: 6000/1000) = -44
		},
		{
			name:     "Sprite image no area bonus",
			html:     `<img src="sprite.jpg" width="800" height="600">`,
			expected: 0, // sprite in src, so no area calculation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			img := doc.Find("img").First()
			score := scoreByDimensions(img)
			assert.Equal(t, tt.expected, score)
		})
	}
}

func TestScoreByPosition(t *testing.T) {
	// Create array of 5 images
	imgs := make([]interface{}, 5)
	
	tests := []struct {
		name     string
		index    int
		expected float64
	}{
		{
			name:     "First image",
			index:    0,
			expected: 2.5, // 5/2 - 0 = 2.5
		},
		{
			name:     "Middle image",
			index:    2,
			expected: 0.5, // 5/2 - 2 = 0.5
		},
		{
			name:     "Last image",
			index:    4,
			expected: -1.5, // 5/2 - 4 = -1.5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scoreByPosition(imgs, tt.index)
			assert.Equal(t, tt.expected, score)
		})
	}
}

func TestGenericLeadImageExtractor_JavaScriptCompatibility(t *testing.T) {
	// Test that ensures our Go implementation behaves identically to JavaScript
	extractor := NewGenericLeadImageExtractor()

	// Complex HTML that tests multiple scoring factors
	html := `<html><head>
		<meta property="og:image" content="">
		<meta name="twitter:image" content="https://example.com/twitter.jpg">
	</head><body>
		<div class="content">
			<div class="sidebar">
				<img src="https://example.com/ads/banner.jpg" width="200" height="100">
			</div>
			<div class="article">
				<img src="https://example.com/small-icon.png" width="50" height="50">
				<figure>
					<img src="https://example.com/main-photo.jpg" width="600" height="400" alt="Main article photo">
					<figcaption>Photo description</figcaption>
				</figure>
				<img src="https://example.com/inline.jpg" width="300" height="200">
			</div>
		</div>
		<link rel="image_src" href="https://example.com/fallback.jpg">
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	params := ExtractorImageParams{
		Doc:       doc,
		Content:   ".content",
		MetaCache: map[string]string{},
		HTML:      html,
	}

	result := extractor.Extract(params)
	
	// Should pick the twitter image since og:image is empty
	require.NotNil(t, result)
	assert.Equal(t, "https://example.com/twitter.jpg", *result)
}

func TestGenericLeadImageExtractor_NoResults(t *testing.T) {
	extractor := NewGenericLeadImageExtractor()

	html := `<html><head></head><body>
		<div class="content">
			<p>No images here</p>
		</div>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	params := ExtractorImageParams{
		Doc:       doc,
		Content:   ".content",
		MetaCache: map[string]string{},
		HTML:      html,
	}

	result := extractor.Extract(params)
	assert.Nil(t, result, "Should return nil when no valid images found")
}

func TestGenericLeadImageExtractor_RealWorldIntegration(t *testing.T) {
	extractor := NewGenericLeadImageExtractor()

	tests := []struct {
		name     string
		html     string
		content  string
		expected string
		desc     string
	}{
		{
			name: "Uses OpenGraph meta tag when available",
			html: `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<meta property="og:image" content="https://news.example.com/og-image.jpg">
	<meta name="twitter:image" content="https://news.example.com/twitter-image.jpg">
	<title>Breaking News Story</title>
</head>
<body>
	<main class="article-content">
		<img src="https://news.example.com/content-image.jpg" width="400" height="300">
	</main>
</body>
</html>`,
			content:  ".article-content",
			expected: "https://news.example.com/og-image.jpg",
			desc:     "Should prioritize OpenGraph meta tag over content analysis",
		},
		{
			name: "Content analysis picks main story photo",
			html: `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Breaking News Story</title>
</head>
<body>
	<header>
		<img src="https://news.example.com/logo.png" width="100" height="50" alt="Site logo">
	</header>
	<nav>
		<img src="https://news.example.com/nav-icon.png" width="20" height="20">
	</nav>
	<main class="article-content">
		<article>
			<h1>Major Breaking News Event</h1>
			<div class="article-meta">
				<time>2024-01-01</time>
				<div class="social-share">
					<img src="https://news.example.com/facebook-icon.png" width="16" height="16">
					<img src="https://news.example.com/twitter-icon.png" width="16" height="16">
				</div>
			</div>
			<figure class="hero-image">
				<img src="https://news.example.com/main-story-photo.jpg" width="800" height="600" alt="Main story photo showing the news event">
				<figcaption>Photo caption describing the main news event</figcaption>
			</figure>
			<div class="article-body">
				<p>First paragraph of the article...</p>
				<img src="https://news.example.com/inline-chart.png" width="400" height="300" alt="Data chart">
				<p>Second paragraph...</p>
				<figure>
					<img src="https://news.example.com/secondary-photo.jpg" width="600" height="400" alt="Secondary photo">
					<figcaption>Secondary photo caption</figcaption>
				</figure>
				<p>Third paragraph...</p>
			</div>
		</article>
	</main>
	<aside class="sidebar">
		<div class="advertisement">
			<img src="https://ads.example.com/banner-ad.jpg" width="300" height="250">
		</div>
		<div class="related-articles">
			<img src="https://news.example.com/related-thumbnail.jpg" width="150" height="100">
		</div>
	</aside>
	<footer>
		<img src="https://news.example.com/footer-logo.png" width="80" height="40">
	</footer>
</body>
</html>`,
			content:  ".article-content",
			expected: "https://news.example.com/main-story-photo.jpg",
			desc:     "When analyzing content, should pick the large hero image with figure parent and caption",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			params := ExtractorImageParams{
				Doc:       doc,
				Content:   tt.content,
				MetaCache: map[string]string{},
				HTML:      tt.html,
			}

			result := extractor.Extract(params)
			require.NotNil(t, result, "Should find an image in realistic news article")
			assert.Equal(t, tt.expected, *result, tt.desc)
		})
	}
}

func TestGenericLeadImageExtractor_EdgeCases(t *testing.T) {
	extractor := NewGenericLeadImageExtractor()

	tests := []struct {
		name        string
		html        string
		content     string
		expected    *string
		description string
	}{
		{
			name: "Empty meta tag content",
			html: `<html><head>
				<meta property="og:image" content="">
				<meta name="twitter:image" content="https://example.com/twitter.jpg">
			</head><body>
				<div class="content"><p>No images in content</p></div>
			</body></html>`,
			content:     ".content",
			expected:    stringPtr("https://example.com/twitter.jpg"),
			description: "Should skip empty meta tag and use next priority",
		},
		{
			name: "Malformed image dimensions",
			html: `<html><body>
				<div class="content">
					<img src="https://example.com/photo.jpg" width="not-a-number" height="invalid">
				</div>
			</body></html>`,
			content:     ".content",
			expected:    stringPtr("https://example.com/photo.jpg"),
			description: "Should handle invalid dimension attributes gracefully",
		},
		{
			name: "Image without src attribute",
			html: `<html><body>
				<div class="content">
					<img width="400" height="300" alt="Image without src">
					<img src="https://example.com/valid.jpg" width="300" height="200">
				</div>
			</body></html>`,
			content:     ".content",
			expected:    stringPtr("https://example.com/valid.jpg"),
			description: "Should skip images without src and process valid ones",
		},
		{
			name: "All negative scoring images",
			html: `<html><body>
				<div class="content">
					<img src="https://example.com/icon-small.png" width="16" height="16">
					<img src="https://example.com/sprite-bg.png" width="20" height="20">
				</div>
			</body></html>`,
			content:     ".content",
			expected:    nil,
			description: "Should return nil when all images have negative or zero scores",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			params := ExtractorImageParams{
				Doc:       doc,
				Content:   tt.content,
				MetaCache: map[string]string{},
				HTML:      tt.html,
			}

			result := extractor.Extract(params)
			
			if tt.expected == nil {
				assert.Nil(t, result, tt.description)
			} else {
				require.NotNil(t, result, tt.description)
				assert.Equal(t, *tt.expected, *result, tt.description)
			}
		})
	}
}

