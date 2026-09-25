package generic

import (
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
			html := "<html><head><title>Ties</title></head><body><article>" + tt.images + "</article></body></html>"

			// Map iteration order varies, so repeat to catch nondeterministic winners.
			for run := range 25 {
				doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
				if err != nil {
					t.Fatalf("failed to parse HTML: %v", err)
				}

				got := extractor.Extract(ExtractorImageParams{Doc: doc, Content: "article"})
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
