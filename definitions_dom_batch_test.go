package hermes

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestDefinitionsDOMMoveBatchInsertionOrder(t *testing.T) {
	source := `<article><div class="sources"><p class="photo" onclick="drop()">First source<script>drop-script</script></p><p class="photo">Second source</p></div><div class="targets"><section><div class="slot">Slot one</div></section><section><div class="slot">Slot two</div></section></div></article>`
	for _, position := range []string{"append", "prepend", "before", "after", "replace"} {
		t.Run(position, func(t *testing.T) {
			s := transformDefinitions(t, `    - target: root
      element.move:
        source: {selector: .photo, select: all}
        target: {selector: .slot, select: all}
        position: `+position+`
        required: true
`)
			for _, op := range []string{"Parse", "ParseHTML"} {
				t.Run(op, func(t *testing.T) {
					content := parseBatchContent(t, s, source, op)
					assertBatchOrder(t, content, position, "First source", "Second source")
					if strings.Contains(content, "onclick") || strings.Contains(content, "drop-script") {
						t.Fatalf("moved content bypassed sanitization: %s", content)
					}
				})
			}
		})
	}
}

func TestDefinitionsNoscriptBatchInsertionOrder(t *testing.T) {
	source := `<article><div class="targets"><section><div class="slot">Slot one</div></section><section><div class="slot">Slot two</div></section></div><noscript>&lt;picture&gt;&lt;img src="/first.jpg" alt="First image" onerror="drop()"&gt;&lt;img src="/second.jpg" alt="Second image"&gt;&lt;/picture&gt;</noscript></article>`
	for _, position := range []string{"append", "prepend", "before", "after", "replace"} {
		t.Run(position, func(t *testing.T) {
			s := transformDefinitions(t, `    - target: root
      noscript.recover:
        source: {selector: noscript, select: all}
        target: {selector: .slot, select: all}
        position: `+position+`
        required: true
`)
			for _, op := range []string{"Parse", "ParseHTML"} {
				t.Run(op, func(t *testing.T) {
					content := parseBatchContent(t, s, source, op)
					assertBatchOrder(t, content, position, "first.jpg", "second.jpg")
					if strings.Contains(content, "onerror") || strings.Contains(content, "<noscript") {
						t.Fatalf("recovered content bypassed sanitization: %s", content)
					}
				})
			}
		})
	}
}

func parseBatchContent(t *testing.T, definitions *Definitions, source, op string) string {
	t.Helper()
	const articleURL = "https://93.184.216.34/story"
	client := New(WithDefinitions(definitions), WithTransport(definitionTransport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"text/html"}}, Body: io.NopCloser(strings.NewReader(source)), Request: r}, nil
	})))
	var result *Result
	var err error
	if op == "Parse" {
		result, err = client.Parse(context.Background(), articleURL)
	} else {
		result, err = client.ParseHTML(context.Background(), source, articleURL)
	}
	if err != nil {
		t.Fatal(err)
	}
	return result.Content
}

func assertBatchOrder(t *testing.T, content, position, first, second string) {
	t.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}
	targets, err := doc.Find(".targets").Html()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(targets, first) != 2 || strings.Count(targets, second) != 2 {
		t.Fatalf("multi-target insertion lost content: %s", targets)
	}
	var ordered []string
	switch position {
	case "prepend", "before":
		ordered = []string{first, second, "Slot one", first, second, "Slot two"}
	case "append", "after":
		ordered = []string{"Slot one", first, second, "Slot two", first, second}
	case "replace":
		ordered = []string{first, second, first, second}
		if strings.Contains(targets, "Slot one") || strings.Contains(targets, "Slot two") {
			t.Fatalf("replacement target was retained: %s", targets)
		}
	}
	offset := 0
	for _, value := range ordered {
		index := strings.Index(targets[offset:], value)
		if index < 0 {
			t.Fatalf("%s did not preserve batch order %q: %s", position, value, targets)
		}
		offset += index + len(value)
	}
}
