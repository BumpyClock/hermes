package dom_test

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BumpyClock/parser-go/pkg/utils/dom"
)

func TestMakeLinksAbsolute(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		baseURL string
		checks  map[string]string // selector -> expected href/src
	}{
		{
			name: "converts relative links",
			html: `<html><body>
				<a href="/page">Relative link</a>
				<a href="http://example.com/abs">Absolute link</a>
				<img src="/image.jpg">
				<img src="https://example.com/abs.jpg">
			</body></html>`,
			baseURL: "https://example.com/article",
			checks: map[string]string{
				`a[href="/page"]`:                          "",
				`a[href="https://example.com/page"]`:       "https://example.com/page",
				`a[href="http://example.com/abs"]`:         "http://example.com/abs",
				`img[src="/image.jpg"]`:                    "",
				`img[src="https://example.com/image.jpg"]`: "https://example.com/image.jpg",
			},
		},
		{
			name: "handles protocol-relative URLs",
			html: `<html><body>
				<a href="//other.com/page">Protocol-relative</a>
				<img src="//cdn.example.com/image.png">
			</body></html>`,
			baseURL: "https://example.com",
			checks: map[string]string{
				`a[href="https://other.com/page"]`:           "https://other.com/page",
				`img[src="https://cdn.example.com/image.png"]`: "https://cdn.example.com/image.png",
			},
		},
		{
			name: "preserves javascript and mailto links",
			html: `<html><body>
				<a href="javascript:void(0)">JS link</a>
				<a href="mailto:test@example.com">Email link</a>
			</body></html>`,
			baseURL: "https://example.com",
			checks: map[string]string{
				`a[href="javascript:void(0)"]`:    "javascript:void(0)",
				`a[href="mailto:test@example.com"]`: "mailto:test@example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			result := dom.MakeLinksAbsolute(doc, tt.baseURL)

			for selector, expectedValue := range tt.checks {
				elements := result.Find(selector)
				if expectedValue == "" {
					// Should not exist
					assert.Equal(t, 0, elements.Length(), "Element should not exist: %s", selector)
				} else {
					// Should exist with expected value
					assert.True(t, elements.Length() > 0, "Element should exist: %s", selector)
					if elements.Length() > 0 {
						var actualValue string
						if strings.Contains(selector, "href") {
							actualValue, _ = elements.Attr("href")
						} else {
							actualValue, _ = elements.Attr("src")
						}
						assert.Equal(t, expectedValue, actualValue, "Value should match for: %s", selector)
					}
				}
			}
		})
	}
}

func TestArticleBaseURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "https://example.com/article?param=value#section",
			expected: "https://example.com/article",
		},
		{
			input:    "http://example.com/path/to/article.html?utm_source=twitter",
			expected: "http://example.com/path/to/article.html",
		},
		{
			input:    "https://example.com/article#top",
			expected: "https://example.com/article",
		},
		{
			input:    "https://example.com/simple",
			expected: "https://example.com/simple",
		},
		{
			input:    "invalid-url",
			expected: "invalid-url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := dom.ArticleBaseURL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoveAnchor(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "https://example.com/article#section",
			expected: "https://example.com/article",
		},
		{
			input:    "https://example.com/article?param=value#section",
			expected: "https://example.com/article?param=value",
		},
		{
			input:    "https://example.com/article",
			expected: "https://example.com/article",
		},
		{
			input:    "invalid-url",
			expected: "invalid-url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := dom.RemoveAnchor(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		url   string
		valid bool
	}{
		{"https://example.com", true},
		{"http://example.com", true},
		{"https://example.com/path", true},
		{"http://subdomain.example.com/path/to/page", true},
		{"ftp://example.com", false},
		{"javascript:void(0)", false},
		{"mailto:test@example.com", false},
		{"//example.com", false},
		{"/relative/path", false},
		{"", false},
		{"not-a-url", false},
		{"https://", false},
		{"http://", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := dom.ValidateURL(tt.url)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestGetDomain(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://example.com", "example.com"},
		{"https://www.example.com", "example.com"},
		{"http://subdomain.example.com", "subdomain.example.com"},
		{"https://www.subdomain.example.com/path", "subdomain.example.com"},
		{"https://example.com:8080", "example.com:8080"},
		{"invalid-url", ""},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := dom.GetDomain(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetBaseDomain(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://example.com", "example.com"},
		{"https://www.example.com", "example.com"},
		{"https://subdomain.example.com", "example.com"},
		{"https://deep.subdomain.example.com", "example.com"},
		{"https://example.co.uk", "co.uk"},
		{"invalid-url", ""},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := dom.GetBaseDomain(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "removes UTM parameters",
			input:    "https://example.com/article?utm_source=twitter&utm_medium=social&content=main",
			expected: "https://example.com/article?content=main",
		},
		{
			name:     "removes Facebook click ID",
			input:    "https://example.com/article?fbclid=abc123&other=value",
			expected: "https://example.com/article?other=value",
		},
		{
			name:     "removes Google click ID",
			input:    "https://example.com/article?gclid=xyz789&keep=this",
			expected: "https://example.com/article?keep=this",
		},
		{
			name:     "preserves non-tracking parameters",
			input:    "https://example.com/search?q=test&page=2",
			expected: "https://example.com/search?page=2&q=test", // Parameter order may vary
		},
		{
			name:     "handles URLs without parameters",
			input:    "https://example.com/article",
			expected: "https://example.com/article",
		},
		{
			name:     "removes all tracking parameters",
			input:    "https://example.com/article?utm_source=google&utm_medium=cpc&utm_campaign=test&fbclid=abc&gclid=xyz&ref=twitter",
			expected: "https://example.com/article",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dom.SanitizeURL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMakeLinksAbsolute_Srcset(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		baseURL string
		checks  map[string]string // selector -> expected srcset
	}{
		{
			name: "basic srcset with descriptors",
			html: `<html><body>
				<img srcset="/small.jpg 1x, /large.jpg 2x">
			</body></html>`,
			baseURL: "https://example.com",
			checks: map[string]string{
				`img[srcset]`: "https://example.com/small.jpg 1x, https://example.com/large.jpg 2x",
			},
		},
		{
			name: "srcset with width descriptors",
			html: `<html><body>
				<img srcset="/image-400.jpg 400w, /image-800.jpg 800w, /image-1200.jpg 1200w">
			</body></html>`,
			baseURL: "https://example.com",
			checks: map[string]string{
				`img[srcset]`: "https://example.com/image-400.jpg 400w, https://example.com/image-800.jpg 800w, https://example.com/image-1200.jpg 1200w",
			},
		},
		{
			name: "srcset with mixed absolute and relative URLs",
			html: `<html><body>
				<img srcset="https://cdn.example.com/abs.jpg 1x, /rel.jpg 2x">
			</body></html>`,
			baseURL: "https://example.com",
			checks: map[string]string{
				`img[srcset]`: "https://cdn.example.com/abs.jpg 1x, https://example.com/rel.jpg 2x",
			},
		},
		{
			name: "srcset with protocol-relative URLs",
			html: `<html><body>
				<img srcset="//cdn.example.com/image1.jpg 1x, //cdn.example.com/image2.jpg 2x">
			</body></html>`,
			baseURL: "https://example.com",
			checks: map[string]string{
				`img[srcset]`: "https://cdn.example.com/image1.jpg 1x, https://cdn.example.com/image2.jpg 2x",
			},
		},
		{
			name: "srcset with extra spaces and commas",
			html: `<html><body>
				<img srcset="  /image1.jpg  1x  ,  /image2.jpg  2x  ">
			</body></html>`,
			baseURL: "https://example.com",
			checks: map[string]string{
				`img[srcset]`: "https://example.com/image1.jpg 1x, https://example.com/image2.jpg 2x",
			},
		},
		{
			name: "srcset with decimal descriptors",
			html: `<html><body>
				<img srcset="/image1.jpg 1.5x, /image2.jpg 2.75x">
			</body></html>`,
			baseURL: "https://example.com",
			checks: map[string]string{
				`img[srcset]`: "https://example.com/image1.jpg 1.5x, https://example.com/image2.jpg 2.75x",
			},
		},
		{
			name: "srcset removes duplicates",
			html: `<html><body>
				<img srcset="/same.jpg 1x, /same.jpg 1x, /different.jpg 2x">
			</body></html>`,
			baseURL: "https://example.com",
			checks: map[string]string{
				`img[srcset]`: "https://example.com/same.jpg 1x, https://example.com/different.jpg 2x",
			},
		},
		{
			name: "multiple images with srcset",
			html: `<html><body>
				<img srcset="/image1-small.jpg 400w, /image1-large.jpg 800w">
				<img srcset="/image2-1x.jpg 1x, /image2-2x.jpg 2x">
			</body></html>`,
			baseURL: "https://example.com",
			checks: map[string]string{
				`img:first-child`: "https://example.com/image1-small.jpg 400w, https://example.com/image1-large.jpg 800w",
				`img:last-child`:  "https://example.com/image2-1x.jpg 1x, https://example.com/image2-2x.jpg 2x",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			result := dom.MakeLinksAbsolute(doc, tt.baseURL)

			for selector, expectedSrcset := range tt.checks {
				element := result.Find(selector)
				assert.True(t, element.Length() > 0, "Element should exist: %s", selector)
				if element.Length() > 0 {
					actualSrcset, exists := element.Attr("srcset")
					assert.True(t, exists, "srcset attribute should exist for: %s", selector)
					assert.Equal(t, expectedSrcset, actualSrcset, "srcset should match for: %s", selector)
				}
			}
		})
	}
}

func TestMakeLinksAbsolute_SrcsetEdgeCases(t *testing.T) {
	t.Run("handles empty srcset", func(t *testing.T) {
		html := `<html><body>
			<img srcset="">
		</body></html>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		// Should not panic
		result := dom.MakeLinksAbsolute(doc, "https://example.com")
		assert.NotNil(t, result)
	})

	t.Run("handles malformed srcset", func(t *testing.T) {
		html := `<html><body>
			<img srcset="not-a-valid-srcset">
		</body></html>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		// Should not panic and leave srcset unchanged if invalid
		result := dom.MakeLinksAbsolute(doc, "https://example.com")
		assert.NotNil(t, result)
	})

	t.Run("handles srcset without descriptors", func(t *testing.T) {
		html := `<html><body>
			<img srcset="/image1.jpg, /image2.jpg">
		</body></html>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		result := dom.MakeLinksAbsolute(doc, "https://example.com")
		
		// Should still convert relative URLs even without descriptors
		img := result.Find("img")
		srcset, _ := img.Attr("srcset")
		assert.Contains(t, srcset, "https://example.com/image1.jpg")
		assert.Contains(t, srcset, "https://example.com/image2.jpg")
	})

	t.Run("preserves base tag for srcset", func(t *testing.T) {
		html := `<html><head>
			<base href="https://cdn.example.com/">
		</head><body>
			<img srcset="/small.jpg 1x, /large.jpg 2x">
		</body></html>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		result := dom.MakeLinksAbsolute(doc, "https://example.com")
		
		// Should use base tag URL instead of provided URL
		img := result.Find("img")
		srcset, _ := img.Attr("srcset")
		assert.Equal(t, "https://cdn.example.com/small.jpg 1x, https://cdn.example.com/large.jpg 2x", srcset)
	})

	t.Run("handles picture element with source srcset", func(t *testing.T) {
		html := `<html><body>
			<picture>
				<source srcset="/mobile.jpg 1x, /mobile-2x.jpg 2x" media="(max-width: 600px)">
				<source srcset="/desktop.jpg 1x, /desktop-2x.jpg 2x">
				<img src="/fallback.jpg" alt="Image">
			</picture>
		</body></html>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		result := dom.MakeLinksAbsolute(doc, "https://example.com")
		
		// Should convert srcset in source elements
		sources := result.Find("source")
		
		// First source (mobile)
		srcset1, _ := sources.Eq(0).Attr("srcset")
		assert.Equal(t, "https://example.com/mobile.jpg 1x, https://example.com/mobile-2x.jpg 2x", srcset1)
		
		// Second source (desktop)
		srcset2, _ := sources.Eq(1).Attr("srcset")
		assert.Equal(t, "https://example.com/desktop.jpg 1x, https://example.com/desktop-2x.jpg 2x", srcset2)
		
		// Fallback img src
		img := result.Find("img")
		src, _ := img.Attr("src")
		assert.Equal(t, "https://example.com/fallback.jpg", src)
	})
}

func TestMakeLinksAbsolute_EdgeCases(t *testing.T) {
	t.Run("handles empty and missing attributes", func(t *testing.T) {
		html := `<html><body>
			<a>No href</a>
			<a href="">Empty href</a>
			<img>No src</img>
			<img src="">Empty src</img>
		</body></html>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		// Should not panic
		result := dom.MakeLinksAbsolute(doc, "https://example.com")
		assert.NotNil(t, result)
	})

	t.Run("handles malformed base URL", func(t *testing.T) {
		html := `<html><body>
			<a href="/test">Test link</a>
		</body></html>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)

		// Should not panic with malformed base URL
		result := dom.MakeLinksAbsolute(doc, "not-a-url")
		assert.NotNil(t, result)

		// Link should remain unchanged
		link := result.Find("a").First()
		href, _ := link.Attr("href")
		assert.Equal(t, "/test", href)
	})
}

func BenchmarkLinkFunctions(b *testing.B) {
	html := `<html><body>
		<a href="/page1">Link 1</a>
		<a href="/page2">Link 2</a>
		<a href="http://external.com">External</a>
		<img src="/image1.jpg">
		<img src="/image2.png">
		<iframe src="/embed"></iframe>
	</body></html>`

	b.Run("MakeLinksAbsolute", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
			dom.MakeLinksAbsolute(doc, "https://example.com")
		}
	})

	url := "https://example.com/article?utm_source=twitter&param=value#section"

	b.Run("ArticleBaseURL", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			dom.ArticleBaseURL(url)
		}
	})

	b.Run("ValidateURL", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			dom.ValidateURL(url)
		}
	})

	b.Run("SanitizeURL", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			dom.SanitizeURL(url)
		}
	})
}