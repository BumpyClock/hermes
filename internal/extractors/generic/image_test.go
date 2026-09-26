package generic

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestGenericLeadImageExtractorBreaksContentScoreTiesByFirstSeenSrc(t *testing.T) {
	// With four images, position scores are 2, 1, 0, and -1. Each extra 10px of
	// height adds one dimension point, so every sized image below scores 12.
	tests := []struct {
		name     string
		images   string
		expected string
	}{
		{
			name: "first tied src wins",
			images: `<img src="https://example.com/first.png" width="100" height="100">
				<img src="https://example.com/second.png" width="100" height="110">
				<img src="https://example.com/third.png" width="100" height="120">
				<img src="https://example.com/fourth.png" width="100" height="130">`,
			expected: "https://example.com/first.png",
		},
		{
			name: "repeated src keeps first position and last score",
			images: `<img src="https://example.com/repeated.png">
				<img src="https://example.com/second.png" width="100" height="110">
				<img src="https://example.com/third.png" width="100" height="120">
				<img src="https://example.com/repeated.png" width="100" height="130">`,
			expected: "https://example.com/repeated.png",
		},
	}

	extractor := NewGenericLeadImageExtractor()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := "<article>" + tt.images + "</article>"
			html := "<html><head><title>Ties</title></head><body>" + content + "</body></html>"

			// Map iteration order varies, so repeat to catch nondeterministic winners.
			for run := range 25 {
				doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
				if err != nil {
					t.Fatalf("failed to parse HTML: %v", err)
				}

				got := extractor.Extract(ExtractorImageParams{Doc: doc, Content: content})
				if got == nil {
					t.Fatalf("run %d: expected %q, got nil", run, tt.expected)
				}
				if *got != tt.expected {
					t.Fatalf("run %d: expected %q, got %q", run, tt.expected, *got)
				}
			}
		})
	}
}

func TestGenericLeadImageExtractorScoresOnlyContentImages(t *testing.T) {
	// The page image outscores the article image, but Mercury scores only
	// images inside the extracted article HTML.
	page := `<html><head><title>Scope</title><link rel="image_src" href="https://example.com/link.png"></head><body>
		<div class="topper"><img src="https://example.com/uploads/page.jpg" width="1200" height="800"></div>
		%s
		</body></html>`
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "article image wins over page image",
			content:  `<div><p>Body</p><img src="https://example.com/article.png" width="400" height="300"></div>`,
			expected: "https://example.com/article.png",
		},
		{
			name:     "article without images falls back to selectors",
			content:  `<div><p>Body</p></div>`,
			expected: "https://example.com/link.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(fmt.Sprintf(page, tt.content)))
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}

			got := NewGenericLeadImageExtractor().Extract(ExtractorImageParams{Doc: doc, Content: tt.content})
			if got == nil || *got != tt.expected {
				t.Fatalf("expected %q, got %v", tt.expected, deref(got))
			}
		})
	}
}

func TestGenericLeadImageExtractorKeepsHalfPointPositionScores(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			// Position scores are 1.5, 0.5, and -0.5. Dimensions add 10 and 11 to
			// the last two images, so both score 10.5 and the first one wins.
			name: "half points tie across the zero crossing",
			content: `<div>
				<img src="https://example.com/icon.png">
				<img src="https://example.com/second.png" width="100" height="100">
				<img src="https://example.com/third.png" width="100" height="110">
				</div>`,
			expected: "https://example.com/second.png",
		},
		{
			// A lone unhinted image scores only its 0.5 position point.
			name:     "half point alone is a positive score",
			content:  `<div><img src="https://example.com/only.png"></div>`,
			expected: "https://example.com/only.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := "<html><head><title>Position</title></head><body>" + tt.content + "</body></html>"
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}

			got := NewGenericLeadImageExtractor().Extract(ExtractorImageParams{Doc: doc, Content: tt.content})
			if got == nil || *got != tt.expected {
				t.Fatalf("expected %q, got %v", tt.expected, deref(got))
			}
		})
	}
}

func deref(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

func TestGenericLeadImageExtractorScoresOnlyNonEmptyAlt(t *testing.T) {
	// Position scores are 1 and 0. Mercury adds 5 only for a truthy alt value.
	tests := []struct {
		name     string
		alt      string
		expected string
	}{
		{name: "empty alt earns no bonus", alt: `alt=""`, expected: "https://example.com/first.png"},
		{name: "non-empty alt earns the bonus", alt: `alt="A photo"`, expected: "https://example.com/second.png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := `<div>
				<img src="https://example.com/first.png" width="100" height="100">
				<img src="https://example.com/second.png" width="100" height="100" ` + tt.alt + `>
				</div>`
			html := "<html><head><title>Alt</title></head><body>" + content + "</body></html>"
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}

			got := NewGenericLeadImageExtractor().Extract(ExtractorImageParams{Doc: doc, Content: content})
			if got == nil || *got != tt.expected {
				t.Fatalf("expected %q, got %v", tt.expected, deref(got))
			}
		})
	}
}

func TestScoreByDimensionsMatchesParseFloat(t *testing.T) {
	// Expected values come from Mercury's scoreByDimensions, which uses
	// parseFloat and checks each dimension for truthiness.
	tests := []struct {
		name  string
		attrs string
		want  float64
	}{
		{name: "both dimensions", attrs: `width="200" height="100"`, want: 20},
		{name: "unit suffixes", attrs: `width="100px" height="100px"`, want: 10},
		{name: "whitespace and decimals", attrs: `width=" 60.6" height="100"`, want: 6},
		{name: "exponent", attrs: `width="1e3" height="10"`, want: -40},
		{name: "narrow width without height", attrs: `width="40"`, want: -50},
		{name: "short height without width", attrs: `height="40"`, want: -50},
		{name: "wide width without height", attrs: `width="400"`, want: 0},
		{name: "small area", attrs: `width="40" height="40"`, want: -200},
		{name: "zero width is ignored", attrs: `width="0" height="100"`, want: 0},
		{name: "non-numeric width is ignored", attrs: `width="auto" height="30"`, want: -50},
		{name: "negative width is truthy", attrs: `width="-10" height="100"`, want: -150},
		{name: "sprite skips area", attrs: `src="https://example.com/sprite.png" width="200" height="100"`, want: 0},
		{name: "infinite area", attrs: `width="1e308" height="1e308"`, want: math.Inf(1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attrs := tt.attrs
			if !strings.Contains(attrs, "src=") {
				attrs = `src="https://example.com/a.png" ` + attrs
			}
			doc, err := goquery.NewDocumentFromReader(strings.NewReader("<img " + attrs + ">"))
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}

			if got := scoreByDimensions(doc.Find("img")); got != tt.want {
				t.Fatalf("scoreByDimensions(%s) = %v, want %v", tt.attrs, got, tt.want)
			}
		})
	}
}
