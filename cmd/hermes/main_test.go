package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"strings"
	"testing"
	"time"

	"github.com/BumpyClock/hermes"
)

func TestRunParseAllFailuresWithTiming(t *testing.T) {
	oldTiming, oldConcurrency, oldTimeout := timing, concurrency, timeout
	t.Cleanup(func() {
		timing, concurrency, timeout = oldTiming, oldConcurrency, oldTimeout
	})
	timing, concurrency, timeout = true, 2, time.Second

	err := runParse(nil, []string{"ftp://example.com/a", "ftp://example.com/b"})
	if err == nil || !strings.Contains(err.Error(), "no URLs were successfully parsed") {
		t.Fatalf("expected all-failed error, got %v", err)
	}
}

func TestBatchOutputPreservesJSONOrder(t *testing.T) {
	oldFormat, oldFile := outputFormat, outputFile
	t.Cleanup(func() { outputFormat, outputFile = oldFormat, oldFile })
	outputFormat, outputFile = "json", filepath.Join(t.TempDir(), "batch.json")
	if err := formatOutput([]ParseResult{{URL: "https://example.com/a", ParseTime: 2 * time.Millisecond}}, false); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}
	want := `[
  {
    "parseTime": "2ms",
    "result": null,
    "url": "https://example.com/a"
  }
]`
	if string(got) != want {
		t.Fatalf("JSON output = %s, want %s", got, want)
	}
	if err = formatOutput(nil, false); err != nil {
		t.Fatal(err)
	}
	got, err = os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "null" {
		t.Fatalf("empty batch output = %s, want null", got)
	}
}

type batchMockParser struct{}

func (batchMockParser) Parse(_ context.Context, targetURL string) (*hermes.Result, error) {
	if strings.HasSuffix(targetURL, "/failure") {
		return nil, errors.New("expected failure")
	}
	return &hermes.Result{URL: targetURL, Title: "Success"}, nil
}

func (p batchMockParser) ParseHTML(ctx context.Context, _ string, targetURL string) (*hermes.Result, error) {
	return p.Parse(ctx, targetURL)
}

func TestBatchParsePreservesInputOrderAndPerURLFailures(t *testing.T) {
	oldConcurrency, oldTimeout := concurrency, timeout
	t.Cleanup(func() { concurrency, timeout = oldConcurrency, oldTimeout })
	concurrency, timeout = 2, time.Second
	urls := []string{"https://example.com/first", "https://example.com/failure", "https://example.com/last"}
	results := batchParse(batchMockParser{}, urls)
	if len(results) != len(urls) {
		t.Fatalf("result count = %d, want %d", len(results), len(urls))
	}
	for i, result := range results {
		if result.URL != urls[i] {
			t.Fatalf("result %d URL = %q, want %q", i, result.URL, urls[i])
		}
		if i == 1 {
			if result.Error == nil || result.Error.Error() != "expected failure" || result.Result != nil {
				t.Fatalf("failure result = %+v", result)
			}
		} else if result.Error != nil || result.Result == nil || result.Result.URL != urls[i] {
			t.Fatalf("success result = %+v", result)
		}
	}
}
