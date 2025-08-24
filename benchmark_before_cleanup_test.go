package hermes

import (
	"context"
	"runtime"
	"testing"
	"time"
)

// BenchmarkMemoryBeforeCleanup documents memory usage before removing orchestration code
func BenchmarkMemoryBeforeCleanup(b *testing.B) {
	// Force GC before starting
	runtime.GC()
	runtime.GC()
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	allocBefore := m.Alloc
	
	client := New()
	ctx := context.Background()
	
	// Simple HTML that should be quick to parse
	html := `<!DOCTYPE html>
<html>
<head><title>Test Article</title></head>
<body>
  <article>
    <h1>Test Article Title</h1>
    <p>This is test content for benchmarking memory usage. It contains enough text to be considered valid article content.</p>
    <p>Another paragraph with more content to ensure extraction works properly.</p>
  </article>
</body>
</html>`
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		result, err := client.ParseHTML(ctx, html, "https://example.com/test")
		if err != nil {
			b.Fatal(err)
		}
		if result.Title == "" {
			b.Fatal("No title extracted")
		}
	}
	
	runtime.GC()
	runtime.ReadMemStats(&m)
	allocAfter := m.Alloc
	
	b.ReportMetric(float64(allocAfter-allocBefore)/float64(b.N), "bytes/op")
	b.ReportMetric(float64(m.NumGC), "GCs")
	b.ReportMetric(float64(m.HeapAlloc), "heap-bytes")
}

// TestMemoryBaseline provides a baseline memory measurement
func TestMemoryBaseline(t *testing.T) {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	t.Logf("Baseline Memory Stats (Phase D - Before Cleanup):")
	t.Logf("  Alloc = %v KB", m.Alloc/1024)
	t.Logf("  TotalAlloc = %v KB", m.TotalAlloc/1024)
	t.Logf("  Sys = %v KB", m.Sys/1024)
	t.Logf("  NumGC = %v", m.NumGC)
	
	// Parse a real URL to see memory usage
	client := New()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	runtime.GC()
	runtime.ReadMemStats(&m)
	beforeParse := m.Alloc
	
	result, err := client.Parse(ctx, "https://www.theverge.com/notepad-microsoft-newsletter/763357/microsoft-asus-xbox-ally-handheld-hands-on-notepad")
	if err != nil {
		t.Logf("Parse error (expected in test): %v", err)
	} else {
		t.Logf("Parsed successfully, title: %s", result.Title)
	}
	
	runtime.GC()
	runtime.ReadMemStats(&m)
	afterParse := m.Alloc
	
	t.Logf("Memory used for parse: %v KB", (afterParse-beforeParse)/1024)
	t.Logf("Final heap alloc: %v KB", m.HeapAlloc/1024)
}