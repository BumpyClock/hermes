package pools

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestDocumentPool(t *testing.T) {
	pool := NewDocumentPool()

	// Test basic get/put cycle
	htmlContent := `<html><body><p>Test content</p></body></html>`
	reader := strings.NewReader(htmlContent)

	doc, err := pool.Get(reader)
	if err != nil {
		t.Fatalf("Failed to get document from pool: %v", err)
	}

	if doc == nil {
		t.Fatal("Got nil document from pool")
	}

	// Verify the document contains expected content
	text := doc.Find("p").Text()
	if text != "Test content" {
		t.Errorf("Expected 'Test content', got '%s'", text)
	}

	// Put document back in pool
	pool.Put(doc)

	// Get another document to ensure pool reuse works
	reader2 := strings.NewReader(`<html><body><h1>Different content</h1></body></html>`)
	doc2, err := pool.Get(reader2)
	if err != nil {
		t.Fatalf("Failed to get second document from pool: %v", err)
	}

	// Verify the new document has the new content
	h1Text := doc2.Find("h1").Text()
	if h1Text != "Different content" {
		t.Errorf("Expected 'Different content', got '%s'", h1Text)
	}

	pool.Put(doc2)
}

func TestDocumentPoolWithInvalidHTML(t *testing.T) {
	pool := NewDocumentPool()

	// Test with invalid HTML
	invalidHTML := `<html><body><p>Unclosed paragraph`
	reader := strings.NewReader(invalidHTML)

	doc, err := pool.Get(reader)
	if err != nil {
		t.Fatalf("Failed to get document with invalid HTML: %v", err)
	}

	if doc == nil {
		t.Fatal("Got nil document from pool with invalid HTML")
	}

	pool.Put(doc)
}

func TestResponseBodyPool(t *testing.T) {
	pool := NewResponseBodyPool()

	// Test basic get/put cycle
	buf1 := pool.Get()
	if buf1 == nil {
		t.Fatal("Got nil buffer from pool")
	}

	// Write some data to the buffer
	testData := []byte("test response body data")
	buf1.Write(testData)

	if buf1.Len() != len(testData) {
		t.Errorf("Expected buffer length %d, got %d", len(testData), buf1.Len())
	}

	// Put buffer back in pool
	pool.Put(buf1)

	// Get another buffer - should be reset
	buf2 := pool.Get()
	if buf2.Len() != 0 {
		t.Errorf("Expected reset buffer length 0, got %d", buf2.Len())
	}

	pool.Put(buf2)
}

func TestResponseBodyPoolReadResponseBody(t *testing.T) {
	pool := NewResponseBodyPool()

	// Create a mock HTTP response
	responseBody := "This is a test response body"
	resp := &http.Response{
		Body: io.NopCloser(strings.NewReader(responseBody)),
	}

	// Read the response body using the pool
	data, err := pool.ReadResponseBody(resp)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if string(data) != responseBody {
		t.Errorf("Expected '%s', got '%s'", responseBody, string(data))
	}
}

func TestResponseBodyPoolWithNilResponse(t *testing.T) {
	pool := NewResponseBodyPool()

	// Test with nil response
	data, err := pool.ReadResponseBody(nil)
	if err != nil {
		t.Errorf("Expected no error with nil response, got: %v", err)
	}

	if data != nil {
		t.Errorf("Expected nil data with nil response, got: %v", data)
	}

	// Test with response with nil body
	resp := &http.Response{Body: nil}
	data, err = pool.ReadResponseBody(resp)
	if err != nil {
		t.Errorf("Expected no error with nil body, got: %v", err)
	}

	if data != nil {
		t.Errorf("Expected nil data with nil body, got: %v", data)
	}
}

func TestBufferPool(t *testing.T) {
	pool := NewBufferPool()

	// Test basic get/put cycle
	buf1 := pool.Get()
	if buf1 == nil {
		t.Fatal("Got nil buffer from pool")
	}

	// Write some data
	testData := "test buffer data"
	buf1.WriteString(testData)

	if buf1.String() != testData {
		t.Errorf("Expected '%s', got '%s'", testData, buf1.String())
	}

	// Put buffer back in pool
	pool.Put(buf1)

	// Get another buffer - should be reset
	buf2 := pool.Get()
	if buf2.Len() != 0 {
		t.Errorf("Expected reset buffer length 0, got %d", buf2.Len())
	}

	if buf2.String() != "" {
		t.Errorf("Expected empty string from reset buffer, got '%s'", buf2.String())
	}

	pool.Put(buf2)
}

func TestBufferPoolSizeLimit(t *testing.T) {
	pool := NewBufferPool()

	// Create a large buffer (over 64KB limit)
	buf := pool.Get()
	largeData := make([]byte, 128*1024) // 128KB
	for i := range largeData {
		largeData[i] = 'A'
	}
	buf.Write(largeData)

	if buf.Cap() < 64*1024 {
		t.Skip("Buffer didn't grow large enough for size limit test")
	}

	// Put it back - should not be returned to pool due to size
	pool.Put(buf)

	// This test is mainly to ensure Put doesn't panic with large buffers
	// The actual size limiting behavior is internal to the pool
}

func TestStringBuilderPool(t *testing.T) {
	pool := NewStringBuilderPool()

	// Test basic get/put cycle
	sb1 := pool.Get()
	if sb1 == nil {
		t.Fatal("Got nil string builder from pool")
	}

	// Write some data
	testData := "test string builder data"
	sb1.WriteString(testData)

	if sb1.String() != testData {
		t.Errorf("Expected '%s', got '%s'", testData, sb1.String())
	}

	// Put builder back in pool
	pool.Put(sb1)

	// Get another builder - should be reset
	sb2 := pool.Get()
	if sb2.Len() != 0 {
		t.Errorf("Expected reset builder length 0, got %d", sb2.Len())
	}

	if sb2.String() != "" {
		t.Errorf("Expected empty string from reset builder, got '%s'", sb2.String())
	}

	pool.Put(sb2)
}

func TestPooledStringBuilder(t *testing.T) {
	psb := NewPooledStringBuilder()
	defer psb.Close()

	// Test writing to the builder
	testData := "test pooled string builder"
	_, err := psb.WriteString(testData)
	if err != nil {
		t.Fatalf("Failed to write to pooled string builder: %v", err)
	}

	if psb.String() != testData {
		t.Errorf("Expected '%s', got '%s'", testData, psb.String())
	}

	// Test reset
	psb.Reset()
	if psb.String() != "" {
		t.Errorf("Expected empty string after reset, got '%s'", psb.String())
	}
}

func TestWithPooledStringBuilder(t *testing.T) {
	result, err := WithPooledStringBuilder(func(sb *strings.Builder) error {
		sb.WriteString("Hello, ")
		sb.WriteString("World!")
		return nil
	})

	if err != nil {
		t.Fatalf("WithPooledStringBuilder failed: %v", err)
	}

	expected := "Hello, World!"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestWithPooledStringBuilderError(t *testing.T) {
	result, err := WithPooledStringBuilder(func(sb *strings.Builder) error {
		sb.WriteString("This should not be returned")
		return bytes.ErrTooLarge
	})

	if err == nil {
		t.Fatal("Expected error from WithPooledStringBuilder")
	}

	if result != "" {
		t.Errorf("Expected empty result on error, got '%s'", result)
	}

	if err != bytes.ErrTooLarge {
		t.Errorf("Expected bytes.ErrTooLarge, got %v", err)
	}
}

func TestWithPooledBuffer(t *testing.T) {
	testData := []byte("test buffer content")

	result, err := WithPooledBuffer(func(buf *bytes.Buffer) error {
		buf.Write(testData)
		return nil
	})

	if err != nil {
		t.Fatalf("WithPooledBuffer failed: %v", err)
	}

	if !bytes.Equal(result, testData) {
		t.Errorf("Expected %v, got %v", testData, result)
	}
}

func TestWithPooledBufferError(t *testing.T) {
	result, err := WithPooledBuffer(func(buf *bytes.Buffer) error {
		buf.WriteString("This should not be returned")
		return io.ErrUnexpectedEOF
	})

	if err == nil {
		t.Fatal("Expected error from WithPooledBuffer")
	}

	if result != nil {
		t.Errorf("Expected nil result on error, got %v", result)
	}

	if err != io.ErrUnexpectedEOF {
		t.Errorf("Expected io.ErrUnexpectedEOF, got %v", err)
	}
}

func TestGlobalPools(t *testing.T) {
	// Test that global pools are initialized and working
	if GlobalDocumentPool == nil {
		t.Error("GlobalDocumentPool is nil")
	}

	if GlobalResponseBodyPool == nil {
		t.Error("GlobalResponseBodyPool is nil")
	}

	if GlobalBufferPool == nil {
		t.Error("GlobalBufferPool is nil")
	}

	if GlobalStringBuilderPool == nil {
		t.Error("GlobalStringBuilderPool is nil")
	}

	// Test using global pools
	buf := GlobalBufferPool.Get()
	buf.WriteString("test")
	GlobalBufferPool.Put(buf)

	sb := GlobalStringBuilderPool.Get()
	sb.WriteString("test")
	GlobalStringBuilderPool.Put(sb)
}

func TestGetPoolStats(t *testing.T) {
	stats := GetPoolStats()

	// Since we can't easily track actual usage with sync.Pool,
	// we just verify the function doesn't panic and returns a valid struct
	if stats.DocumentsInUse < 0 {
		t.Error("DocumentsInUse should not be negative")
	}

	if stats.BuffersInUse < 0 {
		t.Error("BuffersInUse should not be negative")
	}

	if stats.StringBuildersInUse < 0 {
		t.Error("StringBuildersInUse should not be negative")
	}

	if stats.ResponseBuffersInUse < 0 {
		t.Error("ResponseBuffersInUse should not be negative")
	}
}

// Benchmark tests to verify pool performance benefits
func BenchmarkStringBuilderWithPool(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sb := GlobalStringBuilderPool.Get()
		sb.WriteString("test")
		sb.WriteString(" string")
		sb.WriteString(" concatenation")
		_ = sb.String()
		GlobalStringBuilderPool.Put(sb)
	}
}

func BenchmarkStringBuilderWithoutPool(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sb := &strings.Builder{}
		sb.WriteString("test")
		sb.WriteString(" string")
		sb.WriteString(" concatenation")
		_ = sb.String()
	}
}

func BenchmarkBufferWithPool(b *testing.B) {
	data := []byte("test buffer data for benchmarking")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf := GlobalBufferPool.Get()
		buf.Write(data)
		_ = buf.Bytes()
		GlobalBufferPool.Put(buf)
	}
}

func BenchmarkBufferWithoutPool(b *testing.B) {
	data := []byte("test buffer data for benchmarking")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		buf.Write(data)
		_ = buf.Bytes()
	}
}