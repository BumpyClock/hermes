package pools

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestResponseBodyPool(t *testing.T) {
	pool := NewBufferPool()

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
	pool := NewBufferPool()

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
	pool := NewBufferPool()

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

// Benchmark tests to verify pool performance benefits.
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
