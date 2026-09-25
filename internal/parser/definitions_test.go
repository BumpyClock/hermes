package parser

import (
	"context"
	"strings"
	"testing"
)

func TestUnmatchedDefinitionsUseGenericExtraction(t *testing.T) {
	source := `<article><h1>BBC generic fallback</h1><p>A meaningful report provides enough factual text for generic extraction without any compiled site definition.</p></article>`
	opts := &ParserOptions{Fallback: true, ContentType: "html"}
	r, err := New().ParseHTMLWithContext(context.Background(), source, "https://www.bbc.com/news/article", opts)
	if err != nil {
		t.Fatal(err)
	}
	if r.ExtractorUsed != "" || !strings.Contains(r.Content, "meaningful report") {
		t.Fatalf("definition snapshot did not fall back to generic extraction: %+v", r)
	}
}
