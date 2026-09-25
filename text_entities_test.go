package hermes

import (
	"context"
	"testing"
)

// Pre-conversion sanitizing can remove every tag and leave escaped text behind.
func TestTextOutputDecodesEntitiesLeftBySanitizing(t *testing.T) {
	source := `<html><body><textarea>&lt;b&gt;bold&lt;/b&gt; text</textarea></body></html>`
	for name, options := range map[string][]Option{
		"unconfigured": {WithContentType("text")},
		"nil":          {WithContentType("text"), WithDefinitions(nil)},
	} {
		t.Run(name, func(t *testing.T) {
			result, err := New(options...).ParseHTML(context.Background(), source, "https://93.184.216.34/article")
			if err != nil {
				t.Fatal(err)
			}
			if result.Content != "<b>bold</b> text" {
				t.Fatalf("content = %q, want literal text", result.Content)
			}
		})
	}
}
