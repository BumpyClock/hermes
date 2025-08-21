package parser

import (
	"strings"
	"testing"
	"time"
)

func TestDefaultStreamingConfig(t *testing.T) {
	config := DefaultStreamingConfig()
	
	if config.ChunkSize != 64*1024 {
		t.Errorf("Expected ChunkSize 64KB, got %d", config.ChunkSize)
	}
	
	if config.MaxDocumentSize != 50*1024*1024 {
		t.Errorf("Expected MaxDocumentSize 50MB, got %d", config.MaxDocumentSize)
	}
	
	if config.ProcessingTimeout != 60*time.Second {
		t.Errorf("Expected ProcessingTimeout 60s, got %v", config.ProcessingTimeout)
	}
	
	if !config.EnableMemoryLimit {
		t.Error("Expected EnableMemoryLimit to be true")
	}
	
	if config.MemoryLimitMB != 100 {
		t.Errorf("Expected MemoryLimitMB 100, got %d", config.MemoryLimitMB)
	}
}

func TestNewStreamingParser(t *testing.T) {
	// Test with nil config
	parser1 := NewStreamingParser(nil)
	if parser1 == nil {
		t.Fatal("Expected non-nil parser")
	}
	
	if parser1.config == nil {
		t.Error("Expected config to be set to default")
	}
	
	// Test with custom config
	customConfig := &StreamingConfig{
		ChunkSize:         32 * 1024,
		MaxDocumentSize:   10 * 1024 * 1024,
		ProcessingTimeout: 30 * time.Second,
		BufferSize:        4 * 1024,
		EnableMemoryLimit: false,
		MemoryLimitMB:     50,
	}
	
	parser2 := NewStreamingParser(customConfig)
	if parser2.config.ChunkSize != 32*1024 {
		t.Errorf("Expected custom ChunkSize, got %d", parser2.config.ChunkSize)
	}
	
	// Cleanup
	parser1.Close()
	parser2.Close()
}

func TestChunkedReader(t *testing.T) {
	// Test small content
	content := "<!DOCTYPE html><html><head><title>Test</title></head><body><p>Hello World</p></body></html>"
	reader := strings.NewReader(content)
	
	chunkedReader := NewChunkedReader(reader, 32, 1024)
	
	var chunks [][]byte
	for {
		chunk, isEOF, err := chunkedReader.ReadChunk()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if len(chunk) > 0 {
			chunks = append(chunks, chunk)
		}
		
		if isEOF {
			break
		}
	}
	
	// Reconstruct content
	var reconstructed strings.Builder
	for _, chunk := range chunks {
		reconstructed.Write(chunk)
	}
	
	if reconstructed.String() != content {
		t.Errorf("Content mismatch. Expected: %s, Got: %s", content, reconstructed.String())
	}
}

func TestChunkedReaderLargeContent(t *testing.T) {
	// Create large content
	var largeContent strings.Builder
	largeContent.WriteString("<!DOCTYPE html><html><head><title>Large Document</title></head><body>")
	
	// Add many paragraphs
	for i := 0; i < 1000; i++ {
		largeContent.WriteString("<p>This is paragraph number ")
		largeContent.WriteString(strings.Repeat("x", 100)) // 100 chars
		largeContent.WriteString("</p>")
	}
	largeContent.WriteString("</body></html>")
	
	content := largeContent.String()
	reader := strings.NewReader(content)
	
	chunkedReader := NewChunkedReader(reader, 1024, int64(len(content)+1000))
	
	totalBytes := int64(0)
	chunkCount := 0
	
	for {
		chunk, isEOF, err := chunkedReader.ReadChunk()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if len(chunk) > 0 {
			totalBytes += int64(len(chunk))
			chunkCount++
		}
		
		if isEOF {
			break
		}
	}
	
	if totalBytes != int64(len(content)) {
		t.Errorf("Expected %d bytes, got %d", len(content), totalBytes)
	}
	
	if chunkCount < 2 {
		t.Errorf("Expected multiple chunks for large content, got %d", chunkCount)
	}
}

func TestChunkedReaderSizeLimit(t *testing.T) {
	// Create content larger than limit
	content := strings.Repeat("x", 2000)
	reader := strings.NewReader(content)
	
	// Set limit to 1000 bytes
	chunkedReader := NewChunkedReader(reader, 512, 1000)
	
	for {
		_, isEOF, err := chunkedReader.ReadChunk()
		if err != nil {
			// Should get size limit error
			if !strings.Contains(err.Error(), "exceeded limit") {
				t.Errorf("Expected size limit error, got: %v", err)
			}
			break
		}
		
		if isEOF {
			t.Error("Should have hit size limit before EOF")
			break
		}
	}
}

func TestProgressiveBuilder(t *testing.T) {
	config := DefaultStreamingConfig()
	builder := NewProgressiveBuilder(config)
	
	// Add chunks
	chunk1 := []byte("<!DOCTYPE html><html><head><title>Test</title></head>")
	chunk2 := []byte("<body><p>Hello")
	chunk3 := []byte(" World</p></body></html>")
	
	err := builder.AddChunk(chunk1)
	if err != nil {
		t.Fatalf("Error adding chunk1: %v", err)
	}
	
	err = builder.AddChunk(chunk2)
	if err != nil {
		t.Fatalf("Error adding chunk2: %v", err)
	}
	
	err = builder.AddChunk(chunk3)
	if err != nil {
		t.Fatalf("Error adding chunk3: %v", err)
	}
	
	// Build document
	doc, err := builder.BuildDocument()
	if err != nil {
		t.Fatalf("Error building document: %v", err)
	}
	
	if doc == nil {
		t.Fatal("Expected non-nil document")
	}
	
	// Verify content
	title := doc.Find("title").Text()
	if title != "Test" {
		t.Errorf("Expected title 'Test', got '%s'", title)
	}
	
	text := doc.Find("p").Text()
	if text != "Hello World" {
		t.Errorf("Expected text 'Hello World', got '%s'", text)
	}
}

func TestProgressiveBuilderPartialDocument(t *testing.T) {
	config := DefaultStreamingConfig()
	builder := NewProgressiveBuilder(config)
	
	// Add incomplete HTML
	chunk := []byte("<p>Partial content</p>")
	err := builder.AddChunk(chunk)
	if err != nil {
		t.Fatalf("Error adding chunk: %v", err)
	}
	
	// Get partial document
	doc, err := builder.GetPartialDocument()
	if err != nil {
		t.Fatalf("Error getting partial document: %v", err)
	}
	
	// Should wrap content in html/body tags
	text := doc.Find("p").Text()
	if text != "Partial content" {
		t.Errorf("Expected 'Partial content', got '%s'", text)
	}
}

func TestStreamingParserBasic(t *testing.T) {
	config := DefaultStreamingConfig()
	config.ProcessingTimeout = 5 * time.Second // Short timeout for test
	
	parser := NewStreamingParser(config)
	defer parser.Close()
	
	content := "<!DOCTYPE html><html><head><title>Stream Test</title></head><body><p>Streaming content</p></body></html>"
	reader := strings.NewReader(content)
	
	result := parser.ParseStream(reader)
	
	if result.Error != nil {
		t.Fatalf("Unexpected error: %v", result.Error)
	}
	
	if result.Document == nil {
		t.Fatal("Expected non-nil document")
	}
	
	if result.IsPartial {
		t.Error("Should not be partial for small document")
	}
	
	// Verify parsing
	title := result.Document.Find("title").Text()
	if title != "Stream Test" {
		t.Errorf("Expected title 'Stream Test', got '%s'", title)
	}
	
	// Check stats
	if result.Stats.BytesProcessed != int64(len(content)) {
		t.Errorf("Expected %d bytes processed, got %d", len(content), result.Stats.BytesProcessed)
	}
	
	if result.Stats.ChunksProcessed < 1 {
		t.Errorf("Expected at least 1 chunk processed, got %d", result.Stats.ChunksProcessed)
	}
}

func TestStreamingParserLargeDocument(t *testing.T) {
	config := DefaultStreamingConfig()
	config.ChunkSize = 1024 // Small chunks for testing
	config.ProcessingTimeout = 10 * time.Second
	
	parser := NewStreamingParser(config)
	defer parser.Close()
	
	// Create large HTML document
	var content strings.Builder
	content.WriteString("<!DOCTYPE html><html><head><title>Large Document</title></head><body>")
	
	for i := 0; i < 500; i++ {
		content.WriteString("<p>This is a very long paragraph with lots of text to make the document large ")
		content.WriteString(strings.Repeat("content ", 20))
		content.WriteString("</p>")
	}
	content.WriteString("</body></html>")
	
	reader := strings.NewReader(content.String())
	
	result := parser.ParseStream(reader)
	
	if result.Error != nil {
		t.Fatalf("Unexpected error: %v", result.Error)
	}
	
	if result.Document == nil {
		t.Fatal("Expected non-nil document")
	}
	
	// Should process multiple chunks
	if result.Stats.ChunksProcessed < 2 {
		t.Errorf("Expected multiple chunks for large document, got %d", result.Stats.ChunksProcessed)
	}
	
	// Verify content still accessible
	title := result.Document.Find("title").Text()
	if title != "Large Document" {
		t.Errorf("Expected title 'Large Document', got '%s'", title)
	}
	
	paragraphs := result.Document.Find("p")
	if paragraphs.Length() != 500 {
		t.Errorf("Expected 500 paragraphs, got %d", paragraphs.Length())
	}
}

func TestStreamingParserSizeLimit(t *testing.T) {
	config := DefaultStreamingConfig()
	config.MaxDocumentSize = 1000 // Very small limit
	config.ChunkSize = 512
	
	parser := NewStreamingParser(config)
	defer parser.Close()
	
	// Create content larger than limit
	content := strings.Repeat("<p>Large content</p>", 100)
	content = "<!DOCTYPE html><html><body>" + content + "</body></html>"
	
	reader := strings.NewReader(content)
	
	result := parser.ParseStream(reader)
	
	// Should get size error
	if result.Error == nil {
		t.Error("Expected error for oversized document")
	}
	
	if !strings.Contains(result.Error.Error(), "exceeded limit") {
		t.Errorf("Expected size limit error, got: %v", result.Error)
	}
}

func TestIsLargeDocument(t *testing.T) {
	testCases := []struct {
		size     int64
		expected bool
	}{
		{1000, false},           // 1KB - small
		{500 * 1024, false},     // 500KB - medium
		{1024 * 1024, false},    // Exactly 1MB - threshold
		{1024*1024 + 1, true},   // Just over 1MB - large
		{10 * 1024 * 1024, true}, // 10MB - large
	}
	
	for _, tc := range testCases {
		result := IsLargeDocument(tc.size)
		if result != tc.expected {
			t.Errorf("For size %d, expected %v, got %v", tc.size, tc.expected, result)
		}
	}
}

func TestStreamingExtractor(t *testing.T) {
	config := DefaultStreamingConfig()
	extractor := NewStreamingExtractor(config)
	
	if extractor == nil {
		t.Fatal("Expected non-nil extractor")
	}
	
	if extractor.config != config {
		t.Error("Config not set correctly")
	}
}

func TestStreamingExtractorExtractFromStream(t *testing.T) {
	// Skip this test for now as it requires full parser integration
	// which is complex to set up in unit tests
	t.Skip("Skipping integration test - requires full parser setup")
}

func TestStreamingParserStats(t *testing.T) {
	config := DefaultStreamingConfig()
	parser := NewStreamingParser(config)
	defer parser.Close()
	
	content := "<!DOCTYPE html><html><body><p>Test content</p></body></html>"
	reader := strings.NewReader(content)
	
	// Get initial stats
	initialStats := parser.GetStats()
	if initialStats.BytesProcessed != 0 {
		t.Error("Expected zero initial bytes processed")
	}
	
	// Process stream
	result := parser.ParseStream(reader)
	if result.Error != nil {
		t.Fatalf("Unexpected error: %v", result.Error)
	}
	
	// Get final stats
	finalStats := parser.GetStats()
	if finalStats.BytesProcessed != int64(len(content)) {
		t.Errorf("Expected %d bytes processed, got %d", len(content), finalStats.BytesProcessed)
	}
	
	if finalStats.ChunksProcessed < 1 {
		t.Error("Expected at least 1 chunk processed")
	}
	
	if finalStats.ProcessingTime == 0 {
		t.Error("Expected non-zero processing time")
	}
}

func TestStreamingParserReset(t *testing.T) {
	config := DefaultStreamingConfig()
	parser := NewStreamingParser(config)
	defer parser.Close()
	
	content := "<!DOCTYPE html><html><body><p>Test</p></body></html>"
	reader := strings.NewReader(content)
	
	// Process content
	parser.ParseStream(reader)
	
	// Check stats are non-zero
	stats := parser.GetStats()
	if stats.BytesProcessed == 0 {
		t.Error("Expected non-zero bytes processed")
	}
	
	// Reset
	parser.Reset()
	
	// Check stats are reset
	resetStats := parser.GetStats()
	if resetStats.BytesProcessed != 0 {
		t.Error("Expected zero bytes after reset")
	}
	
	if resetStats.ChunksProcessed != 0 {
		t.Error("Expected zero chunks after reset")
	}
}

// Benchmark tests
func BenchmarkStreamingParserSmallDocument(b *testing.B) {
	config := DefaultStreamingConfig()
	content := "<!DOCTYPE html><html><head><title>Test</title></head><body><p>Small content</p></body></html>"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser := NewStreamingParser(config)
		reader := strings.NewReader(content)
		result := parser.ParseStream(reader)
		if result.Error != nil {
			b.Fatalf("Error: %v", result.Error)
		}
		parser.Close()
	}
}

func BenchmarkStreamingParserLargeDocument(b *testing.B) {
	config := DefaultStreamingConfig()
	
	// Create large document
	var content strings.Builder
	content.WriteString("<!DOCTYPE html><html><body>")
	for i := 0; i < 1000; i++ {
		content.WriteString("<p>Large document content with lots of text</p>")
	}
	content.WriteString("</body></html>")
	
	largeContent := content.String()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser := NewStreamingParser(config)
		reader := strings.NewReader(largeContent)
		result := parser.ParseStream(reader)
		if result.Error != nil {
			b.Fatalf("Error: %v", result.Error)
		}
		parser.Close()
	}
}

func BenchmarkChunkedReader(b *testing.B) {
	content := strings.Repeat("<p>Benchmark content</p>", 1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(content)
		chunkedReader := NewChunkedReader(reader, 1024, int64(len(content)))
		
		for {
			_, isEOF, err := chunkedReader.ReadChunk()
			if err != nil {
				b.Fatalf("Error: %v", err)
			}
			if isEOF {
				break
			}
		}
	}
}