package parser

import (
	"strings"
	"testing"

	"github.com/BumpyClock/hermes/internal/utils/security"
)

func TestContentSanitizationBeforeConversion(t *testing.T) {
	const source = "<article>\n" + `<p>Retain &lt;literal&gt; and <strong>article words</strong>.</p>` +
		`<script>unsafe-script-marker</script><p onclick="unsafe-event-marker">` +
		`<a href="javascript:unsafe-url-marker">link label</a>` +
		`<img src="https://example.com/image.jpg" onerror="unsafe-image-marker"></p>` + "\n\n\n</article>"
	clean := security.SanitizeHTML(source)
	for _, format := range []string{"html", "text/html", "", "unknown", "markdown", "md", "text/markdown", "text", "txt", "text/plain", " HTML ", " Text "} {
		t.Run(format, func(t *testing.T) {
			got := formatContent(source, format)
			if want := formatContent(clean, format); got != want {
				t.Fatalf("conversion changed sanitization: got %q, want %q", got, want)
			}
			if strings.Contains(got, "unsafe-") || !strings.Contains(got, "article words") || !strings.Contains(got, "link label") {
				t.Fatalf("unsafe or lossy content: %q", got)
			}
			switch strings.ToLower(strings.TrimSpace(format)) {
			case "html", "text/html", "", "unknown":
				if got != strings.TrimSpace(clean) {
					t.Fatalf("HTML output is not the trimmed sanitized source: %q", got)
				}
			}
		})
	}
}
