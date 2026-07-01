package generic

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestGenericVideoExtractor(t *testing.T) {
	extractor := &GenericVideoExtractor{}

	tests := []struct {
		name     string
		html     string
		expected *VideoMetadata
	}{
		{
			name: "Complete Open Graph video metadata",
			html: `
				<meta property="og:video" content="https://example.com/video.mp4">
				<meta property="og:video:secure_url" content="https://example.com/video.mp4">
				<meta property="og:video:type" content="video/mp4">
				<meta property="og:video:width" content="1280">
				<meta property="og:video:height" content="720">
				<meta property="og:video:duration" content="120">
			`,
			expected: &VideoMetadata{
				URL:       "https://example.com/video.mp4",
				SecureURL: "https://example.com/video.mp4",
				Type:      "video/mp4",
				Width:     1280,
				Height:    720,
				Duration:  120,
			},
		},
		{
			name: "Basic video URL only",
			html: `<meta property="og:video" content="https://example.com/video.mp4">`,
			expected: &VideoMetadata{
				URL: "https://example.com/video.mp4",
			},
		},
		{
			name: "Video with dimensions only",
			html: `
				<meta property="og:video" content="https://example.com/video.mp4">
				<meta property="og:video:width" content="640">
				<meta property="og:video:height" content="480">
			`,
			expected: &VideoMetadata{
				URL:    "https://example.com/video.mp4",
				Width:  640,
				Height: 480,
			},
		},
		{
			name: "Twitter player tags fallback",
			html: `
				<meta name="twitter:player" content="https://player.vimeo.com/video/123456789">
				<meta name="twitter:player:width" content="1920">
				<meta name="twitter:player:height" content="1080">
			`,
			expected: &VideoMetadata{
				URL:    "https://player.vimeo.com/video/123456789",
				Width:  1920,
				Height: 1080,
			},
		},
		{
			name: "Open Graph takes priority over Twitter",
			html: `
				<meta property="og:video" content="https://example.com/og-video.mp4">
				<meta name="twitter:player" content="https://example.com/twitter-video.mp4">
			`,
			expected: &VideoMetadata{
				URL: "https://example.com/og-video.mp4",
			},
		},
		{
			name: "Invalid dimensions ignored",
			html: `
				<meta property="og:video" content="https://example.com/video.mp4">
				<meta property="og:video:width" content="invalid">
				<meta property="og:video:height" content="-100">
				<meta property="og:video:duration" content="abc">
			`,
			expected: &VideoMetadata{
				URL: "https://example.com/video.mp4",
			},
		},
		{
			name: "Protocol-relative URL normalization",
			html: `<meta property="og:video" content="//example.com/video.mp4">`,
			expected: &VideoMetadata{
				URL: "https://example.com/video.mp4",
			},
		},
		{
			name:     "Empty content ignored",
			html:     `<meta property="og:video" content="">`,
			expected: nil,
		},
		{
			name:     "No video metadata",
			html:     `<meta name="description" content="A test page">`,
			expected: nil,
		},
		{
			name: "Zero dimensions ignored",
			html: `
				<meta property="og:video" content="https://example.com/video.mp4">
				<meta property="og:video:width" content="0">
				<meta property="og:video:height" content="0">
			`,
			expected: &VideoMetadata{
				URL: "https://example.com/video.mp4",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			result := extractor.Extract(doc.Selection, "https://example.com", []string{})

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Fatalf("Expected %+v, got nil", tt.expected)
			}

			if result.URL != tt.expected.URL {
				t.Errorf("URL: expected '%s', got '%s'", tt.expected.URL, result.URL)
			}
			if result.SecureURL != tt.expected.SecureURL {
				t.Errorf("SecureURL: expected '%s', got '%s'", tt.expected.SecureURL, result.SecureURL)
			}
			if result.Type != tt.expected.Type {
				t.Errorf("Type: expected '%s', got '%s'", tt.expected.Type, result.Type)
			}
			if result.Width != tt.expected.Width {
				t.Errorf("Width: expected %d, got %d", tt.expected.Width, result.Width)
			}
			if result.Height != tt.expected.Height {
				t.Errorf("Height: expected %d, got %d", tt.expected.Height, result.Height)
			}
			if result.Duration != tt.expected.Duration {
				t.Errorf("Duration: expected %d, got %d", tt.expected.Duration, result.Duration)
			}
		})
	}
}

func TestExtractVideoURL(t *testing.T) {
	extractor := &GenericVideoExtractor{}

	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name: "Prefers secure URL",
			html: `
				<meta property="og:video" content="http://example.com/video.mp4">
				<meta property="og:video:secure_url" content="https://example.com/video.mp4">
			`,
			expected: "https://example.com/video.mp4",
		},
		{
			name:     "Falls back to regular URL",
			html:     `<meta property="og:video" content="https://example.com/video.mp4">`,
			expected: "https://example.com/video.mp4",
		},
		{
			name:     "Returns empty for no video",
			html:     `<meta name="description" content="No video here">`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			result := extractor.ExtractVideoURL(doc.Selection, "https://example.com", []string{})

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestVideoExtractorMetaLookupSupportsNormalizedAndRawMeta(t *testing.T) {
	html := `
		<meta property="og:video" content="https://example.com/video.mp4">
		<meta name="twitter:player" content="https://twitter.com/player">
		<meta property="og:video:width" content="  1280  ">
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	extractor := &GenericVideoExtractor{}
	video := extractor.Extract(doc.Selection, "https://example.com/article", nil)
	if video == nil {
		t.Fatal("expected video metadata")
	}
	if video.URL != "https://example.com/video.mp4" {
		t.Errorf("video URL = %q", video.URL)
	}
	if video.Width != 1280 {
		t.Errorf("video width = %d", video.Width)
	}
}
