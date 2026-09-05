package main

import (
	"strings"
	"testing"
	"time"
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
