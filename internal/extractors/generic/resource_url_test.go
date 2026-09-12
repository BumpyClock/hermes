package generic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourceURLExtraction(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		pageURL  string
		want     string
	}{
		{"absolute", " https://cdn.example.com/asset ", "://bad", "https://cdn.example.com/asset"},
		{"protocol relative", "//cdn.example.com/asset", "http://example.com/article", "https://cdn.example.com/asset"},
		{"relative", "../asset", "https://example.com/news/article", "https://example.com/asset"},
		{"root relative", "/asset", "https://example.com/news/article", "https://example.com/asset"},
		{"query", "?v=2", "https://example.com/article?v=1", "https://example.com/article?v=2"},
		{"invalid resource", " %zz ", "https://example.com/article", "%zz"},
		{"invalid base", " asset ", "://bad", "asset"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := newDoc(`<link rel="icon"><meta property="og:video"><meta property="og:video:secure_url">`, t)
			doc.Find("link").SetAttr("href", tt.resource)
			doc.Find("meta").SetAttr("content", tt.resource)

			favicon := (&GenericFaviconExtractor{}).Extract(doc.Selection, tt.pageURL, nil)
			assert.Equal(t, tt.want, favicon)
			video := (&GenericVideoExtractor{}).Extract(doc.Selection, tt.pageURL, nil)
			if assert.NotNil(t, video) {
				assert.Equal(t, tt.want, video.URL)
				assert.Equal(t, tt.want, video.SecureURL)
			}
		})
	}
}
