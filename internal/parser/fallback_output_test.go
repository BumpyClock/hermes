package parser

import (
	"context"
	"testing"
)

func TestLegacyLastResortTextFormats(t *testing.T) {
	for format, want := range map[string]string{
		"html":          "&lt;img src=x onerror=alert(1)&gt;",
		"markdown":      `\<img src\=x onerror\=alert\(1\)\>`,
		"text/html":     "&lt;img src=x onerror=alert(1)&gt;",
		"text/markdown": `\<img src\=x onerror\=alert\(1\)\>`,
	} {
		t.Run(format, func(t *testing.T) {
			result, err := New().ParseHTMLWithContext(context.Background(),
				"<title>Review article</title><main><iframe><img src=x onerror=alert(1)></iframe></main>",
				"https://93.184.216.34/review", &ParserOptions{Fallback: true, ContentType: format})
			if err != nil {
				t.Fatal(err)
			}
			if result.Content != want {
				t.Errorf("content = %q, want %q", result.Content, want)
			}
		})
	}
}
