package hermes

import (
	"context"
	"strings"
	"testing"
)

func TestDefinitionsExistingOperationsModelRedditImagePreview(t *testing.T) {
	s := transformDefinitions(t, `    - target: descendants
      selector: 'div[role="img"]'
      regex.capture:
        attribute: background-image
        pattern: '^url\([''"]?(https://[^''")]+)[''"]?\)$'
        group: 1
        required: true
    - target: descendants
      selector: 'div[role="img"]'
      element.create:
        target: {selector: img, select: first}
        position: replace
        node:
          tag: img
          attributes:
            src: {attribute: background-image, required: true}
            alt: "Reddit preview"
    - target: descendants
      selector: 'div[role="img"]'
      element.unwrap: {}
`)
	r, err := New(WithDefinitions(s)).ParseHTML(context.Background(), `<article><p>Reddit media preview.</p><div role="img" background-image="url('https://images.example.com/preview.jpg')"><img src="/placeholder.jpg"></div></article>`, "https://93.184.216.34/story")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Content, `src="https://images.example.com/preview.jpg"`) || strings.Contains(r.Content, `role="img"`) || strings.Contains(r.Content, "placeholder.jpg") {
		t.Fatalf("existing typed pipeline did not preserve the preview safely: %s", r.Content)
	}
}
