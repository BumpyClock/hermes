package parser_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/BumpyClock/hermes/internal/parser"
)

func TestParseHTMLDoesNotEmitFetchDebugLogs(t *testing.T) {
	output := captureStdout(t, func() {
		p := parser.New()
		_, err := p.ParseHTML(`
			<!doctype html>
			<html>
				<head><title>Quiet Article</title></head>
				<body><article><p>Content stays local.</p></article></body>
			</html>
		`, "https://example.com/article", parser.DefaultParserOptions())
		if err != nil {
			t.Fatalf("ParseHTML failed: %v", err)
		}
	})

	if strings.Contains(output, "[HERMES FETCH DEBUG]") || strings.Contains(output, "[HERMES HTTP DEBUG]") {
		t.Fatalf("expected ParseHTML to avoid fetch debug output, got %q", output)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}

	os.Stdout = writer
	t.Cleanup(func() {
		os.Stdout = oldStdout
	})

	output := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, reader)
		output <- buf.String()
	}()

	fn()

	os.Stdout = oldStdout
	if err := writer.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	captured := <-output
	if err := reader.Close(); err != nil {
		t.Fatalf("close stdout reader: %v", err)
	}

	return captured
}
